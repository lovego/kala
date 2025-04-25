package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/lovego/goa"
	"github.com/lovego/kala/job"
	"github.com/lovego/kala/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func init() {
	job.SetForTest(&redis.Pool{
		MaxIdle:     2,
		MaxActive:   10,
		IdleTimeout: 600 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(
				"redis://:@localhost:6379/0",
				redis.DialConnectTimeout(3*time.Second),
				redis.DialReadTimeout(3*time.Second),
				redis.DialWriteTimeout(3*time.Second),
			)
		},
	})
}

func generateNewJobMap() map[string]string {
	scheduleTime := time.Now().Add(time.Minute * 5)
	repeat := 1
	delay := "P1DT10M10S"
	parsedTime := scheduleTime.Format(time.RFC3339)
	scheduleStr := fmt.Sprintf("R%d/%s/%s", repeat, parsedTime, delay)

	return map[string]string{
		"schedule": scheduleStr,
		"name":     "mock_job",
		"command":  "bash -c 'date'",
		"owner":    "example@example.com",
	}
}

func generateNewRemoteJobMap() map[string]interface{} {
	return map[string]interface{}{
		"name":     "mock_remote_job",
		"owner":    "example@example.com",
		"job_type": types.RemoteJob,
		"remote_properties": map[string]string{
			"url": "http://example.com",
		},
	}
}

func generateJobAndCache() (*job.LockFreeJobCache, *job.Job) {
	cache := job.NewMockCache()
	j := job.GetMockJobWithGenericSchedule(time.Now())
	j.Init(cache)
	return cache, j
}

type ApiTestSuite struct {
	suite.Suite
}

func TestApiTestSuite(t *testing.T) {
	suite.Run(t, new(ApiTestSuite))
}

func (a *ApiTestSuite) TestHandleAddJob() {
	t := a.T()
	cache := job.NewMockCache()
	jobMap := generateNewJobMap()
	jobMap["owner"] = ""
	defaultOwner := "aj+tester@ajvb.me"
	handler := HandleAddJob(cache, defaultOwner)

	jsonJobMap, err := json.Marshal(jobMap)
	a.NoError(err)
	ctx := setupTestReq(t, "POST", ApiJobPath, jsonJobMap)
	handler(ctx)

	var addJobResp types.AddJobResponse
	err = json.Unmarshal(ctx.ResponseBody(), &addJobResp)
	a.NoError(err)
	retrievedJob, err := cache.Get(addJobResp.Id)
	a.NoError(err)
	a.Equal(jobMap["name"], retrievedJob.Name)
	a.NotEqual(jobMap["owner"], retrievedJob.Owner)
	a.Equal(defaultOwner, retrievedJob.Owner)
	// a.Equal(ctx.Status(), http.StatusCreated)
}

func (a *ApiTestSuite) TestHandleAddRemoteJob() {
	t := a.T()
	cache := job.NewMockCache()
	jobMap := generateNewRemoteJobMap()
	jobMap["owner"] = ""
	defaultOwner := "aj+tester@ajvb.me"
	handler := HandleAddJob(cache, defaultOwner)

	jsonJobMap, err := json.Marshal(jobMap)
	a.NoError(err)
	ctx := setupTestReq(t, "POST", ApiJobPath, jsonJobMap)
	handler(ctx)

	var addJobResp types.AddJobResponse
	err = json.Unmarshal(ctx.ResponseBody(), &addJobResp)
	a.NoError(err)
	retrievedJob, err := cache.Get(addJobResp.Id)
	a.NoError(err)
	a.Equal(jobMap["name"], retrievedJob.Name)
	a.NotEqual(jobMap["owner"], retrievedJob.Owner)
	a.Equal(defaultOwner, retrievedJob.Owner)
	// a.Equal(ctx.Status(), http.StatusCreated)
}

func (a *ApiTestSuite) TestHandleAddJobFailureBadJson() {
	t := a.T()
	cache := job.NewMockCache()
	handler := HandleAddJob(cache, "")

	ctx := setupTestReq(t, "POST", ApiJobPath, []byte("asd"))
	handler(ctx)
	// a.Equal(ctx.Status(), http.StatusBadRequest)
}

