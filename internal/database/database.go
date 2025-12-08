package database

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type Database struct {
	db *sql.DB
}

type ZipCodeRecord struct {
	ID             int
	City           string
	District       string
	Zipcode        string
	Street         string
	Scope          string
	DeliveryOffice string
	Note           string
}

// InitDB initializes the database and creates tables if they don't exist
func InitDB(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Create tables
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS zipcode (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		city TEXT NOT NULL,
		district TEXT NOT NULL,
		zipcode TEXT NOT NULL,
		street TEXT NOT NULL,
		scope TEXT,
		delivery_office TEXT,
		note TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_city ON zipcode(city);
	CREATE INDEX IF NOT EXISTS idx_district ON zipcode(district);
	CREATE INDEX IF NOT EXISTS idx_zipcode ON zipcode(zipcode);
	CREATE INDEX IF NOT EXISTS idx_street ON zipcode(street);

	-- English translation tables
	CREATE TABLE IF NOT EXISTS street_en (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chinese TEXT NOT NULL UNIQUE,
		english TEXT NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_street_en_chinese ON street_en(chinese);

	CREATE TABLE IF NOT EXISTS district_en (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		zipcode3 TEXT NOT NULL,
		chinese TEXT NOT NULL,
		english TEXT NOT NULL,
		UNIQUE(zipcode3, chinese)
	);
	CREATE INDEX IF NOT EXISTS idx_district_en_zipcode ON district_en(zipcode3);

	CREATE TABLE IF NOT EXISTS village_lane_en (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chinese TEXT NOT NULL UNIQUE,
		english TEXT NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_village_lane_en_chinese ON village_lane_en(chinese);
	`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &Database{db: db}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// Search performs a fuzzy search across multiple fields
func (d *Database) Search(keyword string, limit int) ([]ZipCodeRecord, int, error) {
	if keyword == "" {
		return []ZipCodeRecord{}, 0, nil
	}

	likePattern := "%" + keyword + "%"

	// Count total matches
	countSQL := `
		SELECT COUNT(*) FROM zipcode
		WHERE city LIKE ? OR district LIKE ? OR zipcode LIKE ? OR street LIKE ? OR scope LIKE ?
	`
	var total int
	err := d.db.QueryRow(countSQL, likePattern, likePattern, likePattern, likePattern, likePattern).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count results: %w", err)
	}

	// Get results with limit
	querySQL := `
		SELECT id, city, district, zipcode, street, COALESCE(scope, ''), COALESCE(delivery_office, ''), COALESCE(note, '')
		FROM zipcode
		WHERE city LIKE ? OR district LIKE ? OR zipcode LIKE ? OR street LIKE ? OR scope LIKE ?
		ORDER BY city, district, street
		LIMIT ?
	`
	rows, err := d.db.Query(querySQL, likePattern, likePattern, likePattern, likePattern, likePattern, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query: %w", err)
	}
	defer rows.Close()

	var results []ZipCodeRecord
	for rows.Next() {
		var r ZipCodeRecord
		err := rows.Scan(&r.ID, &r.City, &r.District, &r.Zipcode, &r.Street, &r.Scope, &r.DeliveryOffice, &r.Note)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, r)
	}

	return results, total, nil
}

// GetCities returns all distinct cities
func (d *Database) GetCities() ([]string, error) {
	rows, err := d.db.Query("SELECT DISTINCT city FROM zipcode ORDER BY city")
	if err != nil {
		return nil, fmt.Errorf("failed to query cities: %w", err)
	}
	defer rows.Close()

	var cities []string
	for rows.Next() {
		var city string
		if err := rows.Scan(&city); err != nil {
			return nil, fmt.Errorf("failed to scan city: %w", err)
		}
		cities = append(cities, city)
	}
	return cities, nil
}

// GetDistricts returns all districts for a given city
func (d *Database) GetDistricts(city string) ([]string, error) {
	rows, err := d.db.Query("SELECT DISTINCT district FROM zipcode WHERE city = ? ORDER BY district", city)
	if err != nil {
		return nil, fmt.Errorf("failed to query districts: %w", err)
	}
	defer rows.Close()

	var districts []string
	for rows.Next() {
		var district string
		if err := rows.Scan(&district); err != nil {
			return nil, fmt.Errorf("failed to scan district: %w", err)
		}
		districts = append(districts, district)
	}
	return districts, nil
}

// SearchAdvanced performs an advanced search with specific field filters
func (d *Database) SearchAdvanced(city, district, street string, limit int) ([]ZipCodeRecord, int, error) {
	var conditions []string
	var args []interface{}

	if city != "" {
		conditions = append(conditions, "city = ?")
		args = append(args, city)
	}
	if district != "" {
		conditions = append(conditions, "district = ?")
		args = append(args, district)
	}
	if street != "" {
		conditions = append(conditions, "street LIKE ?")
		args = append(args, "%"+street+"%")
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE "
		for i, cond := range conditions {
			if i > 0 {
				whereClause += " AND "
			}
			whereClause += cond
		}
	}

	// Count total
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM zipcode %s", whereClause)
	var total int
	err := d.db.QueryRow(countSQL, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count results: %w", err)
	}

	// Get results
	querySQL := fmt.Sprintf(`
		SELECT id, city, district, zipcode, street, COALESCE(scope, ''), COALESCE(delivery_office, ''), COALESCE(note, '')
		FROM zipcode %s
		ORDER BY city, district, street
		LIMIT ?
	`, whereClause)
	args = append(args, limit)

	rows, err := d.db.Query(querySQL, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query: %w", err)
	}
	defer rows.Close()

	var results []ZipCodeRecord
	for rows.Next() {
		var r ZipCodeRecord
		err := rows.Scan(&r.ID, &r.City, &r.District, &r.Zipcode, &r.Street, &r.Scope, &r.DeliveryOffice, &r.Note)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, r)
	}

	return results, total, nil
}

// InsertRecord inserts a single record into the database
func (d *Database) InsertRecord(r ZipCodeRecord) error {
	_, err := d.db.Exec(`
		INSERT INTO zipcode (city, district, zipcode, street, scope, delivery_office, note)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, r.City, r.District, r.Zipcode, r.Street, r.Scope, r.DeliveryOffice, r.Note)
	return err
}

// BeginTransaction starts a transaction for batch inserts
func (d *Database) BeginTransaction() (*sql.Tx, error) {
	return d.db.Begin()
}

// InsertRecordTx inserts a record within a transaction
func (d *Database) InsertRecordTx(tx *sql.Tx, r ZipCodeRecord) error {
	_, err := tx.Exec(`
		INSERT INTO zipcode (city, district, zipcode, street, scope, delivery_office, note)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, r.City, r.District, r.Zipcode, r.Street, r.Scope, r.DeliveryOffice, r.Note)
	return err
}

// GetRecordCount returns the total number of records in the database
func (d *Database) GetRecordCount() (int, error) {
	var count int
	err := d.db.QueryRow("SELECT COUNT(*) FROM zipcode").Scan(&count)
	return count, err
}

// SearchByStreet searches for records by street name pattern
func (d *Database) SearchByStreet(streetPattern string, limit int) ([]ZipCodeRecord, error) {
	querySQL := `
		SELECT id, city, district, zipcode, street, COALESCE(scope, ''), COALESCE(delivery_office, ''), COALESCE(note, '')
		FROM zipcode
		WHERE street LIKE ?
		ORDER BY city, district, zipcode
		LIMIT ?
	`
	rows, err := d.db.Query(querySQL, "%"+streetPattern+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	defer rows.Close()

	var results []ZipCodeRecord
	for rows.Next() {
		var r ZipCodeRecord
		err := rows.Scan(&r.ID, &r.City, &r.District, &r.Zipcode, &r.Street, &r.Scope, &r.DeliveryOffice, &r.Note)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, r)
	}

	return results, nil
}

// GetStreetEnglish returns the English translation for a Chinese street name
func (d *Database) GetStreetEnglish(chinese string) (string, error) {
	var english string
	err := d.db.QueryRow("SELECT english FROM street_en WHERE chinese = ?", chinese).Scan(&english)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return english, err
}

// GetDistrictEnglish returns the English translation for a Chinese district
func (d *Database) GetDistrictEnglish(zipcode3 string) (string, error) {
	var english string
	err := d.db.QueryRow("SELECT english FROM district_en WHERE zipcode3 = ?", zipcode3).Scan(&english)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return english, err
}

// GetVillageLaneEnglish returns the English translation for village/lane names
func (d *Database) GetVillageLaneEnglish(chinese string) (string, error) {
	var english string
	err := d.db.QueryRow("SELECT english FROM village_lane_en WHERE chinese = ?", chinese).Scan(&english)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return english, err
}

// InsertStreetEnglish inserts a street name translation
func (d *Database) InsertStreetEnglish(tx *sql.Tx, chinese, english string) error {
	_, err := tx.Exec(`INSERT OR IGNORE INTO street_en (chinese, english) VALUES (?, ?)`, chinese, english)
	return err
}

// InsertDistrictEnglish inserts a district translation
func (d *Database) InsertDistrictEnglish(tx *sql.Tx, zipcode3, chinese, english string) error {
	_, err := tx.Exec(`INSERT OR IGNORE INTO district_en (zipcode3, chinese, english) VALUES (?, ?, ?)`, zipcode3, chinese, english)
	return err
}

// InsertVillageLaneEnglish inserts a village/lane translation
func (d *Database) InsertVillageLaneEnglish(tx *sql.Tx, chinese, english string) error {
	_, err := tx.Exec(`INSERT OR IGNORE INTO village_lane_en (chinese, english) VALUES (?, ?)`, chinese, english)
	return err
}

// SearchByStreetAndDistrict searches for records by street and optional district
func (d *Database) SearchByStreetAndDistrict(streetPattern, city, district string, limit int) ([]ZipCodeRecord, error) {
	var conditions []string
	var args []interface{}

	conditions = append(conditions, "street LIKE ?")
	args = append(args, "%"+streetPattern+"%")

	if city != "" {
		conditions = append(conditions, "city LIKE ?")
		args = append(args, "%"+city+"%")
	}
	if district != "" {
		// Use LIKE for fuzzy matching (e.g., 竹東鎮 matches 竹東鎮)
		conditions = append(conditions, "district LIKE ?")
		args = append(args, "%"+district+"%")
	}

	whereClause := "WHERE " + conditions[0]
	for i := 1; i < len(conditions); i++ {
		whereClause += " AND " + conditions[i]
	}

	querySQL := fmt.Sprintf(`
		SELECT id, city, district, zipcode, street, COALESCE(scope, ''), COALESCE(delivery_office, ''), COALESCE(note, '')
		FROM zipcode %s
		ORDER BY city, district, zipcode
		LIMIT ?
	`, whereClause)
	args = append(args, limit)

	rows, err := d.db.Query(querySQL, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	defer rows.Close()

	var results []ZipCodeRecord
	for rows.Next() {
		var r ZipCodeRecord
		err := rows.Scan(&r.ID, &r.City, &r.District, &r.Zipcode, &r.Street, &r.Scope, &r.DeliveryOffice, &r.Note)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, r)
	}

	return results, nil
}
