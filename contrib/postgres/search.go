package postgres

import (
	"fmt"
	"strings"
)

type Expression struct {
	sql string
}

func (e Expression) SQL() string { return e.sql }

func SearchVector(columns ...string) Expression {
	parts := make([]string, len(columns))
	for i, column := range columns {
		parts[i] = fmt.Sprintf(`coalesce(%s, '')`, quote(column))
	}
	return Expression{sql: `to_tsvector('simple', ` + strings.Join(parts, " || ' ' || ") + `)`}
}

func SearchQuery(value string) Expression {
	return Expression{sql: fmt.Sprintf(`plainto_tsquery('simple', '%s')`, escapeSQL(value))}
}

func SearchRank(vector Expression, query Expression) Expression {
	return Expression{sql: fmt.Sprintf("ts_rank(%s, %s)", vector.SQL(), query.SQL())}
}

func SearchHeadline(column string, query Expression) Expression {
	return Expression{sql: fmt.Sprintf("ts_headline(%s, %s)", quote(column), query.SQL())}
}

func Similarity(column string, value string) Expression {
	return Expression{sql: fmt.Sprintf("%s %% '%s'", quote(column), escapeSQL(value))}
}

func Distance(column string, value string) Expression {
	return Expression{sql: fmt.Sprintf("%s <-> '%s'", quote(column), escapeSQL(value))}
}

func WordSimilarity(column string, value string) Expression {
	return Expression{sql: fmt.Sprintf("word_similarity(%s, '%s')", quote(column), escapeSQL(value))}
}

func quote(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}

func escapeSQL(value string) string {
	return strings.ReplaceAll(value, `'`, `''`)
}
