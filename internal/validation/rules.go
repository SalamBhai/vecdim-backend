package validation

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// Default reference date for parity testing with candidate pack.
const ParityReferenceDate = "2026-10-08"

// ValidationSeverity defines severity levels: BLOCK or WARN
type ValidationSeverity string

const (
	SeverityBlock ValidationSeverity = "BLOCK"
	SeverityWarn  ValidationSeverity = "WARN"
)

// ValidationError represents a single validation failure.
type ValidationError struct {
	Rule     string             `json:"rule"`
	Field    string             `json:"field"`
	Severity ValidationSeverity `json:"severity"`
	Message  string             `json:"message"`
}

// RecordInput holds the fields needed to validate an entomological survey record.
type RecordInput struct {
	RecordID           string   `json:"record_id"`
	State              string   `json:"state"`
	LGA                string   `json:"lga"`
	Ward               string   `json:"ward"`
	Latitude           float64  `json:"latitude"`
	Longitude          float64  `json:"longitude"`
	CollectionDate     string   `json:"collection_date"` // YYYY-MM-DD
	Species            *string  `json:"species"`
	Insecticide        *string  `json:"insecticide"`
	NExposed           int      `json:"n_exposed"`
	NDead              int      `json:"n_dead"`
	LarvalDensity      *int     `json:"larval_density,omitempty"`
	RainfallMm         *float64 `json:"rainfall_mm,omitempty"`
	TempC              *float64 `json:"temp_c,omitempty"`
	NDVI               *float64 `json:"ndvi,omitempty"`
	ElevationM         *float64 `json:"elevation_m,omitempty"`
	DistWaterKm        *float64 `json:"dist_water_km,omitempty"`
	Urban              int      `json:"urban"`
	ITNUsePct          *float64 `json:"itn_use_pct,omitempty"`
	IRSCovered         int      `json:"irs_covered"`
	SlidePositivityPct *float64 `json:"slide_positivity_pct,omitempty"`
	GPSAccuracyM       *float64 `json:"gps_accuracy_m,omitempty"`
}

// ValidationFlagsAll stores boolean results for the 11 reference rules in outputs/validation_flags_all.csv.
type ValidationFlagsAll struct {
	RecordID            string `json:"record_id"`
	GPSOutsideNigeria   bool   `json:"GPS_outside_nigeria"`
	GPSZeroZero         bool   `json:"GPS_zero_zero"`
	GPSStateMismatch    bool   `json:"GPS_state_mismatch"`
	GPSFarFromLGA       bool   `json:"GPS_far_from_lga"`
	NDeadGtNExposed     bool   `json:"n_dead_gt_n_exposed"`
	PositivityOutRange  bool   `json:"positivity_out_range"`
	FutureDate          bool   `json:"future_date"`
	TempImplausible     bool   `json:"temp_implausible"`
	RainfallImplausible bool   `json:"rainfall_implausible"`
	Duplicate           bool   `json:"duplicate"`
	MissingCore         bool   `json:"missing_core"`
	AnyFlag             bool   `json:"any_flag"`
}

// CalculateWHOMortalityClass computes WHO mortality percentage rounded to 1 decimal place first,
// and returns (mortalityPct, whoClass).
func CalculateWHOMortalityClass(nExposed, nDead int) (float64, string) {
	if nExposed <= 0 {
		return 0.0, "Resistant"
	}
	rawPct := (float64(nDead) / float64(nExposed)) * 100.0
	roundedPct := math.Round(rawPct*10.0) / 10.0

	var class string
	switch {
	case roundedPct >= 98.0:
		class = "Susceptible"
	case roundedPct >= 90.0:
		class = "Possible resistance"
	default:
		class = "Resistant"
	}
	return roundedPct, class
}

