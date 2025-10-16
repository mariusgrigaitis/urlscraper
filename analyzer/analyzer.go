package analyzer

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// PageAnalysis contains all extracted information about a web page
type PageAnalysis struct {
	URL              string
	Title            string
	HTMLVersion      string
	Headings         map[int]int // h1-h6 counts
	InternalLinks    int
	ExternalLinks    int
	InaccessibleLinks int
	HasLoginForm     bool
	Error            string
	StatusCode       int
}

// AnalyzeURL fetches and analyzes a web page
func AnalyzeURL(urlStr string) *PageAnalysis {
	analysis := &PageAnalysis{
		URL:      urlStr,
		Headings: make(map[int]int),
	}

	// Validate and normalize URL
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		urlStr = "https://" + urlStr
	}

	// Fetch the page with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(urlStr)
	if err != nil {
		analysis.Error = fmt.Sprintf("Failed to fetch URL: %v", err)
		analysis.StatusCode = 0
		return analysis
	}
	defer resp.Body.Close()

	analysis.StatusCode = resp.StatusCode

	// Check for HTTP error status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		analysis.Error = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		return analysis
	}

	// Read response body with size limit to prevent DoS
	limitedBody := io.LimitReader(resp.Body, 10*1024*1024) // 10MB limit
	body, err := io.ReadAll(limitedBody)
	if err != nil {
		analysis.Error = fmt.Sprintf("Failed to read response: %v", err)
		return analysis
	}

	// Parse HTML
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		analysis.Error = fmt.Sprintf("Failed to parse HTML: %v", err)
		return analysis
	}

	// Extract information
	analysis.HTMLVersion = detectHTMLVersion(string(body))
	analysis.Title = extractTitle(doc)
	analysis.Headings = countHeadings(doc)
	analysis.InternalLinks, analysis.ExternalLinks, analysis.InaccessibleLinks = analyzeLinks(doc, urlStr)
	analysis.HasLoginForm = detectLoginForm(doc)

	return analysis
}

// detectHTMLVersion extracts the HTML version from DOCTYPE
func detectHTMLVersion(html string) string {
	htmlLower := strings.ToLower(html)

	if strings.Contains(htmlLower, "<!doctype html>") {
		return "HTML5"
	}
	if strings.Contains(htmlLower, `public "-//w3c//dtd html 4.01`) {
		return "HTML 4.01"
	}
	if strings.Contains(htmlLower, `public "-//w3c//dtd html 4.0`) {
		return "HTML 4.0"
	}
	if strings.Contains(htmlLower, `public "-//w3c//dtd xhtml`) {
		return "XHTML"
	}

	return "Unknown"
}

// extractTitle gets the page title
func extractTitle(doc *html.Node) string {
	return traverseNode(doc, func(n *html.Node) string {
		if n.Type == html.ElementNode && n.Data == "title" {
			if n.FirstChild != nil {
				return strings.TrimSpace(n.FirstChild.Data)
			}
		}
		return ""
	})
}

// countHeadings counts h1-h6 headings
func countHeadings(doc *html.Node) map[int]int {
	headings := make(map[int]int)
	for i := 1; i <= 6; i++ {
		headings[i] = 0
	}

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if len(n.Data) == 2 && n.Data[0] == 'h' && n.Data[1] >= '1' && n.Data[1] <= '6' {
				level := int(n.Data[1] - '0')
				headings[level]++
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	return headings
}

// analyzeLinks counts internal, external, and inaccessible links
func analyzeLinks(doc *html.Node, pageURL string) (int, int, int) {
	internal := 0
	external := 0
	inaccessible := 0

	pageHost := extractHost(pageURL)

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "a" {
				href := getAttr(n, "href")
				if href == "" || strings.HasPrefix(href, "#") {
					inaccessible++
				} else if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
					linkHost := extractHost(href)
					if linkHost == pageHost {
						internal++
					} else {
						external++
					}
				} else if strings.HasPrefix(href, "/") || strings.HasPrefix(href, "./") || !strings.Contains(href, "://") {
					internal++
				} else {
					external++
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	return internal, external, inaccessible
}

// detectLoginForm checks if the page contains a login form
func detectLoginForm(doc *html.Node) bool {
	var traverse func(*html.Node) bool
	traverse = func(n *html.Node) bool {
		if n.Type == html.ElementNode {
			if n.Data == "form" {
				// Look for password input in form
				var hasPassword bool
				var hasUsername bool

				var checkForm func(*html.Node)
				checkForm = func(node *html.Node) {
					if node.Type == html.ElementNode {
						if node.Data == "input" {
							typeAttr := getAttr(node, "type")
							nameAttr := getAttr(node, "name")
							idAttr := getAttr(node, "id")

							if strings.ToLower(typeAttr) == "password" {
								hasPassword = true
							}
							if strings.ToLower(typeAttr) == "text" || strings.ToLower(typeAttr) == "email" {
								hasUsername = true
							}
							if strings.Contains(strings.ToLower(nameAttr), "user") ||
								strings.Contains(strings.ToLower(nameAttr), "login") ||
								strings.Contains(strings.ToLower(nameAttr), "email") ||
								strings.Contains(strings.ToLower(idAttr), "user") ||
								strings.Contains(strings.ToLower(idAttr), "login") {
								hasUsername = true
							}
						}
					}
					for c := node.FirstChild; c != nil; c = c.NextSibling {
						checkForm(c)
					}
				}

				checkForm(n)
				if hasPassword && hasUsername {
					return true
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if traverse(c) {
				return true
			}
		}
		return false
	}

	return traverse(doc)
}

// extractHost extracts the hostname from a URL
func extractHost(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	return u.Host
}

// getAttr gets an attribute value from an HTML node
func getAttr(n *html.Node, name string) string {
	for _, attr := range n.Attr {
		if attr.Key == name {
			return attr.Val
		}
	}
	return ""
}

// traverseNode helper for simple node traversal
func traverseNode(n *html.Node, fn func(*html.Node) string) string {
	result := fn(n)
	if result != "" {
		return result
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result = traverseNode(c, fn)
		if result != "" {
			return result
		}
	}
	return ""
}
