# issue2md

一個命令行和網頁工具，用於將 GitHub issue、discussion 或 pull request 轉換為 Markdown 格式檔案。

>此儲存庫中的大部分內容是由人工智能生成的!

## 命令行模式

### 安裝 issue2md 命令行工具

```
$ go install github.com/bigwhite/issue2md/cmd/issue2md@latest
```

### 將 Issue/Discussion/Pull Request 轉換為 Markdown

```
用法: issue2md [flags] url [markdown-file]
參數:
  url            要轉換的 GitHub issue、discussion 或 pull request 的 URL。
  markdown-file  (選擇性) 輸出的 Markdown 檔案。
標誌:
  -enable-reactions
    	在輸出中包含 reactions。
  -enable-user-links
    	在輸出中包含評論者的 profile 連結
```

## 網頁模式

### 安裝並運行 issue2md web

#### 基於 Docker 鏡像運行(推薦)

```
$docker run -d -p 8080:8080 bigwhite/issue2mdweb
```

#### 從原始碼構建安裝

```
$ git clone https://github.com/bigwhite/issue2md.git
$ make web
$ ./issue2mdweb
伺服器正在運行在 http://0.0.0.0:8080
```

### 將內容轉換為 Markdown

在瀏覽器中打開 localhost:8080：

![](./screen-snapshot.png)

輸入您想要轉換的 issue、discussion 或 pull request 的 URL，然後點擊「Convert」按鈕！