package job

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/cornelk/hashmap"
	"github.com/garyburd/redigo/redis"
	"github.com/lovego/kala/types"
)

var (
	ErrJobDoesntExist = errors.New("The job you requested does not exist")
	pool              *redis.Pool
)

type JobCache interface {
	Get(id string) (*Job, error)
	GetAll() *JobsMap
	Set(j *Job) error
	Delete(id string, logical bool) error
	Persist() error
	Enable(j *Job) error
	Disable(j *Job) error
}

type JobsMap struct {
	Jobs map[string]*Job
	Lock sync.RWMutex
}

func NewJobsMap() *JobsMap {
	return &JobsMap{
		Jobs: map[string]*Job{},
		Lock: sync.RWMutex{},
	}
}

type MemoryJobCache struct {
	// Jobs is a map from Job id's to pointers to the jobs.
	// Used as the main "data store" within this cache implementation.
	jobs           *JobsMap
	jobDB          JobDB
	PersistOnWrite bool
}

func NewMemoryJobCache(jobDB JobDB) *MemoryJobCache {
	return &MemoryJobCache{
		jobs:  NewJobsMap(),
		jobDB: jobDB,
	}
}

func (c *MemoryJobCache) Start(persistWaitTime time.Duration) {
	if persistWaitTime == 0 {
		c.PersistOnWrite = true
	}

	// Prep cache
	allJobs, err := c.jobDB.GetAll()
	if err != nil {
		Logger.Fatal(err)
	}
	for _, j := range allJobs {
		if j.ShouldStartWaiting() {
			j.StartWaiting(c, false)
		}
		err = c.Set(j)
		if err != nil {
			Logger.Error(err)
		}
	}

	// Occasionally, save items in cache to db.
	if persistWaitTime > 0 {
		go c.PersistEvery(persistWaitTime)
	}

	// Process-level defer for shutting down the db.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		s := <-ch
		Logger.Infof("Process got signal: %s", s)
		Logger.Infof("Shutting down....")

		// Persist all jobs to database
		err = c.Persist()
		if err != nil {
			Logger.Error(err)
		}

		// Close the database
		c.jobDB.Close()

		os.Exit(0)
	}()
}

func (c *MemoryJobCache) Get(id string) (*Job, error) {
	c.jobs.Lock.RLock()
	defer c.jobs.Lock.RUnlock()

	j := c.jobs.Jobs[id]
	if j == nil {
		return nil, ErrJobDoesntExist
	}

	return j, nil
}

func (c *MemoryJobCache) GetAll() *JobsMap {
	return c.jobs
}

func (c *MemoryJobCache) Set(j *Job) error {
	c.jobs.Lock.Lock()
	defer c.jobs.Lock.Unlock()
	if j == nil {
		return nil
	}

	if c.PersistOnWrite {
		if err := c.jobDB.Save(j); err != nil {
			return err
		}
	}

	c.jobs.Jobs[j.Id] = j
	return nil
}

func (c *MemoryJobCache) Delete(id string, logical bool) error {
	c.jobs.Lock.Lock()
	defer c.jobs.Lock.Unlock()

	j := c.jobs.Jobs[id]
	if j == nil {
		return ErrJobDoesntExist
	}
	j.lock.Lock()
	defer j.lock.Unlock()

	err := c.jobDB.Delete(id)
	if err != nil {
		err = fmt.Errorf("Error occurred while trying to delete job from db: %s", err)
		if c.PersistOnWrite {
			return err
		}
	}

	j.lock.Unlock()
	j.StopTimer()
	j.lock.Lock()

	go func() {
		if err := j.DeleteFromParentJobs(c); err != nil {
			Logger.Error(err)
		}
	}()

	// Remove itself from dependent jobs as a parent job
	// and possibly delete child jobs if they don't have any other parents.
	go func() {
		if err := j.DeleteFromDependentJobs(c); err != nil {
			Logger.Error(err)
		}
	}()

	delete(c.jobs.Jobs, id)

	return err
}

func (c *MemoryJobCache) Enable(j *Job) error {
	return enable(j, c, c.PersistOnWrite)
}

// Disable stops a job from running by stopping its jobTimer. It also sets Job.Disabled to true,
// which is reflected in the UI.
func (c *MemoryJobCache) Disable(j *Job) error {
	return disable(j, c, c.PersistOnWrite)
}

