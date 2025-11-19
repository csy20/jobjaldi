package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
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
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/jobs", jobsHandler)
	mux.HandleFunc("/jobs/", jobByIDHandler) // e.g. /jobs/1

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
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

	writeJSON(w, filtered)
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
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Println("writeJSON error:", err)
	}
}

func containsIgnoreCase(s, sub string) bool {
	// very simple case-insensitive contains
	return len(sub) == 0 || indexIgnoreCase(s, sub) >= 0
}

func indexIgnoreCase(s, sub string) int {
	// naive implementation to avoid imports; ok for small data
	sLower := []rune{}
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			r += 'a' - 'A'
		}
		sLower = append(sLower, r)
	}
	subLower := []rune{}
	for _, r := range sub {
		if r >= 'A' && r <= 'Z' {
			r += 'a' - 'A'
		}
		subLower = append(subLower, r)
	}

	for i := 0; i+len(subLower) <= len(sLower); i++ {
		match := true
		for j := 0; j < len(subLower); j++ {
			if sLower[i+j] != subLower[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}
