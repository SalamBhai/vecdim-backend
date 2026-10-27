package validation_test

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/salambhai/vecdim-backend/internal/validation"
	"github.com/stretchr/testify/require"
)

func parseFloatPtr(s string) *float64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "NA" {
		return nil
	}
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &val
}

func parseIntPtr(s string) *int {
	s = strings.TrimSpace(s)
	if s == "" || s == "NA" {
		return nil
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &val
}

func parseStringPtr(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" || s == "NA" {
		return nil
	}
	return &s
}

func parseBool(s string) bool {
	s = strings.TrimSpace(strings.ToUpper(s))
	return s == "TRUE" || s == "T" || s == "1"
}

func Test11RulesParityAgainstGroundTruth(t *testing.T) {
	dataPath := filepath.Join("..", "..", "data", "vecdim_dummy_entomology_data.csv")
	flagsPath := filepath.Join("..", "..", "data", "validation_flags_all.csv")

	dataFile, err := os.Open(dataPath)
	require.NoError(t, err, "failed to open dummy entomology data")
	defer dataFile.Close()

	reader := csv.NewReader(dataFile)
	header, err := reader.Read()
	require.NoError(t, err)

	colIdx := make(map[string]int)
	for i, col := range header {
		colIdx[col] = i
	}

	var records []validation.RecordInput
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)

		nExposed, _ := strconv.Atoi(row[colIdx["n_exposed"]])
		nDead, _ := strconv.Atoi(row[colIdx["n_dead"]])
		lat, _ := strconv.ParseFloat(row[colIdx["latitude"]], 64)
		lon, _ := strconv.ParseFloat(row[colIdx["longitude"]], 64)

		urban := 0
		if colIdx["urban"] < len(row) && row[colIdx["urban"]] != "" {
			urban, _ = strconv.Atoi(row[colIdx["urban"]])
		}

		irsCovered := 0
		if colIdx["irs_covered"] < len(row) && row[colIdx["irs_covered"]] != "" {
			irsCovered, _ = strconv.Atoi(row[colIdx["irs_covered"]])
		}

		rec := validation.RecordInput{
			RecordID:           row[colIdx["record_id"]],
			State:              row[colIdx["state"]],
			LGA:                row[colIdx["lga"]],
			Ward:               row[colIdx["ward"]],
			Latitude:           lat,
			Longitude:          lon,
			CollectionDate:     row[colIdx["collection_date"]],
			Species:            parseStringPtr(row[colIdx["species"]]),
			Insecticide:        parseStringPtr(row[colIdx["insecticide"]]),
			NExposed:           nExposed,
			NDead:              nDead,
			LarvalDensity:      parseIntPtr(row[colIdx["larval_density"]]),
			RainfallMm:         parseFloatPtr(row[colIdx["rainfall_mm"]]),
			TempC:              parseFloatPtr(row[colIdx["temp_c"]]),
			NDVI:               parseFloatPtr(row[colIdx["ndvi"]]),
			ElevationM:         parseFloatPtr(row[colIdx["elevation_m"]]),
			DistWaterKm:        parseFloatPtr(row[colIdx["dist_water_km"]]),
			Urban:              urban,
			ITNUsePct:          parseFloatPtr(row[colIdx["itn_use_pct"]]),
			IRSCovered:         irsCovered,
			SlidePositivityPct: parseFloatPtr(row[colIdx["slide_positivity_pct"]]),
		}
		records = append(records, rec)
	}

	require.Equal(t, 1525, len(records), "Dataset must contain exactly 1,525 rows")

	// Run Go parity validator with candidate pack reference date 2026-10-08
	goFlags := validation.ValidateBatchForParity(records, validation.ParityReferenceDate)
	require.Equal(t, 1525, len(goFlags))

	goFlagsByRecordID := make(map[string]validation.ValidationFlagsAll)
	flaggedCount := 0
	cleanCount := 0
	for _, f := range goFlags {
		goFlagsByRecordID[f.RecordID] = f
		if f.AnyFlag {
			flaggedCount++
		} else {
			cleanCount++
		}
	}

	// Read R ground truth validation_flags_all.csv
	flagsFile, err := os.Open(flagsPath)
	require.NoError(t, err, "failed to open validation_flags_all.csv")
	defer flagsFile.Close()

	flagsReader := csv.NewReader(flagsFile)
	flagsHeader, err := flagsReader.Read()
	require.NoError(t, err)

	fColIdx := make(map[string]int)
	for i, col := range flagsHeader {
		fColIdx[col] = i
	}

	mismatchCount := 0
	var mismatchDetails []string

	rTotalRows := 0
	rFlaggedCount := 0

	for {
		row, err := flagsReader.Read()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		rTotalRows++

		recID := row[fColIdx["record_id"]]
		goFlag, exists := goFlagsByRecordID[recID]
		require.True(t, exists, "record_id not found in Go results: %s", recID)

		rGPSOutside := parseBool(row[fColIdx["GPS_outside_nigeria"]])
		rGPSZero := parseBool(row[fColIdx["GPS_zero_zero"]])
		rGPSStateMismatch := parseBool(row[fColIdx["GPS_state_mismatch"]])
		rGPSFar := parseBool(row[fColIdx["GPS_far_from_lga"]])
		rNDeadGt := parseBool(row[fColIdx["n_dead_gt_n_exposed"]])
		rPositivity := parseBool(row[fColIdx["positivity_out_range"]])
		rFutureDate := parseBool(row[fColIdx["future_date"]])
		rTempImplausible := parseBool(row[fColIdx["temp_implausible"]])
		rRainfallImplausible := parseBool(row[fColIdx["rainfall_implausible"]])
		rDuplicate := parseBool(row[fColIdx["duplicate"]])
		rMissingCore := parseBool(row[fColIdx["missing_core"]])
		rAnyFlag := parseBool(row[fColIdx["any_flag"]])

		if rAnyFlag {
			rFlaggedCount++
		}

		// Compare every single rule
		rowMismatch := false
		if goFlag.GPSOutsideNigeria != rGPSOutside {
			rowMismatch = true
		}
		if goFlag.GPSZeroZero != rGPSZero {
			rowMismatch = true
		}
		if goFlag.GPSStateMismatch != rGPSStateMismatch {
			rowMismatch = true
		}
		if goFlag.GPSFarFromLGA != rGPSFar {
			rowMismatch = true
		}
		if goFlag.NDeadGtNExposed != rNDeadGt {
			rowMismatch = true
		}
		if goFlag.PositivityOutRange != rPositivity {
			rowMismatch = true
		}
		if goFlag.FutureDate != rFutureDate {
			rowMismatch = true
		}
		if goFlag.TempImplausible != rTempImplausible {
			rowMismatch = true
		}
		if goFlag.RainfallImplausible != rRainfallImplausible {
			rowMismatch = true
		}
		if goFlag.Duplicate != rDuplicate {
			rowMismatch = true
		}
		if goFlag.MissingCore != rMissingCore {
			rowMismatch = true
		}
		if goFlag.AnyFlag != rAnyFlag {
			rowMismatch = true
		}

		if rowMismatch {
			mismatchCount++
			mismatchDetails = append(mismatchDetails, fmt.Sprintf(
				"Mismatch on %s: Go(any=%t) vs R(any=%t)",
				recID, goFlag.AnyFlag, rAnyFlag,
			))
		}
	}

	report := fmt.Sprintf(`================================================================================
VECDIM VALIDATION ENGINE PARITY TEST REPORT
================================================================================
Total Records Evaluated : %d
Ground Truth Rows in R  : %d
Go Flagged Records      : %d (Expected: 222)
Go Clean Records        : %d (Expected: 1,303)
R Flagged Records       : %d
Differing Rows          : %d (Threshold: <= 5, Target: 0)
Parity Accuracy         : %.2f%%
================================================================================
`, len(records), rTotalRows, flaggedCount, cleanCount, rFlaggedCount, mismatchCount,
		100.0*(1.0-float64(mismatchCount)/float64(len(records))))

	// Save parity report to file for candidate pack submission
	_ = os.WriteFile("parity_test_results.txt", []byte(report), 0644)
	fmt.Println(report)

	require.Equal(t, 222, flaggedCount, "Go validator must flag exactly 222 rows")
	require.Equal(t, 1303, cleanCount, "Go validator must retain exactly 1,303 clean rows")
	require.Equal(t, 0, mismatchCount, "Must have zero mismatched rows against validation_flags_all.csv")
}

