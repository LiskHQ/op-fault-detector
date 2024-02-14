// Package api implements http server, handlers and scaffolding around http server and router
package api

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/LiskHQ/op-fault-detector/pkg/api/middlewares"
	"github.com/LiskHQ/op-fault-detector/pkg/api/routes"
	"github.com/LiskHQ/op-fault-detector/pkg/config"
	"github.com/LiskHQ/op-fault-detector/pkg/log"
	"github.com/gin-gonic/gin"
)

// HTTPServer embeds the http.Server along with the various other properties.
type HTTPServer struct {
	Server    *http.Server
	Router    *gin.Engine
	Ctx       context.Context
	Logger    log.Logger
	Wg        *sync.WaitGroup
	ErrorChan chan error
}

// Start starts the HTTP API server.
func (w *HTTPServer) Start() {
	defer w.Wg.Done()

	w.Logger.Infof("Starting the HTTP server on %s.", w.Server.Addr)
	err := w.Server.ListenAndServe()
	if err != nil {
		w.ErrorChan <- err
	}
}

// Stop gracefully shuts down the HTTP API server.
func (w *HTTPServer) Stop() error {
	err := w.Server.Shutdown(w.Ctx)
	if err == nil {
		w.Logger.Infof("Successfully stopped the HTTP server.")
	}

	return err
}

func (w *HTTPServer) RegisterHandler(httpMethod string, relativePath string, h http.Handler) {
	w.Router.Handle(httpMethod, relativePath, gin.WrapH(h))
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

	// Register handlers for routes following the base path
	basePath := config.Api.BasePath
	baseGroup := router.Group(basePath)
	logger.Debugf("Registering handlers for endpoints under path '%s'.", basePath)
	routes.RegisterHandlersByGroup(logger, baseGroup, config.Api.RegisterVersions)

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
