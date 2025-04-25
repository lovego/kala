package types

const (
	// Base API v1 Path
	ApiUrlPrefix = "/api/v1"
	JobPath      = "/job"
)

const (
	LocalJob JobType = iota
	RemoteJob
)
