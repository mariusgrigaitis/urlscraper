package main

import (
	"log"
	"net/http"
	"text/template"
	"urlscraper/analyzer"
)

var templates *template.Template

func init() {
	var err error
	templates, err = template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatal("Error parsing templates:", err)
	}
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/analyze", analyzeHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Starting server on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templates.ExecuteTemplate(w, "index.html", nil)
}

func analyzeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	urlStr := r.FormValue("url")
	if urlStr == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	analysis := analyzer.AnalyzeURL(urlStr)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templates.ExecuteTemplate(w, "results.html", analysis)
}
