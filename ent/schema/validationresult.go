package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ValidationResult holds the schema definition for the ValidationResult entity.
type ValidationResult struct {
	ent.Schema
}

// Fields of the ValidationResult.
func (ValidationResult) Fields() []ent.Field {
	return []ent.Field{
		field.String("record_id").
			NotEmpty(),
		field.String("rule_id").
			NotEmpty().
			Comment("e.g. GPS_outside_nigeria, n_dead_gt_n_exposed, etc."),
		field.String("field").
			NotEmpty().
			Comment("Target input field name"),
		field.String("severity").
			NotEmpty().
			Comment("BLOCK or WARN"),
		field.String("message").
			NotEmpty().
			Comment("Human-readable correction guidance for enumerator"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Indexes of the ValidationResult.
func (ValidationResult) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("record_id"),
		index.Fields("rule_id"),
	}
}

// Edges of the ValidationResult.
func (ValidationResult) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("record", FieldRecord.Type).
			Ref("validation_results").
			Field("record_id").
			Unique().
			Required(),
	}
}
