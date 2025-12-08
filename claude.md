# 台灣郵遞區號查詢工具 - 專案建置指示

## 專案概述

建立一個跨平台桌面應用程式，用於查詢台灣 3+3 郵遞區號資料。

## 技術棧

- **後端**: Go + Wails v2
- **前端**: Preact + Tailwind CSS + DaisyUI
- **資料庫**: SQLite（內嵌於應用程式）
- **資料來源**: 中華郵政郵遞區號簿 Excel 檔案（2504A + 2504B）

## 資料結構

來源 Excel 檔案包含以下欄位：
- 縣市
- 區域
- 郵遞區號（6碼）
- 街路或聚落名稱
- 投遞範圍
- 投遞局-區段名稱
- 大宗戶或不按址投遞註記

總資料筆數約 80,000 筆（A 檔 65,534 筆 + B 檔 14,327 筆）

---

## 步驟 1：初始化 Wails 專案

```bash
wails init -n zip6 -t preact
cd zip6
```

---

## 步驟 2：專案結構

```
zip6/
├── build/
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   ├── SearchBox.jsx      # 搜尋輸入元件
│   │   │   ├── ResultTable.jsx    # 結果表格元件
│   │   │   └── ThemeToggle.jsx    # 主題切換元件
│   │   ├── app.jsx                # 主應用程式
│   │   ├── app.css                # Tailwind 入口
│   │   └── main.jsx               # Preact 入口
│   ├── index.html
│   ├── package.json
│   ├── tailwind.config.js
│   └── postcss.config.js
├── internal/
│   └── database/
│       └── database.go            # SQLite 操作封裝
├── scripts/
│   └── import_excel.go            # Excel 匯入腳本（獨立執行）
├── app.go                         # Wails 綁定方法
├── main.go                        # 程式入口
├── zipcode.db                     # SQLite 資料庫（由匯入腳本產生）
├── wails.json
└── go.mod
```

---

## 步驟 3：Go 後端實作

### 3.1 go.mod 額外依賴

```bash
go get github.com/mattn/go-sqlite3
go get github.com/xuri/excelize/v2
```

### 3.2 internal/database/database.go

實作以下功能：
- `InitDB(dbPath string) error` - 初始化資料庫，建立 zipcode 表格
- `ImportFromExcel(excelPaths []string) error` - 從 Excel 匯入資料
- `Close()` - 關閉資料庫連線

資料表結構：
```sql
CREATE TABLE IF NOT EXISTS zipcode (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    city TEXT NOT NULL,           -- 縣市
    district TEXT NOT NULL,       -- 區域
    zipcode TEXT NOT NULL,        -- 郵遞區號
    street TEXT NOT NULL,         -- 街路或聚落名稱
    scope TEXT,                   -- 投遞範圍
    delivery_office TEXT,         -- 投遞局-區段名稱
    note TEXT                     -- 大宗戶或不按址投遞註記
);

-- 建立索引加速查詢
CREATE INDEX IF NOT EXISTS idx_city ON zipcode(city);
CREATE INDEX IF NOT EXISTS idx_district ON zipcode(district);
CREATE INDEX IF NOT EXISTS idx_zipcode ON zipcode(zipcode);
CREATE INDEX IF NOT EXISTS idx_street ON zipcode(street);
```

### 3.3 app.go

Wails 綁定結構與方法：

```go
type App struct {
    ctx context.Context
    db  *database.Database
}

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

type SearchResponse struct {
    Results    []ZipCodeResult `json:"results"`
    Total      int             `json:"total"`
    Truncated  bool            `json:"truncated"`
}
```

實作方法：

1. **Search(keyword string) SearchResponse**
   - 跨欄位模糊查詢（city, district, zipcode, street, scope）
   - 使用 SQL LIKE '%keyword%'
   - 限制回傳最多 500 筆，避免 UI 卡頓
   - 若超過 500 筆，設定 truncated = true

2. **GetCities() []string**
   - 回傳所有不重複的縣市列表（供下拉選單使用）

3. **GetDistricts(city string) []string**
   - 根據縣市回傳該縣市的所有區域

4. **SearchAdvanced(city, district, street string) SearchResponse**
   - 進階查詢，可指定縣市、區域、路名
   - 空字串表示不限制該欄位

### 3.4 main.go

```go
package main

import (
    "embed"
    "github.com/wailsapp/wails/v2"
    "github.com/wailsapp/wails/v2/pkg/options"
    "github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
    app := NewApp()

    err := wails.Run(&options.App{
        Title:  "郵遞區號查詢",
        Width:  1024,
        Height: 768,
        AssetServer: &assetserver.Options{
            Assets: assets,
        },
        BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
        OnStartup:        app.startup,
        OnShutdown:       app.shutdown,
        Bind: []interface{}{
            app,
        },
    })

    if err != nil {
        println("Error:", err.Error())
    }
}
```

### 3.5 scripts/import_excel.go

獨立的匯入腳本，可單獨執行：

```bash
go run scripts/import_excel.go -excel "path/to/2504A.xls,path/to/2504B.xls" -db "zipcode.db"
```

功能：
- 讀取一或多個 Excel 檔案
- 合併資料並寫入 SQLite
- 顯示匯入進度
- 處理重複資料（跳過或更新）

---

## 步驟 4：前端實作

### 4.1 安裝依賴

```bash
cd frontend
npm install -D tailwindcss postcss autoprefixer daisyui
```

### 4.2 tailwind.config.js

