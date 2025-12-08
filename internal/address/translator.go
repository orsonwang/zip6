package address

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// TranslatorDB interface for database operations
type TranslatorDB interface {
	GetStreetEnglish(chinese string) (string, error)
	GetDistrictEnglish(zipcode3 string) (string, error)
	GetVillageLaneEnglish(chinese string) (string, error)
}

// Translator handles Chinese to English address translation
type Translator struct {
	db TranslatorDB
}

// NewTranslator creates a new Translator instance
func NewTranslator(db TranslatorDB) *Translator {
	return &Translator{db: db}
}

// TranslateAddress translates a parsed Chinese address to English format
// Taiwan address format: Floor, No., Lane, Alley, Section, Street, District, City, Zipcode, Taiwan (R.O.C.)
func (t *Translator) TranslateAddress(parsed *ParsedAddress, zipcode string) string {
	var parts []string

	// Floor (e.g., 3F, 3F-1)
	if parsed.Floor > 0 {
		floorStr := fmt.Sprintf("%dF", parsed.Floor)
		if parsed.Room > 0 {
			floorStr += fmt.Sprintf("-%d", parsed.Room)
		}
		parts = append(parts, floorStr)
	} else if parsed.Room > 0 {
		parts = append(parts, fmt.Sprintf("Rm. %d", parsed.Room))
	}

	// Number (e.g., No. 172)
	if parsed.Number > 0 {
		numStr := fmt.Sprintf("No. %d", parsed.Number)
		if parsed.NumberSuffix != "" {
			numStr += fmt.Sprintf("-%s", t.translateNumberSuffix(parsed.NumberSuffix))
		}
		parts = append(parts, numStr)
	}

	// Alley (e.g., Aly. 5)
	if parsed.Alley > 0 {
		alleyStr := fmt.Sprintf("Aly. %d", parsed.Alley)
		parts = append(parts, alleyStr)
	}

	// Lane (e.g., Ln. 172)
	if parsed.Lane > 0 {
		laneStr := fmt.Sprintf("Ln. %d", parsed.Lane)
		parts = append(parts, laneStr)
	}

	// Section (e.g., Sec. 1)
	if parsed.Section != "" {
		secNum := t.extractSectionNumber(parsed.Section)
		if secNum > 0 {
			parts = append(parts, fmt.Sprintf("Sec. %d", secNum))
		}
	}

	// Street name
	if parsed.Street != "" {
		streetEn := t.translateStreet(parsed.Street)
		parts = append(parts, streetEn)
	}

	// District
	if zipcode != "" && len(zipcode) >= 3 {
		zipcode3 := zipcode[:3]
		districtEn, err := t.db.GetDistrictEnglish(zipcode3)
		if err == nil && districtEn != "" {
			parts = append(parts, districtEn)
		}
	}

	// Zipcode
	if zipcode != "" {
		parts = append(parts, zipcode)
	}

	// Country
	parts = append(parts, "Taiwan (R.O.C.)")

	return strings.Join(parts, ", ")
}

// translateStreet translates street name from Chinese to English
func (t *Translator) translateStreet(chinese string) string {
	// Try exact match first
	english, err := t.db.GetStreetEnglish(chinese)
	if err == nil && english != "" {
		return english
	}

	// Try without suffixes (路, 街, 大道, etc.)
	suffixes := []string{"路", "街", "大道", "巷", "弄"}
	for _, suffix := range suffixes {
		if strings.HasSuffix(chinese, suffix) {
			base := strings.TrimSuffix(chinese, suffix)
			english, err := t.db.GetStreetEnglish(base + suffix)
			if err == nil && english != "" {
				return english
			}
		}
	}

	// Return romanized version as fallback
	return t.romanizeChinese(chinese)
}

// translateNumberSuffix translates number suffixes like 之1, 附1
func (t *Translator) translateNumberSuffix(suffix string) string {
	// 之1 -> 1, 附1 -> 1
	re := regexp.MustCompile(`[之附]?(\d+)`)
	matches := re.FindStringSubmatch(suffix)
	if len(matches) > 1 {
		return matches[1]
	}
	return suffix
}

// extractSectionNumber extracts the section number from Chinese section string
func (t *Translator) extractSectionNumber(section string) int {
	// 一段 -> 1, 二段 -> 2, etc.
	chineseNums := map[string]int{
		"一": 1, "二": 2, "三": 3, "四": 4, "五": 5,
		"六": 6, "七": 7, "八": 8, "九": 9, "十": 10,
	}

	for cn, num := range chineseNums {
		if strings.Contains(section, cn+"段") {
			return num
		}
	}

	// Try numeric format (1段, 2段)
	re := regexp.MustCompile(`(\d+)段`)
	matches := re.FindStringSubmatch(section)
	if len(matches) > 1 {
		num, _ := strconv.Atoi(matches[1])
		return num
	}

	return 0
}

// romanizeChinese provides basic romanization for Chinese text
// This is a simple fallback when no translation is available
func (t *Translator) romanizeChinese(chinese string) string {
	// Just return the Chinese as-is if no romanization available
	// A more complete implementation would use pinyin conversion
	return chinese
}

// TranslateFloorRoom translates floor and room numbers to English
// Examples: 3樓 -> 3F, 3樓之1 -> 3F-1, 501室 -> Rm. 501
func TranslateFloorRoom(floor, room int, roomSuffix string) string {
	var parts []string

	if floor > 0 {
		floorStr := fmt.Sprintf("%dF", floor)
		if room > 0 {
			floorStr += fmt.Sprintf("-%d", room)
		}
		parts = append(parts, floorStr)
	} else if room > 0 {
		parts = append(parts, fmt.Sprintf("Rm. %d", room))
	}

	return strings.Join(parts, ", ")
}
