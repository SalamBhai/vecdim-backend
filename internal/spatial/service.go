package spatial

import (
	"context"
	"database/sql"

	"github.com/salambhai/vecdim-backend/ent"
	"github.com/salambhai/vecdim-backend/ent/riskpoint"
)

type Service struct {
	client *ent.Client
	db     *sql.DB
}

func NewService(client *ent.Client, db *sql.DB) *Service {
	return &Service{client: client, db: db}
}

type BoundingBox struct {
	MinLon float64
	MinLat float64
	MaxLon float64
	MaxLat float64
}

// GetRiskPointsInBBox queries risk prediction points within a spatial bounding box.
func (s *Service) GetRiskPointsInBBox(ctx context.Context, bbox BoundingBox) ([]*ent.RiskPoint, error) {
	// First attempt PostGIS ST_Intersects spatial query if raw DB pool is available
	if s.db != nil {
		query := `
			SELECT id, state, lga, ward, latitude, longitude, collection_date, p_high, tier
			FROM risk_points
			WHERE latitude >= $1 AND latitude <= $2 AND longitude >= $3 AND longitude <= $4
			ORDER BY p_high DESC
			LIMIT 500;
		`
		rows, err := s.db.QueryContext(ctx, query, bbox.MinLat, bbox.MaxLat, bbox.MinLon, bbox.MaxLon)
		if err == nil {
			defer rows.Close()
			var points []*ent.RiskPoint
			for rows.Next() {
				var p ent.RiskPoint
				if err := rows.Scan(&p.ID, &p.State, &p.Lga, &p.Ward, &p.Latitude, &p.Longitude, &p.CollectionDate, &p.PHigh, &p.Tier); err == nil {
					points = append(points, &p)
				}
			}
			return points, nil
		}
	}

	// Ent ORM fallback query
	return s.client.RiskPoint.Query().
		Where(
			riskpoint.LatitudeGTE(bbox.MinLat),
			riskpoint.LatitudeLTE(bbox.MaxLat),
			riskpoint.LongitudeGTE(bbox.MinLon),
			riskpoint.LongitudeLTE(bbox.MaxLon),
		).
		Order(ent.Desc(riskpoint.FieldPHigh)).
		Limit(500).
		All(ctx)
}
