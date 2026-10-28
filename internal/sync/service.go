package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/salambhai/vecdim-backend/ent"
	"github.com/salambhai/vecdim-backend/ent/fieldrecord"
	"github.com/salambhai/vecdim-backend/internal/auth"
	"github.com/salambhai/vecdim-backend/internal/validation"
)

type Service struct {
	client *ent.Client
}

func NewService(client *ent.Client) *Service {
	return &Service{client: client}
}

// ProcessBatch handles idempotent, partial-failure batch synchronization.
func (s *Service) ProcessBatch(ctx context.Context, req BatchSyncRequest) (*BatchSyncResponse, error) {
	user, _ := auth.GetUser(ctx)
	batchID := uuid.New().String()

	rawJSON, _ := json.Marshal(req)
	rawJSONStr := string(rawJSON)

	syncBatchCreate := s.client.SyncBatch.Create().
		SetID(batchID).
		SetRecordCount(len(req.Records)).
		SetPayloadJSON(rawJSONStr)

	if req.DeviceID != "" {
		syncBatchCreate.SetDeviceID(req.DeviceID)
	}
	if user.UserID != "" {
		syncBatchCreate.SetUserID(user.UserID)
	}

	_, err := syncBatchCreate.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize sync batch record: %w", err)
	}

	results := make([]RecordSyncResult, len(req.Records))
	acceptedCount := 0
	rejectedCount := 0

	for i, rec := range req.Records {
		// Idempotency check: see if record_id already exists in the database
		existing, err := s.client.FieldRecord.Query().
			Where(fieldrecord.ID(rec.RecordID)).
			Only(ctx)

		if err == nil && existing != nil {
			// Record was already processed and stored
			results[i] = RecordSyncResult{
				RecordID: rec.RecordID,
				Status:   "already_received",
				Flags:    []validation.ValidationError{},
				Errors:   []validation.ValidationError{},
			}
			continue
		}

		// Validate record (server clock is used as reference date)
		serverClockDate := time.Now().Format("2006-01-02")
		blocks, warns := validation.ValidateSingleRecord(rec, serverClockDate, true)

		if len(blocks) > 0 {
			// One bad record never blocks the others: mark as rejected, do NOT save to main table
			rejectedCount++
			results[i] = RecordSyncResult{
				RecordID: rec.RecordID,
				Status:   "rejected",
				Flags:    warns,
				Errors:   blocks,
			}
			continue
		}

		// Record passes block rules: compute WHO mortality class
		mortalityPct, whoClass := validation.CalculateWHOMortalityClass(rec.NExposed, rec.NDead)

		status := "accepted"
		if len(warns) > 0 {
			status = "accepted_with_flags"
		}

		// Save record with staged status "UNPROCESSED" (ready for asynchronous R processing)
		createOp := s.client.FieldRecord.Create().
			SetID(rec.RecordID).
			SetSyncBatchID(batchID).
			SetState(rec.State).
			SetLga(rec.LGA).
			SetWard(rec.Ward).
			SetLatitude(rec.Latitude).
			SetLongitude(rec.Longitude).
			SetCollectionDate(rec.CollectionDate).
			SetNExposed(rec.NExposed).
			SetNDead(rec.NDead).
			SetMortalityPct(mortalityPct).
			SetWhoClass(whoClass).
			SetUrban(rec.Urban).
			SetIrsCovered(rec.IRSCovered).
			SetStatus("UNPROCESSED") // Staging architecture: R will update to CLEAN / FLAGGED

		if user.UserID != "" {
			createOp.SetUserID(user.UserID)
		}
		if req.DeviceID != "" {
			createOp.SetDeviceID(req.DeviceID)
		}
		if rec.Species != nil {
			createOp.SetSpecies(*rec.Species)
		}
		if rec.Insecticide != nil {
			createOp.SetInsecticide(*rec.Insecticide)
		}
		if rec.GPSAccuracyM != nil {
			createOp.SetGpsAccuracyM(*rec.GPSAccuracyM)
		}
		if rec.LarvalDensity != nil {
			createOp.SetLarvalDensity(*rec.LarvalDensity)
		}
		if rec.ITNUsePct != nil {
			createOp.SetItnUsePct(*rec.ITNUsePct)
		}

		savedRecord, err := createOp.Save(ctx)
		if err != nil {
			// Check if this is completing an existing Bioassay Pending record (Phase 2 lab update)
			existing, getErr := s.client.FieldRecord.Get(ctx, rec.RecordID)
			if getErr == nil && existing != nil && existing.NExposed == 0 && rec.NExposed > 0 {
				updateOp := s.client.FieldRecord.UpdateOneID(existing.ID).
					SetNExposed(rec.NExposed).
					SetNDead(rec.NDead).
					SetMortalityPct(mortalityPct).
					SetWhoClass(whoClass).
					SetStatus("UNPROCESSED")

				if rec.Insecticide != nil {
					updateOp.SetInsecticide(*rec.Insecticide)
				}
				if rec.Species != nil {
					updateOp.SetSpecies(*rec.Species)
				}

				_, updateErr := updateOp.Save(ctx)
				if updateErr == nil {
					acceptedCount++
					results[i] = RecordSyncResult{
						RecordID: rec.RecordID,
						Status:   "accepted",
						Flags:    warns,
						Errors:   []validation.ValidationError{},
					}
					continue
				}
			}

			// True duplicate / idempotent retry
			results[i] = RecordSyncResult{
				RecordID: rec.RecordID,
				Status:   "already_received",
				Flags:    warns,
				Errors:   []validation.ValidationError{},
			}
			continue
		}

		// Save validation warnings if any
		for _, w := range warns {
			_ = s.client.ValidationResult.Create().
				SetRecordID(savedRecord.ID).
				SetRuleID(w.Rule).
				SetField(w.Field).
				SetSeverity(string(w.Severity)).
				SetMessage(w.Message).
				Exec(ctx)
		}

		acceptedCount++
		results[i] = RecordSyncResult{
			RecordID: rec.RecordID,
			Status:   status,
			Flags:    warns,
			Errors:   []validation.ValidationError{},
		}
	}

	// Update sync batch counts
	_ = s.client.SyncBatch.UpdateOneID(batchID).
		SetAcceptedCount(acceptedCount).
		SetRejectedCount(rejectedCount).
		Exec(ctx)

	return &BatchSyncResponse{Results: results}, nil
}
