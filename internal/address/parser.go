package address

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// ParsedAddress represents a parsed Taiwan address
type ParsedAddress struct {
	City       string // 縣市
	District   string // 區域
	Street     string // 路/街名
	Section    string // 段 (一段、二段...)
	Lane       int    // 巷
	Alley      int    // 弄
	Number     int    // 號
	Floor      int    // 樓
	Room       string // 室
	SubNumber  int    // 之X號
	IsOdd      bool   // 單號
	IsEven     bool   // 雙號
	Raw        string // 原始輸入
}

// ScopeRange represents a parsed scope range from database
type ScopeRange struct {
	Type       string // "all", "odd", "even", "single", "range", "lane_range"
	Start      int    // 起始號碼
	End        int    // 結束號碼
	LaneStart  int    // 巷起始
	LaneEnd    int    // 巷結束
	FloorStart int    // 樓起始
	FloorEnd   int    // 樓結束
	Raw        string // 原始範圍文字
}

var (
	// Chinese number mapping
	chineseNumbers = map[rune]int{
		'零': 0, '一': 1, '二': 2, '三': 3, '四': 4,
		'五': 5, '六': 6, '七': 7, '八': 8, '九': 9,
		'十': 10, '百': 100,
	}

	// Section names
	sectionNames = []string{"一段", "二段", "三段", "四段", "五段", "六段", "七段", "八段", "九段", "十段"}

	// Regex patterns
	numberPattern    = regexp.MustCompile(`(\d+)`)
	lanePattern      = regexp.MustCompile(`(\d+)\s*巷`)
	alleyPattern     = regexp.MustCompile(`(\d+)\s*弄`)
	housePattern     = regexp.MustCompile(`(\d+)\s*號`)
	subNumberPattern = regexp.MustCompile(`之\s*(\d+)`)
	floorPattern     = regexp.MustCompile(`(\d+)\s*樓`)
	roomPattern      = regexp.MustCompile(`(\d+)\s*室`)

	// Scope patterns
	scopeAllPattern       = regexp.MustCompile(`^全$`)
	scopeOddPattern       = regexp.MustCompile(`單`)
	scopeEvenPattern      = regexp.MustCompile(`雙`)
	scopeSinglePattern    = regexp.MustCompile(`^(\d+)號?$`)
	scopeRangePattern     = regexp.MustCompile(`(\d+)\s*號?\s*至\s*(\d+)\s*號?`)
	scopeLanePattern      = regexp.MustCompile(`(\d+)\s*巷`)
	scopeLaneRangePattern = regexp.MustCompile(`(\d+)\s*巷?\s*至\s*(\d+)\s*巷`)
	scopeFloorPattern     = regexp.MustCompile(`(\d+)\s*樓`)
	scopeFloorAbove       = regexp.MustCompile(`(\d+)\s*樓以上`)
	scopeFloorBelow       = regexp.MustCompile(`(\d+)\s*樓以下`)
)

