package templates

import (
	"bytes"
	"embed"
	"fmt"
	"path/filepath"
	"strings"
	texttemplate "text/template"
)

//go:embed project app
var templateFS embed.FS

type ProjectData struct {
	ProjectName       string
	ModulePath        string
	GogoModuleVersion string
}

type AppData struct {
	AppName    string
	AppLabel   string
	ModulePath string
}

type templateFile struct {
	TemplatePath string
	TargetPath   string
}

var projectTemplateFiles = []templateFile{
	{TemplatePath: "project/go.mod.tmpl", TargetPath: "go.mod"},
	{TemplatePath: "project/manage.go.tmpl", TargetPath: "manage.go"},
	{TemplatePath: "project/gitignore.tmpl", TargetPath: ".gitignore"},
	{TemplatePath: "project/env.example.tmpl", TargetPath: ".env.example"},
	{TemplatePath: "project/Makefile.tmpl", TargetPath: "Makefile"},
	{TemplatePath: "project/README.md.tmpl", TargetPath: "README.md"},
	{TemplatePath: "project/agent/rules/gogo.md.tmpl", TargetPath: ".agent/rules/gogo.md"},
	{TemplatePath: "project/agent/rules/gogo/package-feature-index.md.tmpl", TargetPath: ".agent/rules/gogo/package-feature-index.md"},
	{TemplatePath: "project/agent/rules/gogo/project-structure.md.tmpl", TargetPath: ".agent/rules/gogo/project-structure.md"},
	{TemplatePath: "project/agent/rules/gogo/models-orm-migrations.md.tmpl", TargetPath: ".agent/rules/gogo/models-orm-migrations.md"},
	{TemplatePath: "project/agent/rules/gogo/http-admin-api-auth.md.tmpl", TargetPath: ".agent/rules/gogo/http-admin-api-auth.md"},
	{TemplatePath: "project/agent/rules/gogo/forms-templates-static.md.tmpl", TargetPath: ".agent/rules/gogo/forms-templates-static.md"},
	{TemplatePath: "project/agent/rules/gogo/queue-workers.md.tmpl", TargetPath: ".agent/rules/gogo/queue-workers.md"},
	{TemplatePath: "project/agent/rules/gogo/settings-security.md.tmpl", TargetPath: ".agent/rules/gogo/settings-security.md"},
	{TemplatePath: "project/agent/rules/gogo/testing-deployment.md.tmpl", TargetPath: ".agent/rules/gogo/testing-deployment.md"},
	{TemplatePath: "project/app.go.tmpl", TargetPath: "{{.ProjectName}}/app.go"},
	{TemplatePath: "project/settings/base.go.tmpl", TargetPath: "{{.ProjectName}}/settings/base.go"},
	{TemplatePath: "project/settings/local.go.tmpl", TargetPath: "{{.ProjectName}}/settings/local.go"},
	{TemplatePath: "project/settings/test.go.tmpl", TargetPath: "{{.ProjectName}}/settings/test.go"},
	{TemplatePath: "project/settings/production.go.tmpl", TargetPath: "{{.ProjectName}}/settings/production.go"},
	{TemplatePath: "project/urls.go.tmpl", TargetPath: "{{.ProjectName}}/urls.go"},
	{TemplatePath: "project/admin.go.tmpl", TargetPath: "{{.ProjectName}}/admin.go"},
	{TemplatePath: "project/middleware.go.tmpl", TargetPath: "{{.ProjectName}}/middleware.go"},
	{TemplatePath: "project/queue.go.tmpl", TargetPath: "{{.ProjectName}}/queue.go"},
	{TemplatePath: "project/deploy/docker/Dockerfile.tmpl", TargetPath: "deploy/docker/Dockerfile"},
	{TemplatePath: "project/deploy/docker/docker-compose.yml.tmpl", TargetPath: "deploy/docker/docker-compose.yml"},
}

