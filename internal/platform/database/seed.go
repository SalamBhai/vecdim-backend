package database

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/salambhai/vecdim-backend/ent"
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

func parseBool(s string) bool {
	s = strings.TrimSpace(strings.ToUpper(s))
	return s == "TRUE" || s == "T" || s == "1"
}

// SeedAll populates demo users, ground truth alerts, and GIS risk points.
func SeedAll(ctx context.Context, client *ent.Client) error {
	log.Println("Seeding demo users...")
	seedDemoUsers(ctx, client)

	log.Println("Seeding 647 ground truth alerts...")
	seedAlerts(ctx, client)

	log.Println("Seeding GIS risk points...")
	seedRiskPoints(ctx, client)

	log.Println("Database seeding completed.")
	return nil
}

func seedDemoUsers(ctx context.Context, client *ent.Client) {
	path := filepath.Join("data", "demo_users.csv")
	file, err := os.Open(path)
	if err != nil {
		path = filepath.Join("..", "data", "demo_users.csv")
		file, err = os.Open(path)
		if err != nil {
			log.Printf("Warning: cannot open demo_users.csv: %v", err)
			return
		}
	}
	defer file.Close()

	reader := csv.NewReader(file)
	_, _ = reader.Read() // Skip header

	count := 0
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil || len(row) < 3 {
			continue
		}

		userID := row[0]
		name := row[1]
		role := row[2]
		state := ""
		if len(row) > 3 {
			state = row[3]
		}

		uCreate := client.User.Create().
			SetID(userID).
			SetName(name).
			SetRole(role)

		if state != "" {
			uCreate.SetState(state)
		}

		_ = uCreate.Exec(ctx)
		count++
	}
	log.Printf("Seeded %d demo users", count)
}

func seedAlerts(ctx context.Context, client *ent.Client) {
	path := filepath.Join("data", "alerts.csv")
	file, err := os.Open(path)
	if err != nil {
		path = filepath.Join("..", "data", "alerts.csv")
		file, err = os.Open(path)
		if err != nil {
			log.Printf("Warning: cannot open alerts.csv: %v", err)
			return
		}
	}
	defer file.Close()

	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return
	}

	colIdx := make(map[string]int)
	for i, col := range header {
		colIdx[col] = i
	}

	count := 0
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		state := row[colIdx["state"]]
		lga := row[colIdx["lga"]]
		month := row[colIdx["ym"]]
		n, _ := strconv.Atoi(row[colIdx["n"]])
		meanRisk, _ := strconv.ParseFloat(row[colIdx["mean_p_high"]], 64)
		meanMort := parseFloatPtr(row[colIdx["mean_pyrethroid_mort"]])
		tier := row[colIdx["alert"]]
		pyrethroidFlag := parseBool(row[colIdx["pyrethroid_resistance_flag"]])

		alertID := fmt.Sprintf("%s|%s|%s", state, lga, month)

		aCreate := client.Alert.Create().
			SetID(alertID).
			SetState(state).
			SetLga(lga).
			SetMonth(month).
			SetNRecords(n).
			SetMeanRisk(meanRisk).
			SetTier(tier).
			SetPyrethroidResistanceFlag(pyrethroidFlag)

		if meanMort != nil {
			aCreate.SetMeanPyrethroidMort(*meanMort)
		}

		_ = aCreate.Exec(ctx)
		count++
	}
	log.Printf("Seeded %d early warning alerts", count)
}

func seedRiskPoints(ctx context.Context, client *ent.Client) {
	path := filepath.Join("data", "risk_predictions.csv")
	file, err := os.Open(path)
	if err != nil {
		path = filepath.Join("..", "data", "risk_predictions.csv")
		file, err = os.Open(path)
		if err != nil {
			log.Printf("Warning: cannot open risk_predictions.csv: %v", err)
			return
		}
	}
	defer file.Close()

	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return
	}

	colIdx := make(map[string]int)
	for i, col := range header {
		colIdx[col] = i
	}

	count := 0
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		recID := row[colIdx["record_id"]]
		state := row[colIdx["state"]]
		lga := row[colIdx["lga"]]
		ward := row[colIdx["ward"]]
		lat, _ := strconv.ParseFloat(row[colIdx["latitude"]], 64)
		lon, _ := strconv.ParseFloat(row[colIdx["longitude"]], 64)
		collectionDate := row[colIdx["collection_date"]]
		pHigh, _ := strconv.ParseFloat(row[colIdx["p_high"]], 64)
		tier := row[colIdx["tier"]]

		_ = client.RiskPoint.Create().
			SetID(recID).
			SetState(state).
			SetLga(lga).
			SetWard(ward).
			SetLatitude(lat).
			SetLongitude(lon).
			SetCollectionDate(collectionDate).
			SetPHigh(pHigh).
			SetTier(tier).
			Exec(ctx)

		count++
	}
	log.Printf("Seeded %d GIS risk points", count)
}
