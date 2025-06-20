package types

import (
	"time"
)

// JobStat is used to store metrics about a specific Job .Run()
type JobStat struct {
	JobId             string     `json:"job_id"`
	RanAt             time.Time  `json:"ran_at"`
	NumberOfRetries   uint       `json:"number_of_retries"`
	Success           bool       `json:"success"`
	ExecutionDuration int64      `json:"execution_duration"`
	Duration          string     `json:"duration,omitempty"`
	FinishAt          *time.Time `json:"finish_at,omitempty"`
	Error             string     `json:"error,omitempty"`
	Response          string     `json:"response,omitempty"`
}

// KalaStats is the struct for storing app-level metrics
type KalaStats struct {
	ActiveJobs   int `json:"active_jobs"`
	DisabledJobs int `json:"disabled_jobs"`
	Jobs         int `json:"jobs"`

	ErrorCount   uint `json:"error_count"`
	SuccessCount uint `json:"success_count"`

	NextRunAt        time.Time `json:"next_run_at"`
	LastAttemptedRun time.Time `json:"last_attempted_run"`

	CreatedAt time.Time `json:"created"`
}
