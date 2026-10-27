package pipeline

import (
	"net/http"

	"github.com/salambhai/vecdim-backend/internal/platform/response"
)

type Handler struct {
	runner *Runner
}

func NewHandler(runner *Runner) *Handler {
	return &Handler{runner: runner}
}

// HandleRunPipeline handles POST /pipeline/run.
func (h *Handler) HandleRunPipeline(w http.ResponseWriter, r *http.Request) {
	trigger := r.URL.Query().Get("trigger")
	if trigger == "" {
		trigger = "on_demand_api"
	}

	result, err := h.runner.TriggerPipeline(r.Context(), trigger)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, response.Envelope{
		"message":  "R Pipeline executed successfully",
		"pipeline": result,
	})
}