func (a *ApiTestSuite) TestHandleAddJobFailureBadSchedule() {
	t := a.T()
	cache := job.NewMockCache()
	jobMap := generateNewJobMap()
	handler := HandleAddJob(cache, "")

	// Mess up schedule
	jobMap["schedule"] = "asdf"

	jsonJobMap, err := json.Marshal(jobMap)
	a.NoError(err)
	ctx := setupTestReq(t, "POST", ApiJobPath, jsonJobMap)
	handler(ctx)
	// a.Equal(ctx.Status(), http.StatusBadRequest)
	var respErr apiError
	err = json.Unmarshal(ctx.ResponseBody(), &respErr)
	a.NoError(err)
	a.True(strings.Contains(respErr.Error, "Schedule not formatted correctly"))
}

func (a *ApiTestSuite) TestDeleteJobSuccess() {
	t := a.T()
	cache, j := generateJobAndCache()

	router := goa.New()
	router.Delete(types.JobPath+`/(\S{36})`, HandleJobDeleteRequest(cache))
	router.Get(types.JobPath+`/(\S{36})`, HandleJobGetRequest(cache))
	ts := httptest.NewServer(router)

	ctx := setupTestReq(t, "DELETE", ts.URL+types.JobPath+"/"+j.Id, nil)

	client := &http.Client{}
	resp, err := client.Do(ctx.Request)
	a.NoError(err)
	a.Equal(resp.StatusCode, http.StatusNoContent)

	a.Nil(cache.Get(j.Id))
}

func (a *ApiTestSuite) TestDeleteAllJobsSuccess() {
	t := a.T()
	cache, jobOne := generateJobAndCache()
	jobTwo := job.GetMockJobWithGenericSchedule(time.Now())
	jobTwo.Init(cache)

	router := goa.New()
	router.Delete(types.JobPath+"/all", HandleDeleteAllJobs(cache))
	ts := httptest.NewServer(router)

	ctx := setupTestReq(t, "DELETE", ts.URL+types.JobPath+"/all", nil)

	client := &http.Client{}
	resp, err := client.Do(ctx.Request)
	a.NoError(err)
	a.Equal(resp.StatusCode, http.StatusNoContent)

	a.Equal(0, len(cache.GetAll().Jobs))
	a.Nil(cache.Get(jobOne.Id))
	a.Nil(cache.Get(jobTwo.Id))
}

func (a *ApiTestSuite) TestHandleJobRequestJobDoesNotExist() {
	t := a.T()
	cache := job.NewMockCache()

	router := goa.New()
	router.Delete(types.JobPath+`/(\S{36})`, HandleJobDeleteRequest(cache))
	router.Get(types.JobPath+`/(\S{36})`, HandleJobGetRequest(cache))
	ts := httptest.NewServer(router)

	ctx := setupTestReq(t, "DELETE", ts.URL+types.JobPath+"/not-a-real-id", nil)

	client := &http.Client{}
	resp, err := client.Do(ctx.Request)
	a.NoError(err)

	a.Equal(resp.StatusCode, http.StatusNotFound)
}

func (a *ApiTestSuite) TestGetJobSuccess() {
	t := a.T()
	cache, j := generateJobAndCache()

	router := goa.New()
	router.Delete(types.JobPath+`/(\S{36})`, HandleJobDeleteRequest(cache))
	router.Get(types.JobPath+`/(\S{36})`, HandleJobGetRequest(cache))
	ts := httptest.NewServer(router)

	ctx := setupTestReq(t, "GET", ts.URL+types.JobPath+"/"+j.Id, nil)

	client := &http.Client{}
	resp, err := client.Do(ctx.Request)
	a.NoError(err)

	var jobResp types.JobResponse
	body, err := io.ReadAll(resp.Body)
	a.NoError(err)
	resp.Body.Close()
	err = json.Unmarshal(body, &jobResp)
	a.NoError(err)
	a.Equal(j.Id, jobResp.Job.Id)
	a.Equal(j.Owner, jobResp.Job.Owner)
	a.Equal(j.Name, jobResp.Job.Name)
	a.Equal(resp.StatusCode, http.StatusOK)
}

