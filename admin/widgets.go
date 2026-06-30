package admin

import (
	"fmt"
	"html"
	"sort"
	"strings"
	"time"
)

// WidgetChoice describes one select option.
type WidgetChoice struct {
	Value string
	Label string
}

// WidgetConfig configures an admin widget render.
type WidgetConfig struct {
	Name                     string
	Value                    any
	Attrs                    map[string]string
	Choices                  []WidgetChoice
	RelationURL              string
	LinkURL                  string
	LinkLabel                string
	LinkTitle                string
	InitialURL               string
	InitialLabel             string
	RelatedModelName         string
	RelatedModelLabel        string
	AddRelatedURL            string
	ChangeRelatedTemplateURL string
	DeleteRelatedTemplateURL string
	ViewRelatedTemplateURL   string
	URLParams                string
	ViewURLParams            string
	CanAddRelated            bool
	CanChangeRelated         bool
	CanDeleteRelated         bool
	CanViewRelated           bool
	IsHidden                 bool
	ModelHasLimitChoicesTo   bool
}

// TextInput renders a text input.
func TextInput(config WidgetConfig) string {
	return input("text", config, nil)
}

// Textarea renders a textarea.
func Textarea(config WidgetConfig) string {
	return fmt.Sprintf(`<textarea name="%s"%s>%s</textarea>`, esc(config.Name), renderAttrs(config.Attrs), esc(fmt.Sprint(config.Value)))
}

// NumberInput renders a number input.
func NumberInput(config WidgetConfig) string {
	return input("number", config, nil)
}

// EmailInput renders an email input.
func EmailInput(config WidgetConfig) string {
	return input("email", config, nil)
}

// Checkbox renders a checkbox input.
func Checkbox(config WidgetConfig) string {
	extra := map[string]string{}
	if widgetBool(config.Value) {
		extra["checked"] = "checked"
	}
	return inputWithValue("checkbox", config, extra, false)
}

// Select renders a select input.
func Select(config WidgetConfig) string {
	return selectWidget(config, false, selectedValues(config.Value))
}

// SelectMultiple renders a multiple select input.
func SelectMultiple(config WidgetConfig) string {
	return selectWidget(config, true, selectedValues(config.Value))
}

// DateInput renders a date input.
func DateInput(config WidgetConfig) string {
	return `<p class="date">` + adminDateInput(config) + `</p>`
}

// TimeInput renders a time input.
func TimeInput(config WidgetConfig) string {
	return `<p class="time">` + adminTimeInput(config) + `</p>`
}

// DateTimeInput renders a datetime-local input.
func DateTimeInput(config WidgetConfig) string {
	dateValue, timeValue := splitAdminDateTime(config.Value)
	dateConfig := config
	dateConfig.Name = config.Name + "_0"
	dateConfig.Value = dateValue
	dateConfig.Attrs = cloneStringMap(config.Attrs)
	if dateConfig.Attrs == nil {
		dateConfig.Attrs = map[string]string{}
	}
	dateConfig.Attrs["id"] = "id_" + dateConfig.Name
	timeConfig := config
	timeConfig.Name = config.Name + "_1"
	timeConfig.Value = timeValue
	timeConfig.Attrs = cloneStringMap(config.Attrs)
	if timeConfig.Attrs == nil {
		timeConfig.Attrs = map[string]string{}
	}
	timeConfig.Attrs["id"] = "id_" + timeConfig.Name
	return `<p class="datetime">` +
		fmt.Sprintf(`<label for="%s">Date:</label> `, esc("id_"+dateConfig.Name)) + adminDateInput(dateConfig) + `<br>` +
		fmt.Sprintf(`<label for="%s">Time:</label> `, esc("id_"+timeConfig.Name)) + adminTimeInput(timeConfig) +
		`</p>`
}

// FileInput renders a file input.
func FileInput(config WidgetConfig) string {
	return inputWithoutValue("file", config, nil)
}

