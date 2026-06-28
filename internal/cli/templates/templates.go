package templates

import (
	"bytes"
	"embed"
	"fmt"
	"path/filepath"
	"strings"
	texttemplate "text/template"
)

//go:embed project
var templateFS embed.FS

type ProjectData struct {
	ProjectName string
	ModulePath  string
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