// ParseAddress parses a Taiwan address string
func ParseAddress(addr string) *ParsedAddress {
	parsed := &ParsedAddress{
		Raw: addr,
	}

	// Normalize the address
	addr = normalizeAddress(addr)

	// Extract city
	cities := []string{
		"臺北市", "台北市", "新北市", "桃園市", "臺中市", "台中市",
		"臺南市", "台南市", "高雄市", "基隆市", "新竹市", "嘉義市",
		"新竹縣", "苗栗縣", "彰化縣", "南投縣", "雲林縣", "嘉義縣",
		"屏東縣", "宜蘭縣", "花蓮縣", "臺東縣", "台東縣", "澎湖縣",
		"金門縣", "連江縣",
	}
	for _, city := range cities {
		if strings.Contains(addr, city) {
			parsed.City = city
			addr = strings.Replace(addr, city, "", 1)
			break
		}
	}

	// Extract district (ends with 區, 鄉, 鎮, 市)
	districtPattern := regexp.MustCompile(`^([^路街巷弄號]+[區鄉鎮市])`)
	if match := districtPattern.FindStringSubmatch(addr); match != nil {
		parsed.District = match[1]
		addr = strings.Replace(addr, match[1], "", 1)
	}

	// Extract section
	for _, sec := range sectionNames {
		if idx := strings.Index(addr, sec); idx != -1 {
			parsed.Section = sec
			break
		}
	}

	// Extract street name (路 or 街)
	streetPattern := regexp.MustCompile(`([^\d]+(?:路|街|大道|公路))`)
	if match := streetPattern.FindStringSubmatch(addr); match != nil {
		street := match[1]
		// Include section if present
		if parsed.Section != "" && !strings.Contains(street, parsed.Section) {
			// Find where section should be
			for _, sec := range sectionNames {
				if idx := strings.Index(addr, sec); idx != -1 {
					// Insert section into street name
					streetEnd := strings.Index(addr, street) + len(street)
					if idx > streetEnd {
						street = street + sec
					}
					break
				}
			}
		}
		parsed.Street = street
	}

	// Extract lane (巷)
	if match := lanePattern.FindStringSubmatch(addr); match != nil {
		parsed.Lane, _ = strconv.Atoi(match[1])
	}

	// Extract alley (弄)
	if match := alleyPattern.FindStringSubmatch(addr); match != nil {
		parsed.Alley, _ = strconv.Atoi(match[1])
	}

	// Extract house number (號)
	if match := housePattern.FindStringSubmatch(addr); match != nil {
		parsed.Number, _ = strconv.Atoi(match[1])
	}

	// Extract sub-number (之X)
	if match := subNumberPattern.FindStringSubmatch(addr); match != nil {
		parsed.SubNumber, _ = strconv.Atoi(match[1])
	}

	// Extract floor (樓)
	if match := floorPattern.FindStringSubmatch(addr); match != nil {
		parsed.Floor, _ = strconv.Atoi(match[1])
	}

	// Extract room (室)
	if match := roomPattern.FindStringSubmatch(addr); match != nil {
		parsed.Room = match[1]
	}

	// Determine odd/even
	if parsed.Number > 0 {
		parsed.IsOdd = parsed.Number%2 == 1
		parsed.IsEven = parsed.Number%2 == 0
	}

	return parsed
}

// ParseScope parses a scope string from the database
func ParseScope(scope string) *ScopeRange {
	sr := &ScopeRange{
		Raw: scope,
	}

	scope = strings.TrimSpace(scope)

	// Check for "全" (all)
	if scope == "全" || scope == "" {
		sr.Type = "all"
		return sr
	}

	// Check for odd/even
	isOdd := scopeOddPattern.MatchString(scope)
	isEven := scopeEvenPattern.MatchString(scope)

	// Check for lane range (e.g., "126號至 176巷")
	if match := scopeLaneRangePattern.FindStringSubmatch(scope); match != nil {
		sr.LaneStart, _ = strconv.Atoi(match[1])
		sr.LaneEnd, _ = strconv.Atoi(match[2])
		sr.Type = "lane_range"
		return sr
	}

	// Check for single lane
	if match := scopeLanePattern.FindStringSubmatch(scope); match != nil {
		sr.LaneStart, _ = strconv.Atoi(match[1])
		sr.LaneEnd = sr.LaneStart
		// Continue to check for number range within lane
	}

	// Check for number range (e.g., "126號至 176號")
	if match := scopeRangePattern.FindStringSubmatch(scope); match != nil {
		sr.Start, _ = strconv.Atoi(match[1])
		sr.End, _ = strconv.Atoi(match[2])
		if isOdd {
			sr.Type = "odd_range"
		} else if isEven {
			sr.Type = "even_range"
		} else {
			sr.Type = "range"
		}
		return sr
	}

	// Check for "以上" (and above) or "以下" (and below)
	if strings.Contains(scope, "以上") {
		if match := numberPattern.FindStringSubmatch(scope); match != nil {
			sr.Start, _ = strconv.Atoi(match[1])
			sr.End = 99999
			if isOdd {
				sr.Type = "odd_above"
			} else if isEven {
				sr.Type = "even_above"
			} else {
				sr.Type = "above"
			}
			return sr
		}
	}

	if strings.Contains(scope, "以下") {
		if match := numberPattern.FindStringSubmatch(scope); match != nil {
			sr.End, _ = strconv.Atoi(match[1])
			sr.Start = 1
			if isOdd {
				sr.Type = "odd_below"
			} else if isEven {
				sr.Type = "even_below"
			} else {
				sr.Type = "below"
			}
			return sr
		}
	}

	// Check for single number
	if match := scopeSinglePattern.FindStringSubmatch(scope); match != nil {
		sr.Start, _ = strconv.Atoi(match[1])
		sr.End = sr.Start
		sr.Type = "single"
		return sr
	}

	// Check floor patterns
	if match := scopeFloorAbove.FindStringSubmatch(scope); match != nil {
		sr.FloorStart, _ = strconv.Atoi(match[1])
		sr.FloorEnd = 999
	} else if match := scopeFloorBelow.FindStringSubmatch(scope); match != nil {
		sr.FloorStart = 1
		sr.FloorEnd, _ = strconv.Atoi(match[1])
	} else if match := scopeFloorPattern.FindStringSubmatch(scope); match != nil {
		sr.FloorStart, _ = strconv.Atoi(match[1])
		sr.FloorEnd = sr.FloorStart
	}

	// Default: try to extract any numbers
	if sr.Type == "" {
		numbers := numberPattern.FindAllStringSubmatch(scope, -1)
		if len(numbers) >= 2 {
			sr.Start, _ = strconv.Atoi(numbers[0][1])
			sr.End, _ = strconv.Atoi(numbers[1][1])
			if isOdd {
				sr.Type = "odd_range"
			} else if isEven {
				sr.Type = "even_range"
			} else {
				sr.Type = "range"
			}
		} else if len(numbers) == 1 {
			sr.Start, _ = strconv.Atoi(numbers[0][1])
			sr.End = sr.Start
			sr.Type = "single"
		} else {
			sr.Type = "all"
		}
	}

	return sr
}

