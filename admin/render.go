package admin

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"unicode"

	"github.com/cybersaksham/gogo/auth"
	gogohttp "github.com/cybersaksham/gogo/http"
)

type adminBreadcrumb struct {
	URL   string
	Label string
}

type adminSubmitButton struct {
	Name  string
	Value string
	Label string
	Class string
}

type adminListFilter struct {
	Name  string
	Label string
}

type adminPageData struct {
	Site              *Site
	Title             string
	Header            string
	IndexTitle        string
	ContentTitle      string
	BodyClass         string
	UserName          string
	SiteURL           string
	LogoutURL         string
	PasswordChangeURL string
	ContentClass      string
	OmitContentClass  bool
	StaticCSSURL      string
	StaticCSSURLs     []string
	StaticHeadJSURLs  []string
	StaticJSURL       string
	StaticJSURLs      []string
	CSRFToken         string
	Breadcrumbs       []adminBreadcrumb
	Apps              []IndexApp

	AppLabel               string
	ModelName              string
	ModelVerboseName       string
	ModelVerboseNamePlural string
	AddURL                 string
	ChangeListURL          string
	DeleteURL              string
	HistoryURL             string
	SearchQuery            string
	SearchHelpText         string
	ListFilters            []adminListFilter
	ChangeList             ChangeList
	Form                   adminFormData
	Deletion               DeletionSummary
	History                HistoryPage

	Next  string
	Error string

	csrfCookie *http.Cookie
}

type adminFormData struct {
	ID          string
	Fieldsets   []adminFieldsetData
	SaveButtons []adminSubmitButton
	SaveOnTop   bool
	CanDelete   bool
	DeleteURL   string
	HistoryURL  string
}

type adminFieldsetData struct {
	Name   string
	Fields []adminFormFieldData
}

type adminFormFieldData struct {
	Name       string
	Label      string
	FieldID    string
	FieldCSS   string
	Readonly   bool
	HelpText   string
	Errors     string
	WidgetHTML template.HTML
}

func renderAdminTemplate(name string, data adminPageData) gogohttp.Response {
	rendered, err := RenderTemplate(name, data, nil)
	if err != nil {
		return gogohttp.InternalServerError(err)
	}
	response := gogohttp.HTML(http.StatusOK, rendered)
	if data.csrfCookie != nil {
		response.Header().Add("Set-Cookie", data.csrfCookie.String())
	}
	return response
}

func baseAdminPageData(site *Site, request *http.Request, title, contentTitle, bodyClass string) adminPageData {
	site = adminSiteOrDefault(site)
	data := adminPageData{
		Site:              site,
		Title:             title + " | " + site.Title,
		Header:            site.Header,
		IndexTitle:        site.IndexTitle,
		ContentTitle:      contentTitle,
		BodyClass:         strings.TrimSpace(bodyClass),
		SiteURL:           site.URLPrefix + "/",
		LogoutURL:         site.URLPrefix + "/logout/",
		PasswordChangeURL: site.URLPrefix + "/password_change/",
		ContentClass:      "colM",
		StaticCSSURL:      site.URLPrefix + "/static/admin.css",
		StaticCSSURLs: []string{
			site.URLPrefix + "/static/admin/css/base.css",
			site.URLPrefix + "/static/admin/css/dark_mode.css",
			site.URLPrefix + "/static/admin/css/nav_sidebar.css",
			site.URLPrefix + "/static/admin/css/dashboard.css",
			site.URLPrefix + "/static/admin/css/forms.css",
			site.URLPrefix + "/static/admin/css/changelists.css",
			site.URLPrefix + "/static/admin/css/login.css",
			site.URLPrefix + "/static/admin/css/widgets.css",
			site.URLPrefix + "/static/admin/css/responsive.css",
		},
		StaticHeadJSURLs: []string{site.URLPrefix + "/static/admin/js/theme.js"},
		StaticJSURL:      site.URLPrefix + "/static/admin.js",
		StaticJSURLs: []string{
			site.URLPrefix + "/static/admin/js/nav_sidebar.js",
			site.URLPrefix + "/static/admin.js",
		},
		Breadcrumbs: []adminBreadcrumb{{URL: site.URLPrefix + "/", Label: "Home"}},
	}
	data.UserName = adminUserDisplayName(site, request)
	data.CSRFToken, data.csrfCookie = adminCSRFPageToken(request)
	return data
}

