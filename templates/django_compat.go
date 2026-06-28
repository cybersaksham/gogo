package templates

import (
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"math"
	"net/url"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func WithDjangoCompat(config HelperConfig) Option {
	return WithFuncMap(DjangoCompatHelpers(config))
}

func DjangoCompatHelpers(config HelperConfig) template.FuncMap {
	funcs := TemplateHelpers(config)
	for name, fn := range djangoTagHelpers(config) {
		funcs[name] = fn
	}
	for name, fn := range djangoFilterHelpers(config) {
		funcs[name] = fn
	}
	return funcs
}

func djangoTagHelpers(config HelperConfig) template.FuncMap {
	return template.FuncMap{
		"autoescape": func(value any, enabled bool) any {
			if enabled {
				return safeEscape(value)
			}
			return SafeString(fmt.Sprint(value))
		},
		"block":      func(_ string, fallback any) any { return fallback },
		"comment":    func(...any) string { return "" },
		"csrf_token": csrfToken,
		"cycle": func(index int, values ...any) any {
			if len(values) == 0 {
				return ""
			}
			return values[index%len(values)]
		},
		"debug": func(value any) string { return fmt.Sprintf("%#v", value) },
		"extends": func(name string) string {
			return name
		},
		"filter": func(value any) any { return value },
		"firstof": func(values ...any) any {
			for _, value := range values {
				if !helperEmpty(value) {
					return value
				}
			}
			return ""
		},
		"for_empty": func(values any, empty any) any {
			if lengthOf(values) == 0 {
				return empty
			}
			return values
		},
		"ifchanged": func(left any, right any) bool {
			return fmt.Sprint(left) != fmt.Sprint(right)
		},
		"include": func(name string) string { return name },
		"load":    func(...string) string { return "" },
		"lorem":   lorem,
		"now": func(layout string) string {
			return djangoNow(config).Format(layout)
		},
		"querystring": querystring,
		"regroup":     regroup,
		"resetcycle":  func() string { return "" },
		"spaceless":   spaceless,
		"templatetag": templateTag,
		"verbatim":    func(value any) string { return fmt.Sprint(value) },
		"widthratio":  widthRatio,
		"with":        func(value any) any { return value },
	}
}

func djangoFilterHelpers(config HelperConfig) template.FuncMap {
	return template.FuncMap{
		"add":                addValues,
		"addslashes":         addSlashes,
		"capfirst":           capFirst,
		"center":             center,
		"cut":                cut,
		"date":               formatDate,
		"default":            defaultValue,
		"default_if_none":    defaultIfNone,
		"dictsort":           dictSort,
		"divisibleby":        divisibleBy,
		"escape":             safeEscape,
		"escapejs":           escapeJS,
		"filesizeformat":     fileSizeFormat,
		"first":              firstValue,
		"floatformat":        floatFormat,
		"force_escape":       safeEscape,
		"get_digit":          getDigit,
		"join":               joinValues,
		"json_script":        jsonScript,
		"last":               lastValue,
		"length":             lengthOf,
		"length_is":          lengthIs,
		"linebreaks":         linebreaks,
		"linebreaksbr":       linebreaksBR,
		"linebreaks_br":      linebreaksBR,
		"linenumbers":        lineNumbers,
		"ljust":              leftJustify,
		"lower":              strings.ToLower,
		"make_list":          makeList,
		"phone2numeric":      phoneToNumeric,
		"pluralize":          pluralize,
		"pprint":             func(value any) string { return fmt.Sprintf("%#v", value) },
		"random":             firstValue,
		"rjust":              rightJustify,
		"safe":               func(value any) SafeString { return SafeString(fmt.Sprint(value)) },
		"safeseq":            safeSeq,
		"slice":              sliceValue,
		"slugify":            slugify,
		"stringformat":       stringFormat,
		"striptags":          stripTags,
		"time":               formatTimeOnly,
		"timesince":          func(value any) string { return durationPhrase(djangoNow(config).Sub(mustTime(value))) },
		"timeuntil":          func(value any) string { return durationPhrase(mustTime(value).Sub(djangoNow(config))) },
		"title":              strings.Title,
		"truncatechars":      truncateChars,
		"truncatechars_html": truncateCharsHTML,
		"truncatewords":      truncateWords,
		"truncatewords_html": truncateWordsHTML,
		"unordered_list":     unorderedList,
		"upper":              strings.ToUpper,
		"urlencode":          urlEncode,
		"url_encode":         urlEncode,
		"urlize":             urlize,
		"urlizetrunc":        urlizeTrunc,
		"wordcount":          wordCount,
		"wordwrap":           wordWrap,
		"yesno":              yesNo,
	}
}

func djangoNow(config HelperConfig) time.Time {
	if !config.Now.IsZero() {
		return config.Now
	}
	return time.Now()
}

func csrfToken(token string) SafeString {
	return SafeString(`<input type="hidden" name="csrfmiddlewaretoken" value="` + html.EscapeString(token) + `">`)
}

func querystring(base string, params map[string]any) string {
	parsed, err := url.Parse(base)
	if err != nil {
		return base
	}
	values := parsed.Query()
	for key, value := range params {
		values.Set(key, fmt.Sprint(value))
	}
	parsed.RawQuery = values.Encode()
	return parsed.String()
}

func lorem(count int) string {
	words := []string{"lorem", "ipsum", "dolor", "sit", "amet", "consectetur"}
	if count <= 0 {
		return ""
	}
	if count > len(words) {
		count = len(words)
	}
	return strings.Join(words[:count], " ")
}

func regroup(value any, key string) map[string][]any {
	groups := map[string][]any{}
	reflected := reflect.ValueOf(value)
	if reflected.Kind() != reflect.Array && reflected.Kind() != reflect.Slice {
		return groups
	}
	for i := 0; i < reflected.Len(); i++ {
		item := reflected.Index(i).Interface()
		groupKey := fmt.Sprint(fieldOrMapValue(item, key))
		groups[groupKey] = append(groups[groupKey], item)
	}
	return groups
}

func spaceless(value any) SafeString {
	text := strings.TrimSpace(fmt.Sprint(value))
	text = regexp.MustCompile(`>\s+<`).ReplaceAllString(text, "><")
	return SafeString(text)
}

func templateTag(name string) string {
	tags := map[string]string{
		"openblock":     "{%",
		"closeblock":    "%}",
		"openvariable":  "{{",
		"closevariable": "}}",
		"openbrace":     "{",
		"closebrace":    "}",
		"opencomment":   "{#",
		"closecomment":  "#}",
	}
	return tags[name]
}

func widthRatio(value, maxValue, width any) int {
	numerator, _ := toFloat(value)
	denominator, _ := toFloat(maxValue)
	target, _ := toFloat(width)
	if denominator == 0 {
		return 0
	}
	return int(math.Round((numerator / denominator) * target))
}

func addValues(left any, right any) any {
	leftInt, leftIntOK := toInt64(left)
	rightInt, rightIntOK := toInt64(right)
	if leftIntOK && rightIntOK {
		return leftInt + rightInt
	}
	leftFloat, leftFloatOK := toFloat(left)
	rightFloat, rightFloatOK := toFloat(right)
	if leftFloatOK && rightFloatOK {
		return leftFloat + rightFloat
	}
	return fmt.Sprint(left) + fmt.Sprint(right)
}

func addSlashes(value any) SafeString {
	text := fmt.Sprint(value)
	text = strings.ReplaceAll(text, `\`, `\\`)
	text = strings.ReplaceAll(text, `"`, `\"`)
	text = strings.ReplaceAll(text, `'`, `\'`)
	return SafeString(text)
}

func capFirst(value any) string {
	runes := []rune(fmt.Sprint(value))
	if len(runes) == 0 {
		return ""
	}
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func center(value any, width int) string {
	text := fmt.Sprint(value)
	if len(text) >= width {
		return text
	}
	padding := width - len(text)
	left := padding / 2
	right := padding - left
	return strings.Repeat(" ", left) + text + strings.Repeat(" ", right)
}

func cut(value any, remove string) string {
	return strings.ReplaceAll(fmt.Sprint(value), remove, "")
}

func defaultIfNone(value any, fallback any) any {
	if value == nil {
		return fallback
	}
	return value
}

func dictSort(value any) string {
	reflected := reflect.ValueOf(value)
	if reflected.Kind() != reflect.Map {
		return ""
	}
	keys := reflected.MapKeys()
	sort.Slice(keys, func(i, j int) bool {
		return fmt.Sprint(keys[i].Interface()) < fmt.Sprint(keys[j].Interface())
	})
	parts := make([]string, len(keys))
	for i, key := range keys {
		parts[i] = fmt.Sprintf("%v=%v", key.Interface(), reflected.MapIndex(key).Interface())
	}
	return strings.Join(parts, ",")
}

func divisibleBy(value any, divisor any) bool {
	left, leftOK := toInt64(value)
	right, rightOK := toInt64(divisor)
	return leftOK && rightOK && right != 0 && left%right == 0
}

func escapeJS(value any) string {
	replacer := strings.NewReplacer(
		`\`, `\\`,
		"\n", `\n`,
		"\r", `\r`,
		`"`, `\"`,
		`'`, `\'`,
		"<", `\u003C`,
		">", `\u003E`,
		"&", `\u0026`,
	)
	return replacer.Replace(fmt.Sprint(value))
}

func fileSizeFormat(value any) string {
	size, ok := toFloat(value)
	if !ok {
		return "0 bytes"
	}
	units := []string{"bytes", "KB", "MB", "GB", "TB"}
	index := 0
	for size >= 1024 && index < len(units)-1 {
		size /= 1024
		index++
	}
	if index == 0 {
		return fmt.Sprintf("%.0f %s", size, units[index])
	}
	return fmt.Sprintf("%.1f %s", size, units[index])
}

func firstValue(value any) any {
	return indexedValue(value, 0)
}

func lastValue(value any) any {
	reflected := reflect.ValueOf(value)
	if reflected.Kind() == reflect.Array || reflected.Kind() == reflect.Slice || reflected.Kind() == reflect.String {
		if reflected.Len() == 0 {
			return ""
		}
		return indexedValue(value, reflected.Len()-1)
	}
	return ""
}

func indexedValue(value any, index int) any {
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Array, reflect.Slice:
		if index < 0 || index >= reflected.Len() {
			return ""
		}
		return reflected.Index(index).Interface()
	case reflect.String:
		runes := []rune(reflected.String())
		if index < 0 || index >= len(runes) {
			return ""
		}
		return string(runes[index])
	default:
		return ""
	}
}

func floatFormat(value any, precision int) string {
	float, ok := toFloat(value)
	if !ok {
		return ""
	}
	return strconv.FormatFloat(float, 'f', precision, 64)
}

func getDigit(value any, position int) int {
	text := strconv.FormatInt(absInt(toInt64Default(value)), 10)
	if position <= 0 || position > len(text) {
		return 0
	}
	digit := text[len(text)-position]
	return int(digit - '0')
}

func jsonScript(value any, id string) (SafeString, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return SafeString(`<script id="` + html.EscapeString(id) + `" type="application/json">` + string(encoded) + `</script>`), nil
}

func lengthIs(value any, length int) bool {
	return lengthOf(value) == length
}

func linebreaksBR(value any) SafeString {
	escaped := html.EscapeString(fmt.Sprint(value))
	escaped = strings.ReplaceAll(escaped, "\r\n", "\n")
	escaped = strings.ReplaceAll(escaped, "\r", "\n")
	return SafeString(strings.ReplaceAll(escaped, "\n", "<br>"))
}

func lineNumbers(value any) string {
	lines := strings.Split(fmt.Sprint(value), "\n")
	for i := range lines {
		lines[i] = fmt.Sprintf("%d. %s", i+1, lines[i])
	}
	return strings.Join(lines, "\n")
}

func leftJustify(value any, width int) string {
	text := fmt.Sprint(value)
	if len(text) >= width {
		return text
	}
	return text + strings.Repeat(" ", width-len(text))
}

func rightJustify(value any, width int) string {
	text := fmt.Sprint(value)
	if len(text) >= width {
		return text
	}
	return strings.Repeat(" ", width-len(text)) + text
}

func makeList(value any) []string {
	runes := []rune(fmt.Sprint(value))
	values := make([]string, len(runes))
	for i, r := range runes {
		values[i] = string(r)
	}
	return values
}

func phoneToNumeric(value any) string {
	mapping := map[rune]rune{
		'A': '2', 'B': '2', 'C': '2',
		'D': '3', 'E': '3', 'F': '3',
		'G': '4', 'H': '4', 'I': '4',
		'J': '5', 'K': '5', 'L': '5',
		'M': '6', 'N': '6', 'O': '6',
		'P': '7', 'Q': '7', 'R': '7', 'S': '7',
		'T': '8', 'U': '8', 'V': '8',
		'W': '9', 'X': '9', 'Y': '9', 'Z': '9',
	}
	var builder strings.Builder
	for _, r := range strings.ToUpper(fmt.Sprint(value)) {
		if mapped, ok := mapping[r]; ok {
			builder.WriteRune(mapped)
		} else {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func safeSeq(value any) []SafeString {
	reflected := reflect.ValueOf(value)
	if reflected.Kind() != reflect.Array && reflected.Kind() != reflect.Slice {
		return nil
	}
	values := make([]SafeString, reflected.Len())
	for i := 0; i < reflected.Len(); i++ {
		values[i] = SafeString(fmt.Sprint(reflected.Index(i).Interface()))
	}
	return values
}

func sliceValue(value any, spec string) any {
	start, end := parseSliceSpec(spec, lengthOf(value))
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.String:
		runes := []rune(reflected.String())
		return string(runes[start:end])
	case reflect.Array, reflect.Slice:
		return reflected.Slice(start, end).Interface()
	default:
		return value
	}
}

func parseSliceSpec(spec string, length int) (int, int) {
	parts := strings.SplitN(spec, ":", 2)
	start := 0
	end := length
	if parts[0] != "" {
		start, _ = strconv.Atoi(parts[0])
	}
	if len(parts) == 2 && parts[1] != "" {
		end, _ = strconv.Atoi(parts[1])
	}
	if start < 0 {
		start = length + start
	}
	if end < 0 {
		end = length + end
	}
	if start < 0 {
		start = 0
	}
	if end > length {
		end = length
	}
	if start > end {
		start = end
	}
	return start, end
}

func slugify(value any) string {
	var builder strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(fmt.Sprint(value)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash && builder.Len() > 0 {
			builder.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(builder.String(), "-")
}

func stringFormat(value any, format string) string {
	if !strings.HasPrefix(format, "%") {
		format = "%" + format
	}
	return fmt.Sprintf(format, value)
}

var stripTagsRegexp = regexp.MustCompile(`<[^>]*>`)

func stripTags(value any) string {
	return stripTagsRegexp.ReplaceAllString(fmt.Sprint(value), "")
}

func formatTimeOnly(value any, layout string) string {
	timeValue, ok := toTime(value)
	if !ok {
		return ""
	}
	return timeValue.Format(layout)
}

func truncateChars(value any, max int) string {
	return truncateString(fmt.Sprint(value), max)
}

func truncateCharsHTML(value any, max int) SafeString {
	return SafeString(truncateString(stripTags(value), max))
}

func truncateString(value string, max int) string {
	runes := []rune(value)
	if max <= 0 {
		return ""
	}
	if len(runes) <= max {
		return value
	}
	if max <= 3 {
		return strings.Repeat(".", max)
	}
	return string(runes[:max-3]) + "..."
}

func truncateWords(value any, max int) string {
	return truncateWordsFromText(fmt.Sprint(value), max)
}

func truncateWordsHTML(value any, max int) SafeString {
	return SafeString(truncateWordsFromText(stripTags(value), max))
}

func truncateWordsFromText(text string, max int) string {
	words := strings.Fields(text)
	if max <= 0 {
		return ""
	}
	if len(words) <= max {
		return strings.Join(words, " ")
	}
	return strings.Join(words[:max], " ") + " ..."
}

func unorderedList(value any) SafeString {
	reflected := reflect.ValueOf(value)
	if reflected.Kind() != reflect.Array && reflected.Kind() != reflect.Slice {
		return ""
	}
	var builder strings.Builder
	builder.WriteString("<ul>")
	for i := 0; i < reflected.Len(); i++ {
		builder.WriteString("<li>")
		builder.WriteString(html.EscapeString(fmt.Sprint(reflected.Index(i).Interface())))
		builder.WriteString("</li>")
	}
	builder.WriteString("</ul>")
	return SafeString(builder.String())
}

var urlizeRegexp = regexp.MustCompile(`https?://[^\s<]+`)

func urlize(value any) SafeString {
	return urlizeWithLimit(fmt.Sprint(value), 0)
}

func urlizeTrunc(value any, limit int) SafeString {
	return urlizeWithLimit(fmt.Sprint(value), limit)
}

func urlEncode(value any) SafeString {
	return SafeString(url.QueryEscape(fmt.Sprint(value)))
}

func urlizeWithLimit(text string, limit int) SafeString {
	escaped := html.EscapeString(text)
	result := urlizeRegexp.ReplaceAllStringFunc(escaped, func(match string) string {
		label := match
		if limit > 0 {
			label = truncateString(match, limit)
		}
		return `<a href="` + match + `">` + label + `</a>`
	})
	return SafeString(result)
}

func wordCount(value any) int {
	return len(strings.Fields(fmt.Sprint(value)))
}

func wordWrap(value any, width int) string {
	words := strings.Fields(fmt.Sprint(value))
	if width <= 0 || len(words) == 0 {
		return strings.Join(words, " ")
	}
	lines := []string{words[0]}
	for _, word := range words[1:] {
		last := len(lines) - 1
		if len(lines[last])+1+len(word) > width {
			lines = append(lines, word)
			continue
		}
		lines[last] += " " + word
	}
	return strings.Join(lines, "\n")
}

func yesNo(value any, choices string) string {
	parts := strings.Split(choices, ",")
	for len(parts) < 3 {
		parts = append(parts, "")
	}
	if value == nil {
		return parts[2]
	}
	if boolValue, ok := value.(bool); ok {
		if boolValue {
			return parts[0]
		}
		return parts[1]
	}
	if helperEmpty(value) {
		return parts[1]
	}
	return parts[0]
}

func fieldOrMapValue(value any, key string) any {
	reflected := reflect.ValueOf(value)
	if reflected.Kind() == reflect.Map {
		mapValue := reflected.MapIndex(reflect.ValueOf(key))
		if mapValue.IsValid() {
			return mapValue.Interface()
		}
	}
	for reflected.Kind() == reflect.Pointer {
		if reflected.IsNil() {
			return ""
		}
		reflected = reflected.Elem()
	}
	if reflected.Kind() == reflect.Struct {
		field := reflected.FieldByName(capFirst(key))
		if field.IsValid() && field.CanInterface() {
			return field.Interface()
		}
	}
	return ""
}

func toInt64(value any) (int64, bool) {
	switch typed := value.(type) {
	case int:
		return int64(typed), true
	case int8:
		return int64(typed), true
	case int16:
		return int64(typed), true
	case int32:
		return int64(typed), true
	case int64:
		return typed, true
	case uint:
		return int64(typed), true
	case uint8:
		return int64(typed), true
	case uint16:
		return int64(typed), true
	case uint32:
		return int64(typed), true
	case uint64:
		return int64(typed), true
	case string:
		parsed, err := strconv.ParseInt(typed, 10, 64)
		return parsed, err == nil
	default:
		return 0, false
	}
}

func toInt64Default(value any) int64 {
	integer, _ := toInt64(value)
	return integer
}

func absInt(value int64) int64 {
	if value < 0 {
		return -value
	}
	return value
}

func toFloat(value any) (float64, bool) {
	switch typed := value.(type) {
	case float32:
		return float64(typed), true
	case float64:
		return typed, true
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		integer, ok := toInt64(value)
		return float64(integer), ok
	case string:
		parsed, err := strconv.ParseFloat(typed, 64)
		return parsed, err == nil
	default:
		return 0, false
	}
}

func toTime(value any) (time.Time, bool) {
	switch typed := value.(type) {
	case time.Time:
		return typed, true
	case *time.Time:
		if typed == nil {
			return time.Time{}, false
		}
		return *typed, true
	default:
		return time.Time{}, false
	}
}

func mustTime(value any) time.Time {
	parsed, _ := toTime(value)
	return parsed
}

func durationPhrase(duration time.Duration) string {
	if duration < 0 {
		duration = -duration
	}
	units := []struct {
		name string
		size time.Duration
	}{
		{"day", 24 * time.Hour},
		{"hour", time.Hour},
		{"minute", time.Minute},
		{"second", time.Second},
	}
	for _, unit := range units {
		if duration >= unit.size {
			count := int(math.Round(float64(duration) / float64(unit.size)))
			if count == 1 {
				return "1 " + unit.name
			}
			return fmt.Sprintf("%d %ss", count, unit.name)
		}
	}
	return "0 seconds"
}
