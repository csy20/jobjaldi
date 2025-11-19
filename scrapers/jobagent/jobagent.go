package jobagent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/csy/jobagent/jobagent/adapters"
)

const userAgent = "csy-jobboard/1.0 (+personal-use)"

var (
	httpClient = &http.Client{Timeout: 20 * time.Second}

	providerRegistry = map[string]adapters.Fetcher{
		"greenhouse": adapters.FetchGreenhouse,
		"lever":      adapters.FetchLever,
	}
)

// Job re-exports the adapter job structure for gomobile bindings.
type Job = adapters.Job

type scrapeConfig struct {
	Targets []targetSpec `json:"targets"`
}

type targetSpec struct {
	Provider string `json:"provider"`
	Company  string `json:"company"`
}

// ScrapeProvider fetches jobs for a single provider/company pair.
func ScrapeProvider(provider, company string) (string, error) {
	fetcher, company, err := resolveTarget(provider, company)
	if err != nil {
		return "", err
	}

	jobs, err := fetcher(context.Background(), httpClient, userAgent, company)
	if err != nil {
		return "", err
	}

	return marshalJobs(jobs)
}

// ScrapeMany processes a JSON config and fetches from all targets concurrently.
func ScrapeMany(cfgJSON string) (string, error) {
	var cfg scrapeConfig
	if err := json.Unmarshal([]byte(cfgJSON), &cfg); err != nil {
		return "", fmt.Errorf("invalid config json: %w", err)
	}

	if len(cfg.Targets) == 0 {
		return marshalJobs([]Job{})
	}

	ctx := context.Background()
	var wg sync.WaitGroup
	results := make(chan []Job, len(cfg.Targets))
	errorsChan := make(chan error, len(cfg.Targets))

	for _, target := range cfg.Targets {
		wg.Add(1)
		go func(t targetSpec) {
			defer wg.Done()
			fetcher, company, err := resolveTarget(t.Provider, t.Company)
			if err != nil {
				errorsChan <- fmt.Errorf("%s/%s: %v", t.Provider, t.Company, err)
				return
			}

			targetJobs, err := fetcher(ctx, httpClient, userAgent, company)
			if err != nil {
				errorsChan <- fmt.Errorf("%s/%s: %v", t.Provider, t.Company, err)
				return
			}
			results <- targetJobs
		}(target)
	}

	wg.Wait()
	close(results)
	close(errorsChan)

	jobs := make([]Job, 0)
	for res := range results {
		jobs = append(jobs, res...)
	}

	var errs []string
	for err := range errorsChan {
		errs = append(errs, err.Error())
	}

	if len(jobs) == 0 && len(errs) > 0 {
		return "", fmt.Errorf("all targets failed: %s", strings.Join(errs, "; "))
	}

	return marshalJobs(jobs)
}

func resolveTarget(provider, company string) (adapters.Fetcher, string, error) {
	key := normalize(provider)
	fetcher, ok := providerRegistry[key]
	if !ok {
		return nil, "", fmt.Errorf("unsupported provider %q", provider)
	}

	company = strings.TrimSpace(company)
	if company == "" {
		return nil, "", errors.New("company cannot be empty")
	}

	return fetcher, company, nil
}

func marshalJobs(jobs []Job) (string, error) {
	if jobs == nil {
		jobs = make([]Job, 0)
	}

	payload, err := json.Marshal(jobs)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

func normalize(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}
