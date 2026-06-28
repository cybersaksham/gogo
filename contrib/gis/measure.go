package gis

// DistanceMeasure stores distance in meters and exposes common unit conversions.
type DistanceMeasure struct {
	Meters float64
}

// D creates a distance measurement from meters.
func D(meters float64) DistanceMeasure {
	return DistanceMeasure{Meters: meters}
}

func (d DistanceMeasure) Kilometers() float64 {
	return d.Meters / 1000
}

func (d DistanceMeasure) Miles() float64 {
	return d.Meters / 1609.344
}

func (d DistanceMeasure) Feet() float64 {
	return d.Meters / 0.3048
}

// AreaMeasure stores area in square meters and exposes common unit conversions.
type AreaMeasure struct {
	SquareMeters float64
}

// A creates an area measurement from square meters.
func A(squareMeters float64) AreaMeasure {
	return AreaMeasure{SquareMeters: squareMeters}
}

func (a AreaMeasure) SquareKilometers() float64 {
	return a.SquareMeters / 1000000
}

func (a AreaMeasure) Hectares() float64 {
	return a.SquareMeters / 10000
}

func (a AreaMeasure) Acres() float64 {
	return a.SquareMeters / 4046.8564224
}
