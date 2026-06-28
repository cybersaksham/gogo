package gogo_test

import (
	"context"
	"fmt"

	"github.com/cybersaksham/gogo/api"
	"github.com/cybersaksham/gogo/conf"
	"github.com/cybersaksham/gogo/models"
	"github.com/cybersaksham/gogo/orm"
	"github.com/cybersaksham/gogo/orm/dialects/postgres"
	"github.com/cybersaksham/gogo/queue"
)

func Example_referenceCoreAPIs() {
	settings := conf.DefaultSettings()
	settings.SecretKey = "dev-secret"
	settings.DatabaseURL = "sqlite:///tmp/gogo.sqlite3"

	meta := models.Metadata{AppLabel: "blog", ModelName: "Post", TableName: "blog_post", Fields: []models.FieldMeta{{Name: "id"}, {Name: "title"}}}
	compiled, _ := orm.NewCompiler(postgres.New()).CompileSelect(orm.NewQuery(meta).Select("id", "title"))

	serializer := api.NewSerializer(api.StringField("title", api.FieldOptions{Required: true}))
	_, _, valid := serializer.Validate(map[string]any{"title": "Hello"})

	app := queue.NewApp(queue.AppOptions{})
	_, _ = app.RegisterTask("blog.publish", func(context.Context, ...any) (any, error) { return "ok", nil }, queue.TaskOptions{})
	signature := queue.NewSignature("blog.publish", 1).WithQueue("default")

	fmt.Println(settings.Env)
	fmt.Println(meta.Label())
	fmt.Println(compiled.SQL)
	fmt.Println(valid)
	fmt.Println(signature.Name, signature.Options.Queue)
	// Output:
	// development
	// blog.Post
	// SELECT "id", "title" FROM "blog_post"
	// true
	// blog.publish default
}
