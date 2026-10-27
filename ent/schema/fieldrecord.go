package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// FieldRecord holds the schema definition for the FieldRecord entity.
type FieldRecord struct {
	ent.Schema
}

// Fields of the FieldRecord.
func (FieldRecord) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			NotEmpty().
			Unique().
			Comment("Record UUID generated on device or dataset ID"),
		field.String("user_id").
			Optional().
			Nillable(),
		field.String("device_id").
			Optional().
			Nillable(),
		field.String("sync_batch_id").
			Optional().
			Nillable(),
		field.String("state").
			NotEmpty(),
		field.String("lga").
			NotEmpty(),
		field.String("ward").
			NotEmpty(),
		field.Float("latitude"),
		field.Float("longitude"),
		field.Float("gps_accuracy_m").
			Optional().
			Nillable().
			Comment("Horizontal GPS accuracy in metres"),
		field.String("collection_date").
			NotEmpty().
			Comment("ISO Date string YYYY-MM-DD"),
		field.String("species").
			Optional().
			Nillable().
			Comment("Mosquito species e.g. An. gambiae s.s."),
		field.String("insecticide").
			Optional().
			Nillable().
			Comment("Tested insecticide e.g. Permethrin, Deltamethrin"),
		field.Int("n_exposed").
			Default(0),
		field.Int("n_dead").
			Default(0),
		field.Float("mortality_pct").
			Optional().
			Nillable().
			Comment("Rounded to 1 decimal place: 100 * n_dead / n_exposed"),
		field.String("who_class").
			Optional().
			Nillable().
			Comment("Susceptible (>=98.0%), Possible resistance (90.0-97.9%), Resistant (<90.0%)"),
		field.Int("larval_density").
			Optional().
			Nillable(),
		field.Float("rainfall_mm").
			Optional().
			Nillable(),
		field.Float("temp_c").
			Optional().
			Nillable(),
		field.Float("ndvi").
			Optional().
			Nillable(),
		field.Float("elevation_m").
			Optional().
			Nillable(),
		field.Float("dist_water_km").
			Optional().
			Nillable(),
		field.Int("urban").
			Default(0),
		field.Float("itn_use_pct").
			Optional().
			Nillable(),
		field.Int("irs_covered").
			Default(0),
		field.Float("slide_positivity_pct").
			Optional().
			Nillable(),
		field.Int("high_risk").
			Optional().
			Nillable(),
		field.String("status").
			Default("UNPROCESSED").
			Comment("UNPROCESSED, ACCEPTED, ACCEPTED_WITH_FLAGS, FLAGGED, REJECTED"),
		field.Time("captured_at").
			Optional().
			Nillable(),
		field.Time("received_at").
			Default(time.Now),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Indexes of the FieldRecord.
func (FieldRecord) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("state", "lga"),
		index.Fields("collection_date"),
		index.Fields("status"),
		index.Fields("user_id"),
	}
}

// Edges of the FieldRecord.
func (FieldRecord) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("records").
			Field("user_id").
			Unique(),
		edge.From("sync_batch", SyncBatch.Type).
			Ref("records").
			Field("sync_batch_id").
			Unique(),
		edge.To("validation_results", ValidationResult.Type),
	}
}
