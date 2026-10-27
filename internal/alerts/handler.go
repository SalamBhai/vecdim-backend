package alerts

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/salambhai/vecdim-backend/internal/platform/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// HandleGetAlerts handles GET /alerts?state=&lga=&from=&to=&tier=&limit= (must).
func (h *Handler) HandleGetAlerts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	limit := 0
	if lStr := q.Get("limit"); lStr != "" {
		if parsed, err := strconv.Atoi(lStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	filter := AlertFilter{
		State: q.Get("state"),
		LGA:   q.Get("lga"),
		From:  q.Get("from"),
		To:    q.Get("to"),
		Tier:  q.Get("tier"),
		Limit: limit,
	}

	alertsList, err := h.service.GetAlerts(r.Context(), filter)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	// Calculate summary tier totals for dashboard cards
	tierTotals := map[string]int{"ALERT": 0, "WARNING": 0, "WATCH": 0}
	for _, a := range alertsList {
		tierTotals[a.Tier]++
	}

	response.JSON(w, http.StatusOK, response.Envelope{
		"alerts":      alertsList,
		"count":       len(alertsList),
		"tier_totals": tierTotals,
	})
}

// HandleAcknowledgeAlert handles POST /alerts/{id}/acknowledge (bonus).
func (h *Handler) HandleAcknowledgeAlert(w http.ResponseWriter, r *http.Request) {
	alertID := chi.URLParam(r, "id")
	if alertID == "" {
		response.BadRequest(w, "Alert ID is required")
		return
	}

	var req AcknowledgeRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	updated, err := h.service.AcknowledgeAlert(r.Context(), alertID, req)
	if err != nil {
		response.NotFound(w, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, response.Envelope{
		"status": "acknowledged",
		"alert":  updated,
	})
}
