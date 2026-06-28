package admin

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/cybersaksham/gogo/auth"
)

// ChangeFormMode identifies add or edit mode.
type ChangeFormMode string

const (
	ChangeFormAdd  ChangeFormMode = "add"
	ChangeFormEdit ChangeFormMode = "edit"
)

// WidgetKind identifies admin field widgets.
type WidgetKind string

const (
	WidgetText                   WidgetKind = "text"
	WidgetReadonly               WidgetKind = "readonly"
	WidgetRawID                  WidgetKind = "raw_id"
	WidgetAutocomplete           WidgetKind = "autocomplete"
	WidgetRadio                  WidgetKind = "radio"
	WidgetFilteredSelectMultiple WidgetKind = "filtered_select_multiple"
)

// SaveButton identifies visible submit buttons.
type SaveButton string

const (
	SaveButtonSave              SaveButton = "save"
	SaveButtonSaveAndContinue   SaveButton = "save_continue"
	SaveButtonSaveAndAddAnother SaveButton = "save_add_another"
	SaveButtonSaveAsNew         SaveButton = "save_as_new"
)

// SaveIntent identifies the requested save outcome.
type SaveIntent string

const (
	SaveIntentSave       SaveIntent = "save"
	SaveIntentContinue   SaveIntent = "continue"
	SaveIntentAddAnother SaveIntent = "add_another"
	SaveIntentSaveAsNew  SaveIntent = "save_as_new"
)

// ChangeFormInput configures a change form context build.
type ChangeFormInput struct {
	Mode     ChangeFormMode
	ObjectID string
	User     auth.User
	Request  *http.Request
	Values   map[string]any
}

// ChangeFormContext stores render-ready add/edit metadata.
type ChangeFormContext struct {
	Mode                ChangeFormMode
	ObjectID            string
	Fieldsets           []Fieldset
	Fields              map[string]ChangeFormField
	PrepopulatedFields  map[string][]string
	SaveButtons         []SaveButton
	SaveOnTop           bool
	CanDelete           bool
	DeleteURL           string
	JSI18NURL           string
	Popup               bool
	RawIDFields         []string
	AutocompleteFields  []string
	RadioFields         map[string]string
	FilterHorizontal    []string
	FilterVertical      []string
	RelatedPopupEnabled bool
}

// ChangeFormField describes one rendered form field.
type ChangeFormField struct {
	Name     string
	Widget   WidgetKind
	Readonly bool
	Value    any
}

// BuildChangeForm builds add/edit form metadata with permission checks.
func BuildChangeForm(admin ModelAdmin, input ChangeFormInput) (ChangeFormContext, error) {
	user := input.User
	request := input.Request
	if request == nil {
		request, _ = http.NewRequest(http.MethodGet, "/", nil)
	}
	mode := input.Mode
	if mode == "" {
		mode = ChangeFormAdd
	}
	if mode == ChangeFormAdd && !admin.HasAddPermission(request, user) {
		return ChangeFormContext{}, ErrAdminPermissionDenied
	}
	if mode == ChangeFormEdit && !admin.HasChangePermission(request, user) {
		return ChangeFormContext{}, ErrAdminPermissionDenied
	}
	fields := admin.Fields
	if len(fields) == 0 {
		fields = []string{"__all__"}
	}
	context := ChangeFormContext{
		Mode:                mode,
		ObjectID:            input.ObjectID,
		Fieldsets:           cloneFieldsets(admin.Fieldsets),
		Fields:              buildChangeFormFields(admin, fields, input.Values),
		PrepopulatedFields:  cloneStringSliceMap(admin.PrepopulatedFields),
		SaveButtons:         saveButtons(admin),
		SaveOnTop:           admin.SaveOnTop,
		CanDelete:           mode == ChangeFormEdit || admin.HasDeletePermission(request, user),
		DeleteURL:           deleteURL(input.ObjectID),
		JSI18NURL:           "jsi18n/",
		Popup:               request.URL.Query().Get("_popup") == "1",
		RawIDFields:         append([]string(nil), admin.RawIDFields...),
		AutocompleteFields:  append([]string(nil), admin.AutocompleteFields...),
		RadioFields:         cloneStringMap(admin.RadioFields),
		FilterHorizontal:    append([]string(nil), admin.FilterHorizontal...),
		FilterVertical:      append([]string(nil), admin.FilterVertical...),
		RelatedPopupEnabled: true,
	}
	return context, nil
}