func (a *ApiTestSuite) TestHandleListJobStatsRequest() {
	cache, j := generateJobAndCache()
	j.Run(cache)

	router := goa.New()
	router.Get(types.JobPath+`/stats/(\S{36})`, HandleListJobStatsRequest(cache))
	ts := httptest.NewServer(router)

	ctx := setupTestReq(a.T(), "GET", ts.URL+types.JobPath+"/stats/"+j.Id, nil)

	client := &http.Client{}
	resp, err := client.Do(ctx.Request)
	a.NoError(err)

	var jobStatsResp types.ListJobStatsResponse
	body, err := io.ReadAll(resp.Body)
	a.NoError(err)
	resp.Body.Close()
	err = json.Unmarshal(body, &jobStatsResp)
	a.NoError(err)

	a.Equal(len(jobStatsResp.JobStats), 1)
	a.Equal(jobStatsResp.JobStats[0].JobId, j.Id)
	a.Equal(jobStatsResp.JobStats[0].NumberOfRetries, uint(0))
	a.True(jobStatsResp.JobStats[0].Success)
}

func (a *ApiTestSuite) TestHandleListJobStatsRequestNotFound() {
	cache, _ := generateJobAndCache()
	router := goa.New()

	router.Get(types.JobPath+`/stats/(\S{36})`, HandleListJobStatsRequest(cache))
	ts := httptest.NewServer(router)

	ctx := setupTestReq(a.T(), "GET", ts.URL+types.JobPath+"/stats/not-a-real-id", nil)

	client := &http.Client{}
	resp, err := client.Do(ctx.Request)
	a.NoError(err)

	a.Equal(resp.StatusCode, http.StatusNotFound)
}

func (a *ApiTestSuite) TestHandleListJobsRequest() {
	cache, jobOne := generateJobAndCache()
	jobTwo := job.GetMockJobWithGenericSchedule(time.Now())
	jobTwo.Init(cache)

	router := goa.New()
	router.Get(types.JobPath, HandleListJobsRequest(cache))
	ts := httptest.NewServer(router)

	ctx := setupTestReq(a.T(), "GET", ts.URL+types.JobPath, nil)

	client := &http.Client{}
	resp, err := client.Do(ctx.Request)
	a.NoError(err)

	var jobsResp types.ListJobsResponse
	unmarshallRequestBody(a.T(), resp, &jobsResp)

	a.Equal(len(jobsResp.Jobs), 2)
	a.Equal(jobsResp.Jobs[jobOne.Id].Schedule, jobOne.Schedule)
	a.Equal(jobsResp.Jobs[jobOne.Id].Name, jobOne.Name)
	a.Equal(jobsResp.Jobs[jobOne.Id].Owner, jobOne.Owner)
	a.Equal(jobsResp.Jobs[jobOne.Id].Command, jobOne.Command)

	a.Equal(jobsResp.Jobs[jobTwo.Id].Schedule, jobTwo.Schedule)
	a.Equal(jobsResp.Jobs[jobTwo.Id].Name, jobTwo.Name)
	a.Equal(jobsResp.Jobs[jobTwo.Id].Owner, jobTwo.Owner)
	a.Equal(jobsResp.Jobs[jobTwo.Id].Command, jobTwo.Command)
}

func (a *ApiTestSuite) TestHandleStartJobRequest() {
	t := a.T()
	cache, j := generateJobAndCache()
	router := goa.New()
	router.Post(types.JobPath+`/start/(\S{36})`, HandleStartJobRequest(cache))
	ts := httptest.NewServer(router)

	ctx := setupTestReq(t, "POST", ts.URL+types.JobPath+"/start/"+j.Id, nil)

	client := &http.Client{}
	resp, err := client.Do(ctx.Request)
	a.NoError(err)

	now := time.Now()

	a.Equal(resp.StatusCode, http.StatusNoContent)

	a.Equal(j.Metadata.SuccessCount, uint(1))
	a.WithinDuration(j.Metadata.LastSuccess, now, 2*time.Second)
	a.WithinDuration(j.Metadata.LastAttemptedRun, now, 2*time.Second)
}

func (a *ApiTestSuite) TestHandleStartJobRequestNotFound() {
	t := a.T()
	cache := job.NewMockCache()
	handler := HandleStartJobRequest(cache)
	ctx := setupTestReq(t, "POST", ApiJobPath+"/start/asdasd", nil)
	handler(ctx)
	// a.Equal(ctx.Status(), http.StatusNotFound)
}

