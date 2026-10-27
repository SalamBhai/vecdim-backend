package sync

import (
	"encoding/json"
	"net/http"

	"github.com/salambhai/vecdim-backend/internal/platform/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// HandleBatchSync handles POST /records/batch (must).
func (h *Handler) HandleBatchSync(w http.ResponseWriter, r *http.Request) {
	var req BatchSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON payload: "+err.Error())
		return
	}

	res, err := h.service.ProcessBatch(r.Context(), req)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, res)
}
