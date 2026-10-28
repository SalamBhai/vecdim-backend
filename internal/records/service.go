package records

import (
	"context"

	"github.com/salambhai/vecdim-backend/ent"
	"github.com/salambhai/vecdim-backend/ent/fieldrecord"
	"github.com/salambhai/vecdim-backend/ent/predicate"
	"github.com/salambhai/vecdim-backend/internal/auth"
)

type Service struct {
	client *ent.Client
}

func NewService(client *ent.Client) *Service {
	return &Service{client: client}
}

type QueryFilter struct {
	State string `json:"state"`
	LGA   string `json:"lga"`
	Since string `json:"since"`
	Limit int    `json:"limit"`
}

// GetRecords returns records filtered by role access policy and query parameters.
func (s *Service) GetRecords(ctx context.Context, filter QueryFilter) ([]*ent.FieldRecord, error) {
	user, _ := auth.GetUser(ctx)

	var predicates []predicate.FieldRecord

	// Role-based scoping
	switch user.Role {
	case "field":
		predicates = append(predicates, fieldrecord.UserID(user.UserID))
	case "vecdim":
		// State user is locked to their assigned state
		state := user.State
		if state != "" {
			predicates = append(predicates, fieldrecord.StateEQ(state))
		}
	case "uniosun":
		// National user sees everything
	}

	// Apply URL filter overrides (if permitted by role)
	if filter.State != "" && user.Role != "vecdim" {
		predicates = append(predicates, fieldrecord.StateEqualFold(filter.State))
	}
	if filter.LGA != "" {
		predicates = append(predicates, fieldrecord.LgaEqualFold(filter.LGA))
	}
	if filter.Since != "" {
		predicates = append(predicates, fieldrecord.CollectionDateGTE(filter.Since))
	}

	query := s.client.FieldRecord.Query().
		Where(predicates...).
		Order(ent.Desc(fieldrecord.FieldCollectionDate))

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	} else {
		query = query.Limit(100) // Default page limit
	}

	return query.All(ctx)
}
