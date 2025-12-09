package address

import (
	"testing"
)

func TestNormalizeAddress(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// ==========================================
		// 1. 阿拉伯數字段名轉中文 (1段 → 一段)
		// ==========================================
		{"Arabic section 1", "中正路1段", "中正路一段"},
		{"Arabic section 2", "中山路2段", "中山路二段"},
		{"Arabic section 3", "忠孝路3段", "忠孝路三段"},
		{"Arabic section 4", "民生東路4段", "民生東路四段"},
		{"Arabic section 5", "仁愛路5段", "仁愛路五段"},
		{"Arabic section 6", "信義路6段", "信義路六段"},
		{"Arabic section 7", "和平東路7段", "和平東路七段"},
		{"Arabic section 8", "南京東路8段", "南京東路八段"},
		{"Arabic section 9", "復興北路9段", "復興北路九段"},
		{"Arabic section 10", "忠孝東路10段", "忠孝東路十段"},
		// 段名在路名中間
		{"Section in middle", "忠孝東路1段100號", "忠孝東路一段100號"},

		// ==========================================
		// 2. 中文數字轉阿拉伯數字 (一號 → 1號)
		// ==========================================
		// 基本個位數
		{"Chinese number 一號", "中正路一號", "中正路1號"},
		{"Chinese number 二號", "中正路二號", "中正路2號"},
		{"Chinese number 三號", "中正路三號", "中正路3號"},
		{"Chinese number 四號", "中正路四號", "中正路4號"},
		{"Chinese number 五號", "中正路五號", "中正路5號"},
		{"Chinese number 六號", "中正路六號", "中正路6號"},
		{"Chinese number 七號", "中正路七號", "中正路7號"},
		{"Chinese number 八號", "中正路八號", "中正路8號"},
		{"Chinese number 九號", "中正路九號", "中正路9號"},

		// 十位數
		{"Chinese number 十號", "中正路十號", "中正路10號"},
		{"Chinese number 十一號", "中正路十一號", "中正路11號"},
		{"Chinese number 十二號", "中正路十二號", "中正路12號"},
		{"Chinese number 十九號", "中正路十九號", "中正路19號"},
		{"Chinese number 二十號", "中正路二十號", "中正路20號"},
		{"Chinese number 二十一號", "中正路二十一號", "中正路21號"},
		{"Chinese number 三十五號", "中正路三十五號", "中正路35號"},
		{"Chinese number 九十九號", "中正路九十九號", "中正路99號"},

		// 百位數
		{"Chinese number 一百號", "中正路一百號", "中正路100號"},
		{"Chinese number 一百零一號", "中正路一百零一號", "中正路101號"},
		{"Chinese number 一百二十三號", "中正路一百二十三號", "中正路123號"},
		// 連續中文數字 (如 二二六 = 226)
		{"Chinese consecutive 二二六號", "基隆路一段二二六號", "基隆路一段226號"},
		{"Chinese consecutive with Arabic section", "基隆路1段二二六號", "基隆路一段226號"},
		{"Chinese consecutive 一一號", "中正路一一號", "中正路11號"},
		{"Chinese consecutive 三三號", "中正路三三號", "中正路33號"},
		{"Chinese consecutive 五六號", "中正路五六號", "中正路56號"},
		{"Chinese consecutive 一二三號", "中正路一二三號", "中正路123號"},

		// ==========================================
		// 3. 巷弄轉換
		// ==========================================
		{"Chinese lane 一巷", "中正路一巷", "中正路1巷"},
		{"Chinese lane 十二巷", "中正路十二巷", "中正路12巷"},
		{"Chinese lane 二十巷", "中正路二十巷", "中正路20巷"},
		{"Chinese alley 一弄", "中正路1巷一弄", "中正路1巷1弄"},
		{"Chinese alley 三弄", "中正路一巷三弄", "中正路1巷3弄"},
		{"Chinese lane and alley", "中正路十二巷五弄", "中正路12巷5弄"},

		// ==========================================
		// 4. 樓層轉換
		// ==========================================
		{"Chinese floor 一樓", "一樓", "1樓"},
		{"Chinese floor 二樓", "二樓", "2樓"},
		{"Chinese floor 三樓", "三樓", "3樓"},
		{"Chinese floor 十樓", "十樓", "10樓"},
		{"Chinese floor 十一樓", "十一樓", "11樓"},
		{"Chinese floor 十二樓", "十二樓", "12樓"},
		{"Chinese floor 二十樓", "二十樓", "20樓"},
		{"Chinese floor 二十一樓", "二十一樓", "21樓"},

		// ==========================================
		// 5. 樓之X 轉換 (一樓之一 → 1樓之1)
		// ==========================================
		{"Chinese floor sub 一樓之一", "一樓之一", "1樓之1"},
		{"Chinese floor sub 三樓之一", "三樓之一", "3樓之1"},
		{"Chinese floor sub 三樓之二", "三樓之二", "3樓之2"},
		{"Chinese floor sub 十樓之二", "十樓之二", "10樓之2"},
		{"Chinese floor sub 一樓之十一", "一樓之十一", "1樓之11"},
		{"Chinese floor sub 十二樓之三", "十二樓之三", "12樓之3"},

		// ==========================================
		// 6. 號之X 轉換 (十號之一 → 10號之1)
		// ==========================================
		{"Chinese number sub 一號之一", "一號之一", "1號之1"},
		{"Chinese number sub 十號之一", "十號之一", "10號之1"},
		{"Chinese number sub 二十一號之三", "二十一號之三", "21號之3"},
		{"Chinese number sub 一百號之二", "一百號之二", "100號之2"},

		// ==========================================
		// 7. 室 轉換
		// ==========================================
		{"Chinese room 一室", "一室", "1室"},
		{"Chinese room 五室", "五室", "5室"},
		{"Chinese room 十室", "十室", "10室"},

		// ==========================================
		// 8. 大寫中文數字 (壹貳參...)
		// ==========================================
		{"Traditional 壹號", "壹號", "1號"},
		{"Traditional 貳號", "貳號", "2號"},
		{"Traditional 參號", "參號", "3號"},
		{"Traditional 肆號", "肆號", "4號"},
		{"Traditional 伍號", "伍號", "5號"},
		{"Traditional 陸號", "陸號", "6號"},
		{"Traditional 柒號", "柒號", "7號"},
		{"Traditional 捌號", "捌號", "8號"},
		{"Traditional 玖號", "玖號", "9號"},
		{"Traditional 拾號", "拾號", "10號"},
		{"Traditional 貳樓", "貳樓", "2樓"},
		{"Traditional 拾貳樓", "拾貳樓", "12樓"},

		// ==========================================
		// 9. 完整地址組合測試
		// ==========================================
		// 中文數字 + 阿拉伯數字段名
		{"Full address 1", "中正路1段一號三樓之一", "中正路一段1號3樓之1"},
		{"Full address 2", "忠孝東路2段十二巷三弄五號", "忠孝東路二段12巷3弄5號"},
		{"Full address 3", "民生東路3段一百號十樓", "民生東路三段100號10樓"},
		{"Full address 4", "仁愛路4段二十巷十號之一 五樓之二", "仁愛路四段20巷10號之1 5樓之2"},
		// 完整台灣地址
		{"Full Taiwan address", "台北市中正區中正路1段一號三樓", "臺北市中正區中正路一段1號3樓"},
		{"Full Taiwan address 2", "台中市西區民生路二段十五號", "臺中市西區民生路二段15號"},

		// ==========================================
		// 10. 全形轉半形
		// ==========================================
		{"Full-width numbers", "中正路１２３號", "中正路123號"},
		{"Full-width mixed", "中正路１段１２號", "中正路一段12號"},
		{"Full-width letters", "Ａ棟Ｂ室", "A棟B室"},

		// ==========================================
		// 11. 台 → 臺 轉換
		// ==========================================
		{"Taipei", "台北市中正區", "臺北市中正區"},
		{"Taichung", "台中市西區", "臺中市西區"},
		{"Tainan", "台南市東區", "臺南市東區"},
		{"Taitung", "台東縣", "臺東縣"},
		// 不應轉換的台字
		{"Keep 台 in street", "平台路", "平台路"},

		// ==========================================
		// 12. 邊界情況
		// ==========================================
		{"Empty string", "", ""},
		{"Only spaces", "   ", "   "},
		{"No conversion needed", "中正路123號5樓", "中正路123號5樓"},
		{"Mixed formats", "中正路1段123號", "中正路一段123號"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeAddress(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeAddress(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseChineseNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
		ok       bool
	}{
		// 個位數
		{"零", "零", 0, true},
		{"〇", "〇", 0, true},
		{"一", "一", 1, true},
		{"二", "二", 2, true},
		{"三", "三", 3, true},
		{"四", "四", 4, true},
		{"五", "五", 5, true},
		{"六", "六", 6, true},
		{"七", "七", 7, true},
		{"八", "八", 8, true},
		{"九", "九", 9, true},

		// 十位數
		{"十", "十", 10, true},
		{"十一", "十一", 11, true},
		{"十二", "十二", 12, true},
		{"十九", "十九", 19, true},
		{"二十", "二十", 20, true},
		{"二十一", "二十一", 21, true},
		{"三十", "三十", 30, true},
		{"三十五", "三十五", 35, true},
		{"九十", "九十", 90, true},
		{"九十九", "九十九", 99, true},

		// 百位數
		{"一百", "一百", 100, true},
		{"一百零一", "一百零一", 101, true},
		{"一百一十", "一百一十", 110, true},
		{"一百二十三", "一百二十三", 123, true},
		{"二百", "二百", 200, true},
		{"三百四十五", "三百四十五", 345, true},
		{"九百九十九", "九百九十九", 999, true},

		// 大寫數字
		{"壹", "壹", 1, true},
		{"貳", "貳", 2, true},
		{"參", "參", 3, true},
		{"肆", "肆", 4, true},
		{"伍", "伍", 5, true},
		{"陸", "陸", 6, true},
		{"柒", "柒", 7, true},
		{"捌", "捌", 8, true},
		{"玖", "玖", 9, true},
		{"拾", "拾", 10, true},
		{"拾壹", "拾壹", 11, true},
		{"貳拾壹", "貳拾壹", 21, true},
		{"佰", "佰", 100, true},
		{"壹佰貳拾參", "壹佰貳拾參", 123, true},

		// 特殊寫法
		{"兩", "兩", 2, true},

		// 邊界情況
		{"empty", "", 0, false},
		{"non-chinese", "abc", 0, false},
		{"arabic", "123", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := parseChineseNumber(tt.input)
			if result != tt.expected || ok != tt.ok {
				t.Errorf("parseChineseNumber(%q) = (%d, %v), want (%d, %v)", tt.input, result, ok, tt.expected, tt.ok)
			}
		})
	}
}

func TestParseAddress(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ParsedAddress
	}{
		// ==========================================
		// 中文數字地址解析
		// ==========================================
		{
			name:  "Full address with Chinese numbers",
			input: "臺北市中正區中正路一段十二巷三弄五號三樓之一",
			expected: ParsedAddress{
				City:         "臺北市",
				District:     "中正區",
				Street:       "中正路",
				Section:      "一段",
				Lane:         12,
				Alley:        3,
				Number:       5,
				NumberSuffix: "之1", // 樓之一 的 之一 會被解析為 NumberSuffix
				Floor:        3,
				Room:         1,
			},
		},
		{
			name:  "Address with Arabic section converted",
			input: "臺北市中正區中正路1段12號",
			expected: ParsedAddress{
				City:     "臺北市",
				District: "中正區",
				Street:   "中正路",
				Section:  "一段",
				Number:   12,
			},
		},
		{
			name:  "Address with Chinese house number",
			input: "臺北市大安區忠孝東路三段一百號",
			expected: ParsedAddress{
				City:     "臺北市",
				District: "大安區",
				Street:   "忠孝東路",
				Section:  "三段",
				Number:   100,
			},
		},
		{
			name:  "Address with Chinese floor",
			input: "臺北市信義區信義路五段7號十二樓",
			expected: ParsedAddress{
				City:     "臺北市",
				District: "信義區",
				Street:   "信義路",
				Section:  "五段",
				Number:   7,
				Floor:    12,
			},
		},
		{
			name:  "Address with Chinese floor sub-number",
			input: "新北市板橋區中山路一段50號三樓之一",
			expected: ParsedAddress{
				City:         "新北市",
				District:     "板橋區",
				Street:       "中山路",
				Section:      "一段",
				Number:       50,
				NumberSuffix: "之1", // 樓之一 的 之一 會被解析為 NumberSuffix
				Floor:        3,
				Room:         1,
			},
		},

		// ==========================================
		// 台 → 臺 轉換後的解析
		// ==========================================
		{
			name:  "Taipei with 台",
			input: "台北市中正區重慶南路一段122號",
			expected: ParsedAddress{
				City:     "臺北市",
				District: "中正區",
				Street:   "重慶南路",
				Section:  "一段",
				Number:   122,
			},
		},
		{
			name:  "Taichung with 台",
			input: "台中市西區民生路100號",
			expected: ParsedAddress{
				City:     "臺中市",
				District: "西區",
				Street:   "民生路",
				Number:   100,
			},
		},

		// ==========================================
		// 各種號碼格式
		// ==========================================
		{
			name:  "Number with 之X suffix",
			input: "臺北市中正區中正路100號之1",
			expected: ParsedAddress{
				City:         "臺北市",
				District:     "中正區",
				Street:       "中正路",
				Number:       100,
				NumberSuffix: "之1",
			},
		},
		{
			name:  "Chinese number with 之X",
			input: "臺北市中正區中正路十號之一",
			expected: ParsedAddress{
				City:         "臺北市",
				District:     "中正區",
				Street:       "中正路",
				Number:       10,
				NumberSuffix: "之1",
			},
		},

		// ==========================================
		// 巷弄組合
		// ==========================================
		{
			name:  "Lane only",
			input: "臺北市中正區中正路100巷5號",
			expected: ParsedAddress{
				City:     "臺北市",
				District: "中正區",
				Street:   "中正路",
				Lane:     100,
				Number:   5,
			},
		},
		{
			name:  "Lane and alley",
			input: "臺北市中正區中正路50巷10弄3號",
			expected: ParsedAddress{
				City:     "臺北市",
				District: "中正區",
				Street:   "中正路",
				Lane:     50,
				Alley:    10,
				Number:   3,
			},
		},
		{
			name:  "Chinese lane and alley",
			input: "臺北市中正區中正路十二巷三弄五號",
			expected: ParsedAddress{
				City:     "臺北市",
				District: "中正區",
				Street:   "中正路",
				Lane:     12,
				Alley:    3,
				Number:   5,
			},
		},
		{
			name:  "Consecutive Chinese digits with Arabic section",
			input: "基隆路1段二二六號",
			expected: ParsedAddress{
				Street:  "基隆路",
				Section: "一段",
				Number:  226,
			},
		},

		// ==========================================
		// 全形數字轉換
		// ==========================================
		{
			name:  "Full-width numbers",
			input: "臺北市中正區中正路１２３號",
			expected: ParsedAddress{
				City:     "臺北市",
				District: "中正區",
				Street:   "中正路",
				Number:   123,
			},
		},

		// ==========================================
		// 複雜完整地址
		// ==========================================
		{
			name:  "Complete complex address",
			input: "台北市大安區忠孝東路4段二十巷十號之一 十二樓之三",
			expected: ParsedAddress{
				City:         "臺北市",
				District:     "大安區",
				Street:       "忠孝東路",
				Section:      "四段",
				Lane:         20,
				Number:       10,
				NumberSuffix: "之1",
				Floor:        12,
				Room:         3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseAddress(tt.input)
			if result.City != tt.expected.City {
				t.Errorf("City = %q, want %q", result.City, tt.expected.City)
			}
			if result.District != tt.expected.District {
				t.Errorf("District = %q, want %q", result.District, tt.expected.District)
			}
			if result.Section != tt.expected.Section {
				t.Errorf("Section = %q, want %q", result.Section, tt.expected.Section)
			}
			if result.Lane != tt.expected.Lane {
				t.Errorf("Lane = %d, want %d", result.Lane, tt.expected.Lane)
			}
			if result.Alley != tt.expected.Alley {
				t.Errorf("Alley = %d, want %d", result.Alley, tt.expected.Alley)
			}
			if result.Number != tt.expected.Number {
				t.Errorf("Number = %d, want %d", result.Number, tt.expected.Number)
			}
			if result.NumberSuffix != tt.expected.NumberSuffix {
				t.Errorf("NumberSuffix = %q, want %q", result.NumberSuffix, tt.expected.NumberSuffix)
			}
			if result.Floor != tt.expected.Floor {
				t.Errorf("Floor = %d, want %d", result.Floor, tt.expected.Floor)
			}
			if result.Room != tt.expected.Room {
				t.Errorf("Room = %d, want %d", result.Room, tt.expected.Room)
			}
		})
	}
}

