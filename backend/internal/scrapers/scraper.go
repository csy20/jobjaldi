package scrapers

import "jobjaldi/backend/internal/jobs"

// Scraper defines the interface for a job board scraper
type Scraper interface {
	Scrape() ([]jobs.Job, error)
}

// FetchAll runs all registered scrapers and returns combined results
func FetchAll() ([]jobs.Job, error) {
	// TODO: Register and run scrapers
	return []jobs.Job{}, nil
}