func (c *MemoryJobCache) Persist() error {
	c.jobs.Lock.RLock()
	defer c.jobs.Lock.RUnlock()
	for _, j := range c.jobs.Jobs {
		err := c.jobDB.Save(j)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *MemoryJobCache) PersistEvery(persistWaitTime time.Duration) {
	wait := time.NewTicker(persistWaitTime).C
	var err error
	for {
		<-wait
		err = c.Persist()
		if err != nil {
			Logger.Errorf("Error occurred persisting the database. Err: %s", err)
		}
	}
}

type LockFreeJobCache struct {
	jobs            *hashmap.HashMap
	jobDB           JobDB
	retentionPeriod time.Duration
	PersistOnWrite  bool
	Clock
}

func NewLockFreeJobCache(jobDB JobDB) *LockFreeJobCache {
	return &LockFreeJobCache{
		jobs:            hashmap.New(8), //nolint:gomnd
		jobDB:           jobDB,
		retentionPeriod: -1,
	}
}

func (c *LockFreeJobCache) Start(redisPool *redis.Pool, persistWaitTime time.Duration, jobstatTtl time.Duration) {
	if persistWaitTime == 0 {
		c.PersistOnWrite = true
	}
	pool = redisPool

	// Prep cache
	allJobs, err := c.jobDB.GetAll()
	if err != nil {
		Logger.Fatal(err)
	}
	// Clear all jobs running stat
	err = clear(allJobs...)
	if err != nil {
		Logger.Fatal(err)
	}
	for _, j := range allJobs {
		if j.Schedule == "" {
			Logger.Infof("Job %s:%s skipped.", j.Name, j.Id)
			continue
		}
		if j.ShouldStartWaiting() {
			j.StartWaiting(c, false)
		}
		// Logger.Infof("Job %s:%s added to cache.", j.Name, j.Id)
		err := c.Set(j)
		if err != nil {
			Logger.Error(err)
		}
	}
	// Occasionally, save items in cache to db.
	if persistWaitTime > 0 {
		go c.PersistEvery(persistWaitTime)
	}

	// Run retention every minute to clean up old job stats entries
	if jobstatTtl > 0 {
		c.retentionPeriod = jobstatTtl
		go c.RetainEvery(1 * time.Minute)
	}
}

func (c *LockFreeJobCache) Get(id string) (*Job, error) {
	val, exists := c.jobs.GetStringKey(id)
	if val == nil || !exists {
		return nil, ErrJobDoesntExist
	}
	j := val.(*Job)
	if j == nil {
		return nil, ErrJobDoesntExist
	}
	return j, nil
}

func (c *LockFreeJobCache) GetAll() *JobsMap {
	jm := NewJobsMap()
	for el := range c.jobs.Iter() {
		jm.Jobs[el.Key.(string)] = el.Value.(*Job)
	}
	return jm
}

func (c *LockFreeJobCache) Set(j *Job) error {
	if j == nil {
		return nil
	}

	if c.PersistOnWrite {
		if err := c.jobDB.Save(j); err != nil {
			return err
		}
	}

	c.jobs.Set(j.Id, j)
	return nil
}

func (c *LockFreeJobCache) Delete(id string, logical bool) error {
	j, err := c.Get(id)
	if err != nil {
		return ErrJobDoesntExist
	}
	if logical && j.Deleted {
		return nil
	}
	running, err := j.isRunning()
	if err != nil {
		return err
	}
	if running {
		return errors.New("running job can not delete")
	}
	j.lock.Lock()
	defer j.lock.Unlock()

	if logical {
		j.Deleted = true
		err = c.Set(j)
	} else {
		err = c.jobDB.Delete(id)
	}
	if err != nil {
		err = fmt.Errorf("Error occurred while trying to delete job from db: %s", err)
		if c.PersistOnWrite {
			return err
		}
	}

	j.lock.Unlock()
	j.StopTimer()
	j.lock.Lock()

	go func() {
		Logger.Error(j.DeleteFromParentJobs(c)) // todo: review
	}()
	// Remove itself from dependent jobs as a parent job
	// and possibly delete child jobs if they don't have any other parents.
	go func() {
		Logger.Error(j.DeleteFromDependentJobs(c)) // todo: review
	}()
	Logger.Infof("Deleting %s", id)
	if !logical {
		c.jobs.Del(id)
	}
	return err
}

func (c *LockFreeJobCache) Enable(j *Job) error {
	return enable(j, c, c.PersistOnWrite)
}

// Disable stops a job from running by stopping its jobTimer. It also sets Job.Disabled to true,
// which is reflected in the UI.
func (c *LockFreeJobCache) Disable(j *Job) error {
	return disable(j, c, c.PersistOnWrite)
}

func (c *LockFreeJobCache) Persist() error {
	jm := c.GetAll()
	for _, j := range jm.Jobs {
		j.lock.RLock()
		err := c.jobDB.Save(j)
		if err != nil {
			j.lock.RUnlock()
			return err
		}
		j.lock.RUnlock()
	}
	return nil
}

func (c *LockFreeJobCache) PersistEvery(persistWaitTime time.Duration) {
	wait := time.NewTicker(persistWaitTime).C
	var err error
	for {
		<-wait
		err = c.Persist()
		if err != nil {
			Logger.Errorf("Error occurred persisting the database. Err: %s", err)
		}
	}
}

func (c *LockFreeJobCache) locateJobStatsIndexForRetention(stats []*types.JobStat) (marker int) {
	now := time.Now()
	expiresAt := now.Add(-c.retentionPeriod)
	pos := -1
	for i, el := range stats {
		diff := el.RanAt.Sub(expiresAt)
		if diff < 0 {
			pos = i
		}
	}
	return pos
}

func (c *LockFreeJobCache) Retain() error {
	for el := range c.jobs.Iter() {
		job := el.Value.(*Job)
		c.compactJobStats(job)
	}
	return nil
}

func (c *LockFreeJobCache) compactJobStats(job *Job) {
	job.lock.Lock()
	defer job.lock.Unlock()
	pos := c.locateJobStatsIndexForRetention(job.Stats)
	if pos >= 0 {
		Logger.Infof("JobStats TTL: removing %d items", pos+1)
		tmp := make([]*types.JobStat, len(job.Stats)-pos-1)
		copy(tmp, job.Stats[pos+1:])
		job.Stats = tmp
	}
}

func (c *LockFreeJobCache) RetainEvery(retentionWaitTime time.Duration) {
	wait := time.NewTicker(retentionWaitTime).C
	var err error
	for {
		<-wait
		err = c.Retain()
		if err != nil {
			Logger.Errorf("Error occurred during invoking retention. Err: %s", err)
		}
	}
}

func SetForTest(redisPool *redis.Pool) {
	pool = redisPool
}
