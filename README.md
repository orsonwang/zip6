# 台灣郵遞區號查詢工具 (zip6)

台灣 3+3 六碼郵遞區號查詢工具，支援地址解析、投遞範圍匹配與英文地址翻譯。

**線上 Demo**: <https://zip6.orson.tw>

## 開發緣由

現有的郵遞區號查詢工具各有不足：

| 工具 | 問題 |
|------|------|
| [中華郵政官網](https://www.post.gov.tw/post/internet/Postal/index.jsp?ID=208) | 介面老舊、操作繁瑣、需要逐層選擇縣市區域 |
| 第三方查詢網站 | 資料更新不即時，常有過期資料 |
| 其他 App | 廣告多、功能受限、隱私疑慮 |

**乾脆自己寫一個。**

直接使用中華郵政官方資料（2025年4月版），確保資料正確且即時。

## 功能特色

- **智慧地址解析** - 輸入完整或部分地址，自動解析路名、巷弄、門牌
- **精準範圍匹配** - 支援單雙號、號碼範圍、巷弄範圍等投遞規則
- **英文地址翻譯** - 完全匹配時提供標準英文地址格式
- **離線使用** - 資料內建於程式，無需網路連線
- **跨平台支援** - Windows / macOS / Linux

## 資料來源

- 郵遞區號資料：中華郵政郵遞區號簿 2504A + 2504B（約 80,000 筆）
- 英文翻譯資料：中華郵政中英對照檔（街路 30,030 筆、鄉鎮市區 371 筆、村里文字巷 8,369 筆）

## 使用方式

### 桌面版 (Wails)

需要 WebView2 運行環境（Windows 10/11 內建）

```bash
# 開發模式
wails dev

# 建置
wails build
```

### 網頁版 (純 Go)

不依賴 WebView，適合伺服器部署或無 GUI 環境

```bash
# 建置
go build -ldflags="-s -w" -o zip6-web ./cmd/web/

# 執行（預設 port 8080）
./zip6-web

# 自訂 port
PORT=3000 ./zip6-web
```

### 跨平台編譯

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o build/linux/zip6-web ./cmd/web/

# macOS
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o build/darwin/zip6-web ./cmd/web/

# Windows
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o build/windows/zip6-web.exe ./cmd/web/
```

## 技術架構

- **後端**: Go + [Wails v2](https://wails.io/)（桌面版）/ 純 Go HTTP Server（網頁版）
- **前端**: Preact + Tailwind CSS + DaisyUI v4
- **資料庫**: SQLite（使用 `modernc.org/sqlite`，純 Go 實作，無需 CGO）

## 專案結構

```text
zip6/
├── cmd/web/              # 純網頁版
│   ├── main.go
│   └── static/
├── internal/
│   ├── address/          # 地址解析與翻譯
│   └── database/         # SQLite 操作
├── frontend/             # Wails 前端 (Preact)
├── scripts/              # 資料匯入腳本
├── app.go                # Wails 綁定
├── main.go               # Wails 入口
└── zipcode.db            # SQLite 資料庫
```

## 授權

Apache License 2.0
