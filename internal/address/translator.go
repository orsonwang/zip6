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

	// Street name (translate first to check if section is already included)
	var streetEn string
	if parsed.Street != "" {
		streetEn = t.translateStreet(parsed.Street, parsed.Section)
	}

	// Section (e.g., Sec. 1) - only add if not already in street translation
	if parsed.Section != "" && !strings.Contains(streetEn, "Sec.") {
		secNum := t.extractSectionNumber(parsed.Section)
		if secNum > 0 {
			parts = append(parts, fmt.Sprintf("Sec. %d", secNum))
		}
	}

	// Street name
	if streetEn != "" {
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
func (t *Translator) translateStreet(chinese string, section string) string {
	// Try exact match first
	english, err := t.db.GetStreetEnglish(chinese)
	if err == nil && english != "" {
		return english
	}

	// Try with section appended (e.g., "基隆路" + "二段" = "基隆路二段")
	// Database may have entries like "基隆路一段", "基隆路二段" etc.
	if section != "" {
		streetWithSection := chinese + section
		english, err := t.db.GetStreetEnglish(streetWithSection)
		if err == nil && english != "" {
			return english
		}
	}

	// Try without section (一段, 二段, etc.)
	// Street name in parsed result may include section like "基隆路二段"
	// but database only has "基隆路"
	sectionPattern := regexp.MustCompile(`([一二三四五六七八九十]+段)$`)
	if sectionPattern.MatchString(chinese) {
		baseStreet := sectionPattern.ReplaceAllString(chinese, "")
		english, err := t.db.GetStreetEnglish(baseStreet)
		if err == nil && english != "" {
			return english
		}
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

// HanyuToTongyong converts Hanyu Pinyin to Tongyong Pinyin
// Main differences:
// - zh → jh (e.g., Zhongshan → Jhongshan)
// - x → s (e.g., Xinyi → Sinyi)
// - q → c (e.g., Qixian → Cisian)
// - c → ts (e.g., Cixin → Tsisin)
// - shi → shih, chi → chih, zhi → jhih, ri → rih
// - si → sih, zi → zih, ci → tsih
// - iu → iou (e.g., Liu → Liou)
// - Special ending conversions
func HanyuToTongyong(hanyu string) string {
	if hanyu == "" {
		return ""
	}

	result := hanyu

	// Skip phrases that should not be converted (must check before word splitting)
	skipPhrases := map[string]string{
		"Taiwan (R.O.C.)": "___TAIWAN_ROC___",
	}
	for phrase, placeholder := range skipPhrases {
		result = strings.ReplaceAll(result, phrase, placeholder)
	}

	// Process word by word to handle capitalization
	words := strings.Split(result, " ")
	for i, word := range words {
		// Skip placeholder words
		isPlaceholder := false
		for _, placeholder := range skipPhrases {
			if word == placeholder {
				isPlaceholder = true
				break
			}
		}
		if isPlaceholder {
			continue
		}
		words[i] = convertWordToTongyong(word)
	}

	result = strings.Join(words, " ")

	// Restore skipped phrases
	for phrase, placeholder := range skipPhrases {
		result = strings.ReplaceAll(result, placeholder, phrase)
	}

	return result
}

// convertWordToTongyong converts a single word from Hanyu to Tongyong Pinyin
func convertWordToTongyong(word string) string {
	if word == "" {
		return ""
	}

	// Skip non-pinyin words (numbers, abbreviations, etc.)
	if !containsLetter(word) {
		return word
	}

	// Handle words with punctuation (e.g., "Rd.", "St.", "(R.O.C.)")
	prefix := ""
	suffix := ""
	for strings.HasPrefix(word, "(") {
		prefix = prefix + word[:1]
		word = word[1:]
	}
	for strings.HasSuffix(word, ".") || strings.HasSuffix(word, ",") || strings.HasSuffix(word, ")") {
		suffix = word[len(word)-1:] + suffix
		word = word[:len(word)-1]
	}

	// Skip common English words that shouldn't be converted
	skipWords := map[string]bool{
		"Rd": true, "St": true, "Ln": true, "Aly": true, "Sec": true,
		"No": true, "Rm": true, "F": true, "Dist": true, "City": true,
		"Taiwan": true, "R": true, "O": true, "C": true, "N": true,
		"S": true, "E": true, "W": true, "and": true, "the": true,
		"New": true, "Vil": true, "Village": true, "Township": true,
		"ROC": true,
	}
	if skipWords[word] {
		return prefix + word + suffix
	}

	// Check if first letter is uppercase
	isCapitalized := len(word) > 0 && word[0] >= 'A' && word[0] <= 'Z'

	// Convert to lowercase for processing
	lower := strings.ToLower(word)

	// Apply Tongyong Pinyin conversions (order matters!)
	// Handle special syllable endings first
	conversions := []struct {
		from string
		to   string
	}{
		// Special syllable patterns (longer patterns first)
		{"zhi", "jhih"},
		{"chi", "chih"},
		{"shi", "shih"},
		{"ri", "rih"},
		{"zi", "zih"},
		{"ci", "tsih"},
		{"si", "sih"},

		// Consonant clusters
		{"zh", "jh"},

		// iu → iou (but not after q/j/x which become c/j/s)
		{"niu", "niou"},
		{"liu", "liou"},
		{"diu", "diou"},
		{"miu", "miou"},

		// x → s (before vowels)
		{"xia", "sia"},
		{"xian", "sian"},
		{"xiang", "siang"},
		{"xiao", "siao"},
		{"xie", "sie"},
		{"xin", "sin"},
		{"xing", "sing"},
		{"xiong", "siong"},
		{"xiu", "siou"},
		{"xu", "syu"},
		{"xuan", "syuan"},
		{"xue", "syue"},
		{"xun", "syun"},

		// q → c (before vowels)
		{"qia", "cia"},
		{"qian", "cian"},
		{"qiang", "ciang"},
		{"qiao", "ciao"},
		{"qie", "cie"},
		{"qin", "cin"},
		{"qing", "cing"},
		{"qiong", "ciong"},
		{"qiu", "ciou"},
		{"qu", "cyu"},
		{"quan", "cyuan"},
		{"que", "cyue"},
		{"qun", "cyun"},

		// c → ts (before vowels, but not in ch)
		{"cai", "tsai"},
		{"can", "tsan"},
		{"cang", "tsang"},
		{"cao", "tsao"},
		{"ce", "tse"},
		{"cen", "tsen"},
		{"ceng", "tseng"},
		{"cou", "tsou"},
		{"cu", "tsu"},
		{"cuan", "tsuan"},
		{"cui", "tsuei"},
		{"cun", "tsun"},
		{"cuo", "tsuo"},

		// ü related (after j, q, x, y the u is actually ü)
		{"ju", "jyu"},
		{"juan", "jyuan"},
		{"jue", "jyue"},
		{"jun", "jyun"},
		{"yu", "yu"}, // yu stays the same
		{"yue", "yue"},
		{"yuan", "yuan"},
		{"yun", "yun"},

		// nü, lü
		{"nv", "nyu"},
		{"lv", "lyu"},

		// ong after certain consonants
		// These generally stay the same in Tongyong
	}

	for _, conv := range conversions {
		lower = strings.ReplaceAll(lower, conv.from, conv.to)
	}

	// Restore capitalization
	if isCapitalized && len(lower) > 0 {
		lower = strings.ToUpper(string(lower[0])) + lower[1:]
	}

	return prefix + lower + suffix
}

// containsLetter checks if a string contains at least one letter
func containsLetter(s string) bool {
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			return true
		}
	}
	return false
}
