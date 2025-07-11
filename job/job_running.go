// Use Redis to prevent duplicate task execution.
package job

import (
	"errors"

	"github.com/garyburd/redigo/redis"
	"github.com/lovego/kala/types"
)

var (
	runningKeyPrefix      = "kala-job-running"
	concurrency           = 2 // max concurrency jobs for every group name
	ErrJobIsRunning       = errors.New("job is running")
	ErrBeyoundConcurrency = errors.New("beyound job concurrency")
)

// clear all running jobs (used for app start)
func clear(jobs ...*Job) error {
	if len(jobs) == 0 {
		return nil
	}
	runningKeys := make([]interface{}, len(jobs))
	for i := range jobs {
		runningKeys[i] = runningKey(jobs[i].GroupName, jobs[i].Id)
	}
	conn := pool.Get()
	defer conn.Close()
	_, err := conn.Do("DEL", runningKeys...)
	return err
}

// job start
func (j *Job) start() error {
	conn := pool.Get()
	defer conn.Close()
	jobIds, err := groupRunningJobs(conn, j.GroupName)
	if err != nil {
		return err
	}
	if len(jobIds) >= concurrency {
		return ErrBeyoundConcurrency
	}
	ok, err := j.setRunning(conn)
	if err != nil {
		return err
	}
	if !ok {
		return ErrJobIsRunning
	}
	return nil
}

// job finished
func (j *Job) finish() error {
	conn := pool.Get()
	defer conn.Close()
	_, err := conn.Do("DEL", runningKey(j.GroupName, j.Id))
	return err
}

// set job running
func (j *Job) setRunning(conn redis.Conn) (bool, error) {
	reply, err := redis.String(conn.Do("SET", runningKey(j.GroupName, j.Id), j.Id, "NX"))
	if err != nil && err != redis.ErrNil {
		return false, err
	}
	return reply == "OK", nil
}

// job running stat
func (j *Job) isRunning() (bool, error) {
	conn := pool.Get()
	defer conn.Close()
	reply, err := redis.Int(conn.Do("EXISTS", runningKey(j.GroupName, j.Id)))
	if err != nil && err != redis.ErrNil {
		return false, err
	}
	return reply == 1, nil
}

// get running jobs
func groupRunningJobs(conn redis.Conn, group string) ([]string, error) {
	jobs, err := redis.Strings(conn.Do("KEYS", runningKey(group, "")))
	if err != nil && err != redis.ErrNil {
		return nil, err
	}
	return jobs, nil
}

func runningKey(groupName, id string) string {
	var key = runningKeyPrefix
	if groupName != "" {
		key += "-" + groupName
	}
	if id != "" {
		key += "-" + id
	}
	return key
}

func JobsRunning(jobs map[string]*types.Job) error {
	if len(jobs) == 0 {
		return nil
	}
	runningKeys := make([]interface{}, 0, len(jobs))
	for k := range jobs {
		if jobs[k].Disabled || jobs[k].Deleted || jobs[k].IsDone {
			continue
		}
		runningKeys = append(runningKeys, runningKey(jobs[k].GroupName, jobs[k].Id))
	}
	if len(runningKeys) == 0 {
		return nil
	}
	conn := pool.Get()
	defer conn.Close()
	running, err := redis.Strings(conn.Do("MGET", runningKeys...))
	if err != nil && err != redis.ErrNil {
		return err
	}
	for i := range running {
		if _, ok := jobs[running[i]]; !ok {
			continue
		}
		jobs[running[i]].IsRunning = true
	}
	return nil
}
