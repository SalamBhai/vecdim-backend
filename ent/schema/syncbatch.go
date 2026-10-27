package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// SyncBatch holds the schema definition for the SyncBatch entity (sync outbox audit log).
type SyncBatch struct {
	ent.Schema
}

// Fields of the SyncBatch.
func (SyncBatch) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			DefaultFunc(func() string {
				return uuid.New().String()
			}).
			Unique(),
		field.String("device_id").
			Optional().
			Nillable(),
		field.String("user_id").
			Optional().
			Nillable(),
		field.Int("record_count").
			Default(0),
		field.Int("accepted_count").
			Default(0),
		field.Int("rejected_count").
			Default(0),
		field.Time("received_at").
			Default(time.Now).
			Immutable(),
		field.Text("payload_json").
			Optional().
			Nillable(),
	}
}

// Edges of the SyncBatch.
func (SyncBatch) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("records", FieldRecord.Type),
	}
}
