package spatial

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/salambhai/vecdim-backend/internal/platform/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// HandleGetRiskPoints handles GET /risk-points?bbox=minLon,minLat,maxLon,maxLat (bonus).
func (h *Handler) HandleGetRiskPoints(w http.ResponseWriter, r *http.Request) {
	bboxStr := r.URL.Query().Get("bbox")
	if bboxStr == "" {
		// Default to entire Nigeria bounding box if not provided
		bboxStr = "2.5,4.0,15.0,14.0"
	}

	parts := strings.Split(bboxStr, ",")
	if len(parts) != 4 {
		response.BadRequest(w, "Invalid bbox format. Expected: minLon,minLat,maxLon,maxLat")
		return
	}

	minLon, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	minLat, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	maxLon, err3 := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
	maxLat, err4 := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		response.BadRequest(w, "Bounding box coordinates must be valid floating-point numbers")
		return
	}

	bbox := BoundingBox{
		MinLon: minLon,
		MinLat: minLat,
		MaxLon: maxLon,
		MaxLat: maxLat,
	}

	points, err := h.service.GetRiskPointsInBBox(r.Context(), bbox)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	// Format as GeoJSON FeatureCollection for direct Leaflet / Mapbox rendering!
	features := make([]response.Envelope, len(points))
	for i, p := range points {
		features[i] = response.Envelope{
			"type": "Feature",
			"geometry": response.Envelope{
				"type":        "Point",
				"coordinates": []float64{p.Longitude, p.Latitude},
			},
			"properties": response.Envelope{
				"record_id":       p.ID,
				"state":           p.State,
				"lga":             p.Lga,
				"ward":            p.Ward,
				"collection_date": p.CollectionDate,
				"p_high":          p.PHigh,
				"tier":            p.Tier,
			},
		}
	}

	response.JSON(w, http.StatusOK, response.Envelope{
		"type":        "FeatureCollection",
		"bbox":        []float64{minLon, minLat, maxLon, maxLat},
		"count":       len(points),
		"features":    features,
		"risk_points": points,
	})
}
