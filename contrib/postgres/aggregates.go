package postgres

import "fmt"

type Aggregate struct {
	sql string
}

func (a Aggregate) SQL() string { return a.sql }

func ArrayAgg(column string) Aggregate {
	return Aggregate{sql: fmt.Sprintf("array_agg(%s)", quote(column))}
}

func JSONObjectAgg(key string, value string) Aggregate {
	return Aggregate{sql: fmt.Sprintf("jsonb_object_agg(%s, %s)", quote(key), quote(value))}
}

func StringAgg(column string, delimiter string) Aggregate {
	return Aggregate{sql: fmt.Sprintf("string_agg(%s, '%s')", quote(column), escapeSQL(delimiter))}
}

func BoolAnd(column string) Aggregate {
	return Aggregate{sql: fmt.Sprintf("bool_and(%s)", quote(column))}
}

func BoolOr(column string) Aggregate {
	return Aggregate{sql: fmt.Sprintf("bool_or(%s)", quote(column))}
}

func BitAnd(column string) Aggregate {
	return Aggregate{sql: fmt.Sprintf("bit_and(%s)", quote(column))}
}

func BitOr(column string) Aggregate {
	return Aggregate{sql: fmt.Sprintf("bit_or(%s)", quote(column))}
}

func BitXor(column string) Aggregate {
	return Aggregate{sql: fmt.Sprintf("bit_xor(%s)", quote(column))}
}
