package types

type KalaStatsResponse struct {
	Stats *KalaStats
}

type ListJobStatsResponse struct {
	JobStats []*JobStat `json:"job_stats"`
}

type ListJobsResponse struct {
	Jobs map[string]*Job `json:"jobs"`
}

type AddJobResponse struct {
	Id string `json:"id"`
}

type JobResponse struct {
	Job *Job `json:"job"`
}
