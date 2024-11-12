package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ScaAuthUserDevice holds the schema definition for the ScaAuthUserDevice entity.
type ScaAuthUserDevice struct {
	ent.Schema
}

// Fields of the ScaAuthUserDevice.
func (ScaAuthUserDevice) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			SchemaType(map[string]string{
				dialect.MySQL: "bigint(20)",
			}).
			Unique().
			Comment("主键ID"),
		field.String("user_id").
			MaxLen(20).
			Comment("用户ID"),
		field.String("ip").
			MaxLen(20).
			Comment("登录IP"),
		field.String("location").
			MaxLen(20).
			Comment("地址"),
		field.String("agent").
			MaxLen(255).
			Comment("设备信息"),
		field.Time("created_at").
			Default(time.Now).
			Optional().
			Immutable().
			Comment("创建时间"),
		field.Time("update_at").
			Default(time.Now).UpdateDefault(time.Now).
			Nillable().
			Comment("更新时间"),
		field.Int("deleted").
			Default(0).
			Comment("是否删除"),
		field.String("browser").
			MaxLen(20).
			Comment("浏览器"),
		field.String("operating_system").
			MaxLen(20).
			Comment("操作系统"),
		field.String("browser_version").
			MaxLen(20).
			Comment("浏览器版本"),
		field.Int("mobile").
			Comment("是否为手机 0否1是"),
		field.Int("bot").
			Comment("是否为bot 0否1是"),
		field.String("mozilla").
			MaxLen(10).
			Comment("火狐版本"),
		field.String("platform").
			MaxLen(20).
			Comment("平台"),
		field.String("engine_name").
			MaxLen(20).
			Comment("引擎名称"),
		field.String("engine_version").
			MaxLen(20).
			Comment("引擎版本"),
	}
}

// Edges of the ScaAuthUserDevice.
func (ScaAuthUserDevice) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("sca_auth_user", ScaAuthUser.Type).
			Ref("sca_auth_user_device").
			Unique(),
	}
}

// Indexes of the ScaAuthUserDevice.
func (ScaAuthUserDevice) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id").
			Unique(),
	}
}