func (a *ApiTestSuite) TestHandleEnableJobRequest() {
	t := a.T()
	cache, j := generateJobAndCache()
	router := goa.New()
	router.Post(types.JobPath+`/enable/(\S{36})`, HandleEnableJobRequest(cache))
	ts := httptest.NewServer(router)

	a.NoError(j.Disable(cache))

	ctx := setupTestReq(t, "POST", ts.URL+types.JobPath+"/enable/"+j.Id, nil)

	client := &http.Client{}
	resp, err := client.Do(ctx.Request)
	a.NoError(err)

	a.Equal(http.StatusNoContent, resp.StatusCode)

	a.Equal(false, j.Disabled)
}

func (a *ApiTestSuite) TestHandleEnableJobRequestNotFound() {
	t := a.T()
	cache := job.NewMockCache()
	handler := HandleEnableJobRequest(cache)
	ctx := setupTestReq(t, "POST", types.JobPath+"/enable/asdasd", nil)
	handler(ctx)
	// a.Equal(ctx.Status(), http.StatusNotFound)
}

func (a *ApiTestSuite) TestHandleDisableJobRequest() {
	t := a.T()
	cache, j := generateJobAndCache()
	router := goa.New()
	router.Post(types.JobPath+`/disable/(\S{36})`, HandleDisableJobRequest(cache))
	ts := httptest.NewServer(router)

	ctx := setupTestReq(t, "POST", ts.URL+types.JobPath+"/disable/"+j.Id, nil)

	client := &http.Client{}
	resp, err := client.Do(ctx.Request)
	a.NoError(err)

	a.Equal(http.StatusNoContent, resp.StatusCode)

	a.Equal(true, j.Disabled)
}

func (a *ApiTestSuite) TestHandleDisableJobRequestNotFound() {
	t := a.T()
	cache := job.NewMockCache()
	handler := HandleDisableJobRequest(cache)
	ctx := setupTestReq(t, "POST", ApiJobPath+"/disable/asdasd", nil)
	handler(ctx)
	// a.Equal(ctx.Status(), http.StatusNotFound)
}

func (a *ApiTestSuite) TestHandleKalaStatsRequest() {
	cache, _ := generateJobAndCache()
	jobTwo := job.GetMockJobWithGenericSchedule(time.Now())
	jobTwo.Init(cache)
	jobTwo.Run(cache)

	router := goa.New()
	router.Get(`/stats`, HandleKalaStatsRequest(cache))
	ts := httptest.NewServer(router)

	ctx := setupTestReq(a.T(), "GET", ts.URL+"/stats", nil)

	client := &http.Client{}
	resp, err := client.Do(ctx.Request)
	a.NoError(err)

	now := time.Now()

	var statsResp types.KalaStatsResponse
	unmarshallRequestBody(a.T(), resp, &statsResp)

	a.Equal(statsResp.Stats.Jobs, 2)
	a.Equal(statsResp.Stats.ActiveJobs, 2)
	a.Equal(statsResp.Stats.DisabledJobs, 0)

	a.Equal(statsResp.Stats.ErrorCount, uint(0))
	a.Equal(statsResp.Stats.SuccessCount, uint(1))

	a.WithinDuration(statsResp.Stats.LastAttemptedRun, now, 2*time.Second)
	a.WithinDuration(statsResp.Stats.CreatedAt, now, 2*time.Second)
}

func (a *ApiTestSuite) TestSetupApiRoutes() {
	cache := job.NewMockCache()
	router := goa.New()

	SetupApiRoutes(&router.RouterGroup, cache, "")

	a.NotNil(router)
	a.IsType(router, goa.New())
}

// setupTestReq constructs the writer recorder and request obj for use in tests
func setupTestReq(t assert.TestingT, method, path string, data []byte) *goa.Context {
	w := httptest.NewRecorder()
	req, err := http.NewRequest(method, path, bytes.NewReader(data))
	assert.NoError(t, err)
	return &goa.Context{
		ContextBeforeLookup: goa.ContextBeforeLookup{
			Request:        req,
			ResponseWriter: w,
		},
	}
}

func unmarshallRequestBody(t assert.TestingT, resp *http.Response, obj interface{}) {
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()
	err = json.Unmarshal(body, obj)
	assert.NoError(t, err)
}