// ValidateSingleRecord evaluates all rules for a single record against an optional reference date.
// If isServerOnlyEnrichmentPending is true, missing ndvi/rainfall is treated as pending server enrichment.
func ValidateSingleRecord(rec RecordInput, refDate string, isServerOnlyEnrichmentPending bool) ([]ValidationError, []ValidationError) {
	var blocks []ValidationError
	var warns []ValidationError

	if refDate == "" {
		refDate = time.Now().Format("2006-01-02")
	}

	// 1. GPS_outside_nigeria (Block)
	if !(rec.Latitude >= NigeriaLatMin && rec.Latitude <= NigeriaLatMax &&
		rec.Longitude >= NigeriaLonMin && rec.Longitude <= NigeriaLonMax) {
		blocks = append(blocks, ValidationError{
			Rule:     "GPS_outside_nigeria",
			Field:    "latitude",
			Severity: SeverityBlock,
			Message:  fmt.Sprintf("GPS coordinates (%.5f, %.5f) outside Nigeria bounding box.", rec.Latitude, rec.Longitude),
		})
	}

	// 2. GPS_zero_zero (Block)
	if rec.Latitude == 0.0 && rec.Longitude == 0.0 {
		blocks = append(blocks, ValidationError{
			Rule:     "GPS_zero_zero",
			Field:    "latitude",
			Severity: SeverityBlock,
			Message:  "GPS coordinates cannot be (0, 0).",
		})
	}

	// 3. GPS_state_mismatch (Block)
	box, stateKnown := StateBoxes[rec.State]
	if !stateKnown || !(rec.Latitude >= box.LatMin && rec.Latitude <= box.LatMax &&
		rec.Longitude >= box.LonMin && rec.Longitude <= box.LonMax) {
		blocks = append(blocks, ValidationError{
			Rule:     "GPS_state_mismatch",
			Field:    "state",
			Severity: SeverityBlock,
			Message:  fmt.Sprintf("GPS coordinates (%.5f, %.5f) outside bounding box for state %s.", rec.Latitude, rec.Longitude, rec.State),
		})
	}

	// 4. GPS_far_from_lga (Warn)
	lgaKey := fmt.Sprintf("%s|%s", rec.State, rec.LGA)
	lgaRef, lgaKnown := LGAReferences[lgaKey]
	if !lgaKnown {
		warns = append(warns, ValidationError{
			Rule:     "GPS_far_from_lga",
			Field:    "lga",
			Severity: SeverityWarn,
			Message:  fmt.Sprintf("Unknown LGA reference for state: %s, LGA: %s.", rec.State, rec.LGA),
		})
	} else {
		distKm := HaversineDistanceKm(rec.Latitude, rec.Longitude, lgaRef.RefLat, lgaRef.RefLon)
		if distKm > LgaMaxKm {
			warns = append(warns, ValidationError{
				Rule:     "GPS_far_from_lga",
				Field:    "lga",
				Severity: SeverityWarn,
				Message:  fmt.Sprintf("GPS location is %.1f km from %s reference point (limit %d km).", distKm, rec.LGA, int(LgaMaxKm)),
			})
		}
	}

	// 5. GPS accuracy rules (Device & Server)
	if rec.GPSAccuracyM != nil {
		acc := *rec.GPSAccuracyM
		if acc > 10.0 && acc <= 50.0 {
			warns = append(warns, ValidationError{
				Rule:     "GPS_accuracy_poor",
				Field:    "gps_accuracy_m",
				Severity: SeverityWarn,
				Message:  fmt.Sprintf("GPS accuracy (%.1fm) is between 10m and 50m.", acc),
			})
		} else if acc > 50.0 {
			blocks = append(blocks, ValidationError{
				Rule:     "GPS_accuracy_unusable",
				Field:    "gps_accuracy_m",
				Severity: SeverityBlock,
				Message:  fmt.Sprintf("GPS accuracy (%.1fm) exceeds usable limit of 50m.", acc),
			})
		}
	}

	// 6. n_dead_gt_n_exposed (Block)
	if rec.NDead > rec.NExposed {
		blocks = append(blocks, ValidationError{
			Rule:     "n_dead_gt_n_exposed",
			Field:    "n_dead",
			Severity: SeverityBlock,
			Message:  fmt.Sprintf("n_dead (%d) is more than n_exposed (%d). Correct the number of dead mosquitoes.", rec.NDead, rec.NExposed),
		})
	}

	// 7. positivity_out_range (Block on server)
	if rec.SlidePositivityPct != nil {
		pos := *rec.SlidePositivityPct
		if pos < 0.0 || pos > 100.0 {
			blocks = append(blocks, ValidationError{
				Rule:     "positivity_out_range",
				Field:    "slide_positivity_pct",
				Severity: SeverityBlock,
				Message:  fmt.Sprintf("slide_positivity_pct (%.1f) must be between 0 and 100.", pos),
			})
		}
	}

	// 8. future_date (Block)
	if rec.CollectionDate > refDate {
		blocks = append(blocks, ValidationError{
			Rule:     "future_date",
			Field:    "collection_date",
			Severity: SeverityBlock,
			Message:  fmt.Sprintf("collection_date (%s) is later than reference date (%s).", rec.CollectionDate, refDate),
		})
	}

	// 9. temp_implausible (Block on server)
	if rec.TempC != nil {
		temp := *rec.TempC
		if temp < 10.0 || temp > 50.0 {
			blocks = append(blocks, ValidationError{
				Rule:     "temp_implausible",
				Field:    "temp_c",
				Severity: SeverityBlock,
				Message:  fmt.Sprintf("temp_c (%.1f) is implausible (must be between 10°C and 50°C).", temp),
			})
		}
	}

	// 10. rainfall_implausible (Block on server)
	if rec.RainfallMm != nil {
		rain := *rec.RainfallMm
		if rain < 0.0 || rain > 1500.0 {
			blocks = append(blocks, ValidationError{
				Rule:     "rainfall_implausible",
				Field:    "rainfall_mm",
				Severity: SeverityBlock,
				Message:  fmt.Sprintf("rainfall_mm (%.1f) is implausible (must be between 0mm and 1500mm).", rain),
			})
		}
	}

	// 11. missing_core (Block)
	var speciesVal string
	if rec.Species != nil {
		speciesVal = strings.TrimSpace(*rec.Species)
	}

	hasMissingCore := false
	if !isServerOnlyEnrichmentPending {
		if rec.NDVI == nil || rec.RainfallMm == nil || rec.ITNUsePct == nil || speciesVal == "" {
			hasMissingCore = true
		}
	} else {
		// On device or mobile submission before environmental enrichment
		if rec.ITNUsePct == nil || speciesVal == "" {
			hasMissingCore = true
		}
	}

	if hasMissingCore {
		blocks = append(blocks, ValidationError{
			Rule:     "missing_core",
			Field:    "missing_core",
			Severity: SeverityBlock,
			Message:  "Core variables (species, ITN use, NDVI, or rainfall) are missing.",
		})
	}

	return blocks, warns
}