func modelAdminPageData(site *Site, request *http.Request, modelAdmin ModelAdmin, title, contentTitle, actionClass string) adminPageData {
	site = adminSiteOrDefault(site)
	appLabel := strings.ToLower(modelAdmin.Model.AppLabel)
	modelName := strings.ToLower(modelAdmin.Model.ModelName)
	modelURL := site.URLPrefix + "/" + appLabel + "/" + modelName + "/"
	verboseName := modelVerboseName(modelAdmin)
	verbosePlural := modelVerboseNamePlural(modelAdmin)
	data := baseAdminPageData(site, request, title, contentTitle, strings.Join([]string{
		"dashboard",
		"app-" + adminClassName(appLabel),
		"model-" + adminClassName(modelName),
		actionClass,
	}, " "))
	data.AppLabel = appLabel
	data.ModelName = modelAdmin.Model.ModelName
	data.ModelVerboseName = verboseName
	data.ModelVerboseNamePlural = verbosePlural
	data.AddURL = modelURL + "add/"
	data.ChangeListURL = modelURL
	data.SearchQuery = request.URL.Query().Get("q")
	data.SearchHelpText = modelAdmin.SearchHelpText
	data.ListFilters = listFilters(modelAdmin)
	data.Breadcrumbs = []adminBreadcrumb{
		{URL: site.URLPrefix + "/", Label: "Home"},
		{URL: site.URLPrefix + "/" + appLabel + "/", Label: appLabel},
		{URL: modelURL, Label: modelAdmin.Model.ModelName},
	}
	return data
}

func changeFormViewData(modelAdmin ModelAdmin, context ChangeFormContext) adminFormData {
	form := adminFormData{
		ID:          strings.ToLower(modelAdmin.Model.ModelName) + "_form",
		SaveButtons: submitButtons(context.SaveButtons),
		SaveOnTop:   context.SaveOnTop,
		CanDelete:   context.CanDelete,
		DeleteURL:   context.DeleteURL,
	}
	if context.ObjectID != "" {
		form.HistoryURL = context.ObjectID + "/history/"
	}
	for _, fieldset := range context.Fieldsets {
		fieldsetData := adminFieldsetData{Name: fieldset.Name}
		for _, fieldName := range fieldset.Fields {
			field, ok := context.Fields[fieldName]
			if !ok {
				field = ChangeFormField{Name: fieldName, Widget: WidgetText}
			}
			fieldsetData.Fields = append(fieldsetData.Fields, adminFormFieldData{
				Name:       fieldName,
				Label:      adminLabel(fieldName),
				FieldID:    "id_" + fieldName,
				FieldCSS:   "form-row field-" + adminClassName(fieldName),
				Readonly:   field.Readonly,
				WidgetHTML: template.HTML(renderAdminFormWidget(field)),
			})
		}
		form.Fieldsets = append(form.Fieldsets, fieldsetData)
	}
	return form
}

func renderAdminFormWidget(field ChangeFormField) string {
	value := field.Value
	if value == nil {
		value = ""
		if field.Readonly {
			value = "-"
		}
	}
	config := WidgetConfig{
		Name:  field.Name,
		Value: value,
		Attrs: map[string]string{
			"id":    "id_" + field.Name,
			"class": "vTextField",
		},
		RelationURL: "autocomplete/",
	}
	switch field.Widget {
	case WidgetReadonly:
		return ReadonlyDisplay(config)
	case WidgetRawID:
		config.Attrs["class"] = "vForeignKeyRawIdAdminField"
		return RawIDRelationWidget(config)
	case WidgetAutocomplete:
		config.Attrs["class"] = "admin-autocomplete"
		return AutocompleteWidget(config)
	case WidgetRadio:
		return Select(config)
	case WidgetFilteredSelectMultiple:
		return FilteredSelectMultiple(config)
	default:
		return TextInput(config)
	}
}