func TestGPSAccuracyEdgeCases(t *testing.T) {
	// Unit test for the two GPS accuracy rules not in the dataset
	// Edge cases: 10, 10.1, 50, 50.1m
	cases := []struct {
		accuracy   float64
		expectWarn bool
		expectFail bool
	}{
		{accuracy: 5.0, expectWarn: false, expectFail: false},
		{accuracy: 10.0, expectWarn: false, expectFail: false},
		{accuracy: 10.1, expectWarn: true, expectFail: false},
		{accuracy: 50.0, expectWarn: true, expectFail: false},
		{accuracy: 50.1, expectWarn: false, expectFail: true},
		{accuracy: 75.0, expectWarn: false, expectFail: true},
	}

	for _, tc := range cases {
		acc := tc.accuracy
		rec := validation.RecordInput{
			RecordID:       "TEST_ACC",
			State:          "Osun",
			LGA:            "Osogbo",
			Ward:           "Ward 1",
			Latitude:       7.77,
			Longitude:      4.56,
			CollectionDate: "2026-10-01",
			GPSAccuracyM:   &acc,
			NExposed:       100,
			NDead:          50,
		}
		blocks, warns := validation.ValidateSingleRecord(rec, "2026-10-08", true)

		hasPoorWarn := false
		for _, w := range warns {
			if w.Rule == "GPS_accuracy_poor" {
				hasPoorWarn = true
			}
		}

		hasUnusableBlock := false
		for _, b := range blocks {
			if b.Rule == "GPS_accuracy_unusable" {
				hasUnusableBlock = true
			}
		}

		require.Equal(t, tc.expectWarn, hasPoorWarn, "Failed on accuracy %.1f for warn", tc.accuracy)
		require.Equal(t, tc.expectFail, hasUnusableBlock, "Failed on accuracy %.1f for block", tc.accuracy)
	}
}

func TestWHOMortalityRoundingClasses(t *testing.T) {
	cases := []struct {
		nExposed      int
		nDead         int
		expectedPct   float64
		expectedClass string
	}{
		{100, 98, 98.0, "Susceptible"},
		{100, 100, 100.0, "Susceptible"},
		{100, 90, 90.0, "Possible resistance"},
		{1000, 975, 97.5, "Possible resistance"},
		{1000, 979, 97.9, "Possible resistance"},
		{1000, 899, 89.9, "Resistant"},
		{100, 50, 50.0, "Resistant"},
	}

	for _, tc := range cases {
		pct, class := validation.CalculateWHOMortalityClass(tc.nExposed, tc.nDead)
		require.Equal(t, tc.expectedPct, pct)
		require.Equal(t, tc.expectedClass, class)
	}
}
