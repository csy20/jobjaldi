package jobs

type Job struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Company  string `json:"company"`
	Location string `json:"location"`
	Source   string `json:"source"`
	URL      string `json:"url"`
}

// Repository defines the interface for job storage
type Repository interface {
	SaveBatch(jobs []Job) error
	GetAll() ([]Job, error)
}
