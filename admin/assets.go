package admin

import (
	"bytes"
	"embed"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

//go:embed templates/*.html static/*
var embeddedAssets embed.FS

// AssetNames returns embedded admin asset paths.
func AssetNames() []string {
	var names []string
	_ = fs.WalkDir(embeddedAssets, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || path == "." {
			return err
		}
		names = append(names, path)
		return nil
	})
	sort.Strings(names)
	return names
}

// ReadAsset returns one embedded admin asset.
func ReadAsset(name string) ([]byte, bool) {
	body, err := embeddedAssets.ReadFile(name)
	return body, err == nil
}

// RenderTemplate renders an admin template, preferring project overrides.
func RenderTemplate(name string, data any, overrideDirs []string) (string, error) {
	if body, ok, err := readTemplateOverride(name, overrideDirs); ok || err != nil {
		if err != nil {
			return "", err
		}
		tpl, err := template.New(name).Parse(string(body))
		if err != nil {
			return "", err
		}
		var buffer bytes.Buffer
		if err := tpl.Execute(&buffer, data); err != nil {
			return "", err
		}
		return buffer.String(), nil
	}

	tpl, err := template.ParseFS(embeddedAssets, "templates/base.html", "templates/"+name)
	if err != nil {
		return "", err
	}
	var buffer bytes.Buffer
	if err := tpl.ExecuteTemplate(&buffer, "base", data); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func readTemplateOverride(name string, overrideDirs []string) ([]byte, bool, error) {
	for _, dir := range overrideDirs {
		for _, candidate := range []string{
			filepath.Join(dir, name),
			filepath.Join(dir, "admin", "templates", name),
		} {
			body, err := os.ReadFile(candidate)
			if err == nil {
				return body, true, nil
			}
			if !os.IsNotExist(err) {
				return nil, false, err
			}
		}
	}
	return nil, false, nil
}