```javascript
/** @type {import('tailwindcss').Config} */
export default {
  content: [
    './index.html',
    './src/**/*.{js,jsx,ts,tsx}',
  ],
  theme: {
    extend: {},
  },
  plugins: [require('daisyui')],
  daisyui: {
    themes: ['light', 'dark', 'nord', 'cupcake'],
  },
}
```

### 4.3 postcss.config.js

```javascript
export default {
  plugins: {
    tailwindcss: {},
    autoprefixer: {},
  },
}
```

### 4.4 src/app.css

```css
@tailwind base;
@tailwind components;
@tailwind utilities;

/* 自訂樣式 */
.search-highlight {
  @apply bg-yellow-200 dark:bg-yellow-700 rounded px-0.5;
}
```

### 4.5 src/components/SearchBox.jsx

功能：
- 單一輸入框，支援即時搜尋（debounce 300ms）
- 顯示搜尋圖示
- 清除按鈕
- 搜尋中顯示 loading 狀態

Props：
- `onSearch: (keyword: string) => void`
- `loading: boolean`

### 4.6 src/components/ResultTable.jsx

功能：
- 顯示查詢結果表格
- 欄位：郵遞區號、縣市、區域、路名、投遞範圍
- 支援關鍵字高亮（將搜尋關鍵字以不同底色標示）
- 無結果時顯示提示訊息
- 結果被截斷時顯示提示

Props：
- `results: ZipCodeResult[]`
- `keyword: string`
- `truncated: boolean`
- `total: number`

樣式：
- 使用 DaisyUI 的 table、table-zebra
- 固定表頭，內容可捲動
- hover 效果

### 4.7 src/components/ThemeToggle.jsx

功能：
- 主題切換按鈕（亮色/暗色）
- 記住使用者偏好（localStorage）

### 4.8 src/app.jsx

主應用程式邏輯：

```jsx
import { useState, useCallback } from 'preact/hooks'
import { Search } from '../wailsjs/go/main/App'
import SearchBox from './components/SearchBox'
import ResultTable from './components/ResultTable'
import ThemeToggle from './components/ThemeToggle'

export function App() {
  const [results, setResults] = useState([])
  const [loading, setLoading] = useState(false)
  const [keyword, setKeyword] = useState('')
  const [truncated, setTruncated] = useState(false)
  const [total, setTotal] = useState(0)

  const handleSearch = useCallback(async (kw) => {
    setKeyword(kw)
    if (!kw.trim()) {
      setResults([])
      setTotal(0)
      setTruncated(false)
      return
    }
    
    setLoading(true)
    try {
      const response = await Search(kw)
      setResults(response.results || [])
      setTotal(response.total)
      setTruncated(response.truncated)
    } catch (err) {
      console.error('Search error:', err)
    } finally {
      setLoading(false)
    }
  }, [])

  return (
    <div class="min-h-screen bg-base-200 p-4">
      <div class="max-w-6xl mx-auto">
        {/* 標題列 */}
        <div class="flex justify-between items-center mb-4">
          <h1 class="text-2xl font-bold">台灣郵遞區號查詢</h1>
          <ThemeToggle />
        </div>
        
        {/* 搜尋區 */}
        <div class="card bg-base-100 shadow-xl mb-4">
          <div class="card-body">
            <SearchBox onSearch={handleSearch} loading={loading} />
            {total > 0 && (
              <div class="text-sm text-base-content/70 mt-2">
                找到 {total} 筆結果
                {truncated && '（僅顯示前 500 筆）'}
              </div>
            )}
          </div>
        </div>
        
        {/* 結果表格 */}
        <div class="card bg-base-100 shadow-xl">
          <div class="card-body p-0">
            <ResultTable 
              results={results} 
              keyword={keyword}
              truncated={truncated}
              total={total}
            />
          </div>
        </div>
      </div>
    </div>
  )
}
```

---

## 步驟 5：資料庫初始化流程

應用程式啟動時：

1. 檢查 `zipcode.db` 是否存在於執行檔同目錄
2. 若不存在，檢查是否有內嵌的資料庫（可選：用 embed 內嵌）
3. 開啟資料庫連線
4. 應用程式關閉時，關閉連線

建議做法：將 `zipcode.db` 用 `//go:embed` 內嵌，首次執行時解壓縮到使用者資料目錄。

---

## 步驟 6：建置與打包

### 開發模式

```bash
wails dev
```

### 正式建置

```bash
# Windows
wails build -platform windows/amd64

# macOS
wails build -platform darwin/universal

# Linux
wails build -platform linux/amd64
```

---

## 步驟 7：額外功能（可選）

1. **匯出功能** - 將查詢結果匯出為 CSV
2. **複製功能** - 點擊郵遞區號可複製到剪貼簿
3. **進階篩選** - 下拉選單選擇縣市、區域
4. **快捷鍵** - Ctrl+F 聚焦搜尋框、Esc 清除
5. **記錄查詢歷史** - 最近 10 筆查詢

---

## 注意事項

1. **中文編碼** - 確保 Excel 讀取時使用正確編碼
2. **SQLite 中文排序** - 可能需要設定 COLLATE
3. **CGO** - go-sqlite3 需要 CGO，Windows 需安裝 gcc（建議用 mingw-w64）
4. **首次查詢效能** - 可在啟動時預熱資料庫連線

---

## 驗收標準

- [ ] 可輸入關鍵字進行模糊查詢
- [ ] 查詢結果正確顯示於表格
- [ ] 關鍵字高亮顯示
- [ ] 支援跨欄位搜尋（輸入「中正」可找到縣市、區域、路名含此關鍵字的所有結果）
- [ ] 主題切換正常運作
- [ ] 可成功打包為單一執行檔
- [ ] 執行檔體積 < 20MB
