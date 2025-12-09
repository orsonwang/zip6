package address

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// ParsedAddress represents a parsed Taiwan address
type ParsedAddress struct {
	City         string // 縣市
	District     string // 區域
	Street       string // 路/街名
	Section      string // 段 (一段、二段...)
	Lane         int    // 巷
	Alley        int    // 弄
	Number       int    // 號
	NumberSuffix string // 之X (e.g., 之1, 之2)
	Floor        int    // 樓
	Room         int    // 室 (e.g., 501室)
	SubNumber    int    // 之X號 (deprecated, use NumberSuffix)
	IsOdd        bool   // 單號
	IsEven       bool   // 雙號
	Raw          string // 原始輸入
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

	// Extract street name (路, 街, 大道, 公路, 莊, 村, 坑, 寮, 厝, 埔, 坪, 灣, etc.)
	// First try common road suffixes
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

	// If no street found, try village/settlement names (莊, 村, 坑, 寮, 厝, 埔, 坪, 灣, 崙, 窩, etc.)
	if parsed.Street == "" {
		villagePattern := regexp.MustCompile(`([^\d\s]+(?:莊|村|坑|寮|厝|埔|坪|灣|崙|窩|嶺|園|圍|底|頂|腳|尾|頭|角|口|井|塘|洲|洋|湖|潭|溪|溝|港|澳|山|嶼|島))`)
		if match := villagePattern.FindStringSubmatch(addr); match != nil {
			parsed.Street = match[1]
		}
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

	// Extract sub-number (之X) - stored as NumberSuffix
	if match := subNumberPattern.FindStringSubmatch(addr); match != nil {
		parsed.SubNumber, _ = strconv.Atoi(match[1])
		parsed.NumberSuffix = "之" + match[1]
	}

	// Extract floor (樓)
	if match := floorPattern.FindStringSubmatch(addr); match != nil {
		parsed.Floor, _ = strconv.Atoi(match[1])
	}

	// Check for floor suffix like 3樓之1
	floorSubPattern := regexp.MustCompile(`(\d+)\s*樓之(\d+)`)
	if match := floorSubPattern.FindStringSubmatch(addr); match != nil {
		parsed.Floor, _ = strconv.Atoi(match[1])
		parsed.Room, _ = strconv.Atoi(match[2])
	}

	// Extract room (室) - e.g., 501室
	if match := roomPattern.FindStringSubmatch(addr); match != nil {
		parsed.Room, _ = strconv.Atoi(match[1])
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
		// No number specified, consider it a match for any scope
		// (user just searching by street name without specific number)
		return true
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

// Arabic to Chinese numeral mapping for section names (1段 → 一段)
var arabicToChineseSection = map[string]string{
	"1段": "一段", "2段": "二段", "3段": "三段", "4段": "四段", "5段": "五段",
	"6段": "六段", "7段": "七段", "8段": "八段", "9段": "九段", "10段": "十段",
}

// Chinese numeral to Arabic mapping for numbers (一 → 1, 二 → 2, etc.)
var chineseToArabic = map[rune]int{
	'零': 0, '〇': 0,
	'一': 1, '壹': 1,
	'二': 2, '貳': 2, '兩': 2,
	'三': 3, '參': 3,
	'四': 4, '肆': 4,
	'五': 5, '伍': 5,
	'六': 6, '陸': 6,
	'七': 7, '柒': 7,
	'八': 8, '捌': 8,
	'九': 9, '玖': 9,
	'十': 10, '拾': 10,
	'百': 100, '佰': 100,
}

// parseChineseNumber converts a Chinese number string to Arabic number
// Supports two formats:
// 1. Traditional: 一, 十, 十一, 二十, 二十一, 一百, 一百二十三, etc.
// 2. Consecutive: 二二六 (226), 一一 (11), 五六 (56), etc.
func parseChineseNumber(s string) (int, bool) {
	if len(s) == 0 {
		return 0, false
	}

	runes := []rune(s)

	// Check if it contains positional characters (十, 百, 拾, 佰)
	hasPositional := false
	for _, r := range runes {
		if r == '十' || r == '拾' || r == '百' || r == '佰' {
			hasPositional = true
			break
		}
	}

	// If no positional characters, treat as consecutive digits (二二六 = 226)
	if !hasPositional {
		return parseConsecutiveChineseNumber(runes)
	}

	// Traditional parsing with positional characters
	result := 0
	temp := 0
	hasNumber := false
	hundredPart := 0

	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if val, ok := chineseToArabic[r]; ok {
			hasNumber = true
			if r == '百' || r == '佰' {
				if temp == 0 {
					temp = 1
				}
				hundredPart = temp * 100
				temp = 0
			} else if r == '十' || r == '拾' {
				if temp == 0 {
					temp = 1 // 十 means 10, 十一 means 11
				}
				result += temp * 10
				temp = 0
			} else {
				temp = val
			}
		}
	}
	result += temp + hundredPart

	return result, hasNumber
}

// parseConsecutiveChineseNumber parses consecutive Chinese digits like 二二六 = 226
func parseConsecutiveChineseNumber(runes []rune) (int, bool) {
	result := 0
	hasNumber := false

	for _, r := range runes {
		if val, ok := chineseToArabic[r]; ok {
			// Only accept single digits (0-9) for consecutive format
			if val <= 9 {
				hasNumber = true
				result = result*10 + val
			}
		}
	}

	return result, hasNumber
}

// convertChineseNumbersInAddress converts Chinese numbers to Arabic in specific contexts
// e.g., 一號 → 1號, 十二巷 → 12巷, 一樓 → 1樓, 一樓之一 → 1樓之1
func convertChineseNumbersInAddress(addr string) string {
	// Patterns for Chinese numbers followed by address units
	// Order matters: longer patterns first
	patterns := []struct {
		suffix string
		regex  *regexp.Regexp
	}{
		{"樓之", regexp.MustCompile(`([零〇一二三四五六七八九十百壹貳參肆伍陸柒捌玖拾兩]+)樓之([零〇一二三四五六七八九十百壹貳參肆伍陸柒捌玖拾兩]+)`)},
		{"號之", regexp.MustCompile(`([零〇一二三四五六七八九十百壹貳參肆伍陸柒捌玖拾兩]+)號之([零〇一二三四五六七八九十百壹貳參肆伍陸柒捌玖拾兩]+)`)},
		{"號", regexp.MustCompile(`([零〇一二三四五六七八九十百壹貳參肆伍陸柒捌玖拾兩]+)號`)},
		{"巷", regexp.MustCompile(`([零〇一二三四五六七八九十百壹貳參肆伍陸柒捌玖拾兩]+)巷`)},
		{"弄", regexp.MustCompile(`([零〇一二三四五六七八九十百壹貳參肆伍陸柒捌玖拾兩]+)弄`)},
		{"樓", regexp.MustCompile(`([零〇一二三四五六七八九十百壹貳參肆伍陸柒捌玖拾兩]+)樓`)},
		{"室", regexp.MustCompile(`([零〇一二三四五六七八九十百壹貳參肆伍陸柒捌玖拾兩]+)室`)},
		{"之", regexp.MustCompile(`之([零〇一二三四五六七八九十百壹貳參肆伍陸柒捌玖拾兩]+)`)},
	}

	result := addr

	// Handle 樓之X pattern first (e.g., 三樓之一 → 3樓之1)
	floorSubPattern := patterns[0]
	result = floorSubPattern.regex.ReplaceAllStringFunc(result, func(match string) string {
		subMatches := floorSubPattern.regex.FindStringSubmatch(match)
		if len(subMatches) >= 3 {
			floor, ok1 := parseChineseNumber(subMatches[1])
			sub, ok2 := parseChineseNumber(subMatches[2])
			if ok1 && ok2 {
				return strconv.Itoa(floor) + "樓之" + strconv.Itoa(sub)
			}
		}
		return match
	})

	// Handle 號之X pattern (e.g., 十號之一 → 10號之1)
	numSubPattern := patterns[1]
	result = numSubPattern.regex.ReplaceAllStringFunc(result, func(match string) string {
		subMatches := numSubPattern.regex.FindStringSubmatch(match)
		if len(subMatches) >= 3 {
			num, ok1 := parseChineseNumber(subMatches[1])
			sub, ok2 := parseChineseNumber(subMatches[2])
			if ok1 && ok2 {
				return strconv.Itoa(num) + "號之" + strconv.Itoa(sub)
			}
		}
		return match
	})

	// Handle other patterns
	for _, p := range patterns[2:] {
		if p.suffix == "之" {
			// Special handling for standalone 之X
			result = p.regex.ReplaceAllStringFunc(result, func(match string) string {
				subMatches := p.regex.FindStringSubmatch(match)
				if len(subMatches) >= 2 {
					num, ok := parseChineseNumber(subMatches[1])
					if ok {
						return "之" + strconv.Itoa(num)
					}
				}
				return match
			})
		} else {
			result = p.regex.ReplaceAllStringFunc(result, func(match string) string {
				subMatches := p.regex.FindStringSubmatch(match)
				if len(subMatches) >= 2 {
					num, ok := parseChineseNumber(subMatches[1])
					if ok {
						return strconv.Itoa(num) + p.suffix
					}
				}
				return match
			})
		}
	}

	return result
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

	// Convert Arabic numerals to Chinese for section names (1段 → 一段)
	for arabic, chinese := range arabicToChineseSection {
		s = strings.ReplaceAll(s, arabic, chinese)
	}

	// Convert Chinese numerals to Arabic for numbers (一號 → 1號, 一樓之一 → 1樓之1)
	s = convertChineseNumbersInAddress(s)

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
