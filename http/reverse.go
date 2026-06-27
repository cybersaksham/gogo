package http

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// Reverse resolves a named route into a URL path.
func (r *Router) Reverse(name string, args map[string]any) (string, error) {
	for _, route := range r.routes {
		if route.Name != name {
			continue
		}
		return reversePattern(route.Pattern, args)
	}

	return "", fmt.Errorf("%w: route %q not found", ErrReverse, name)
}

func reversePattern(pattern string, args map[string]any) (string, error) {
	required := make(map[string]struct{})
	var builder strings.Builder

	last := 0
	matches := tokenPattern.FindAllStringSubmatchIndex(pattern, -1)
	for _, match := range matches {
		builder.WriteString(pattern[last:match[0]])

		converterName := pattern[match[2]:match[3]]
		paramName := pattern[match[4]:match[5]]
		required[paramName] = struct{}{}

		value, ok := args[paramName]
		if !ok {
			return "", fmt.Errorf("%w: missing argument %q", ErrReverse, paramName)
		}
		stringValue := fmt.Sprint(value)
		if err := validateReverseValue(converterName, paramName, stringValue); err != nil {
			return "", err
		}

		builder.WriteString(escapeReverseValue(converterName, stringValue))
		last = match[1]
	}
	builder.WriteString(pattern[last:])

	for name := range args {
		if _, ok := required[name]; !ok {
			return "", fmt.Errorf("%w: extra argument %q", ErrReverse, name)
		}
	}

	return builder.String(), nil
}

func validateReverseValue(converterName, paramName, value string) error {
	regex, ok := converters[converterName]
	if !ok {
		return fmt.Errorf("%w: unknown converter %q", ErrReverse, converterName)
	}

	matcher, err := regexp.Compile("^(?:" + regex + ")$")
	if err != nil {
		return fmt.Errorf("%w: invalid converter %q: %v", ErrReverse, converterName, err)
	}
	if !matcher.MatchString(value) {
		return fmt.Errorf("%w: argument %q does not match %s converter", ErrReverse, paramName, converterName)
	}
	return nil
}

func escapeReverseValue(converterName, value string) string {
	if converterName != "path" {
		return url.PathEscape(value)
	}

	segments := strings.Split(value, "/")
	for i, segment := range segments {
		segments[i] = url.PathEscape(segment)
	}
	return strings.Join(segments, "/")
}
