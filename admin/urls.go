package admin

import (
	"context"
	"strings"

	gogohttp "github.com/cybersaksham/gogo/http"
)

// URLs builds the namespaced admin router for this site.
func (s *Site) URLs() (*gogohttp.Router, error) {
	router := gogohttp.NewRouter()
	routes := []struct {
		name    string
		pattern string
	}{
		{"admin:index", s.URLPrefix + "/"},
		{"admin:login", s.URLPrefix + "/login/"},
		{"admin:logout", s.URLPrefix + "/logout/"},
		{"admin:password_change", s.URLPrefix + "/password_change/"},
		{"admin:app_list", s.URLPrefix + "/<str:app_label>/"},
	}
	for _, route := range routes {
		if err := router.Handle(route.name, route.pattern, placeholderView(route.name), "GET", "POST"); err != nil {
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
