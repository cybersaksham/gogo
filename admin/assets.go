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

//go:embed templates static/*
var embeddedAssets embed.FS

var adminTemplatePartials = []string{
	"templates/actions.html",
	"templates/app_list.html",
	"templates/change_form_object_tools.html",
	"templates/change_list_object_tools.html",
	"templates/change_list_results.html",
	"templates/color_theme_toggle.html",
	"templates/date_hierarchy.html",
	"templates/filter.html",
	"templates/includes/fieldset.html",
	"templates/includes/object_delete_summary.html",
	"templates/nav_sidebar.html",
	"templates/pagination.html",
	"templates/prepopulated_fields_js.html",
	"templates/search_form.html",
	"templates/submit_line.html",
	"templates/widgets/clearable_file_input.html",
	"templates/widgets/date.html",
	"templates/widgets/foreign_key_raw_id.html",
	"templates/widgets/many_to_many_raw_id.html",
	"templates/widgets/radio.html",
	"templates/widgets/related_widget_wrapper.html",
	"templates/widgets/split_datetime.html",
	"templates/widgets/time.html",
	"templates/widgets/url.html",
	"templates/edit_inline/stacked.html",
	"templates/edit_inline/tabular.html",
}

var standaloneAdminTemplates = map[string]struct{}{
	"popup_response.html": {},
}

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

	if _, ok := standaloneAdminTemplates[name]; ok {
		tpl, err := template.ParseFS(embeddedAssets, "templates/"+name)
		if err != nil {
			return "", err
		}
		var buffer bytes.Buffer
		if err := tpl.Execute(&buffer, data); err != nil {
			return "", err
		}
		return buffer.String(), nil
	}

	tpl, err := template.ParseFS(embeddedAssets, templateFilesFor(name)...)
	if err != nil {
		return "", err
	}
	var buffer bytes.Buffer
	if err := tpl.ExecuteTemplate(&buffer, "base", data); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func templateFilesFor(name string) []string {
	target := "templates/" + name
	files := []string{"templates/base.html"}
	seen := map[string]struct{}{"templates/base.html": {}}
	for _, partial := range adminTemplatePartials {
		if _, ok := seen[partial]; ok {
			continue
		}
		files = append(files, partial)
		seen[partial] = struct{}{}
	}
	if _, ok := seen[target]; !ok {
		files = append(files, target)
	}
	return files
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