// MatchAddress checks if an address matches a scope range
func MatchAddress(addr *ParsedAddress, scope *ScopeRange) bool {
	// Handle lane matching
	if addr.Lane > 0 && scope.LaneStart > 0 {
		if scope.Type == "lane_range" {
			if addr.Lane < scope.LaneStart || addr.Lane > scope.LaneEnd {
				return false
			}
		} else if addr.Lane != scope.LaneStart {
			return false
		}
	}

	// Handle floor matching
	if addr.Floor > 0 && scope.FloorStart > 0 {
		if addr.Floor < scope.FloorStart || addr.Floor > scope.FloorEnd {
			return false
		}
	}

	// Handle number matching
	num := addr.Number
	if num == 0 {
		// No number specified, consider it a match if scope is "all"
		return scope.Type == "all"
	}

	switch scope.Type {
	case "all":
		return true
	case "single":
		return num == scope.Start
	case "range":
		return num >= scope.Start && num <= scope.End
	case "odd_range", "odd_above", "odd_below":
		if num%2 != 1 {
			return false
		}
		return num >= scope.Start && num <= scope.End
	case "even_range", "even_above", "even_below":
		if num%2 != 0 {
			return false
		}
		return num >= scope.Start && num <= scope.End
	case "above":
		return num >= scope.Start
	case "below":
		return num <= scope.End
	case "lane_range":
		// Lane range without specific number match
		return true
	}

	return false
}

// normalizeAddress normalizes address string
func normalizeAddress(addr string) string {
	// Replace full-width characters with half-width
	var result strings.Builder
	for _, r := range addr {
		if r >= '０' && r <= '９' {
			result.WriteRune(r - '０' + '0')
		} else if r >= 'Ａ' && r <= 'Ｚ' {
			result.WriteRune(r - 'Ａ' + 'A')
		} else if r >= 'ａ' && r <= 'ｚ' {
			result.WriteRune(r - 'ａ' + 'a')
		} else if unicode.IsSpace(r) {
			result.WriteRune(' ')
		} else {
			result.WriteRune(r)
		}
	}

	// Replace 台 with 臺
	s := result.String()
	s = strings.ReplaceAll(s, "台北", "臺北")
	s = strings.ReplaceAll(s, "台中", "臺中")
	s = strings.ReplaceAll(s, "台南", "臺南")
	s = strings.ReplaceAll(s, "台東", "臺東")

	return s
}

// GetStreetSearchPattern returns a pattern for database search
func GetStreetSearchPattern(addr *ParsedAddress) string {
	if addr.Street == "" {
		return ""
	}

	// Build search pattern
	pattern := addr.Street
	if addr.Section != "" && !strings.Contains(pattern, addr.Section) {
		pattern += addr.Section
	}

	return pattern
}
