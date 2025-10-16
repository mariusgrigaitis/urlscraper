# Page Insight Tool

## Objective

Build a web application that accepts a URL input, analyzes the contents of the corresponding web page, and displays specific information about it.

## Functionality Requirements

The application should:

1. Display a form with a single input field for a URL and a submit button.
2. Upon form submission:
   - Fetch the HTML content of the provided URL.
   - Analyze and extract the following information:
     - HTML version (e.g., HTML5, HTML 4.01)
     - Page title
     - Count of headings per level (h1 through h6)
     - Number of internal and external links
     - Count of inaccessible links
     - Whether the page contains a login form
3. Handle unreachable URLs gracefully:
   - Display a clear error message that includes the HTTP status code and a brief explanation.

## Technical Requirements

- Use Golang for the backend.
- Frontend is flexible: choose server-side rendering or your preferred frontend stack.
- You may use any third-party libraries/tools.
- The project must be version-controlled using Git.

## Deliverables

Submit a Git repository (hosted or as a downloadable archive) that includes:

1. Complete source code.
2. A short README (README.md) containing:
   - Build and run instructions
   - Assumptions and design decisions made
   - Suggestions for future improvements
