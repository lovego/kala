package job

import (
	"time"

	"github.com/lovego/kala/types"
)

// NewKalaStats is used to easily generate a current app-level metrics report.
func NewKalaStats(cache JobCache) *types.KalaStats {
	ks := &types.KalaStats{
		CreatedAt: time.Now(),
	}
	jobs := cache.GetAll()
	jobs.Lock.RLock()
	defer jobs.Lock.RUnlock()

	ks.Jobs = len(jobs.Jobs)
	if ks.Jobs == 0 {
		return ks
	}

	nextRun := time.Time{}
	lastRun := time.Time{}
	for _, job := range jobs.Jobs {
		func() {
			job.lock.RLock()
			defer job.lock.RUnlock()

			if job.Disabled {
				ks.DisabledJobs += 1
			} else {
				ks.ActiveJobs += 1
			}

			if nextRun.IsZero() {
				nextRun = job.NextRunAt
			} else if (nextRun.UnixNano() - job.NextRunAt.UnixNano()) > 0 {
				nextRun = job.NextRunAt
			}

			if lastRun.IsZero() {
				if !job.Metadata.LastAttemptedRun.IsZero() {
					lastRun = job.Metadata.LastAttemptedRun
				}
			} else if (lastRun.UnixNano() - job.Metadata.LastAttemptedRun.UnixNano()) < 0 {
				lastRun = job.Metadata.LastAttemptedRun
			}

			ks.ErrorCount += job.Metadata.ErrorCount
			ks.SuccessCount += job.Metadata.SuccessCount
		}()
	}
	ks.NextRunAt = nextRun
	ks.LastAttemptedRun = lastRun

	return ks
}

func NewJobStat(id string) *types.JobStat {
	return &types.JobStat{
		JobId: id,
		RanAt: time.Now(),
	}
}
