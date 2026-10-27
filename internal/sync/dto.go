package sync

import "github.com/salambhai/vecdim-backend/internal/validation"

// BatchSyncRequest represents the incoming mobile payload.
type BatchSyncRequest struct {
	DeviceID string                   `json:"device_id,omitempty"`
	Records  []validation.RecordInput `json:"records"`
}

// RecordSyncResult holds the per-record sync status.
type RecordSyncResult struct {
	RecordID string                      `json:"record_id"`
	Status   string                      `json:"status"` // accepted, accepted_with_flags, rejected, already_received
	Flags    []validation.ValidationError `json:"flags"`
	Errors   []validation.ValidationError `json:"errors"`
}

// BatchSyncResponse is the response returned to the mobile app.
type BatchSyncResponse struct {
	Results []RecordSyncResult `json:"results"`
}
