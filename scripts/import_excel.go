package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/extrame/xls"
	"github.com/xuri/excelize/v2"

	"zip6/internal/database"
)

func main() {
	// Parse command line arguments
	excelFiles := flag.String("excel", "", "Comma-separated list of Excel files to import")
	dbPath := flag.String("db", "zipcode.db", "Path to SQLite database file")
	flag.Parse()

	if *excelFiles == "" {
		fmt.Println("Usage: go run scripts/import_excel.go -excel \"path/to/file1.xlsx,path/to/file2.xlsx\" -db \"zipcode.db\"")
		os.Exit(1)
	}

	// Initialize database
	db, err := database.InitDB(*dbPath)
	if err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Process each Excel file
	files := strings.Split(*excelFiles, ",")
	totalImported := 0

	for _, filePath := range files {
		filePath = strings.TrimSpace(filePath)
		if filePath == "" {
			continue
		}

		fmt.Printf("Processing: %s\n", filePath)
		var count int
		var err error

		ext := strings.ToLower(filepath.Ext(filePath))
		if ext == ".xls" {
			count, err = importXlsFile(db, filePath)
		} else {
			count, err = importXlsxFile(db, filePath)
		}

		if err != nil {
			fmt.Printf("Error importing %s: %v\n", filePath, err)
			continue
		}
		totalImported += count
		fmt.Printf("Imported %d records from %s\n", count, filePath)
	}

	// Show final count
	finalCount, _ := db.GetRecordCount()
	fmt.Printf("\n=== Import Complete ===\n")
	fmt.Printf("Total records imported: %d\n", totalImported)
	fmt.Printf("Total records in database: %d\n", finalCount)
}

// importXlsFile imports data from old .xls format (Excel 97-2003)
func importXlsFile(db *database.Database, filePath string) (int, error) {
	// Try Big5 encoding first (code page 950)
	xlFile, err := xls.Open(filePath, "big5")
	if err != nil {
		return 0, fmt.Errorf("failed to open XLS file: %w", err)
	}

	sheet := xlFile.GetSheet(0)
	if sheet == nil {
		return 0, fmt.Errorf("no sheet found in XLS file")
	}

	maxRow := int(sheet.MaxRow)
	if maxRow < 2 {
		return 0, fmt.Errorf("XLS file has no data rows")
	}

	// Get header row to find column indices
	headerRow := sheet.Row(0)
	colMap := make(map[string]int)
	for i := 0; i <= int(headerRow.LastCol()); i++ {
		cell := headerRow.Col(i)
		colMap[strings.TrimSpace(cell)] = i
	}

	// Find column indices
	cityCol := findColumn(colMap, "縣市")
	districtCol := findColumn(colMap, "區域")
	zipcodeCol := findColumn(colMap, "郵遞區號")
	streetCol := findColumn(colMap, "街路或聚落名稱", "路名")
	scopeCol := findColumn(colMap, "投遞範圍")
	officeCol := findColumn(colMap, "投遞局-區段名稱", "投遞局")
	noteCol := findColumn(colMap, "大宗戶或不按址投遞註記", "註記")

	if cityCol == -1 || districtCol == -1 || zipcodeCol == -1 || streetCol == -1 {
		var headers []string
		for k := range colMap {
			headers = append(headers, k)
		}
		return 0, fmt.Errorf("missing required columns. Found headers: %v", headers)
	}

	// Start transaction
	tx, err := db.BeginTransaction()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}

	count := 0
	for i := 1; i <= maxRow; i++ {
		row := sheet.Row(i)
		if row == nil {
			continue
		}

		record := database.ZipCodeRecord{
			City:           strings.TrimSpace(row.Col(cityCol)),
			District:       strings.TrimSpace(row.Col(districtCol)),
			Zipcode:        strings.TrimSpace(row.Col(zipcodeCol)),
			Street:         strings.TrimSpace(row.Col(streetCol)),
			Scope:          getColSafe(row, scopeCol),
			DeliveryOffice: getColSafe(row, officeCol),
			Note:           getColSafe(row, noteCol),
		}

		// Skip empty records
		if record.City == "" && record.District == "" && record.Zipcode == "" {
			continue
		}

		err := db.InsertRecordTx(tx, record)
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("failed to insert record at row %d: %w", i+1, err)
		}
		count++

		// Progress indicator
		if count%10000 == 0 {
			fmt.Printf("  Processed %d records...\n", count)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return count, nil
}

// getColSafe safely gets a column value from xls row
func getColSafe(row *xls.Row, idx int) string {
	if idx < 0 {
		return ""
	}
	return strings.TrimSpace(row.Col(idx))
}

// importXlsxFile imports data from .xlsx format
func importXlsxFile(db *database.Database, filePath string) (int, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	// Get the first sheet
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return 0, fmt.Errorf("no sheets found in Excel file")
	}
	sheetName := sheets[0]

	// Get all rows
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return 0, fmt.Errorf("failed to get rows: %w", err)
	}

	if len(rows) < 2 {
		return 0, fmt.Errorf("Excel file has no data rows")
	}

	// Start transaction for better performance
	tx, err := db.BeginTransaction()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Find column indices by header names
	headerRow := rows[0]
	colMap := make(map[string]int)
	for i, cell := range headerRow {
		colMap[strings.TrimSpace(cell)] = i
	}

	// Expected column names
	cityCol := findColumn(colMap, "縣市")
	districtCol := findColumn(colMap, "區域")
	zipcodeCol := findColumn(colMap, "郵遞區號")
	streetCol := findColumn(colMap, "街路或聚落名稱", "路名")
	scopeCol := findColumn(colMap, "投遞範圍")
	officeCol := findColumn(colMap, "投遞局-區段名稱", "投遞局")
	noteCol := findColumn(colMap, "大宗戶或不按址投遞註記", "註記")

	if cityCol == -1 || districtCol == -1 || zipcodeCol == -1 || streetCol == -1 {
		return 0, fmt.Errorf("missing required columns. Found headers: %v", headerRow)
	}

	count := 0
	for i, row := range rows[1:] {
		if len(row) == 0 {
			continue
		}

		record := database.ZipCodeRecord{
			City:           getCell(row, cityCol),
			District:       getCell(row, districtCol),
			Zipcode:        getCell(row, zipcodeCol),
			Street:         getCell(row, streetCol),
			Scope:          getCell(row, scopeCol),
			DeliveryOffice: getCell(row, officeCol),
			Note:           getCell(row, noteCol),
		}

		// Skip empty records
		if record.City == "" && record.District == "" && record.Zipcode == "" {
			continue
		}

		err := db.InsertRecordTx(tx, record)
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("failed to insert record at row %d: %w", i+2, err)
		}
		count++

		// Progress indicator
		if count%10000 == 0 {
			fmt.Printf("  Processed %d records...\n", count)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return count, nil
}

// findColumn finds a column index by trying multiple possible header names
func findColumn(colMap map[string]int, names ...string) int {
	for _, name := range names {
		if idx, ok := colMap[name]; ok {
			return idx
		}
	}
	return -1
}

// getCell safely gets a cell value from a row
func getCell(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}
