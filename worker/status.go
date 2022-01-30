package worker

// Status can be used to report the current Status of a Manager.
type Status struct {
	ActiveWorkers int `json:"activeWorkers"`
	// Number of errors occurred in all workers thus far.
	Errors int `json:"errors"`
}
