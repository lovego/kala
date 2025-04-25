package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/lovego/errs"
	"github.com/lovego/goa"
	"github.com/lovego/kala/job"
	"github.com/lovego/kala/types"
)

const (
	ApiJobPath = types.ApiUrlPrefix + types.JobPath

	contentType     = "Content-Type"
	jsonContentType = "application/json;charset=UTF-8"

	MAX_BODY_SIZE       = 1048576
	READ_HEADER_TIMEOUT = 0
)

// HandleKalaStatsRequest is the handler for getting system-level metrics
// /api/v1/stats
func HandleKalaStatsRequest(cache job.JobCache) func(c *goa.Context) {
	return func(c *goa.Context) {
		c.StatusJson(http.StatusOK, &types.KalaStatsResponse{Stats: job.NewKalaStats(cache)})
	}
}

// HandleListJobStatsRequest is the handler for getting job-specific stats
// /api/v1/job/stats/{id}
func HandleListJobStatsRequest(cache job.JobCache) func(c *goa.Context) {
	return func(c *goa.Context) {
		id := c.Param(0)
		j, err := cache.Get(id)
		if err != nil || j == nil {
			c.WriteHeader(http.StatusNotFound)
			return
		}
		c.StatusJson(http.StatusOK, &types.ListJobStatsResponse{JobStats: j.Stats})
	}
}

// HandleListJobs responds with an array of all Jobs within the server,
// active or disabled.
func HandleListJobsRequest(cache job.JobCache) func(c *goa.Context) {
	return func(c *goa.Context) {
		allJobs := cache.GetAll()
		allJobs.Lock.RLock()
		defer allJobs.Lock.RUnlock()

		resp := &types.ListJobsResponse{Jobs: make(map[string]*types.Job)}
		for k := range allJobs.Jobs {
			resp.Jobs[k] = allJobs.Jobs[k].Job
		}
		c.StatusJson(http.StatusOK, resp)
	}
}

// HandleAddJob takes a job object and unmarshals it to a Job type,
// and then throws the job in the schedulers.
func HandleAddJob(cache job.JobCache, defaultOwner string) func(*goa.Context) {
	return func(c *goa.Context) {
		body, err := c.RequestBody()
		if err != nil {
			errorEncodeJSON(errs.Trace(err), http.StatusBadRequest, c.ResponseWriter)
			return
		}
		newJob := &job.Job{}
		if err := json.Unmarshal(body, newJob); err != nil {
			errorEncodeJSON(errs.Trace(err), http.StatusBadRequest, c.ResponseWriter)
			return
		}
		if defaultOwner != "" && newJob.Owner == "" {
			newJob.Owner = defaultOwner
		}
		err = newJob.Init(cache)
		if err != nil {
			errorEncodeJSON(errs.Trace(err), http.StatusInternalServerError, c.ResponseWriter)
			return
		}
		c.StatusJson(http.StatusCreated, &types.AddJobResponse{Id: newJob.Id})
	}
}

// HandleJobGetRequest routes requests to /api/v1/job/{id} to GET a job.
func HandleJobGetRequest(cache job.JobCache) func(c *goa.Context) {
	return func(c *goa.Context) {
		id := c.Param(0)

		j, err := cache.Get(id)
		if err != nil {
			errorEncodeJSON(errs.Trace(err), http.StatusNotFound, c.ResponseWriter)
			return
		}
		if j == nil {
			c.WriteHeader(http.StatusNoContent)
			return
		}
		c.StatusJson(http.StatusOK, &types.JobResponse{Job: j.Job})
	}
}

// HandleJobRequest routes requests to /api/v1/job/{id} to DELETE a job.
func HandleJobDeleteRequest(cache job.JobCache) func(c *goa.Context) {
	return func(c *goa.Context) {
		id := c.Param(0)

		j, err := cache.Get(id)
		if err != nil {
			errorEncodeJSON(errs.Trace(err), http.StatusBadRequest, c.ResponseWriter)
			return
		}
		if j == nil {
			c.WriteHeader(http.StatusNotFound)
			return
		}
		if err := j.Delete(cache); err != nil {
			errorEncodeJSON(errs.Trace(err), http.StatusInternalServerError, c.ResponseWriter)
		} else {
			c.WriteHeader(http.StatusNoContent)
		}
	}
}

