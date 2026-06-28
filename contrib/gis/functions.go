package gis

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrUnsupportedDialect = errors.New("unsupported dialect")

type Expression struct {
	sql string
}

func (e Expression) SQL(dialect string) (string, error) {
	if !isPostgres(dialect) {
		return "", ErrUnsupportedDialect
	}
	return e.sql, nil
}

func Area(expr string) Expression {
	return unaryFunction("ST_Area", expr)
}

func AsGeoJSON(expr string) Expression {
	return unaryFunction("ST_AsGeoJSON", expr)
}

func AsKML(expr string) Expression {
	return unaryFunction("ST_AsKML", expr)
}

func AsSVG(expr string) Expression {
	return unaryFunction("ST_AsSVG", expr)
}

func Centroid(expr string) Expression {
	return unaryFunction("ST_Centroid", expr)
}

func Difference(left string, right string) Expression {
	return binaryFunction("ST_Difference", left, right)
}

func Distance(left string, right string) Expression {
	return binaryFunction("ST_Distance", left, right)
}

func Envelope(expr string) Expression {
	return unaryFunction("ST_Envelope", expr)
}

func Intersection(left string, right string) Expression {
	return binaryFunction("ST_Intersection", left, right)
}

func Length(expr string) Expression {
	return unaryFunction("ST_Length", expr)
}

func Perimeter(expr string) Expression {
	return unaryFunction("ST_Perimeter", expr)
}

func PointOnSurface(expr string) Expression {
	return unaryFunction("ST_PointOnSurface", expr)
}

func Scale(expr string, x float64, y float64) Expression {
	return Expression{sql: fmt.Sprintf("ST_Scale(%s, %s, %s)", spatialArg(expr), formatFloat(x), formatFloat(y))}
}

func SnapToGrid(expr string, size float64) Expression {
	return Expression{sql: fmt.Sprintf("ST_SnapToGrid(%s, %s)", spatialArg(expr), formatFloat(size))}
}

func SymDifference(left string, right string) Expression {
	return binaryFunction("ST_SymDifference", left, right)
}

func Transform(expr string, srid int) Expression {
	return Expression{sql: fmt.Sprintf("ST_Transform(%s, %d)", spatialArg(expr), srid)}
}

func Translate(expr string, x float64, y float64) Expression {
	return Expression{sql: fmt.Sprintf("ST_Translate(%s, %s, %s)", spatialArg(expr), formatFloat(x), formatFloat(y))}
}

func Union(left string, right string) Expression {
	return binaryFunction("ST_Union", left, right)
}

func unaryFunction(name string, expr string) Expression {
	return Expression{sql: fmt.Sprintf("%s(%s)", name, spatialArg(expr))}
}

func binaryFunction(name string, left string, right string) Expression {
	return Expression{sql: fmt.Sprintf("%s(%s, %s)", name, spatialArg(left), spatialArg(right))}
}

func spatialArg(value string) string {
	if value == "" {
		return `""`
	}
	if strings.HasPrefix(value, "$") || strings.HasPrefix(value, "?") || strings.HasPrefix(value, "'") || strings.HasPrefix(value, `"`) {
		return value
	}
	if strings.Contains(value, "(") || strings.Contains(value, ".") || strings.Contains(value, "::") {
		return value
	}
	return quoteIdentifier(value)
}

func quoteIdentifier(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}

func quoteLiteral(value string) string {
	return `'` + strings.ReplaceAll(value, `'`, `''`) + `'`
}

func isPostgres(dialect string) bool {
	return dialect == "postgres" || dialect == "postgresql"
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
