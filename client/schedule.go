package client

import (
	"fmt"
	"net/http"
	"time"

	"github.com/lovego/kala/types"
	"github.com/lovego/kala/utils/iso8601"
)

type Schedule struct {
	After    time.Duration    `json:"after" comment:"Start datetime after now seconds"`
	Interval iso8601.Duration `json:"interval" comment:"Interval Between Runs"`
	Repeat   uint             `json:"repeat" comment:"Number of times to repeat"`
}

// Make Schedule for create job.
// Number of times to repeat/Start Datetime/Interval Between Runs
func (s Schedule) String() string {
	schedule := fmt.Sprintf("R%d/%s", s.Repeat, time.Now().Add(s.After*time.Second).Format(time.RFC3339))
	if !s.Interval.IsZero() {
		schedule += "/" + s.Interval.String()
	} else {
		schedule += "/PT0S"
	}
	return schedule
}

type Scheduler struct {
	Name      string                 `json:"name" comment:"job name"`
	Owner     string                 `json:"owner" comment:"job owner"`
	GroupName string                 `json:"group_name" comment:"job group name"`
	Content   string                 `json:"content" comment:"job content"`
	Remote    types.RemoteProperties `json:"remote" comment:"job remote properties"`
	Retries   uint                   `json:"retries" comment:"job retry times when failed"`
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
	return job
}
