package validation_test

import (
	"encoding/csv"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/salambhai/vecdim-backend/internal/validation"
)

func BenchmarkValidationAllRecords(b *testing.B) {
	records, err := loadBenchmarkRecords()
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validation.ValidateBatchForParity(records, validation.ParityReferenceDate)
	}
}

func loadBenchmarkRecords() ([]validation.RecordInput, error) {
	file, err := os.Open(filepath.Join("..", "..", "data", "vecdim_dummy_entomology_data.csv"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return nil, err
	}

	columns := make(map[string]int, len(header))
	for i, name := range header {
		columns[name] = i
	}

	var records []validation.RecordInput
	for {
		row, err := reader.Read()
		if err == io.EOF {
			return records, nil
		}
		if err != nil {
			return nil, err
		}

		nExposed, err := strconv.Atoi(row[columns["n_exposed"]])
		if err != nil {
			return nil, err
		}
		nDead, err := strconv.Atoi(row[columns["n_dead"]])
		if err != nil {
			return nil, err
		}
		latitude, err := strconv.ParseFloat(row[columns["latitude"]], 64)
		if err != nil {
			return nil, err
		}
		longitude, err := strconv.ParseFloat(row[columns["longitude"]], 64)
		if err != nil {
			return nil, err
		}

		records = append(records, validation.RecordInput{
			RecordID:       row[columns["record_id"]],
			State:          row[columns["state"]],
			LGA:            row[columns["lga"]],
			Ward:           row[columns["ward"]],
			Latitude:       latitude,
			Longitude:      longitude,
			CollectionDate: row[columns["collection_date"]],
			Species:        benchmarkStringPtr(row[columns["species"]]),
			Insecticide:    benchmarkStringPtr(row[columns["insecticide"]]),
			NExposed:       nExposed,
			NDead:          nDead,
			LarvalDensity:  benchmarkIntPtr(row[columns["larval_density"]]),
			RainfallMm:     benchmarkFloatPtr(row[columns["rainfall_mm"]]),
			TempC:          benchmarkFloatPtr(row[columns["temp_c"]]),
			NDVI:           benchmarkFloatPtr(row[columns["ndvi"]]),
			ElevationM:     benchmarkFloatPtr(row[columns["elevation_m"]]),
			DistWaterKm:    benchmarkFloatPtr(row[columns["dist_water_km"]]),
			Urban:          benchmarkInt(row[columns["urban"]]),
			ITNUsePct:      benchmarkFloatPtr(row[columns["itn_use_pct"]]),
			IRSCovered:     benchmarkInt(row[columns["irs_covered"]]),
			SlidePositivityPct: benchmarkFloatPtr(
				row[columns["slide_positivity_pct"]],
			),
		})
	}
}

func benchmarkStringPtr(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" || value == "NA" {
		return nil
	}
	return &value
}

func benchmarkFloatPtr(value string) *float64 {
	value = strings.TrimSpace(value)
	if value == "" || value == "NA" {
		return nil
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil
	}
	return &parsed
}

func benchmarkIntPtr(value string) *int {
	value = strings.TrimSpace(value)
	if value == "" || value == "NA" {
		return nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return nil
	}
	return &parsed
}

func benchmarkInt(value string) int {
	parsed, _ := strconv.Atoi(strings.TrimSpace(value))
	return parsed
}
