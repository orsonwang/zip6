package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"zip6/internal/address"
	"zip6/internal/database"
)

// App struct
type App struct {
	ctx context.Context
	db  *database.Database
}

// ZipCodeResult represents a single search result
type ZipCodeResult struct {
	ID             int    `json:"id"`
	City           string `json:"city"`
	District       string `json:"district"`
	Zipcode        string `json:"zipcode"`
	Street         string `json:"street"`
	Scope          string `json:"scope"`
	DeliveryOffice string `json:"deliveryOffice"`
	Note           string `json:"note"`
}

// SearchResponse represents the response from a search operation
type SearchResponse struct {
	Results   []ZipCodeResult `json:"results"`
	Total     int             `json:"total"`
	Truncated bool            `json:"truncated"`
}

// AddressSearchResult represents a matched zipcode result with match info
type AddressSearchResult struct {
	Zipcode  string `json:"zipcode"`
	City     string `json:"city"`
	District string `json:"district"`
	Street   string `json:"street"`
	Scope    string `json:"scope"`
	Matched  bool   `json:"matched"` // true if scope matches the address
}

// AddressSearchResponse represents the response from address search
type AddressSearchResponse struct {
	ParsedAddress   string                `json:"parsedAddress"`
	EnglishAddress  string                `json:"englishAddress"`
	Results         []AddressSearchResult `json:"results"`
	BestMatch       *AddressSearchResult  `json:"bestMatch"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Get executable directory
	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("Error getting executable path: %v\n", err)
		return
	}
	execDir := filepath.Dir(execPath)

	// Try to find database in several locations
	dbPaths := []string{
		filepath.Join(execDir, "zipcode.db"),
		"zipcode.db",
		filepath.Join(".", "zipcode.db"),
	}

	var dbPath string
	for _, p := range dbPaths {
		if _, err := os.Stat(p); err == nil {
			dbPath = p
			break
		}
	}

	if dbPath == "" {
		// Use default path, database will be created if it doesn't exist
		dbPath = filepath.Join(execDir, "zipcode.db")
	}

	// Initialize database
	db, err := database.InitDB(dbPath)
	if err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		return
	}
	a.db = db

	fmt.Printf("Database initialized at: %s\n", dbPath)
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	if a.db != nil {
		a.db.Close()
	}
}

// Search performs a fuzzy search across all fields
func (a *App) Search(keyword string) SearchResponse {
	if a.db == nil {
		return SearchResponse{Results: []ZipCodeResult{}, Total: 0, Truncated: false}
	}

	const maxResults = 500
	records, total, err := a.db.Search(keyword, maxResults)
	if err != nil {
		fmt.Printf("Search error: %v\n", err)
		return SearchResponse{Results: []ZipCodeResult{}, Total: 0, Truncated: false}
	}

	results := make([]ZipCodeResult, len(records))
	for i, r := range records {
		results[i] = ZipCodeResult{
			ID:             r.ID,
			City:           r.City,
			District:       r.District,
			Zipcode:        r.Zipcode,
			Street:         r.Street,
			Scope:          r.Scope,
			DeliveryOffice: r.DeliveryOffice,
			Note:           r.Note,
		}
	}

	return SearchResponse{
		Results:   results,
		Total:     total,
		Truncated: total > maxResults,
	}
}

// GetCities returns all distinct cities
func (a *App) GetCities() []string {
	if a.db == nil {
		return []string{}
	}

	cities, err := a.db.GetCities()
	if err != nil {
		fmt.Printf("GetCities error: %v\n", err)
		return []string{}
	}
	return cities
}

// GetDistricts returns all districts for a given city
func (a *App) GetDistricts(city string) []string {
	if a.db == nil {
		return []string{}
	}

	districts, err := a.db.GetDistricts(city)
	if err != nil {
		fmt.Printf("GetDistricts error: %v\n", err)
		return []string{}
	}
	return districts
}

// SearchAdvanced performs an advanced search with specific field filters
func (a *App) SearchAdvanced(city, district, street string) SearchResponse {
	if a.db == nil {
		return SearchResponse{Results: []ZipCodeResult{}, Total: 0, Truncated: false}
	}

	const maxResults = 500
	records, total, err := a.db.SearchAdvanced(city, district, street, maxResults)
	if err != nil {
		fmt.Printf("SearchAdvanced error: %v\n", err)
		return SearchResponse{Results: []ZipCodeResult{}, Total: 0, Truncated: false}
	}

	results := make([]ZipCodeResult, len(records))
	for i, r := range records {
		results[i] = ZipCodeResult{
			ID:             r.ID,
			City:           r.City,
			District:       r.District,
			Zipcode:        r.Zipcode,
			Street:         r.Street,
			Scope:          r.Scope,
			DeliveryOffice: r.DeliveryOffice,
			Note:           r.Note,
		}
	}

	return SearchResponse{
		Results:   results,
		Total:     total,
		Truncated: total > maxResults,
	}
}

// SearchByAddress searches for zipcode by parsing a full address
func (a *App) SearchByAddress(fullAddress string) AddressSearchResponse {
	if a.db == nil {
		return AddressSearchResponse{ParsedAddress: fullAddress, Results: []AddressSearchResult{}}
	}

	// Parse the address
	parsed := address.ParseAddress(fullAddress)

	// Build parsed address description
	parsedDesc := fmt.Sprintf("路名:%s", parsed.Street)
	if parsed.Section != "" {
		parsedDesc += parsed.Section
	}
	if parsed.Lane > 0 {
		parsedDesc += fmt.Sprintf(" %d巷", parsed.Lane)
	}
	if parsed.Alley > 0 {
		parsedDesc += fmt.Sprintf(" %d弄", parsed.Alley)
	}
	if parsed.Number > 0 {
		parsedDesc += fmt.Sprintf(" %d號", parsed.Number)
	}
	if parsed.Floor > 0 {
		parsedDesc += fmt.Sprintf(" %d樓", parsed.Floor)
	}
	if parsed.Room > 0 {
		parsedDesc += fmt.Sprintf("之%d", parsed.Room)
	}

	response := AddressSearchResponse{
		ParsedAddress: parsedDesc,
		Results:       []AddressSearchResult{},
	}

	// Get street search pattern
	streetPattern := address.GetStreetSearchPattern(parsed)
	if streetPattern == "" {
		// Try to extract any street-like pattern from the input
		streetPattern = fullAddress
	}

	// Search database
	records, err := a.db.SearchByStreetAndDistrict(streetPattern, parsed.City, parsed.District, 1000)
	if err != nil {
		fmt.Printf("SearchByAddress error: %v\n", err)
		return response
	}

	// Match each record's scope against the parsed address
	var bestMatch *AddressSearchResult
	var bestMatchZipcode string
	for _, r := range records {
		scope := address.ParseScope(r.Scope)
		matched := address.MatchAddress(parsed, scope)

		result := AddressSearchResult{
			Zipcode:  r.Zipcode,
			City:     r.City,
			District: r.District,
			Street:   r.Street,
			Scope:    r.Scope,
			Matched:  matched,
		}

		if matched && bestMatch == nil {
			bestMatch = &result
			bestMatchZipcode = r.Zipcode
		}

		response.Results = append(response.Results, result)
	}

	response.BestMatch = bestMatch

	// Generate English address if we have a best match
	if bestMatch != nil {
		translator := address.NewTranslator(a.db)
		response.EnglishAddress = translator.TranslateAddress(parsed, bestMatchZipcode)
	}

	// Sort results: matched first, then by zipcode
	matchedResults := []AddressSearchResult{}
	unmatchedResults := []AddressSearchResult{}
	for _, r := range response.Results {
		if r.Matched {
			matchedResults = append(matchedResults, r)
		} else {
			unmatchedResults = append(unmatchedResults, r)
		}
	}
	response.Results = append(matchedResults, unmatchedResults...)

	// Limit results
	if len(response.Results) > 100 {
		response.Results = response.Results[:100]
	}

	return response
}