func buildChangeFormFields(admin ModelAdmin, fields []string, values map[string]any) map[string]ChangeFormField {
	readonly := setFromSlice(admin.ReadonlyFields)
	rawID := setFromSlice(admin.RawIDFields)
	autocomplete := setFromSlice(admin.AutocompleteFields)
	radio := setFromSlice(keys(admin.RadioFields))
	filtered := setFromSlice(append(append([]string(nil), admin.FilterHorizontal...), admin.FilterVertical...))
	result := make(map[string]ChangeFormField, len(fields)+len(readonly))
	for _, field := range fields {
		result[field] = ChangeFormField{Name: field, Widget: widgetForField(field, readonly, rawID, autocomplete, radio, filtered), Readonly: hasKey(readonly, field), Value: values[field]}
	}
	for field := range readonly {
		if _, ok := result[field]; !ok {
			result[field] = ChangeFormField{Name: field, Widget: WidgetReadonly, Readonly: true, Value: values[field]}
		}
	}
	return result
}

func widgetForField(field string, readonly, rawID, autocomplete, radio, filtered map[string]struct{}) WidgetKind {
	switch {
	case hasKey(readonly, field):
		return WidgetReadonly
	case hasKey(rawID, field):
		return WidgetRawID
	case hasKey(autocomplete, field):
		return WidgetAutocomplete
	case hasKey(radio, field):
		return WidgetRadio
	case hasKey(filtered, field):
		return WidgetFilteredSelectMultiple
	default:
		return WidgetText
	}
}

func saveButtons(admin ModelAdmin) []SaveButton {
	buttons := []SaveButton{SaveButtonSave, SaveButtonSaveAndContinue, SaveButtonSaveAndAddAnother}
	if admin.SaveAs {
		buttons = append(buttons, SaveButtonSaveAsNew)
	}
	return buttons
}

// ResolveSaveIntent parses admin submit button intent.
func ResolveSaveIntent(values url.Values) SaveIntent {
	switch {
	case values.Get("_continue") != "":
		return SaveIntentContinue
	case values.Get("_addanother") != "":
		return SaveIntentAddAnother
	case values.Get("_saveasnew") != "":
		return SaveIntentSaveAsNew
	default:
		return SaveIntentSave
	}
}

// RelatedPopup stores add/change related popup response metadata.
type RelatedPopup struct {
	Action     string
	ObjectID   string
	ObjectRepr string
}

// RelatedPopupResponse returns metadata consumed by admin popup JavaScript.
func RelatedPopupResponse(objectID, objectRepr string) RelatedPopup {
	return RelatedPopup{Action: "change", ObjectID: objectID, ObjectRepr: objectRepr}
}

// JavaScriptCatalogResponse stores admin widget translation JavaScript.
type JavaScriptCatalogResponse struct {
	ContentType string
	Body        string
}

// JavaScriptCatalog renders a tiny admin JavaScript translation catalog.
func JavaScriptCatalog(messages map[string]string) JavaScriptCatalogResponse {
	body, _ := json.Marshal(messages)
	return JavaScriptCatalogResponse{ContentType: "application/javascript", Body: "window.gogoAdminCatalog=" + string(body) + ";"}
}

func deleteURL(objectID string) string {
	if objectID == "" {
		return ""
	}
	return objectID + "/delete/"
}

func hasKey(set map[string]struct{}, key string) bool {
	_, ok := set[key]
	return ok
}

func keys(values map[string]string) []string {
	result := make([]string, 0, len(values))
	for key := range values {
		result = append(result, key)
	}
	return result
}
