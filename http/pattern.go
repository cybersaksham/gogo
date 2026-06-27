package http

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var (
	tokenPattern = regexp.MustCompile(`<([A-Za-z_][A-Za-z0-9_]*):([A-Za-z_][A-Za-z0-9_]*)>`)
	converters   = map[string]string{
		"str":  `[^/]+`,
		"int":  `[0-9]+`,
		"slug": `[-A-Za-z0-9_]+`,
		"path": `.+`,
		"uuid": `[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`,
	}
)

// Pattern matches URL paths and extracts named parameters.
type Pattern struct {
	raw        string
	regex      *regexp.Regexp
	paramNames []string
}

// RegisterConverter registers a path converter by name.
func RegisterConverter(name, regex string) {
	converters[name] = regex
}

// CompilePattern compiles a Django-style route pattern.
func CompilePattern(pattern string) (Pattern, error) {
	seen := make(map[string]struct{})
	names := make([]string, 0)
	var builder strings.Builder
	builder.WriteString("^")

	last := 0
	matches := tokenPattern.FindAllStringSubmatchIndex(pattern, -1)
	for _, match := range matches {
		builder.WriteString(regexp.QuoteMeta(pattern[last:match[0]]))
		converterName := pattern[match[2]:match[3]]
		paramName := pattern[match[4]:match[5]]

		regex, ok := converters[converterName]
		if !ok {
			return Pattern{}, fmt.Errorf("%w: unknown converter %q", ErrInvalidPattern, converterName)
		}
		if _, exists := seen[paramName]; exists {
			return Pattern{}, fmt.Errorf("%w: duplicate parameter %q", ErrInvalidPattern, paramName)
		}

		seen[paramName] = struct{}{}
		names = append(names, paramName)
		builder.WriteString("(?P<")
		builder.WriteString(paramName)
		builder.WriteString(">")
		builder.WriteString(regex)
		builder.WriteString(")")
		last = match[1]
	}
	builder.WriteString(regexp.QuoteMeta(pattern[last:]))
	builder.WriteString("$")

	return CompileRegexPatternWithRaw(pattern, builder.String(), names)
}

// CompileRegexPattern compiles a regex-backed pattern with named groups.
func CompileRegexPattern(regex string) (Pattern, error) {
	return CompileRegexPatternWithRaw(regex, regex, nil)
}

// CompileRegexPatternWithRaw compiles a regex-backed pattern with a raw display value.
func CompileRegexPatternWithRaw(raw, regex string, names []string) (Pattern, error) {
	compiled, err := regexp.Compile(regex)
	if err != nil {
		return Pattern{}, fmt.Errorf("%w: %v", ErrInvalidPattern, err)
	}

	paramNames := names
	if paramNames == nil {
		paramNames = make([]string, 0)
		for _, name := range compiled.SubexpNames()[1:] {
			if name != "" {
				paramNames = append(paramNames, name)
			}
		}
	}

	return Pattern{
		raw:        raw,
		regex:      compiled,
		paramNames: paramNames,
	}, nil
}

// Match matches a path and returns decoded named parameters.
func (p Pattern) Match(path string) (map[string]string, bool) {
	matches := p.regex.FindStringSubmatch(path)
	if matches == nil {
		return nil, false
	}

	params := make(map[string]string, len(p.paramNames))
	for _, name := range p.paramNames {
		index := p.regex.SubexpIndex(name)
		if index <= 0 || index >= len(matches) {
			continue
		}
		value, err := url.PathUnescape(matches[index])
		if err != nil {
			return nil, false
		}
		params[name] = value
	}

	return params, true
}
