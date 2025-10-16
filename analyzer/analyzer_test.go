package analyzer

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

// TestDetectHTMLVersion tests HTML version detection
func TestDetectHTMLVersion(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "HTML5",
			html:     "<!DOCTYPE html><html></html>",
			expected: "HTML5",
		},
		{
			name:     "HTML 4.01",
			html:     `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01//EN"><html></html>`,
			expected: "HTML 4.01",
		},
		{
			name:     "HTML 4.0",
			html:     `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.0//EN"><html></html>`,
			expected: "HTML 4.0",
		},
		{
			name:     "XHTML",
			html:     `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0//EN"><html></html>`,
			expected: "XHTML",
		},
		{
			name:     "Unknown",
			html:     "<html></html>",
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectHTMLVersion(tt.html)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestExtractTitle tests title extraction
func TestExtractTitle(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "Simple title",
			html:     "<html><head><title>My Page</title></head></html>",
			expected: "My Page",
		},
		{
			name:     "Title with spaces",
			html:     "<html><head><title>  Spaced Title  </title></head></html>",
			expected: "Spaced Title",
		},
		{
			name:     "No title",
			html:     "<html><head></head></html>",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := parseHTML(tt.html)
			if err != nil {
				t.Fatalf("failed to parse HTML: %v", err)
			}
			result := extractTitle(doc)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestCountHeadings tests heading count extraction
func TestCountHeadings(t *testing.T) {
	html := `
		<html>
		<body>
			<h1>Main</h1>
			<h2>Sub</h2>
			<h2>Sub2</h2>
			<h3>SubSub</h3>
		</body>
		</html>
	`

	doc, err := parseHTML(html)
	if err != nil {
		t.Fatalf("failed to parse HTML: %v", err)
	}

	result := countHeadings(doc)

	if result[1] != 1 {
		t.Errorf("h1 count: got %d, want 1", result[1])
	}
	if result[2] != 2 {
		t.Errorf("h2 count: got %d, want 2", result[2])
	}
	if result[3] != 1 {
		t.Errorf("h3 count: got %d, want 1", result[3])
	}
	if result[4] != 0 {
		t.Errorf("h4 count: got %d, want 0", result[4])
	}
}

// TestAnalyzeLinks tests link counting and categorization
func TestAnalyzeLinks(t *testing.T) {
	tests := []struct {
		name                string
		html                string
		pageURL             string
		expectedInternal    int
		expectedExternal    int
		expectedInaccessible int
	}{
		{
			name: "Mix of links",
			html: `
				<html>
				<body>
					<a href="/page">Internal</a>
					<a href="https://other.com/page">External</a>
					<a href="#">Anchor</a>
					<a href="">Empty</a>
					<a href="/another">Another internal</a>
				</body>
				</html>
			`,
			pageURL:              "https://example.com",
			expectedInternal:     2,
			expectedExternal:     1,
			expectedInaccessible: 2,
		},
		{
			name: "External link same domain but different scheme",
			html: `
				<html>
				<body>
					<a href="http://example.com/page">Same domain</a>
				</body>
				</html>
			`,
			pageURL:              "https://example.com",
			expectedInternal:     1,
			expectedExternal:     0,
			expectedInaccessible: 0,
		},
		{
			name: "Relative links",
			html: `
				<html>
				<body>
					<a href="./page">Relative</a>
					<a href="../page">Parent</a>
				</body>
				</html>
			`,
			pageURL:              "https://example.com",
			expectedInternal:     2,
			expectedExternal:     0,
			expectedInaccessible: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := parseHTML(tt.html)
			if err != nil {
				t.Fatalf("failed to parse HTML: %v", err)
			}

			internal, external, inaccessible := analyzeLinks(doc, tt.pageURL)

			if internal != tt.expectedInternal {
				t.Errorf("internal links: got %d, want %d", internal, tt.expectedInternal)
			}
			if external != tt.expectedExternal {
				t.Errorf("external links: got %d, want %d", external, tt.expectedExternal)
			}
			if inaccessible != tt.expectedInaccessible {
				t.Errorf("inaccessible links: got %d, want %d", inaccessible, tt.expectedInaccessible)
			}
		})
	}
}

// TestDetectLoginForm tests login form detection
func TestDetectLoginForm(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected bool
	}{
		{
			name: "Simple login form",
			html: `
				<html>
				<body>
					<form>
						<input type="text" name="username">
						<input type="password" name="password">
						<button>Login</button>
					</form>
				</body>
				</html>
			`,
			expected: true,
		},
		{
			name: "Login form with email",
			html: `
				<html>
				<body>
					<form>
						<input type="email" name="email">
						<input type="password" name="password">
					</form>
				</body>
				</html>
			`,
			expected: true,
		},
		{
			name: "Form without password",
			html: `
				<html>
				<body>
					<form>
						<input type="text" name="search">
						<button>Search</button>
					</form>
				</body>
				</html>
			`,
			expected: false,
		},
		{
			name: "No forms",
			html: `
				<html>
				<body>
					<p>Just text</p>
				</body>
				</html>
			`,
			expected: false,
		},
		{
			name: "Form with id containing login",
			html: `
				<html>
				<body>
					<form>
						<input type="text" id="login-username">
						<input type="password" id="login-password">
					</form>
				</body>
				</html>
			`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := parseHTML(tt.html)
			if err != nil {
				t.Fatalf("failed to parse HTML: %v", err)
			}

			result := detectLoginForm(doc)
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestAnalyzeURLWithMockServer tests the full analyze function with a mock server
func TestAnalyzeURLWithMockServer(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		htmlContent    string
		shouldError    bool
		expectedTitle  string
	}{
		{
			name:        "Successful analysis",
			statusCode:  200,
			htmlContent: "<!DOCTYPE html><html><head><title>Test Page</title></head></html>",
			shouldError: false,
			expectedTitle: "Test Page",
		},
		{
			name:        "404 error",
			statusCode:  404,
			htmlContent: "<!DOCTYPE html><html><head><title>Not Found</title></head></html>",
			shouldError: true,
		},
		{
			name:        "500 error",
			statusCode:  500,
			htmlContent: "<!DOCTYPE html><html><head><title>Server Error</title></head></html>",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.htmlContent))
			}))
			defer server.Close()

			// Analyze the mock server URL
			result := AnalyzeURL(server.URL)

			if tt.shouldError {
				if result.Error == "" {
					t.Errorf("expected error, but got none")
				}
			} else {
				if result.Error != "" {
					t.Errorf("unexpected error: %s", result.Error)
				}
				if result.Title != tt.expectedTitle {
					t.Errorf("title: got %q, want %q", result.Title, tt.expectedTitle)
				}
			}
		})
	}
}

