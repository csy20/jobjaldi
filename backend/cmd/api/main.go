package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Job struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Company  string `json:"company"`
	Location string `json:"location"`
	Source   string `json:"source"`
	URL      string `json:"url"`
}

// TODO: replace this with real DB queries.
var jobs = []Job{
	{ID: 1, Title: "Backend Engineer", Company: "ExampleCorp", Location: "Remote", Source: "Naukri", URL: "https://example.com/job/1"},
	{ID: 2, Title: "Go Developer", Company: "StartupX", Location: "Bangalore", Source: "LinkedIn", URL: "https://example.com/job/2"},
	{ID: 3, Title: "Flutter Developer", Company: "AppStudio", Location: "Delhi", Source: "Indeed", URL: "https://example.com/job/3"},
	{ID: 4, Title: "DevOps Engineer", Company: "CloudScale", Location: "Remote", Source: "Naukri", URL: "https://example.com/job/4"},
	{ID: 5, Title: "Data Scientist", Company: "AI Labs", Location: "Bangalore", Source: "LinkedIn", URL: "https://example.com/job/5"},
}

// Buffer pool for JSON encoding to reduce allocations
var jsonBufferPool = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}


func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/jobs", jobsHandler)
	mux.HandleFunc("/jobs/", jobByIDHandler) // e.g. /jobs/1

	// Wrap mux with GZIP middleware
	handler := gzipMiddleware(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Listening on :%s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}

func gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		acceptEncoding := r.Header.Get("Accept-Encoding")
		if !strings.Contains(acceptEncoding, "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")
		
		// Create new gzip writer (can't be pooled due to Reset limitations)
		gz := gzip.NewWriter(w)
		defer gz.Close()

		gw := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		next.ServeHTTP(gw, r)
	})
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func jobsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Simple filters by query parameters (optional)
	q := r.URL.Query()
	location := q.Get("location")
	role := q.Get("role")

	// Pagination
	pageStr := q.Get("page")
	limitStr := q.Get("limit")

	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	filtered := make([]Job, 0, len(jobs))
	for _, j := range jobs {
		if location != "" && j.Location != location {
			continue
		}
		if role != "" && !containsIgnoreCase(j.Title, role) {
			continue
		}
		filtered = append(filtered, j)
	}

	// Apply pagination slice
	start := (page - 1) * limit
	if start > len(filtered) {
		start = len(filtered)
	}
	end := start + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	paginated := filtered[start:end]

	writeJSON(w, paginated)
}

func jobByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// URL pattern: /jobs/{id}
	// trim "/jobs/"
	idStr := r.URL.Path[len("/jobs/"):]
	if idStr == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	for _, j := range jobs {
		if j.ID == id {
			writeJSON(w, j)
			return
		}
	}

	http.NotFound(w, r)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	
	// Use buffer pool for better performance
	buf := jsonBufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		jsonBufferPool.Put(buf)
	}()

	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false) // Faster encoding
	if err := encoder.Encode(v); err != nil {
		log.Println("writeJSON error:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Remove trailing newline
	data := buf.Bytes()
	if len(data) > 0 && data[len(data)-1] == '\n' {
		data = data[:len(data)-1]
	}
	
	w.Write(data)
}

func containsIgnoreCase(s, sub string) bool {
	if len(sub) == 0 {
		return true
	}
	if len(sub) > len(s) {
		return false
	}
	// Optimized case-insensitive contains using bytes
	sBytes := []byte(s)
	subBytes := []byte(sub)
	
	for i := 0; i <= len(sBytes)-len(subBytes); i++ {
		match := true
		for j := 0; j < len(subBytes); j++ {
			c1 := sBytes[i+j]
			c2 := subBytes[j]
			// Fast ASCII case-insensitive comparison
			if c1 >= 'A' && c1 <= 'Z' {
				c1 += 'a' - 'A'
			}
			if c2 >= 'A' && c2 <= 'Z' {
				c2 += 'a' - 'A'
			}
			if c1 != c2 {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