// ClearableFileInput renders a file input with a clear checkbox.
func ClearableFileInput(config WidgetConfig) string {
	if fmt.Sprint(config.Value) == "" {
		return FileInput(config)
	}
	label := config.InitialLabel
	if label == "" {
		label = fmt.Sprint(config.Value)
	}
	url := config.InitialURL
	if url == "" {
		url = label
	}
	clearID := config.Name + "-clear_id"
	return `<p class="file-upload">Currently: ` +
		fmt.Sprintf(`<a href="%s">%s</a>`, esc(url), esc(label)) +
		`<span class="clearable-file-input">` +
		fmt.Sprintf(`<input type="checkbox" name="%s-clear" id="%s">`, esc(config.Name), esc(clearID)) +
		fmt.Sprintf(`<label for="%s">Clear</label></span><br>`, esc(clearID)) +
		`Change:` + FileInput(config) + `</p>`
}

// RawIDRelationWidget renders a raw ID relation input.
func RawIDRelationWidget(config WidgetConfig) string {
	rendered := input("text", config, nil)
	if config.RelationURL == "" {
		return rendered
	}
	title := config.LinkTitle
	if title == "" {
		title = "Lookup"
	}
	rendered += fmt.Sprintf(`<a href="%s" class="related-lookup" id="lookup_id_%s" title="%s"></a>`, esc(config.RelationURL), esc(config.Name), esc(title))
	if config.LinkLabel != "" {
		label := esc(config.LinkLabel)
		if config.LinkURL != "" {
			label = fmt.Sprintf(`<a href="%s">%s</a>`, esc(config.LinkURL), label)
		}
		rendered += `<strong>` + label + `</strong>`
	}
	return `<div>` + rendered + `</div>`
}

// AutocompleteWidget renders an autocomplete relation input.
func AutocompleteWidget(config WidgetConfig) string {
	return input("text", config, map[string]string{"data-autocomplete-url": config.RelationURL})
}

// FilteredSelectMultiple renders a Django-style filtered multiple select.
func FilteredSelectMultiple(config WidgetConfig) string {
	attrs := cloneStringMap(config.Attrs)
	if attrs == nil {
		attrs = map[string]string{}
	}
	attrs["class"] = strings.TrimSpace(attrs["class"] + " filtered-select-multiple selectfilter")
	config.Attrs = attrs
	return SelectMultiple(config)
}

// ReadonlyDisplay renders an escaped readonly value.
func ReadonlyDisplay(config WidgetConfig) string {
	return fmt.Sprintf(`<span class="readonly" data-field="%s">%s</span>`, esc(config.Name), esc(fmt.Sprint(config.Value)))
}

// PasswordHashDisplay renders Django's read-only password hash summary.
func PasswordHashDisplay(config WidgetConfig) string {
	algorithm, iterations, salt, hash := splitPasswordHash(fmt.Sprint(config.Value))
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf(`<div aria-describedby="%s_helptext" disabled id="%s">`, esc("id_"+config.Name), esc("id_"+config.Name)))
	if algorithm != "" {
		builder.WriteString(`<p>`)
		builder.WriteString(fmt.Sprintf(`<strong>algorithm</strong>: <bdi>%s</bdi> `, esc(algorithm)))
		if iterations != "" {
			builder.WriteString(fmt.Sprintf(`<strong>iterations</strong>: <bdi>%s</bdi> `, esc(iterations)))
		}
		if salt != "" {
			builder.WriteString(fmt.Sprintf(`<strong>salt</strong>: <bdi>%s</bdi> `, esc(maskHashPart(salt, 6))))
		}
		if hash != "" {
			builder.WriteString(fmt.Sprintf(`<strong>hash</strong>: <bdi>%s</bdi>`, esc(maskHashPart(hash, 6))))
		}
		builder.WriteString(`</p>`)
	} else {
		builder.WriteString(`<p><strong>No password set.</strong></p>`)
	}
	builder.WriteString(`<p><a class="button" href="../password/" role="button">Reset password</a></p>`)
	builder.WriteString(`</div>`)
	return builder.String()
}

