# Page Insight Tool

A web application that analyzes web pages and extracts key insights about their structure and content. Built with Go backend and responsive HTML/CSS frontend.

## Features

- **URL Analysis**: Fetch and analyze any web page
- **HTML Version Detection**: Automatically detects the HTML version (HTML5, HTML 4.01, XHTML, etc.)
- **Page Title Extraction**: Extracts the page title
- **Heading Analysis**: Counts headings at each level (H1-H6)
- **Link Analysis**: Categorizes links as internal, external, or inaccessible
- **Login Form Detection**: Identifies if the page contains a login form
- **Error Handling**: Graceful error handling with HTTP status codes
- **Responsive Design**: Works on desktop and mobile devices

## Requirements

- Go 1.21 or higher
- No external dependencies beyond standard library and `golang.org/x/net`

## Build Instructions

### Prerequisites
Ensure you have Go installed:
```bash
go version
```

### Building

1. Clone the repository:
```bash
git clone <repository-url>
cd urlscraper
```

2. Download dependencies:
```bash
go mod download
```

3. Build the application:
```bash
go build -o urlscraper
```

### Running

Start the server:
```bash
./urlscraper
```

The application will be available at `http://localhost:8080`

## Usage

1. Open your browser and navigate to `http://localhost:8080`
2. Enter a URL in the input field (with or without `https://` protocol)
3. Click "Analyze Page"
4. View the results showing:
   - Page title
   - HTML version
   - Heading counts (H1-H6)
   - Internal/External/Inaccessible link counts
   - Login form detection status

## Project Structure

```
urlscraper/
├── main.go                 # Server setup and route handlers
├── go.mod                  # Go module definition
├── go.sum                  # Go module checksums
├── README.md              # This file
├── .gitignore             # Git ignore rules
├── analyzer/
│   ├── analyzer.go        # Core URL analysis logic
│   └── analyzer_test.go   # Comprehensive test suite
├── templates/
│   ├── index.html         # Home page with form
│   └── results.html       # Results display page
└── static/
    └── style.css          # Responsive styling
```

## Design Decisions

### 1. **HTML Parsing**
- Used `golang.org/x/net/html` for robust HTML parsing
- Supports malformed HTML gracefully
- Efficient traversal of DOM tree

### 2. **Link Categorization**
- **Internal Links**: Same domain as the analyzed page
- **External Links**: Different domain
- **Inaccessible Links**: Anchors (#), empty hrefs, or invalid URLs

### 3. **Login Form Detection**
- Detects presence of both password input field AND either:
  - Username/email input field
  - Input with name/id containing "user", "login", or "email"
- Reduces false positives compared to password-only detection

### 4. **Error Handling**
- Network errors: Display connection error message
- HTTP errors (4xx, 5xx): Display status code and description
- Parsing errors: Display parsing failure message
- Size limit: Maximum 10MB page size to prevent DoS

### 5. **Frontend**
- Server-side rendering with Go's `text/template`
- Modern responsive CSS with flexbox and CSS Grid
- Gradient design for visual appeal
- Mobile-first responsive design

### 6. **Testing**
- Unit tests for all analyzer functions
- Integration tests with mock HTTP servers
- Mock server tests for HTTP error codes
- Comprehensive test coverage for edge cases

## API Endpoints

### GET `/`
Returns the home page with the URL input form.

### POST `/analyze`
Analyzes a URL provided in the form data.
- **Parameter**: `url` (required) - The URL to analyze
- **Returns**: HTML page with analysis results or error message

### GET `/static/*`
Serves static files (CSS, etc.)

## Error Handling

The application handles various error scenarios:

| Error Type | Response |
|-----------|----------|
| Invalid URL | Displays error message explaining the issue |
| Connection Failed | Shows connection error with details |
| HTTP 4xx/5xx | Displays HTTP status code and meaning |
| Parse Error | Shows HTML parsing error |
| Large Pages | Returns error if page exceeds 10MB limit |
| Missing URL Parameter | Returns 400 Bad Request |

## Testing

Run the test suite:
```bash
go test ./analyzer -v
```

Run with coverage:
```bash
go test ./analyzer -cover
```

Run benchmarks:
```bash
go test ./analyzer -bench=. -benchmem
```

## Performance Considerations

- **Request Timeout**: 10 seconds per URL fetch
- **Size Limit**: Maximum 10MB page content
- **Efficient DOM Traversal**: Single-pass analysis
- **Connection Pooling**: Reusable HTTP client

## Security Considerations

- **Input Validation**: URLs are validated before processing
- **Size Limits**: Prevents memory exhaustion attacks
- **Timeout Protection**: Prevents hanging connections
- **Safe HTML Parsing**: Uses standard HTML parser
- **No Code Execution**: Pure static analysis only

## Future Improvements

1. **Database Integration**
   - Store analysis history
   - Track frequently analyzed sites
   - Provide analytics dashboard

2. **Advanced Features**
   - Page performance metrics (load time, resource sizes)
   - SEO analysis (meta tags, keywords, readability)
   - Accessibility audit (WCAG compliance)
   - Security headers detection
   - CDN detection

3. **User Interface Enhancements**
   - Analysis history/favorites
   - Export results (JSON, CSV, PDF)
   - Comparison mode (analyze multiple URLs side-by-side)
   - Dark mode toggle
   - Advanced filtering options

4. **Backend Improvements**
   - Caching layer for frequently analyzed sites
   - Rate limiting per IP
   - Webhook integration for batch processing
   - REST API endpoints

5. **Monitoring & Analytics**
   - Application metrics and health checks
   - Request logging and analysis
   - Error tracking and reporting
   - Performance monitoring

6. **Deployment**
   - Docker containerization
   - Kubernetes deployment manifests
   - CI/CD pipeline setup
   - Configuration management

## Assumptions

1. **URL Format**: URLs without a scheme default to HTTPS
2. **Link Categorization**: Links are internal if they share the same domain (host) as the page being analyzed
3. **HTML Version**: Detected from DOCTYPE declaration; defaults to "Unknown"
4. **Login Form**: Requires both a password field and a username/email indicator
5. **Timeout**: All URL fetches have a 10-second timeout
6. **Page Size**: Maximum analyzed page size is 10MB
7. **Character Encoding**: Assumes UTF-8 or compatible encoding

## License

This project is provided as-is for educational purposes.

## Support

For issues or questions, please refer to the documentation or review the inline code comments.
