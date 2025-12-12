package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// httpStatusError represents an HTTP status error
type httpStatusError struct {
	statusCode int
	url        string
	message    string
}

func (e *httpStatusError) Error() string {
	return e.message
}

func (e *httpStatusError) StatusCode() int {
	return e.statusCode
}

const (
	greenhouseBaseURL    = "https://boards.greenhouse.io"
	greenhouseAPIBaseURL = "https://boards-api.greenhouse.io/v1/boards"
	greenhouseSource     = "greenhouse"
)

// Job describes a single scraped opening.
type Job struct {
	Title     string `json:"title"`
	Company   string `json:"company"`
	URL       string `json:"url"`
	Location  string `json:"location,omitempty"`
	Source    string `json:"source"`
	Level     string `json:"level"`
	UpdatedAt int64  `json:"updated_at,omitempty"` // Unix timestamp for date filtering
}

// Fetcher defines the adapter contract so callers can register providers dynamically.
type Fetcher func(ctx context.Context, client *http.Client, ua, company string) ([]Job, error)

// FetchGreenhouse scrapes the Greenhouse job board for the provided company slug.
func FetchGreenhouse(ctx context.Context, client *http.Client, ua, company string) ([]Job, error) {
	if client == nil {
		return nil, fmt.Errorf("http client is required")
	}

	company = strings.TrimSpace(company)
	if company == "" {
		return nil, fmt.Errorf("company cannot be empty")
	}

	url := fmt.Sprintf("%s/%s/jobs?content=false", greenhouseAPIBaseURL, company)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if ua != "" {
		req.Header.Set("User-Agent", ua)
	}
	req.Header.Set("Accept", "application/json")
	// Do not manually set Accept-Encoding, let http.Transport handle it

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("greenhouse request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []Job{}, nil
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		// Return error that can be checked by retry logic
		return nil, &httpStatusError{
			statusCode: resp.StatusCode,
			url:        url,
			message:    fmt.Sprintf("unexpected status %d from %s", resp.StatusCode, url),
		}
	}

	var payload greenhouseResponse
	// Use json.Decoder directly for better performance
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode greenhouse response: %w", err)
	}

	// Pre-allocate with exact capacity
	jobs := make([]Job, 0, len(payload.Jobs))
	for i := range payload.Jobs {
		item := &payload.Jobs[i]
		title := strings.TrimSpace(item.Title)
		href := strings.TrimSpace(item.AbsoluteURL)
		if title == "" || href == "" {
			continue
		}

		jobCompany := strings.TrimSpace(item.CompanyName)
		if jobCompany == "" {
			jobCompany = company
		}

		location := strings.TrimSpace(item.Location.Name)

		// Parse updated_at timestamp
		var updatedAt int64
		if item.UpdatedAt != "" {
			if t, err := time.Parse(time.RFC3339, item.UpdatedAt); err == nil {
				updatedAt = t.Unix()
			}
		}

		jobs = append(jobs, Job{
			Title:     title,
			Company:   jobCompany,
			URL:       href,
			Location:  location,
			Source:    greenhouseSource,
			Level:     inferLevel(title),
			UpdatedAt: updatedAt,
		})
	}
	return jobs, nil
}

type greenhouseResponse struct {
	Jobs []struct {
		Title       string `json:"title"`
		AbsoluteURL string `json:"absolute_url"`
		Location    struct {
			Name string `json:"name"`
		} `json:"location"`
		CompanyName string `json:"company_name"`
		UpdatedAt   string `json:"updated_at"` // ISO 8601 format
	} `json:"jobs"`
}

func fetchDocument(ctx context.Context, client *http.Client, ua, url string) (*goquery.Document, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}
	if ua != "" {
		req.Header.Set("User-Agent", ua)
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch document failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, &httpStatusError{
			statusCode: resp.StatusCode,
			url:        url,
			message:    fmt.Sprintf("unexpected status %d from %s", resp.StatusCode, url),
		}
	}

	return goquery.NewDocumentFromReader(resp.Body)
}

func ensureAbsoluteURL(baseURL, href string) string {
	href = strings.TrimSpace(href)
	if href == "" {
		return ""
	}
	// Fast path: already absolute
	if len(href) >= 7 && (href[:7] == "http://" || (len(href) >= 8 && href[:8] == "https://")) {
		return href
	}
	// Use strings.Builder for better performance
	var builder strings.Builder
	builder.Grow(len(baseURL) + len(href) + 1)
	if strings.HasPrefix(href, "/") {
		baseURL = strings.TrimRight(baseURL, "/")
		builder.WriteString(baseURL)
		builder.WriteString(href)
	} else {
		baseURL = strings.TrimRight(baseURL, "/")
		builder.WriteString(baseURL)
		builder.WriteByte('/')
		builder.WriteString(strings.TrimLeft(href, "/"))
	}
	return builder.String()
}

// inferLevel optimizes level detection with early returns and better string matching
func inferLevel(title string) string {
	// Convert to lowercase once
	t := strings.ToLower(title)

	// Check in order of likelihood (most common first)
	if strings.Contains(t, "intern") {
		return "intern"
	}
	if strings.Contains(t, "new grad") || strings.Contains(t, "newgrad") {
		return "new grad"
	}
	if strings.Contains(t, "entry") {
		return "entry"
	}

	return ""
}
