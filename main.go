package main

import (
	"net/http"

	"github.com/lovego/goa"
	"github.com/lovego/goa/server"
	"github.com/lovego/kala/api"
	"github.com/lovego/kala/job"
	"github.com/lovego/kala/job/storage/postgres"
	"github.com/lovego/kala/types"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.InfoLevel)
}

func main() {
	// Example db
	var db = postgres.New("postgres://postgres:postgres@localhost/database")

	// Create cache
	log.Infof("Preparing cache")
	cache := job.NewLockFreeJobCache(db)

	// Startup cache
	cache.Start(0, 0)

	router := goa.New()
	// Setup middlewares

	api.SetupApiRoutes(router.Group(types.ApiUrlPrefix), cache, "default")
	router.Use(func(c *goa.Context) {
		http.StripPrefix("/webui/", http.FileServer(api.AssetFS())).ServeHTTP(c.ResponseWriter, c.Request)
	})

	// Launch API server
	server.ListenAndServe(router)
}
