package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

type SearchResult struct {
	Index       int
	Title       string
	URL         string
	Description string
}

type MinimalSearchResult struct {
	Index int
	Title string
	URL   string
}

func extractResultURL(href string) string {
	u, err := url.Parse("https:" + href)
	if err != nil {
		return href
	}

	uddg := u.Query().Get("uddg")
	if uddg == "" {
		return href
	}

	decodedURL, err := url.QueryUnescape(uddg)
	if err != nil {
		return uddg
	}

	return decodedURL
}

func hasClass(n *html.Node, className string) bool {
	for _, attr := range n.Attr {
		if attr.Key != "class" {
			continue
		}

		for _, value := range strings.Fields(attr.Val) {
			if value == className {
				return true
			}
		}
	}
	return false
}

func getAttr(n *html.Node, key string) (string, bool) {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val, true
		}
	}
	return "", false
}

func nodeText(n *html.Node) string {
	var b strings.Builder
	var f func(*html.Node)
	f = func(cur *html.Node) {
		if cur == nil {
			return
		}
		if cur.Type == html.TextNode {
			b.WriteString(cur.Data)
		}
		for c := cur.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return b.String()
}

func firstChildElement(n *html.Node, tagName string) *html.Node {
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == tagName {
			return child
		}
	}
	return nil
}

func nextSiblingElement(n *html.Node, tagName string) *html.Node {
	for sibling := n.NextSibling; sibling != nil; sibling = sibling.NextSibling {
		if sibling.Type == html.ElementNode && sibling.Data == tagName {
			return sibling
		}
	}
	return nil
}

func findResultRows(n *html.Node) []*html.Node {
	var rows []*html.Node

	var f func(*html.Node)
	f = func(cur *html.Node) {
		if cur == nil {
			return
		}

		if cur.Type == html.ElementNode && cur.Data == "a" && hasClass(cur, "result-link") {
			row := cur.Parent
			for row != nil && !(row.Type == html.ElementNode && row.Data == "tr") {
				row = row.Parent
			}
			if row != nil {
				rows = append(rows, row)
			}
		}

		for c := cur.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(n)
	return rows
}

func ScrapeDuckDuckGo(query string, page int, limit int, region string, safeSearch int) ([]SearchResult, error) {
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36"

	escapedQuery := url.QueryEscape(query)
	requestURL := fmt.Sprintf("https://lite.duckduckgo.com/lite/?q=%s&kl=%s&kp=%d", escapedQuery, region, safeSearch)

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	var results []SearchResult
	index := 1

	for _, row := range findResultRows(doc) {
		if len(results) >= limit {
			break
		}

		leftCell := firstChildElement(row, "td")
		if leftCell == nil {
			continue
		}

		linkContainer := nextSiblingElement(leftCell, "td")
		if linkContainer == nil {
			continue
		}

		link := firstChildElement(linkContainer, "a")
		if link == nil || !hasClass(link, "result-link") {
			continue
		}

		title := strings.TrimSpace(nodeText(link))
		href, _ := getAttr(link, "href")
		actualURL := extractResultURL(href)

		description := ""
		snippetRow := nextSiblingElement(row, "tr")
		if snippetRow != nil {
			snippetLeft := firstChildElement(snippetRow, "td")
			if snippetLeft != nil {
				snippetCell := nextSiblingElement(snippetLeft, "td")
				if snippetCell != nil && hasClass(snippetCell, "result-snippet") {
					description = strings.TrimSpace(nodeText(snippetCell))
				}
			}
		}

		results = append(results, SearchResult{
			Index:       index,
			Title:       title,
			URL:         actualURL,
			Description: description,
		})
		index++
	}

	return results, nil
}
