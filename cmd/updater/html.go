package main

import (
	"html"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func formatHTMLForLogging(htmlContent string, statusCode int) string {
	// EUC-JPからUTF-8に変換
	decoder := japanese.EUCJP.NewDecoder()
	utf8Content, _, err := transform.String(decoder, htmlContent)
	if err != nil {
		// 変換に失敗した場合は元の文字列をそのまま使用
		utf8Content = htmlContent
	}

	// HTMLエンティティをデコード
	decoded := html.UnescapeString(utf8Content)

	// 200レスポンスの場合はmain要素を抽出
	if statusCode == 200 {
		return extractMainContent(decoded)
	}

	// その他のレスポンスコードの場合は全体を返す
	return decoded
}

func extractMainContent(htmlContent string) string {
	// HTMLをパース
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		// パースに失敗した場合は元のHTMLを返す
		return htmlContent
	}

	// main要素内の.contentsクラスのdiv要素を検索
	var texts []string
	doc.Find("main .contents").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if text != "" {
			texts = append(texts, text)
		}
	})

	// .contentsクラスが見つからない場合はmain要素全体を使用
	if len(texts) == 0 {
		doc.Find("main").Each(func(i int, s *goquery.Selection) {
			text := s.Text()
			if text != "" {
				texts = append(texts, text)
			}
		})
	}

	// main要素も見つからない場合は元のHTMLを返す
	if len(texts) == 0 {
		return htmlContent
	}

	// 全てのテキストを結合
	combinedText := strings.Join(texts, "\n\n")

	// 連続する空白や改行を整理
	whitespaceRegex := regexp.MustCompile(`\s+`)
	combinedText = whitespaceRegex.ReplaceAllString(combinedText, " ")

	// 前後の空白を削除
	combinedText = strings.TrimSpace(combinedText)

	return combinedText
}
