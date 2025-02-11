package main

import (
	"os"
	"path/filepath"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
)

const MySQLDSN = "root:LDQ20020618xxx@tcp(1.95.0.111:3306)/schisandra-cloud-album?charset=utf8mb4&parseTime=True&loc=Local"

func main() {

	// 连接数据库
	db, err := gorm.Open(mysql.Open(MySQLDSN))
	if err != nil {
		panic(err)
	}

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	path := filepath.Join(dir, "app/auth/model/mysql/", "query")
	// 生成实例
	g := gen.NewGenerator(gen.Config{
		// 相对执行`go run`时的路径, 会自动创建目录
		OutPath: path,
		// 生成的文件名，默认gen.go
		OutFile: "gen.go",
		// 生成DAO代码的包名，默认是：model
		ModelPkgPath: "model",
		// 是否为DAO包生成单元测试代码，默认：false
		WithUnitTest: false,

		// WithDefaultQuery 生成默认查询结构体(作为全局变量使用), 即`Q`结构体和其字段(各表模型)
		// WithoutContext 生成没有context调用限制的代码供查询
		// WithQueryInterface 生成interface形式的查询代码(可导出), 如`Where()`方法返回的就是一个可导出的接口类型
		Mode: gen.WithDefaultQuery | gen.WithQueryInterface | gen.WithoutContext,

		// 表字段可为 null 值时, 对应结体字段使用指针类型
		FieldNullable: false, // generate pointer when field is nullable

		// 表字段默认值与模型结构体字段零值不一致的字段, 在插入数据时需要赋值该字段值为零值的, 结构体字段须是指针类型才能成功, 即`FieldCoverable:true`配置下生成的结构体字段.
		// 因为在插入时遇到字段为零值的会被GORM赋予默认值. 如字段`age`表默认值为10, 即使你显式设置为0最后也会被GORM设为10提交.
		// 如果该字段没有上面提到的插入时赋零值的特殊需要, 则字段为非指针类型使用起来会比较方便.
		FieldCoverable: true,

		// 模型结构体字段的数字类型的符号表示是否与表字段的一致, `false`指示都用有符号类型
		FieldSignable: false,
		// 生成 gorm 标签的字段索引属性
		FieldWithIndexTag: true,
		// 生成 gorm 标签的字段类型属性
		FieldWithTypeTag: true,
	})
	// 设置目标 db
	g.UseDB(db)

	// 自定义字段的数据类型
	// 统一数字类型为int64,兼容protobuf
	dataMap := map[string]func(columnType gorm.ColumnType) (dataType string){
		"tinyint":   func(columnType gorm.ColumnType) (dataType string) { return "int64" },
		"smallint":  func(columnType gorm.ColumnType) (dataType string) { return "int64" },
		"mediumint": func(columnType gorm.ColumnType) (dataType string) { return "int64" },
		"bigint":    func(columnType gorm.ColumnType) (dataType string) { return "int64" },
		"int":       func(columnType gorm.ColumnType) (dataType string) { return "int64" },
	}
	// 要先于`ApplyBasic`执行
	g.WithDataTypeMap(dataMap)

	// 自定义模型结体字段的标签
	// 将特定字段名的 json 标签加上`string`属性,即 MarshalJSON 时该字段由数字类型转成字符串类型
	jsonField := gen.FieldJSONTagWithNS(func(columnName string) (tagContent string) {
		toStringField := `id, `
		if strings.Contains(toStringField, columnName) {
			return columnName + ",string"
		}
		return columnName
	})
	// 将非默认字段名的字段定义为自动时间戳和软删除字段;
	// 自动时间戳默认字段名为:`updated_at`、`created_at, 表字段数据类型为: INT 或 DATETIME
	// 软删除默认字段名为:`deleted_at`, 表字段数据类型为: DATETIME
	idField := gen.FieldGORMTag("id", func(tag field.GormTag) field.GormTag {
		return tag.Append("primary_key")
	})
	autoUpdateTimeField := gen.FieldGORMTag("updated_at", func(tag field.GormTag) field.GormTag {
		return tag.Append("autoUpdateTime")
	})
	autoCreateTimeField := gen.FieldGORMTag("created_at", func(tag field.GormTag) field.GormTag {
		return tag.Append("autoCreateTime")
	})
	softDeleteField := gen.FieldType("delete_at", "gorm.DeletedAt")
	versionField := gen.FieldType("version", "optimisticlock.Version")
	// 模型自定义选项组
	fieldOpts := []gen.ModelOpt{jsonField, idField, autoUpdateTimeField, autoCreateTimeField, softDeleteField, versionField}

	// 创建全部模型文件, 并覆盖前面创建的同名模型
	scaAuthMenu := g.GenerateModel("sca_auth_menu", fieldOpts...)
	scaAuthPermissionRule := g.GenerateModel("sca_auth_permission_rule", fieldOpts...)
	scaAuthRole := g.GenerateModel("sca_auth_role", fieldOpts...)
	scaAuthUser := g.GenerateModel("sca_auth_user", fieldOpts...)
	scaAuthUserDevice := g.GenerateModel("sca_auth_user_device", fieldOpts...)
	scaAuthUserSocial := g.GenerateModel("sca_auth_user_social", fieldOpts...)
	scaCommentReply := g.GenerateModel("sca_comment_reply", fieldOpts...)
	scaCommentLikes := g.GenerateModel("sca_comment_likes", fieldOpts...)
	scaMessageReport := g.GenerateModel("sca_message_report", fieldOpts...)
	scaStorageConfig := g.GenerateModel("sca_storage_config", fieldOpts...)
	scaStorageInfo := g.GenerateModel("sca_storage_info", fieldOpts...)
	scaStorageTag := g.GenerateModel("sca_storage_tag", fieldOpts...)
	scaStorageTagInfo := g.GenerateModel("sca_storage_tag_info", fieldOpts...)
	scaUserFollows := g.GenerateModel("sca_user_follows", fieldOpts...)
	scaUserLevel := g.GenerateModel("sca_user_level", fieldOpts...)
	scaUserMessage := g.GenerateModel("sca_user_message", fieldOpts...)
	scaStorageAlbum := g.GenerateModel("sca_storage_album", fieldOpts...)
	scaStorageLocation := g.GenerateModel("sca_storage_location", fieldOpts...)

	g.ApplyBasic(
		scaAuthMenu,
		scaAuthPermissionRule,
		scaAuthRole,
		scaAuthUser,
		scaAuthUserDevice,
		scaAuthUserSocial,
		scaCommentReply,
		scaCommentLikes,
		scaMessageReport,
		scaStorageConfig,
		scaStorageInfo,
		scaStorageTag,
		scaStorageTagInfo,
		scaUserFollows,
		scaUserLevel,
		scaUserMessage,
		scaStorageAlbum,
		scaStorageLocation,
	)

	g.Execute()
}
