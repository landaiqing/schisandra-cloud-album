package mixin

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

type DefaultMixin struct {
	mixin.Schema
}

func (DefaultMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			Comment("创建时间"),
		field.Time("updated_at").
			Default(time.Now).
			Comment("更新时间").
			UpdateDefault(time.Now),
		field.Int8("deleted").
			Default(0).
			Max(1).
			Optional().
			Comment("是否删除 0 未删除 1 已删除"),
	}
}
