package alerts

import (
	"context"
	"fmt"
	"time"

	"github.com/salambhai/vecdim-backend/ent"
	"github.com/salambhai/vecdim-backend/ent/alert"
	"github.com/salambhai/vecdim-backend/ent/predicate"
	"github.com/salambhai/vecdim-backend/internal/auth"
)

type Service struct {
	client *ent.Client
}

func NewService(client *ent.Client) *Service {
	return &Service{client: client}
}

type AlertFilter struct {
	State string `json:"state"`
	LGA   string `json:"lga"`
	From  string `json:"from"`
	To    string `json:"to"`
	Tier  string `json:"tier"`
	Limit int    `json:"limit"`
}

type AcknowledgeRequest struct {
	Note string `json:"note"`
}

// GetAlerts returns alerts filtered by state, lga, month range, and role.
func (s *Service) GetAlerts(ctx context.Context, filter AlertFilter) ([]*ent.Alert, error) {
	user, _ := auth.GetUser(ctx)

	var predicates []predicate.Alert

	// Role scoping: vecdim state user only sees their own state alerts
	if user.Role == "vecdim" && user.State != "" {
		predicates = append(predicates, alert.StateEQ(user.State))
	} else if filter.State != "" {
		predicates = append(predicates, alert.StateEQ(filter.State))
	}

	if filter.LGA != "" {
		predicates = append(predicates, alert.LgaEQ(filter.LGA))
	}
	if filter.From != "" {
		predicates = append(predicates, alert.MonthGTE(filter.From))
	}
	if filter.To != "" {
		predicates = append(predicates, alert.MonthLTE(filter.To))
	}
	if filter.Tier != "" {
		predicates = append(predicates, alert.TierEQ(filter.Tier))
	}

	query := s.client.Alert.Query().
		Where(predicates...).
		Order(ent.Desc(alert.FieldMonth), ent.Desc(alert.FieldMeanRisk))

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}

	return query.All(ctx)
}

// AcknowledgeAlert handles POST /alerts/{id}/acknowledge.
func (s *Service) AcknowledgeAlert(ctx context.Context, alertID string, req AcknowledgeRequest) (*ent.Alert, error) {
	user, _ := auth.GetUser(ctx)
	now := time.Now()

	a, err := s.client.Alert.Query().
		Where(alert.IDEQ(alertID)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("alert not found: %s", alertID)
	}

	updater := s.client.Alert.UpdateOne(a).
		SetAcknowledged(true).
		SetAcknowledgedAt(now)

	if user.UserID != "" {
		updater.SetAcknowledgedBy(user.UserID)
	} else {
		updater.SetAcknowledgedBy("anonymous_operator")
	}

	if req.Note != "" {
		updater.SetAcknowledgementNote(req.Note)
	}

	return updater.Save(ctx)
}
