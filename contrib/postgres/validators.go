package postgres

import (
	"fmt"
	"reflect"
)

func ValidateArrayLength(value any, min int, max int) error {
	length := reflect.ValueOf(value).Len()
	if min > 0 && length < min {
		return fmt.Errorf("array length %d below minimum %d", length, min)
	}
	if max > 0 && length > max {
		return fmt.Errorf("array length %d above maximum %d", length, max)
	}
	return nil
}

func ValidateRangeBounds[T ~int | ~int64 | ~float64](lower T, upper T, lowerInclusive bool, upperInclusive bool) error {
	if lower > upper || lower == upper && (!lowerInclusive || !upperInclusive) {
		return fmt.Errorf("invalid range bounds")
	}
	return nil
}

func ValidateJSONStructure(value map[string]any, requiredKeys []string) error {
	for _, key := range requiredKeys {
		if _, ok := value[key]; !ok {
			return fmt.Errorf("missing JSON key %s", key)
		}
	}
	return nil
}
