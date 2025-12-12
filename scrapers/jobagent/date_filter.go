package jobagent

import (
	"time"

	"github.com/csy/jobagent/jobagent/adapters"
)

// MaxJobAgeDays is the maximum age of jobs to include (30 days = 1 month)
const MaxJobAgeDays = 30

// filterJobsByDate filters jobs to only include those updated within the last MaxJobAgeDays
// Jobs without an updated_at timestamp are included by default
func filterJobsByDate(jobs []adapters.Job) []adapters.Job {
	if len(jobs) == 0 {
		return jobs
	}

	cutoff := time.Now().AddDate(0, 0, -MaxJobAgeDays).Unix()
	filtered := make([]adapters.Job, 0, len(jobs))

	for _, job := range jobs {
		// Include job if:
		// 1. No UpdatedAt timestamp (can't determine age, include it)
		// 2. UpdatedAt is within the last MaxJobAgeDays
		if job.UpdatedAt == 0 || job.UpdatedAt >= cutoff {
			filtered = append(filtered, job)
		}
	}

	return filtered
}
