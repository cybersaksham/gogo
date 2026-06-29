package admin

import (
	"context"
	"html"
	"strings"

	gogohttp "github.com/cybersaksham/gogo/http"
)

// URLs builds the namespaced admin router for this site.
func (s *Site) URLs() (*gogohttp.Router, error) {
	router := gogohttp.NewRouter()
	routes := []struct {
		name    string
		pattern string
		view    gogohttp.View
	}{
		{"admin:index", s.URLPrefix + "/", adminIndexView(s)},
		{"admin:login", s.URLPrefix + "/login/", placeholderView("admin:login")},
		{"admin:logout", s.URLPrefix + "/logout/", placeholderView("admin:logout")},
		{"admin:password_change", s.URLPrefix + "/password_change/", placeholderView("admin:password_change")},
		{"admin:app_list", s.URLPrefix + "/<str:app_label>/", adminAppListView(s)},
	}
	for _, route := range routes {
		if err := router.Handle(route.name, route.pattern, route.view, "GET", "POST"); err != nil {
			return nil, err
		}
	}

	for _, label := range s.ModelRegistry.RegisteredModels() {
		modelAdmin, ok := s.ModelRegistry.GetAdmin(label)
		if !ok {
			continue
		}
		if err := registerModelURLs(router, s, modelAdmin); err != nil {
			return nil, err
		}
	}
	return router, nil
}

func registerModelURLs(router *gogohttp.Router, site *Site, admin ModelAdmin) error {
	appLabel := strings.ToLower(admin.Model.AppLabel)
	modelName := strings.ToLower(admin.Model.ModelName)
	prefix := site.URLPrefix + "/" + appLabel + "/" + modelName
	namePrefix := "admin:" + appLabel + "_" + modelName
	routes := []struct {
		name    string
		pattern string
	}{
		{namePrefix + "_changelist", prefix + "/"},
		{namePrefix + "_add", prefix + "/add/"},
		{namePrefix + "_change", prefix + "/<path:object_id>/change/"},
		{namePrefix + "_delete", prefix + "/<path:object_id>/delete/"},
		{namePrefix + "_history", prefix + "/<path:object_id>/history/"},
		{namePrefix + "_autocomplete", prefix + "/autocomplete/"},
		{namePrefix + "_jsi18n", prefix + "/jsi18n/"},
	}
	for _, route := range routes {
		if err := router.Handle(route.name, route.pattern, placeholderView(route.name), "GET", "POST"); err != nil {
			return err
		}
	}
	for _, custom := range admin.GetURLs(nil) {
		pattern := prefix + "/" + strings.TrimLeft(custom.Path, "/")
		if !strings.HasSuffix(pattern, "/") {
			pattern += "/"
		}
		if err := router.Handle(namePrefix+"_"+custom.Name, pattern, placeholderView(custom.Name), "GET", "POST"); err != nil {
			return err
		}
	}
	return nil
}

func placeholderView(name string) gogohttp.View {
	return func(context.Context, *gogohttp.Request) gogohttp.Response {
		return gogohttp.Text(200, name)
	}
}

func adminIndexView(site *Site) gogohttp.View {
	return func(context.Context, *gogohttp.Request) gogohttp.Response {
		return gogohttp.HTML(200, renderAdminIndex(site, ""))
	}
}

func adminAppListView(site *Site) gogohttp.View {
	return func(_ context.Context, request *gogohttp.Request) gogohttp.Response {
		return gogohttp.HTML(200, renderAdminIndex(site, strings.ToLower(request.PathParam("app_label"))))
	}
}

func renderAdminIndex(site *Site, onlyApp string) string {
	if site == nil {
		site = DefaultSite()
	}
	var builder strings.Builder
	builder.WriteString("<!doctype html><html><head><meta charset=\"utf-8\"><title>")
	builder.WriteString(html.EscapeString(site.Title))
	builder.WriteString("</title></head><body><header><h1>")
	builder.WriteString(html.EscapeString(site.Header))
	builder.WriteString("</h1></header><main><h2>")
	builder.WriteString(html.EscapeString(site.IndexTitle))
	builder.WriteString("</h2>")

	apps := groupedAdminModels(site, onlyApp)
	if len(apps) == 0 {
		builder.WriteString("<p>No admin models registered.</p>")
	} else {
		for _, app := range apps {
			builder.WriteString("<section><h3><a href=\"")
			builder.WriteString(html.EscapeString(site.URLPrefix + "/" + app.AppLabel + "/"))
			builder.WriteString("\">")
			builder.WriteString(html.EscapeString(app.AppLabel))
			builder.WriteString("</a></h3><ul>")
			for _, model := range app.Models {
				modelPath := site.URLPrefix + "/" + app.AppLabel + "/" + strings.ToLower(model.Name) + "/"
				builder.WriteString("<li><a href=\"")
				builder.WriteString(html.EscapeString(modelPath))
				builder.WriteString("\">")
				builder.WriteString(html.EscapeString(model.Name))
				builder.WriteString("</a> <a href=\"")
				builder.WriteString(html.EscapeString(modelPath + "add/"))
				builder.WriteString("\">Add</a></li>")
			}
			builder.WriteString("</ul></section>")
		}
	}
	builder.WriteString("</main></body></html>")
	return builder.String()
}

func groupedAdminModels(site *Site, onlyApp string) []IndexApp {
	appPositions := map[string]int{}
	var apps []IndexApp
	for _, label := range site.ModelRegistry.RegisteredModels() {
		modelAdmin, ok := site.ModelRegistry.GetAdmin(label)
		if !ok {
			continue
		}
		appLabel := strings.ToLower(modelAdmin.Model.AppLabel)
		if onlyApp != "" && appLabel != onlyApp {
			continue
		}
		position, ok := appPositions[appLabel]
		if !ok {
			apps = append(apps, IndexApp{AppLabel: appLabel})
			position = len(apps) - 1
			appPositions[appLabel] = position
		}
		apps[position].Models = append(apps[position].Models, IndexModel{
			AppLabel: appLabel,
			Name:     modelAdmin.Model.ModelName,
		})
	}
	return apps
}
