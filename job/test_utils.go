package job

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/lovego/kala/types"
	"github.com/lovego/kala/utils/iso8601"
)

type MockDBGetAll struct {
	MockDB
	response []*Job
}

func (d *MockDBGetAll) GetAll() ([]*Job, error) {
	return d.response, nil
}

type MockDB struct{}

func (m *MockDB) GetAll() ([]*Job, error) {
	return nil, nil
}
func (m *MockDB) Get(id string) (*Job, error) {
	return nil, nil
}
func (m *MockDB) Delete(id string) error {
	return nil
}
func (m *MockDB) Save(job *Job) error {
	return nil
}
func (m *MockDB) Close() error {
	return nil
}

func NewMockCache() *LockFreeJobCache {
	return NewLockFreeJobCache(&MockDB{})
}

func GetMockJob() *Job {
	return &Job{
		Job: &types.Job{
			Name:    "mock_job",
			Command: "bash -c 'date'",
			Owner:   "example@example.com",
			Retries: 2,
		},
	}
}

func GetMockFailingJob() *Job {
	return &Job{
		Job: &types.Job{
			Name:    "mock_failing_job",
			Command: "asdf",
			Owner:   "example@example.com",
			Retries: 2,
		},
	}
}

func GetMockRemoteJob(props types.RemoteProperties) *Job {
	return &Job{
		Job: &types.Job{
			Name:             "mock_remote_job",
			GroupName:        fmt.Sprintf("%d", time.Now().UnixNano()),
			Command:          "",
			JobType:          types.RemoteJob,
			RemoteProperties: props,
		},
	}
}

func GetMockJobWithSchedule(repeat int, scheduleTime time.Time, delay string) *Job {
	genericMockJob := GetMockJob()

	parsedTime := scheduleTime.Format(time.RFC3339)
	scheduleStr := fmt.Sprintf("R%d/%s/%s", repeat, parsedTime, delay)
	genericMockJob.Schedule = scheduleStr

	return genericMockJob
}

func GetMockRecurringJobWithSchedule(scheduleTime time.Time, delay string) *Job {
	genericMockJob := GetMockJob()

	parsedTime := scheduleTime.Format(time.RFC3339)
	scheduleStr := fmt.Sprintf("R/%s/%s", parsedTime, delay)
	genericMockJob.Schedule = scheduleStr
	genericMockJob.timesToRepeat = -1

	parsedDuration, _ := iso8601.FromString(delay)

	genericMockJob.delayDuration = parsedDuration

	return genericMockJob
}

func GetMockJobStats(oldDate time.Time, count int) []*types.JobStat {
	stats := make([]*types.JobStat, 0)
	for i := 1; i <= count; i++ {
		el := &types.JobStat{
			JobId:             "stats-id-" + fmt.Sprint(i),
			NumberOfRetries:   0,
			ExecutionDuration: 10000,
			Success:           true,
			RanAt:             oldDate,
		}
		stats = append(stats, el)
	}
	return stats
}

func GetMockJobWithGenericSchedule(now time.Time) *Job {
	fiveMinutesFromNow := now.Add(time.Minute * 5)
	return GetMockJobWithSchedule(2, fiveMinutesFromNow, "P1DT10M10S")
}

func parseTime(t *testing.T, value string) time.Time {
	now, err := time.Parse("2006-Jan-02 15:04", value)
	if err != nil {
		t.Fatal(err)
	}
	return now
}

func parseTimeInLocation(t *testing.T, value string, location string) time.Time {
	loc, err := time.LoadLocation(location)
	if err != nil {
		t.Fatal(err)
	}
	now, err := time.ParseInLocation("2006-Jan-02 15:04", value, loc)
	if err != nil {
		t.Fatal(err)
	}
	return now
}

// Used to hand off execution briefly so that jobs can run and so on.
func briefPause() {
	time.Sleep(time.Millisecond * 100)
}

func awaitJobRan(t *testing.T, j *Job, timeout time.Duration) {
	t.Helper()
	select {
	case <-j.ranChan:
	case <-time.After(timeout):
		t.Fatal("Job failed to run")
	}
}

var _ JobDB = (*MemoryDB)(nil)

type MemoryDB struct {
	m    map[string]*Job
	lock sync.RWMutex
}

func NewMemoryDB() *MemoryDB {
	return &MemoryDB{
		m: map[string]*Job{},
	}
}

func (m *MemoryDB) GetAll() (ret []*Job, _ error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	for _, v := range m.m {
		ret = append(ret, v)
	}
	return
}

func (m *MemoryDB) Get(id string) (*Job, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	j, exist := m.m[id]
	if !exist {
		return nil, ErrJobNotFound(id)
	}
	return j, nil
}

func (m *MemoryDB) Delete(id string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, exists := m.m[id]; !exists {
		return errors.New("Doesn't exist") // Used for testing
	}
	delete(m.m, id)
	// log.Printf("After delete: %+v", m)
	return nil
}

func (m *MemoryDB) Save(j *Job) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.m[j.Id] = j
	return nil
}

func (m *MemoryDB) Close() error {
	return nil
}
