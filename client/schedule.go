package client

import (
	"fmt"
	"net/http"
	"time"

	"github.com/lovego/kala/types"
	"github.com/lovego/kala/utils/iso8601"
)

type Schedule struct {
	After    time.Duration
	Interval iso8601.Duration
	Repeat   uint
}

// Make Schedule for create job.
// Number of times to repeat/Start Datetime/Interval Between Runs
func (s Schedule) String() string {
	schedule := fmt.Sprintf("R%d/%s", s.Repeat, time.Now().Add(s.After).Format(time.RFC3339))
	if !s.Interval.IsZero() {
		schedule += "/" + s.Interval.String()
	} else {
		schedule += "/PT0S"
	}
	return schedule
}

type Scheduler struct {
	Name      string                 // job name
	Owner     string                 // job owner
	GroupName string                 // job group name
	Content   string                 // job content
	Remote    types.RemoteProperties // job remote properties
	Retries   uint                   // job retry times when failed
	Schedule
}

func (s Scheduler) Job() *types.Job {
	job := &types.Job{
		Name:             s.Name,
		Owner:            s.Owner,
		GroupName:        s.GroupName,
		Content:          s.Content,
		Schedule:         s.Schedule.String(),
		Retries:          s.Retries,
		JobType:          types.RemoteJob,
		RemoteProperties: s.Remote,
	}
	if job.RemoteProperties.Headers == nil {
		job.RemoteProperties.Headers = make(http.Header)
	}
	if job.Retries == 0 {
		job.Retries = 1 // default retry 1 times
	}
	return job
}