// TestConvertChineseNumbersInAddress tests the convertChineseNumbersInAddress function
func TestConvertChineseNumbersInAddress(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// 單一轉換
		{"號", "一號", "1號"},
		{"巷", "十二巷", "12巷"},
		{"弄", "三弄", "3弄"},
		{"樓", "五樓", "5樓"},
		{"室", "八室", "8室"},

		// 複合轉換
		{"樓之X", "三樓之一", "3樓之1"},
		{"號之X", "十號之二", "10號之2"},

		// 多重轉換
		{"Multiple", "十二巷三弄五號", "12巷3弄5號"},
		{"Full", "一巷二弄三號四樓之五", "1巷2弄3號4樓之5"},

		// 不轉換的情況
		{"Arabic", "123號", "123號"},
		{"Mixed", "12巷三號", "12巷3號"},
		{"No suffix", "一二三", "一二三"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertChineseNumbersInAddress(tt.input)
			if result != tt.expected {
				t.Errorf("convertChineseNumbersInAddress(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestScopeMatching tests scope matching with normalized addresses
func TestScopeMatching(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		scope    string
		expected bool
	}{
		// 全範圍
		{"All scope", "中正路一段10號", "全", true},
		{"Empty scope", "中正路10號", "", true},

		// 單號雙號範圍
		{"Odd range match", "中正路11號", "單1號至99號", true},
		{"Odd range no match", "中正路10號", "單1號至99號", false},
		{"Even range match", "中正路10號", "雙2號至100號", true},
		{"Even range no match", "中正路11號", "雙2號至100號", false},

		// 數字範圍
		{"Range match", "中正路50號", "1號至100號", true},
		{"Range no match", "中正路150號", "1號至100號", false},
		{"Range boundary start", "中正路1號", "1號至100號", true},
		{"Range boundary end", "中正路100號", "1號至100號", true},

		// 單一號碼
		{"Single match", "中正路50號", "50號", true},
		{"Single no match", "中正路51號", "50號", false},

		// 以上/以下
		{"Above match", "中正路150號", "100號以上", true},
		{"Above no match", "中正路99號", "100號以上", false},
		{"Below match", "中正路50號", "100號以下", true},
		{"Below no match", "中正路150號", "100號以下", false},

		// 無號碼地址 (只有路名，應匹配任何範圍)
		{"No number matches all", "中正路", "1號至100號", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := ParseAddress(tt.address)
			scope := ParseScope(tt.scope)
			result := MatchAddress(parsed, scope)
			if result != tt.expected {
				t.Errorf("MatchAddress(%q, %q) = %v, want %v", tt.address, tt.scope, result, tt.expected)
			}
		})
	}
}
