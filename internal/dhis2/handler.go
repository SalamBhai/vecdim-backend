package dhis2

import (
	"net/http"

	"github.com/salambhai/vecdim-backend/ent"
	"github.com/salambhai/vecdim-backend/ent/fieldrecord"
	"github.com/salambhai/vecdim-backend/ent/predicate"
	"github.com/salambhai/vecdim-backend/internal/platform/response"
)

type Handler struct {
	client *ent.Client
}

func NewHandler(client *ent.Client) *Handler {
	return &Handler{client: client}
}

// HandleExportDHIS2 handles GET /export/dhis2?state=&month= (bonus).
func (h *Handler) HandleExportDHIS2(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	state := q.Get("state")
	month := q.Get("month") // e.g. "2026-10"

	var predicates []predicate.FieldRecord

	if state != "" {
		predicates = append(predicates, fieldrecord.StateEQ(state))
	}
	if month != "" {
		predicates = append(predicates, fieldrecord.CollectionDateHasPrefix(month))
	}

	records, err := h.client.FieldRecord.Query().
		Where(predicates...).
		Limit(500).
		All(r.Context())

	if err != nil {
		response.InternalError(w, err)
		return
	}

	dhis2Payload := ConvertRecordsToDHIS2(records)
	response.JSON(w, http.StatusOK, dhis2Payload)
}