func submitButtons(buttons []SaveButton) []adminSubmitButton {
	result := make([]adminSubmitButton, 0, len(buttons))
	for _, button := range buttons {
		switch button {
		case SaveButtonSave:
			result = append(result, adminSubmitButton{Name: "_save", Value: "Save", Label: "Save", Class: "default"})
		case SaveButtonSaveAndContinue:
			result = append(result, adminSubmitButton{Name: "_continue", Value: "Save and continue editing", Label: "Save and continue editing"})
		case SaveButtonSaveAndAddAnother:
			result = append(result, adminSubmitButton{Name: "_addanother", Value: "Save and add another", Label: "Save and add another"})
		case SaveButtonSaveAsNew:
			result = append(result, adminSubmitButton{Name: "_saveasnew", Value: "Save as new", Label: "Save as new"})
		}
	}
	return result
}

func listFilters(modelAdmin ModelAdmin) []adminListFilter {
	filters := make([]adminListFilter, 0, len(modelAdmin.ListFilter))
	for _, name := range modelAdmin.ListFilter {
		filters = append(filters, adminListFilter{Name: name, Label: adminLabel(name)})
	}
	return filters
}

func modelVerboseName(modelAdmin ModelAdmin) string {
	if modelAdmin.Model.VerboseName != "" {
		return modelAdmin.Model.VerboseName
	}
	return strings.ToLower(modelAdmin.Model.ModelName)
}

func modelVerboseNamePlural(modelAdmin ModelAdmin) string {
	if modelAdmin.Model.VerboseNamePlural != "" {
		return modelAdmin.Model.VerboseNamePlural
	}
	return modelVerboseName(modelAdmin) + "s"
}

func rowsFromModelAdmin(modelAdmin ModelAdmin, request *http.Request) []map[string]any {
	if modelAdmin.Hooks.GetQuerySet == nil {
		return nil
	}
	switch rows := modelAdmin.Hooks.GetQuerySet(request).(type) {
	case []map[string]any:
		return cloneRows(rows)
	default:
		return nil
	}
}

func adminRequestUser(site *Site, request *http.Request) (auth.User, bool) {
	if user, ok := auth.UserFromContext(request.Context()); ok && user.IsAuthenticated() && !user.IsAnonymous() {
		return user, true
	}
	if site != nil {
		if provider, ok := site.PermissionPolicy.(interface {
			UserForRequest(*http.Request) (auth.User, bool)
		}); ok {
			return provider.UserForRequest(request)
		}
	}
	return auth.User{}, false
}

func adminUserDisplayName(site *Site, request *http.Request) string {
	user, ok := adminRequestUser(site, request)
	if !ok {
		return ""
	}
	if user.Username != "" {
		return user.Username
	}
	if user.Email != "" {
		return user.Email
	}
	if user.ID != 0 {
		return fmt.Sprint(user.ID)
	}
	return ""
}

func adminSiteOrDefault(site *Site) *Site {
	if site != nil {
		return site
	}
	return DefaultSite()
}

func adminLabel(name string) string {
	name = strings.TrimSpace(strings.ReplaceAll(name, "_", " "))
	if name == "" {
		return ""
	}
	words := strings.Fields(name)
	for i, word := range words {
		runes := []rune(strings.ToLower(word))
		if len(runes) == 0 {
			continue
		}
		runes[0] = unicode.ToUpper(runes[0])
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}

func adminClassName(value string) string {
	value = strings.ToLower(value)
	var builder strings.Builder
	previousDash := false
	for _, char := range value {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			builder.WriteRune(char)
			previousDash = false
			continue
		}
		if !previousDash {
			builder.WriteByte('-')
			previousDash = true
		}
	}
	return strings.Trim(builder.String(), "-")
}