// HandleDeleteAllJobs is the handler for deleting all jobs
// DELETE /api/v1/job/all
func HandleDeleteAllJobs(cache job.JobCache) func(c *goa.Context) {
	return func(c *goa.Context) {
		if err := job.DeleteAll(cache); err != nil {
			errorEncodeJSON(errs.Trace(err), http.StatusInternalServerError, c.ResponseWriter)
		} else {
			c.WriteHeader(http.StatusNoContent)
		}
	}
}

// HandleStartJobRequest is the handler for manually starting jobs
// /api/v1/job/start/{id}
func HandleStartJobRequest(cache job.JobCache) func(c *goa.Context) {
	return func(c *goa.Context) {
		id := c.Param(0)
		j, err := cache.Get(id)
		if err != nil || j == nil {
			c.WriteHeader(http.StatusNotFound)
			return
		}

		j.StopTimer()
		j.Run(cache)
		c.WriteHeader(http.StatusNoContent)
	}
}

// HandleDisableJobRequest is the handler for mdisabling jobs
// /api/v1/job/disable/{id}
func HandleDisableJobRequest(cache job.JobCache) func(c *goa.Context) {
	return func(c *goa.Context) {
		id := c.Param(0)
		j, err := cache.Get(id)
		if err != nil || j == nil {
			c.WriteHeader(http.StatusNotFound)
			return
		}

		if err := j.Disable(cache); err != nil {
			errorEncodeJSON(errs.Trace(err), http.StatusInternalServerError, c.ResponseWriter)
			return
		}
		c.WriteHeader(http.StatusNoContent)
	}
}

// HandleEnableJobRequest is the handler for enable jobs
// /api/v1/job/enable/{id}
func HandleEnableJobRequest(cache job.JobCache) func(c *goa.Context) {
	return func(c *goa.Context) {
		id := c.Param(0)
		j, err := cache.Get(id)
		if err != nil || j == nil {
			c.WriteHeader(http.StatusNotFound)
			return
		}

		if err := j.Enable(cache); err != nil {
			errorEncodeJSON(errs.Trace(err), http.StatusInternalServerError, c.ResponseWriter)
			return
		}
		c.WriteHeader(http.StatusNoContent)
	}
}

type apiError struct {
	Error string `json:"error"`
}

func errorEncodeJSON(errToEncode error, status int, w http.ResponseWriter) {
	log.Println(errs.WithStack(errToEncode))
	js, err := json.Marshal(apiError{Error: errToEncode.Error()})
	if err != nil {
		return
	}
	w.Header().Set(contentType, jsonContentType)
	http.Error(w, string(js), status)
}

// SetupApiRoutes is used within main to initialize all of the routes
func SetupApiRoutes(router *goa.RouterGroup, cache job.JobCache, defaultOwner string) {
	// Route for creating a job
	router.Post(types.JobPath, HandleAddJob(cache, defaultOwner))
	// Route for deleting all jobs
	router.Delete(types.JobPath+"/all", HandleDeleteAllJobs(cache))
	// Route for deleting a job
	router.Delete(types.JobPath+`/(\S{36})`, HandleJobDeleteRequest(cache))
	// Route for getting a job
	router.Get(types.JobPath+`/(\S{36})`, HandleJobGetRequest(cache))
	// Route for getting job stats
	router.Get(types.JobPath+`/stats/(\S{36})`, HandleListJobStatsRequest(cache))
	// Route for listing all jops
	router.Get(types.JobPath, HandleListJobsRequest(cache))
	// Route for manually start a job
	router.Post(types.JobPath+`/start/(\S{36})`, HandleStartJobRequest(cache))
	// Route for manually start a job
	router.Post(types.JobPath+`/enable/(\S{36})`, HandleEnableJobRequest(cache))
	// Route for manually disable a job
	router.Post(types.JobPath+`/disable/(\S{36})`, HandleDisableJobRequest(cache))
	// Route for getting app-level metrics
	router.Get(`/stats`, HandleKalaStatsRequest(cache))
}
