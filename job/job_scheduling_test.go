// Tests for scheduling code.
package job

import (
	"fmt"
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
					Schedule: "R/2025-01-31T11:44:54.389361+08:00/P1M",
					Metadata: types.Metadata{
						LastAttemptedRun: time.Date(2026, 1, 31, 12, 0, 0, 0, time.UTC),
					},
				},
			}
		},
		ExpectedDuration: 24 * time.Hour,
		Name:             "Passed time, 1 month",
	},
}

func TestGetWatiDuration(t *testing.T) {
	for _, testStruct := range getWaitDurationTableTests {
		testStruct.Job = testStruct.JobFunc()
		err := testStruct.Job.InitDelayDuration(false)
		assert.NoError(t, err)
		actualDuration := testStruct.Job.GetWaitDuration()
		fmt.Println(`222`, actualDuration)
		Logger.Errorf("LastAttempted: %s", testStruct.Job.Metadata.LastAttemptedRun)
		assert.InDelta(t, float64(testStruct.ExpectedDuration), float64(actualDuration), float64(time.Millisecond*50), "Test of "+testStruct.Name)
	}
}
