// Tests for scheduling code.
package job

import (
	"testing"
	"time"

	"github.com/lovego/kala/types"
	"github.com/stretchr/testify/assert"
)

var getWaitDurationTableTests = []struct {
	Job              *Job
	JobFunc          func() *Job
	ExpectedDuration time.Duration
	Name             string
}{
	{
		JobFunc: func() *Job {
			return &Job{
				Job: &types.Job{
					// Schedule time is passed
					Schedule: "R/2015-10-17T11:44:54.389361-07:00/PT10S",
					Metadata: types.Metadata{
						LastAttemptedRun: time.Now(),
					},
				},
			}
		},
		ExpectedDuration: 10 * time.Second,
		Name:             "Passed time, 10 seconds",
	},
	{
		JobFunc: func() *Job {
			return &Job{
				Job: &types.Job{
					// Schedule time is passed
					Schedule: "R/2015-10-17T11:44:54.389361-07:00/PT1M",
					Metadata: types.Metadata{
						LastAttemptedRun: time.Now(),
					},
				},
			}
		},
		ExpectedDuration: time.Minute,
		Name:             "Passed time, 1 minute",
	},
	{
		JobFunc: func() *Job {
			return &Job{
				Job: &types.Job{
					// Schedule time is passed, no repetitions
					Schedule: "R0/2015-10-17T11:44:54.389361-07:00/",
					Metadata: types.Metadata{
						LastAttemptedRun: time.Time{},
					},
				},
			}
		},
		ExpectedDuration: 0 * time.Second,
		Name:             "Passed time, 0 seconds",
	},
	{
		JobFunc: func() *Job {
			return &Job{
				Job: &types.Job{
					// Schedule time is passed
					Schedule: "R/2015-10-17T11:44:54.389361-07:00/P1D",
					Metadata: types.Metadata{
						LastAttemptedRun: time.Now(),
					},
				},
			}
		},
		ExpectedDuration: 24 * time.Hour,
		Name:             "Passed time, 1 day",
	},
}

func TestGetWatiDuration(t *testing.T) {
	for _, testStruct := range getWaitDurationTableTests {
		testStruct.Job = testStruct.JobFunc()
		err := testStruct.Job.InitDelayDuration(false)
		assert.NoError(t, err)
		actualDuration := testStruct.Job.GetWaitDuration()
		Logger.Errorf("LastAttempted: %s", testStruct.Job.Metadata.LastAttemptedRun)
		assert.InDelta(t, float64(testStruct.ExpectedDuration), float64(actualDuration), float64(time.Millisecond*50), "Test of "+testStruct.Name)
	}
}