// RelatedWidgetWrapper renders Django admin's relation action shell.
func RelatedWidgetWrapper(config WidgetConfig, renderedWidget string) string {
	var builder strings.Builder
	builder.WriteString(`<div class="related-widget-wrapper"`)
	if !config.ModelHasLimitChoicesTo && config.RelatedModelName != "" {
		builder.WriteString(fmt.Sprintf(` data-model-ref="%s"`, esc(config.RelatedModelName)))
	}
	builder.WriteString(`>`)
	builder.WriteString(renderedWidget)
	if !config.IsHidden {
		model := config.RelatedModelLabel
		if model == "" {
			model = config.RelatedModelName
		}
		if config.CanChangeRelated {
			builder.WriteString(relatedActionLink("change", config.Name, "", appendQuery(config.ChangeRelatedTemplateURL, config.URLParams), "Change selected "+model, "icon-changelink.svg"))
		}
		if config.CanAddRelated {
			builder.WriteString(relatedActionLink("add", config.Name, appendQuery(config.AddRelatedURL, config.URLParams), "", "Add another "+model, "icon-addlink.svg"))
		}
		if config.CanDeleteRelated {
			builder.WriteString(relatedActionLink("delete", config.Name, "", appendQuery(config.DeleteRelatedTemplateURL, config.URLParams), "Delete selected "+model, "icon-deletelink.svg"))
		}
		if config.CanViewRelated {
			params := config.ViewURLParams
			if params == "" {
				params = config.URLParams
			}
			templateURL := config.ViewRelatedTemplateURL
			if templateURL == "" {
				templateURL = config.ChangeRelatedTemplateURL
			}
			builder.WriteString(relatedActionLink("view", config.Name, "", appendQuery(templateURL, params), "View selected "+model, "icon-viewlink.svg"))
		}
	}
	builder.WriteString(`</div>`)
	return builder.String()
}

func input(inputType string, config WidgetConfig, extra map[string]string) string {
	return inputWithValue(inputType, config, extra, true)
}

func inputWithoutValue(inputType string, config WidgetConfig, extra map[string]string) string {
	return inputWithValue(inputType, config, extra, false)
}

func inputWithValue(inputType string, config WidgetConfig, extra map[string]string, includeValue bool) string {
	attrs := cloneStringMap(config.Attrs)
	if attrs == nil {
		attrs = map[string]string{}
	}
	for key, value := range extra {
		attrs[key] = value
	}
	value := ""
	if includeValue {
		value = fmt.Sprintf(` value="%s"`, esc(fmt.Sprint(config.Value)))
	}
	return fmt.Sprintf(`<input type="%s" name="%s"%s%s>`, esc(inputType), esc(config.Name), value, renderAttrs(attrs))
}

func adminDateInput(config WidgetConfig) string {
	config = withWidgetClass(config, "vDateField")
	config = withAttr(config, "size", "10")
	return input("text", config, nil)
}

func adminTimeInput(config WidgetConfig) string {
	config = withWidgetClass(config, "vTimeField")
	config = withAttr(config, "size", "8")
	return input("text", config, nil)
}

func withWidgetClass(config WidgetConfig, className string) WidgetConfig {
	attrs := cloneStringMap(config.Attrs)
	if attrs == nil {
		attrs = map[string]string{}
	}
	attrs["class"] = strings.TrimSpace(attrs["class"] + " " + className)
	config.Attrs = attrs
	return config
}

func withAttr(config WidgetConfig, name, value string) WidgetConfig {
	attrs := cloneStringMap(config.Attrs)
	if attrs == nil {
		attrs = map[string]string{}
	}
	attrs[name] = value
	config.Attrs = attrs
	return config
}

func selectWidget(config WidgetConfig, multiple bool, selected map[string]struct{}) string {
	attrs := cloneStringMap(config.Attrs)
	if attrs == nil {
		attrs = map[string]string{}
	}
	if multiple {
		attrs["multiple"] = "multiple"
	}
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf(`<select name="%s"%s>`, esc(config.Name), renderAttrs(attrs)))
	for _, choice := range config.Choices {
		selectedAttr := ""
		if _, ok := selected[choice.Value]; ok {
			selectedAttr = " selected"
		}
		builder.WriteString(fmt.Sprintf(`<option value="%s"%s>%s</option>`, esc(choice.Value), selectedAttr, esc(choice.Label)))
	}
	builder.WriteString(`</select>`)
	return builder.String()
}

