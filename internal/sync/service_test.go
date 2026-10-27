package sync_test

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/salambhai/vecdim-backend/ent/enttest"
	"github.com/salambhai/vecdim-backend/internal/sync"
	"github.com/salambhai/vecdim-backend/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatchSyncIdempotencyAndPartialFailure(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	svc := sync.NewService(client)
	ctx := context.Background()

	species := "An. gambiae s.s."
	insecticide := "Permethrin"
	itn := 65.0

	// Batch with 1 VALID record and 1 INVALID record (n_dead > n_exposed)
	req1 := sync.BatchSyncRequest{
		DeviceID: "device_phone_001",
		Records: []validation.RecordInput{
			{
				RecordID:       "rec-valid-01",
				State:          "Osun",
				LGA:            "Osogbo",
				Ward:           "Ward 1",
				Latitude:       7.77,
				Longitude:      4.56,
				CollectionDate: "2026-10-01",
				Species:        &species,
				Insecticide:    &insecticide,
				NExposed:       100,
				NDead:          50,
				ITNUsePct:      &itn,
			},
			{
				RecordID:       "rec-invalid-02",
				State:          "Osun",
				LGA:            "Osogbo",
				Ward:           "Ward 2",
				Latitude:       7.77,
				Longitude:      4.56,
				CollectionDate: "2026-10-01",
				Species:        &species,
				Insecticide:    &insecticide,
				NExposed:       100,
				NDead:          150, // Bad: n_dead > n_exposed
				ITNUsePct:      &itn,
			},
		},
	}

	res1, err := svc.ProcessBatch(ctx, req1)
	require.NoError(t, err)
	require.Len(t, res1.Results, 2)

	// Partial failure isolation: rec-valid-01 accepted, rec-invalid-02 rejected
	assert.Equal(t, "rec-valid-01", res1.Results[0].RecordID)
	assert.Equal(t, "accepted", res1.Results[0].Status)
	assert.Empty(t, res1.Results[0].Errors)

	assert.Equal(t, "rec-invalid-02", res1.Results[1].RecordID)
	assert.Equal(t, "rejected", res1.Results[1].Status)
	assert.NotEmpty(t, res1.Results[1].Errors)
	assert.Equal(t, "n_dead_gt_n_exposed", res1.Results[1].Errors[0].Rule)

	// Idempotency: resubmitting exact same batch
	res2, err := svc.ProcessBatch(ctx, req1)
	require.NoError(t, err)
	require.Len(t, res2.Results, 2)

	// rec-valid-01 is now "already_received" and NOT duplicated
	assert.Equal(t, "already_received", res2.Results[0].Status)
	// rec-invalid-02 was never stored, so it is evaluated afresh and rejected again
	assert.Equal(t, "rejected", res2.Results[1].Status)

	// Resubmission: enumerator corrects rec-invalid-02 (n_dead = 40) under the SAME record_id
	correctedReq := sync.BatchSyncRequest{
		DeviceID: "device_phone_001",
		Records: []validation.RecordInput{
			{
				RecordID:       "rec-invalid-02", // SAME record_id
				State:          "Osun",
				LGA:            "Osogbo",
				Ward:           "Ward 2",
				Latitude:       7.77,
				Longitude:      4.56,
				CollectionDate: "2026-10-01",
				Species:        &species,
				Insecticide:    &insecticide,
				NExposed:       100,
				NDead:          40, // CORRECTED dead mosquitoes
				ITNUsePct:      &itn,
			},
		},
	}

	res3, err := svc.ProcessBatch(ctx, correctedReq)
	require.NoError(t, err)
	require.Len(t, res3.Results, 1)

	// It now succeeds with "accepted"!
	assert.Equal(t, "rec-invalid-02", res3.Results[0].RecordID)
	assert.Equal(t, "accepted", res3.Results[0].Status)
	assert.Empty(t, res3.Results[0].Errors)

	// Verify database count is exactly 2 rows
	count, err := client.FieldRecord.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, count, "Database must contain exactly 2 unique records")
}
