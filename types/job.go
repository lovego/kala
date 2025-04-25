package types

import (
	"net/http"
	"time"
)

type Job struct {
	Id   string `json:"id"`
	Name string `json:"name"`

	// Owner of this job
	Owner string `json:"owner"`

	// Group of this job
	GroupName string `json:"groupName"`

	// Content of this job
	Content string `json:"content"`

	// Command to run
	// e.g. "bash /path/to/my/script.sh"
	Command string `json:"command"`

	// Is this job disabled?
	Disabled bool `json:"disabled"`

	// Jobs that are dependent upon this one will be run after this job runs.
	DependentJobs []string `json:"dependent_jobs"`

	// List of ids of jobs that this job is dependent upon.
	ParentJobs []string `json:"parent_jobs"`

	// Job that gets run after all retries have failed consecutively
	OnFailureJob string `json:"on_failure_job"`

	// ISO 8601 String
	// e.g. "R/2014-03-08T20:00:00.000Z/PT2H"
	Schedule string `json:"schedule"`

	// Number of times to retry on failed attempt for each run.
	Retries uint `json:"retries"`

	// Duration in which it is safe to retry the Job.
	Epsilon string `json:"epsilon"`

	NextRunAt time.Time `json:"next_run_at"`

	// Templating delimiters, the left & right separated by space,
	// for example `{{ }}` or `${ }`.
	//
	// If this field is non-empty, then each time this
	// job is executed, Kala will template its main
	// content as a Go Template with the job itself as data.
	//
	// The Command is templated for local jobs,
	// and Url and Body in RemoteProperties.
	TemplateDelimiters string

	// If the job is disabled (or the system inoperative) and we pass
	// the scheduled run point, when the job becomes active again,
	// normally the job will run immediately.
	// With this setting on, it will not run immediately, but will wait
	// until the next scheduled run time comes along.
	ResumeAtNextScheduledTime bool `json:"resume_at_next_scheduled_time"`

	// Meta data about successful and failed runs.
	Metadata Metadata `json:"metadata"`

	// Type of the job
	JobType JobType `json:"job_type"`

	// Custom properties for the remote job type
	RemoteProperties RemoteProperties `json:"remote_properties"`

	// Collection of Job Stats
	Stats []*JobStat `json:"stats"`

	// Says if a job has been executed right numbers of time
	// and should not been executed again in the future
	IsDone bool `json:"is_done"`
}

// RemoteProperties Custom properties for the remote job type
type RemoteProperties struct {
	Url    string `json:"url"`
	Method string `json:"method"`

	// A body to attach to the http request
	Body string `json:"body"`

	// A list of headers to add to http request (e.g. [{"key": "charset", "value": "UTF-8"}])
	Headers http.Header `json:"headers"`

	// A timeout property for the http request in seconds
	Timeout int `json:"timeout"`

	// A list of expected response codes (e.g. [200, 201])
	ExpectedResponseCodes []int `json:"expected_response_codes"`
}

type Metadata struct {
	SuccessCount         uint      `json:"success_count"`
	LastSuccess          time.Time `json:"last_success"`
	ErrorCount           uint      `json:"error_count"`
	LastError            time.Time `json:"last_error"`
	LastAttemptedRun     time.Time `json:"last_attempted_run"`
	NumberOfFinishedRuns uint      `json:"number_of_finished_runs"`
}

type JobType int
