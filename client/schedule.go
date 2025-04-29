package client

import (
	"fmt"
	"time"

	"github.com/lovego/kala/utils/iso8601"
)

type Schedule struct {
	After    time.Duration
	Interval iso8601.Duration
	Repeat   int
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