// ValidateBatchForParity processes a full slice of records and returns ValidationFlagsAll
// matching the exact R helpers.R validate_records() semantics.
func ValidateBatchForParity(records []RecordInput, refDate string) []ValidationFlagsAll {
	if refDate == "" {
		refDate = ParityReferenceDate
	}

	// Duplicate detection key logic in R:
	// key <- paste(d$ward, d$collection_date, d$species, d$insecticide, d$n_exposed, d$n_dead)
	// ord <- order(d$record_id)
	// is_dup[ord] <- duplicated(key[ord])

	type indexedRecord struct {
		index    int
		recordID string
		dupKey   string
	}

	indexed := make([]indexedRecord, len(records))
	for i, r := range records {
		var speciesStr, insecticideStr string
		if r.Species != nil {
			speciesStr = *r.Species
		}
		if r.Insecticide != nil {
			insecticideStr = *r.Insecticide
		}
		key := fmt.Sprintf("%s %s %s %s %d %d", r.Ward, r.CollectionDate, speciesStr, insecticideStr, r.NExposed, r.NDead)
		indexed[i] = indexedRecord{
			index:    i,
			recordID: r.RecordID,
			dupKey:   key,
		}
	}

	// Sort by recordID (matches R's order(d$record_id))
	sort.Slice(indexed, func(i, j int) bool {
		return indexed[i].recordID < indexed[j].recordID
	})

	isDupByIndex := make([]bool, len(records))
	seenKeys := make(map[string]bool)
	for _, item := range indexed {
		if seenKeys[item.dupKey] {
			isDupByIndex[item.index] = true
		} else {
			seenKeys[item.dupKey] = true
		}
	}

	results := make([]ValidationFlagsAll, len(records))

	for i, r := range records {
		flags := ValidationFlagsAll{
			RecordID: r.RecordID,
		}

		// 1. GPS_outside_nigeria
		flags.GPSOutsideNigeria = !(r.Latitude >= NigeriaLatMin && r.Latitude <= NigeriaLatMax &&
			r.Longitude >= NigeriaLonMin && r.Longitude <= NigeriaLonMax)

		// 2. GPS_zero_zero
		flags.GPSZeroZero = (r.Latitude == 0.0 && r.Longitude == 0.0)

		// 3. GPS_state_mismatch
		box, ok := StateBoxes[r.State]
		flags.GPSStateMismatch = !ok || !(r.Latitude >= box.LatMin && r.Latitude <= box.LatMax &&
			r.Longitude >= box.LonMin && r.Longitude <= box.LonMax)

		// 4. GPS_far_from_lga
		lgaKey := fmt.Sprintf("%s|%s", r.State, r.LGA)
		lgaRef, lgaOk := LGAReferences[lgaKey]
		if !lgaOk {
			flags.GPSFarFromLGA = true
		} else {
			dist := HaversineDistanceKm(r.Latitude, r.Longitude, lgaRef.RefLat, lgaRef.RefLon)
			flags.GPSFarFromLGA = (dist > LgaMaxKm)
		}

		// 5. n_dead_gt_n_exposed
		flags.NDeadGtNExposed = (r.NDead > r.NExposed)

		// 6. positivity_out_range
		if r.SlidePositivityPct != nil {
			flags.PositivityOutRange = (*r.SlidePositivityPct < 0.0 || *r.SlidePositivityPct > 100.0)
		}

		// 7. future_date
		flags.FutureDate = (r.CollectionDate > refDate)

		// 8. temp_implausible
		if r.TempC != nil {
			flags.TempImplausible = (*r.TempC < 10.0 || *r.TempC > 50.0)
		}

		// 9. rainfall_implausible
		if r.RainfallMm != nil {
			flags.RainfallImplausible = (*r.RainfallMm < 0.0 || *r.RainfallMm > 1500.0)
		}

		// 10. duplicate
		flags.Duplicate = isDupByIndex[i]

		// 11. missing_core
		var sp string
		if r.Species != nil {
			sp = *r.Species
		}
		flags.MissingCore = (r.NDVI == nil || r.RainfallMm == nil || r.ITNUsePct == nil || sp == "")

		// any_flag
		flags.AnyFlag = flags.GPSOutsideNigeria || flags.GPSZeroZero || flags.GPSStateMismatch ||
			flags.GPSFarFromLGA || flags.NDeadGtNExposed || flags.PositivityOutRange ||
			flags.FutureDate || flags.TempImplausible || flags.RainfallImplausible ||
			flags.Duplicate || flags.MissingCore

		results[i] = flags
	}

	return results
}