var projectKeepFiles = []string{
	"apps/.keep",
	"fixtures/.keep",
	"media/.keep",
	"static/.keep",
	"tests/integration/.keep",
}

var appTemplateFiles = []templateFile{
	{TemplatePath: "app/app.go.tmpl", TargetPath: "app.go"},
	{TemplatePath: "app/models.go.tmpl", TargetPath: "models.go"},
	{TemplatePath: "app/admin.go.tmpl", TargetPath: "admin.go"},
	{TemplatePath: "app/urls.go.tmpl", TargetPath: "urls.go"},
	{TemplatePath: "app/api.go.tmpl", TargetPath: "api.go"},
	{TemplatePath: "app/serializers.go.tmpl", TargetPath: "serializers.go"},
	{TemplatePath: "app/forms.go.tmpl", TargetPath: "forms.go"},
	{TemplatePath: "app/services.go.tmpl", TargetPath: "services.go"},
	{TemplatePath: "app/tasks.go.tmpl", TargetPath: "tasks.go"},
	{TemplatePath: "app/permissions.go.tmpl", TargetPath: "permissions.go"},
	{TemplatePath: "app/tests.go.tmpl", TargetPath: "tests/{{.AppLabel}}_test.go"},
}

var appKeepFiles = []string{
	"migrations/.keep",
	"templates/{{.AppLabel}}/.keep",
	"static/{{.AppLabel}}/.keep",
}

func ProjectFiles(data ProjectData) (map[string]string, error) {
	if strings.TrimSpace(data.ProjectName) == "" {
		return nil, fmt.Errorf("project name is required")
	}
	if strings.TrimSpace(data.ModulePath) == "" {
		data.ModulePath = data.ProjectName
	}
	files := make(map[string]string, len(projectTemplateFiles)+len(projectKeepFiles)+1)
	for _, item := range projectTemplateFiles {
		targetPath, err := renderString(item.TargetPath, data)
		if err != nil {
			return nil, err
		}
		contents, err := renderTemplate(item.TemplatePath, data)
		if err != nil {
			return nil, err
		}
		files[filepath.FromSlash(targetPath)] = contents
	}
	files["templates/base.html"] = "<!doctype html>\n<html><body>{{ block \"content\" . }}{{ end }}</body></html>\n"
	for _, keep := range projectKeepFiles {
		files[filepath.FromSlash(keep)] = ""
	}
	return files, nil
}

func AppFiles(data AppData) (map[string]string, error) {
	if strings.TrimSpace(data.AppName) == "" {
		return nil, fmt.Errorf("app name is required")
	}
	if strings.TrimSpace(data.AppLabel) == "" {
		data.AppLabel = data.AppName
	}
	files := make(map[string]string, len(appTemplateFiles)+len(appKeepFiles))
	for _, item := range appTemplateFiles {
		targetPath, err := renderString(item.TargetPath, data)
		if err != nil {
			return nil, err
		}
		contents, err := renderTemplate(item.TemplatePath, data)
		if err != nil {
			return nil, err
		}
		files[filepath.FromSlash(targetPath)] = contents
	}
	for _, keep := range appKeepFiles {
		targetPath, err := renderString(keep, data)
		if err != nil {
			return nil, err
		}
		files[filepath.FromSlash(targetPath)] = ""
	}
	return files, nil
}

func renderTemplate(path string, data any) (string, error) {
	parsed, err := texttemplate.ParseFS(templateFS, path)
	if err != nil {
		return "", err
	}
	var output bytes.Buffer
	if err := parsed.Execute(&output, data); err != nil {
		return "", err
	}
	return output.String(), nil
}

func renderString(value string, data any) (string, error) {
	parsed, err := texttemplate.New("path").Parse(value)
	if err != nil {
		return "", err
	}
	var output bytes.Buffer
	if err := parsed.Execute(&output, data); err != nil {
		return "", err
	}
	return output.String(), nil
}
