// Package api implements http server, handlers and scaffolding around http server and router
package api

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	v1 "github.com/LiskHQ/op-fault-detector/pkg/api/handlers/v1"
	"github.com/LiskHQ/op-fault-detector/pkg/api/middlewares"
	"github.com/LiskHQ/op-fault-detector/pkg/api/routes"
	"github.com/LiskHQ/op-fault-detector/pkg/config"
	"github.com/LiskHQ/op-fault-detector/pkg/faultdetector"
	"github.com/LiskHQ/op-fault-detector/pkg/log"
	"github.com/gin-gonic/gin"
)

var areVersionsHandlerRegistered bool = false

// HTTPServer embeds the http.Server along with the various other properties.
type HTTPServer struct {
	server    *http.Server
	router    *gin.Engine
	ctx       context.Context
	logger    log.Logger
	wg        *sync.WaitGroup
	errorChan chan error
}

// RegisterHandlersForVersion is responsible to register API version specific route handlers.
func (w *HTTPServer) RegisterHandlersForVersion(fd *faultdetector.FaultDetector, versions []string, basePath string) {
	baseGroup := w.router.Group(basePath)
	for _, version := range versions {
		group := baseGroup.Group(version)
		switch version {
		case "v1":
			group.GET("/status", func(c *gin.Context) {
				v1.GetStatus(c, fd.IsFaultDetected())
			})

		default:
			w.logger.Warningf("No routes and handlers defined for version %s. Please verify the API config.", version)
		}
	}
	areVersionsHandlerRegistered = true
}

// Start starts the HTTP API server.
func (w *HTTPServer) Start() {
	defer w.wg.Done()

	if !areVersionsHandlerRegistered {
		w.errorChan <- fmt.Errorf("API specific versions handler are not registered")
	}

	w.logger.Infof("Starting the HTTP server on %s.", w.server.Addr)
	err := w.server.ListenAndServe()
	if err != nil {
		w.errorChan <- err
	}
}

// Stop gracefully shuts down the HTTP API server.
func (w *HTTPServer) Stop() error {
	err := w.server.Shutdown(w.ctx)
	if err == nil {
		w.logger.Infof("Successfully stopped the HTTP server.")
	}

	return err
}

func (w *HTTPServer) RegisterHandler(httpMethod string, relativePath string, h http.Handler) {
	w.router.Handle(httpMethod, relativePath, gin.WrapH(h))
}

func getGinModeFromSysLogLevel(sysLogLevel string) string {
	ginMode := gin.DebugMode // Default mode

	if sysLogLevel != "debug" && sysLogLevel != "trace" {
		ginMode = gin.ReleaseMode
	}

	return ginMode
}

// NewHTTPServer creates a router instance and sets up the necessary routes/handlers.
func NewHTTPServer(ctx context.Context, logger log.Logger, wg *sync.WaitGroup, config *config.Config, errorChan chan error) *HTTPServer {
	gin.SetMode(getGinModeFromSysLogLevel(config.System.LogLevel))

	router := gin.Default()

	// Global middlewares
	router.Use(middlewares.Authenticate())

	// Register handlers for routes without any base path
	logger.Debug("Registering handlers for non-versioned endpoints.")

	routes.RegisterHandlers(logger, router)

	host := config.Api.Server.Host
	port := config.Api.Server.Port
	addr := fmt.Sprintf("%s:%d", host, port)

	server := &HTTPServer{
		&http.Server{
			Addr:              addr,
			Handler:           router,
			ReadHeaderTimeout: 10 * time.Second,
		},
		router,
		ctx,
		logger,
		wg,
		errorChan,
	}

	return server
}
