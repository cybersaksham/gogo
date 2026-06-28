package gis

import "fmt"

type Lookup struct {
	sql string
}

func (l Lookup) SQL(dialect string) (string, error) {
	if !isPostgres(dialect) {
		return "", ErrUnsupportedDialect
	}
	return l.sql, nil
}

func Equals(left string, right string) Lookup {
	return spatialLookup("ST_Equals", left, right)
}

func Contains(left string, right string) Lookup {
	return spatialLookup("ST_Contains", left, right)
}

func CoveredBy(left string, right string) Lookup {
	return spatialLookup("ST_CoveredBy", left, right)
}

func Covers(left string, right string) Lookup {
	return spatialLookup("ST_Covers", left, right)
}

func Crosses(left string, right string) Lookup {
	return spatialLookup("ST_Crosses", left, right)
}

func Disjoint(left string, right string) Lookup {
	return spatialLookup("ST_Disjoint", left, right)
}

func Intersects(left string, right string) Lookup {
	return spatialLookup("ST_Intersects", left, right)
}

func Overlaps(left string, right string) Lookup {
	return spatialLookup("ST_Overlaps", left, right)
}

func Relate(left string, right string, pattern string) Lookup {
	return Lookup{sql: fmt.Sprintf("ST_Relate(%s, %s, %s)", spatialArg(left), spatialArg(right), quoteLiteral(pattern))}
}

func Touches(left string, right string) Lookup {
	return spatialLookup("ST_Touches", left, right)
}

func Within(left string, right string) Lookup {
	return spatialLookup("ST_Within", left, right)
}

func DistanceLT(left string, right string, distance DistanceMeasure) Lookup {
	return Lookup{sql: fmt.Sprintf("ST_Distance(%s, %s) < %s", spatialArg(left), spatialArg(right), formatFloat(distance.Meters))}
}

func DistanceLTE(left string, right string, distance DistanceMeasure) Lookup {
	return Lookup{sql: fmt.Sprintf("ST_DWithin(%s, %s, %s)", spatialArg(left), spatialArg(right), formatFloat(distance.Meters))}
}

func DistanceGT(left string, right string, distance DistanceMeasure) Lookup {
	return Lookup{sql: fmt.Sprintf("ST_Distance(%s, %s) > %s", spatialArg(left), spatialArg(right), formatFloat(distance.Meters))}
}

func DistanceGTE(left string, right string, distance DistanceMeasure) Lookup {
	return Lookup{sql: fmt.Sprintf("ST_Distance(%s, %s) >= %s", spatialArg(left), spatialArg(right), formatFloat(distance.Meters))}
}

func spatialLookup(name string, left string, right string) Lookup {
	return Lookup{sql: fmt.Sprintf("%s(%s, %s)", name, spatialArg(left), spatialArg(right))}
}
