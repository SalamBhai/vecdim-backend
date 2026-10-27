package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// RiskPoint holds the schema definition for the RiskPoint entity (used for GIS spatial bounding-box queries).
type RiskPoint struct {
	ent.Schema
}

// Fields of the RiskPoint.
func (RiskPoint) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			NotEmpty().
			Unique().
			Comment("Record ID e.g. VD01324"),
		field.String("state").
			NotEmpty(),
		field.String("lga").
			NotEmpty(),
		field.String("ward").
			NotEmpty(),
		field.Float("latitude"),
		field.Float("longitude"),
		field.String("collection_date").
			NotEmpty(),
		field.Float("p_high").
			Comment("Predicted risk score"),
		field.String("tier").
			NotEmpty().
			Comment("ALERT, WARNING, WATCH"),
	}
}

// Indexes of the RiskPoint.
func (RiskPoint) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("state", "lga"),
		index.Fields("tier"),
		index.Fields("latitude", "longitude"),
	}
}
