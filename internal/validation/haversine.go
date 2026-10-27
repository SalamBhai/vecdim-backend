package validation

import (
	"math"
)

const (
	EarthRadiusKm = 6371.0
	LgaMaxKm      = 25.0
)

// LGAReference holds coordinates for an LGA centroid.
type LGAReference struct {
	State  string
	LGA    string
	RefLat float64
	RefLon float64
}

// LGAReferences map: "State|LGA" -> LGAReference
var LGAReferences = map[string]LGAReference{
	"Lagos|Ikorodu":     {State: "Lagos", LGA: "Ikorodu", RefLat: 6.62, RefLon: 3.51},
	"Lagos|Alimosho":    {State: "Lagos", LGA: "Alimosho", RefLat: 6.60, RefLon: 3.25},
	"Lagos|Epe":         {State: "Lagos", LGA: "Epe", RefLat: 6.58, RefLon: 3.98},
	"Lagos|Badagry":     {State: "Lagos", LGA: "Badagry", RefLat: 6.43, RefLon: 2.88},
	"Lagos|Eti-Osa":     {State: "Lagos", LGA: "Eti-Osa", RefLat: 6.45, RefLon: 3.60},
	"Kano|Dala":         {State: "Kano", LGA: "Dala", RefLat: 12.00, RefLon: 8.52},
	"Kano|Nasarawa":     {State: "Kano", LGA: "Nasarawa", RefLat: 12.00, RefLon: 8.55},
	"Kano|Gwale":        {State: "Kano", LGA: "Gwale", RefLat: 11.98, RefLon: 8.49},
	"Kano|Tarauni":      {State: "Kano", LGA: "Tarauni", RefLat: 11.97, RefLon: 8.54},
	"Kano|Ungogo":       {State: "Kano", LGA: "Ungogo", RefLat: 12.05, RefLon: 8.48},
	"Osun|Osogbo":       {State: "Osun", LGA: "Osogbo", RefLat: 7.77, RefLon: 4.56},
	"Osun|Ife Central":  {State: "Osun", LGA: "Ife Central", RefLat: 7.48, RefLon: 4.56},
	"Osun|Ilesa East":   {State: "Osun", LGA: "Ilesa East", RefLat: 7.62, RefLon: 4.73},
	"Osun|Iwo":          {State: "Osun", LGA: "Iwo", RefLat: 7.63, RefLon: 4.18},
	"Osun|Ede North":    {State: "Osun", LGA: "Ede North", RefLat: 7.74, RefLon: 4.43},
}

// StateBoundingBox defines the pilot state bounding boxes in WGS84 decimal degrees.
type StateBoundingBox struct {
	LatMin float64
	LatMax float64
	LonMin float64
	LonMax float64
}

var StateBoxes = map[string]StateBoundingBox{
	"Lagos": {LatMin: 6.3, LatMax: 6.8, LonMin: 2.7, LonMax: 4.4},
	"Kano":  {LatMin: 11.5, LatMax: 12.6, LonMin: 8.0, LonMax: 9.2},
	"Osun":  {LatMin: 7.0, LatMax: 8.2, LonMin: 4.0, LonMax: 5.0},
}

// Nigeria Bounding Box: Lat [4, 14], Lon [2.5, 15]
const (
	NigeriaLatMin = 4.0
	NigeriaLatMax = 14.0
	NigeriaLonMin = 2.5
	NigeriaLonMax = 15.0
)

// HaversineDistanceKm calculates great-circle distance between two points in kilometres.
// Matches R/helpers.R: haversine_km
func HaversineDistanceKm(lat1, lon1, lat2, lon2 float64) float64 {
	rad := math.Pi / 180.0
	dLat := (lat2 - lat1) * rad
	dLon := (lon2 - lon1) * rad

	lat1Rad := lat1 * rad
	lat2Rad := lat2 * rad

	a := math.Sin(dLat/2.0)*math.Sin(dLat/2.0) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dLon/2.0)*math.Sin(dLon/2.0)

	sqrtA := math.Sqrt(a)
	if sqrtA > 1.0 {
		sqrtA = 1.0
	}
	return 2.0 * EarthRadiusKm * math.Asin(sqrtA)
}
