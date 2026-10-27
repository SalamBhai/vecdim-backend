package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			NotEmpty().
			Unique().
			Comment("User identifier e.g. u_field_01, u_vecdim_osun, u_uniosun_01"),
		field.String("name").
			NotEmpty(),
		field.String("role").
			NotEmpty().
			Comment("Role: field, vecdim, or uniosun"),
		field.String("state").
			Optional().
			Nillable().
			Comment("State jurisdiction: Lagos, Kano, Osun, or empty for national"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("records", FieldRecord.Type),
	}
}
