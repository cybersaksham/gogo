package constraints

import (
	"crypto/sha1"
	"encoding/hex"
	"strings"
	"unicode"
)

// MaxNameLength keeps generated names portable across common SQL backends.
const MaxNameLength = 63

func deterministicName(table, suffix string, parts ...string) string {
	segments := append([]string{table}, parts...)
	readable := sanitizeName(strings.Join(segments, "_"))
	if readable == "" {
		readable = suffix
	}

	hashInput := table + "|" + suffix + "|" + strings.Join(parts, "|")
	sum := sha1.Sum([]byte(hashInput))
	hash := hex.EncodeToString(sum[:])[:8]

	trailer := "_" + suffix + "_" + hash
	limit := MaxNameLength - len(trailer)
	if limit < 1 {
		limit = 1
	}
	if len(readable) > limit {
		readable = strings.Trim(readable[:limit], "_")
	}
	return readable + trailer
}

func sanitizeName(value string) string {
	var builder strings.Builder
	lastUnderscore := false
	for _, char := range strings.ToLower(value) {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			builder.WriteRune(char)
			lastUnderscore = false
			continue
		}
		if !lastUnderscore {
			builder.WriteByte('_')
			lastUnderscore = true
		}
	}
	return strings.Trim(builder.String(), "_")
}