func selectedValues(value any) map[string]struct{} {
	result := map[string]struct{}{}
	switch typed := value.(type) {
	case []string:
		for _, item := range typed {
			result[item] = struct{}{}
		}
	default:
		if value != nil {
			result[fmt.Sprint(value)] = struct{}{}
		}
	}
	return result
}

func widgetBool(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		normalized := strings.ToLower(strings.TrimSpace(typed))
		return normalized == "1" || normalized == "true" || normalized == "on" || normalized == "yes"
	case int:
		return typed != 0
	case int64:
		return typed != 0
	default:
		return false
	}
}

func splitPasswordHash(value string) (algorithm, iterations, salt, hash string) {
	parts := strings.Split(value, "$")
	if len(parts) > 0 {
		algorithm = parts[0]
	}
	if len(parts) > 1 {
		iterations = parts[1]
	}
	if len(parts) > 2 {
		salt = parts[2]
	}
	if len(parts) > 3 {
		hash = parts[3]
	}
	return algorithm, iterations, salt, hash
}

func maskHashPart(value string, keep int) string {
	if value == "" {
		return ""
	}
	if keep <= 0 || len(value) <= keep {
		return value
	}
	return value[:keep] + strings.Repeat("*", min(16, len(value)-keep))
}

func splitAdminDateTime(value any) (string, string) {
	switch typed := value.(type) {
	case time.Time:
		if typed.IsZero() {
			return "", ""
		}
		return typed.Format("2006-01-02"), typed.Format("15:04:05")
	case string:
		value := strings.TrimSpace(typed)
		if value == "" {
			return "", ""
		}
		parts := strings.FieldsFunc(value, func(r rune) bool {
			return r == 'T' || r == ' '
		})
		if len(parts) >= 2 {
			return parts[0], strings.TrimSuffix(parts[1], "Z")
		}
		return value, ""
	default:
		if value == nil {
			return "", ""
		}
		return fmt.Sprint(value), ""
	}
}

func relatedActionLink(action, name, href, dataHrefTemplate, title, icon string) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf(`<a class="related-widget-wrapper-link %s-related" id="%s_id_%s"`, esc(action), esc(action), esc(name)))
	if dataHrefTemplate != "" {
		builder.WriteString(fmt.Sprintf(` data-href-template="%s"`, esc(dataHrefTemplate)))
	}
	if action == "add" {
		builder.WriteString(` data-popup="yes"`)
	}
	if href != "" {
		builder.WriteString(fmt.Sprintf(` href="%s"`, esc(href)))
	}
	if action == "change" || action == "delete" {
		builder.WriteString(` data-popup="yes"`)
	}
	builder.WriteString(fmt.Sprintf(` title="%s">`, esc(title)))
	builder.WriteString(fmt.Sprintf(`<img src="/admin/static/admin/img/%s" alt="" width="24" height="24">`, esc(icon)))
	builder.WriteString(`</a>`)
	return builder.String()
}

func appendQuery(baseURL, query string) string {
	if baseURL == "" || query == "" {
		return baseURL
	}
	separator := "?"
	if strings.Contains(baseURL, "?") {
		separator = "&"
	}
	return baseURL + separator + query
}

func renderAttrs(attrs map[string]string) string {
	if len(attrs) == 0 {
		return ""
	}
	keys := make([]string, 0, len(attrs))
	for key := range attrs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var builder strings.Builder
	for _, key := range keys {
		if attrs[key] == key {
			builder.WriteString(" ")
			builder.WriteString(esc(key))
			continue
		}
		builder.WriteString(fmt.Sprintf(` %s="%s"`, esc(key), esc(attrs[key])))
	}
	return builder.String()
}

func esc(value string) string {
	return html.EscapeString(value)
}
