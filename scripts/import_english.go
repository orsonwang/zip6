package main

import (
	"flag"
	"fmt"
	"strings"

	"zip6/internal/database"

	"github.com/xuri/excelize/v2"
)

func main() {
	dbPath := flag.String("db", "zipcode.db", "Path to SQLite database")
	streetFile := flag.String("street", "中英文街路名稱對照檔1130401.xlsx", "Street name translation file")
	districtFile := flag.String("district", "鄉鎮市區中英對照.xlsx", "District translation file")
	villageLaneFile := flag.String("village", "村里文字巷中英對照.xlsx", "Village/Lane translation file")
	flag.Parse()

	// Initialize database
	db, err := database.InitDB(*dbPath)
	if err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		return
	}
	defer db.Close()

	// Import street names
	if *streetFile != "" {
		count, err := importStreetNames(db, *streetFile)
		if err != nil {
			fmt.Printf("Error importing street names: %v\n", err)
		} else {
			fmt.Printf("Imported %d street name translations\n", count)
		}
	}

	// Import district names
	if *districtFile != "" {
		count, err := importDistrictNames(db, *districtFile)
		if err != nil {
			fmt.Printf("Error importing district names: %v\n", err)
		} else {
			fmt.Printf("Imported %d district translations\n", count)
		}
	}

	// Import village/lane names
	if *villageLaneFile != "" {
		count, err := importVillageLaneNames(db, *villageLaneFile)
		if err != nil {
			fmt.Printf("Error importing village/lane names: %v\n", err)
		} else {
			fmt.Printf("Imported %d village/lane translations\n", count)
		}
	}

	fmt.Println("English translation import complete!")
}

func importStreetNames(db *database.Database, filePath string) (int, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return 0, fmt.Errorf("failed to get rows: %w", err)
	}

	tx, err := db.BeginTransaction()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}

	count := 0
	for i, row := range rows {
		// Skip header row
		if i == 0 {
			continue
		}
		if len(row) < 2 {
			continue
		}

		chinese := strings.TrimSpace(row[0])
		english := strings.TrimSpace(row[1])

		if chinese == "" || english == "" {
			continue
		}

		err := db.InsertStreetEnglish(tx, chinese, english)
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("failed to insert street: %w", err)
		}
		count++
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit: %w", err)
	}

	return count, nil
}

func importDistrictNames(db *database.Database, filePath string) (int, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return 0, fmt.Errorf("failed to get rows: %w", err)
	}

	tx, err := db.BeginTransaction()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}

	count := 0
	for _, row := range rows {
		if len(row) < 3 {
			continue
		}

		zipcode3 := strings.TrimSpace(row[0])
		chinese := strings.TrimSpace(row[1])
		english := strings.TrimSpace(row[2])

		if zipcode3 == "" || chinese == "" || english == "" {
			continue
		}

		// Skip if first column is not a number (header row)
		if len(zipcode3) != 3 {
			continue
		}

		err := db.InsertDistrictEnglish(tx, zipcode3, chinese, english)
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("failed to insert district: %w", err)
		}
		count++
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit: %w", err)
	}

	return count, nil
}

func importVillageLaneNames(db *database.Database, filePath string) (int, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return 0, fmt.Errorf("failed to get rows: %w", err)
	}

	tx, err := db.BeginTransaction()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}

	count := 0
	for _, row := range rows {
		if len(row) < 1 {
			continue
		}

		// This file has CSV format in single column: 中文,"English"
		cell := strings.TrimSpace(row[0])
		if cell == "" {
			continue
		}

		// Parse CSV format: 一心里,"Yixin Vil."
		parts := strings.SplitN(cell, ",", 2)
		if len(parts) != 2 {
			continue
		}

		chinese := strings.TrimSpace(parts[0])
		english := strings.TrimSpace(parts[1])
		// Remove quotes from English
		english = strings.Trim(english, "\"")

		if chinese == "" || english == "" {
			continue
		}

		err := db.InsertVillageLaneEnglish(tx, chinese, english)
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("failed to insert village/lane: %w", err)
		}
		count++
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit: %w", err)
	}

	return count, nil
}
