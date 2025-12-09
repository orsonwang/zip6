package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"zip6/internal/address"
	"zip6/internal/database"
)

// Version info - set at build time using ldflags
var (
	Version     = "dev"
	BuildDate   = "unknown"
	DataVersion = "2504" // 郵遞區號資料版本 (年月)
)

//go:embed static
var staticFiles embed.FS

var db *database.Database

// AddressSearchResult represents a single search result
type AddressSearchResult struct {
	Zipcode  string `json:"zipcode"`
	City     string `json:"city"`
	District string `json:"district"`
	Street   string `json:"street"`
	Scope    string `json:"scope"`
	Matched  bool   `json:"matched"`
}

// AddressSearchResponse represents the response from address search
type AddressSearchResponse struct {
	ParsedAddress   string                `json:"parsedAddress"`
	EnglishAddress  string                `json:"englishAddress"`
	TongyongAddress string                `json:"tongyongAddress"`
	Results         []AddressSearchResult `json:"results"`
	BestMatch       *AddressSearchResult  `json:"bestMatch"`
}

func main() {
	// Initialize database
	if err := initDatabase(); err != nil {
		log.Fatalf("Error initializing database: %v\n", err)
	}
	defer db.Close()

	// API routes
	http.HandleFunc("/api/search", handleSearch)

	// Serve static files with version injection for index.html
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Serve index.html with version info injected
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			indexData, err := fs.ReadFile(staticFiles, "static/index.html")
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			versionInfo := fmt.Sprintf("v%s (%s) | 資料版本: %s", Version, BuildDate, DataVersion)
			content := strings.Replace(string(indexData), "{{VERSION_INFO}}", versionInfo, 1)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(content))
			return
		}
		// Serve other static files
		http.FileServer(http.FS(staticFS)).ServeHTTP(w, r)
	})

	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	fmt.Printf("Server starting at http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	keyword := r.URL.Query().Get("q")
	if keyword == "" {
		json.NewEncoder(w).Encode(AddressSearchResponse{})
		return
	}

	// Parse address
	parsed := address.ParseAddress(keyword)

	// Build parsed description
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
		streetPattern = keyword
	}

	records, err := db.SearchByStreetAndDistrict(streetPattern, parsed.City, parsed.District, 1000)
	if err != nil {
		fmt.Printf("Search error: %v\n", err)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Match each record's scope against the parsed address
	var bestMatch *AddressSearchResult
	var bestMatchZipcode string
	var matchedResults []AddressSearchResult
	var unmatchedResults []AddressSearchResult

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

		if matched {
			if bestMatch == nil {
				bestMatch = &result
				bestMatchZipcode = r.Zipcode
			}
			matchedResults = append(matchedResults, result)
		} else {
			unmatchedResults = append(unmatchedResults, result)
		}
	}

	response.BestMatch = bestMatch
	response.Results = append(matchedResults, unmatchedResults...)

	// Generate English address if we have a best match
	if bestMatch != nil {
		translator := address.NewTranslator(db)
		response.EnglishAddress = translator.TranslateAddress(parsed, bestMatchZipcode)
		// Convert Hanyu Pinyin to Tongyong Pinyin
		response.TongyongAddress = address.HanyuToTongyong(response.EnglishAddress)
	}

	// Limit results
	if len(response.Results) > 100 {
		response.Results = response.Results[:100]
	}

	json.NewEncoder(w).Encode(response)
}

func initDatabase() error {
	// Get executable directory
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	execDir := filepath.Dir(execPath)

	// Try to find database in several locations
	dbPaths := []string{
		filepath.Join(execDir, "zipcode.db"),
		"zipcode.db",
		filepath.Join(".", "zipcode.db"),
		filepath.Join("..", "..", "zipcode.db"), // For development
	}

	var dbPath string
	for _, p := range dbPaths {
		if _, err := os.Stat(p); err == nil {
			dbPath = p
			break
		}
	}

	if dbPath == "" {
		return fmt.Errorf("zipcode.db not found")
	}

	db, err = database.InitDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	fmt.Printf("Database initialized at: %s\n", dbPath)
	return nil
}
