package records

import (
	"net/http"
	"strconv"

	"github.com/salambhai/vecdim-backend/internal/platform/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// HandleGetRecords handles GET /records?state=&lga=&since=&limit= (must).
func (h *Handler) HandleGetRecords(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	limit := 100
	if lStr := q.Get("limit"); lStr != "" {
		if parsed, err := strconv.Atoi(lStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	filter := QueryFilter{
		State: q.Get("state"),
		LGA:   q.Get("lga"),
		Since: q.Get("since"),
		Limit: limit,
	}

	records, err := h.service.GetRecords(r.Context(), filter)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, response.Envelope{
		"records": records,
		"count":   len(records),
	})
}
