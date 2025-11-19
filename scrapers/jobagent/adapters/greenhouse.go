package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	greenhouseBaseURL    = "https://boards.greenhouse.io"
	greenhouseAPIBaseURL = "https://boards-api.greenhouse.io/v1/boards"
	greenhouseSource     = "greenhouse"
)

// Job describes a single scraped opening.
type Job struct {
	Title    string `json:"title"`
	Company  string `json:"company"`
	URL      string `json:"url"`
	Location string `json:"location,omitempty"`
	Source   string `json:"source"`
	Level    string `json:"level"`
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

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []Job{}, nil
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("unexpected status %d from %s", resp.StatusCode, url)
	}

	var payload greenhouseResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode greenhouse response: %w", err)
	}

	jobs := make([]Job, 0, len(payload.Jobs))
	for _, item := range payload.Jobs {
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
		jobs = append(jobs, Job{
			Title:    title,
			Company:  jobCompany,
			URL:      href,
			Location: location,
			Source:   greenhouseSource,
			Level:    inferLevel(title),
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
	} `json:"jobs"`
}

func fetchDocument(ctx context.Context, client *http.Client, ua, url string) (*goquery.Document, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if ua != "" {
		req.Header.Set("User-Agent", ua)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("unexpected status %d from %s", resp.StatusCode, url)
	}

	return goquery.NewDocumentFromReader(resp.Body)
}

func ensureAbsoluteURL(baseURL, href string) string {
	href = strings.TrimSpace(href)
	if href == "" {
		return ""
	}
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	if strings.HasPrefix(href, "/") {
		return strings.TrimRight(baseURL, "/") + href
	}
	return strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(href, "/")
}

func inferLevel(title string) string {
	t := strings.ToLower(title)
	switch {
	case strings.Contains(t, "intern"):
		return "intern"
	case strings.Contains(t, "new grad"):
		return "new grad"
	case strings.Contains(t, "entry"):
		return "entry"
	default:
		return ""
	}
}