// TestAnalyzeURLWithInvalidURL tests handling of invalid URLs
func TestAnalyzeURLWithInvalidURL(t *testing.T) {
	// This test uses a deliberately non-existent domain
	result := AnalyzeURL("https://definitely-invalid-domain-that-does-not-exist-12345.com")

	if result.Error == "" {
		t.Errorf("expected error for invalid URL, but got none")
	}

	if result.StatusCode != 0 {
		t.Errorf("expected status code 0 for connection error, got %d", result.StatusCode)
	}
}

// TestCompletePageAnalysis tests a comprehensive page analysis
func TestCompletePageAnalysis(t *testing.T) {
	htmlContent := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Test Page</title>
		</head>
		<body>
			<h1>Main Title</h1>
			<p>Some content</p>

			<h2>Section 1</h2>
			<p>Content 1</p>
			<a href="/">Home</a>
			<a href="https://external.com">External</a>
			<a href="#">Anchor</a>

			<h2>Section 2</h2>
			<form>
				<input type="text" id="username">
				<input type="password" id="password">
				<button>Login</button>
			</form>
			<a href="/about">About</a>
		</body>
		</html>
	`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(htmlContent))
	}))
	defer server.Close()

	result := AnalyzeURL(server.URL)

	// Verify all expected values
	if result.Error != "" {
		t.Errorf("unexpected error: %s", result.Error)
	}

	if result.Title != "Test Page" {
		t.Errorf("title: got %q, want %q", result.Title, "Test Page")
	}

	if result.HTMLVersion != "HTML5" {
		t.Errorf("HTML version: got %q, want %q", result.HTMLVersion, "HTML5")
	}

	if result.Headings[1] != 1 {
		t.Errorf("h1 count: got %d, want 1", result.Headings[1])
	}

	if result.Headings[2] != 2 {
		t.Errorf("h2 count: got %d, want 2", result.Headings[2])
	}

	if result.InternalLinks != 2 {
		t.Errorf("internal links: got %d, want 2", result.InternalLinks)
	}

	if result.ExternalLinks != 1 {
		t.Errorf("external links: got %d, want 1", result.ExternalLinks)
	}

	if result.InaccessibleLinks != 1 {
		t.Errorf("inaccessible links: got %d, want 1", result.InaccessibleLinks)
	}

	if !result.HasLoginForm {
		t.Errorf("login form detection: got %v, want %v", result.HasLoginForm, true)
	}
}

// Helper function to parse HTML for testing
func parseHTML(htmlStr string) (*html.Node, error) {
	return html.Parse(strings.NewReader(htmlStr))
}

// Benchmark tests
func BenchmarkAnalyzeURL(b *testing.B) {
	htmlContent := `
		<!DOCTYPE html>
		<html>
		<head><title>Benchmark Page</title></head>
		<body>
			<h1>Title</h1>
			<a href="/">Home</a>
			<form><input type="password"></form>
		</body>
		</html>
	`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(htmlContent))
	}))
	defer server.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AnalyzeURL(server.URL)
	}
}

