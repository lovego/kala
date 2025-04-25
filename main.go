package main

import (
	"net/http"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/lovego/goa"
	"github.com/lovego/goa/server"
	"github.com/lovego/kala/api"
	"github.com/lovego/kala/job"
	"github.com/lovego/kala/job/storage/postgres"
	"github.com/lovego/kala/types"
	"github.com/lovego/logger"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.InfoLevel)
}

func main() {
	// Example db
	var db = postgres.New("postgres://postgres:postgres@localhost/database")
	// Set job Logger
	job.Logger = logger.New(nil)

	// Create cache
	log.Infof("Preparing cache")
	cache := job.NewLockFreeJobCache(db)
	redisPool := &redis.Pool{
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
	}

	// Startup cache
	cache.Start(redisPool, 0, 0)

	router := goa.New()
	// Setup middlewares

	api.SetupApiRoutes(router.Group(types.ApiUrlPrefix), cache, "default")
	router.Use(func(c *goa.Context) {
		http.StripPrefix("/webui/", http.FileServer(api.AssetFS())).ServeHTTP(c.ResponseWriter, c.Request)
	})

	// Launch API server
	server.ListenAndServe(router)
}
