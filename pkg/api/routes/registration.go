// Package routes implements the logic to register the API endpoints against their corresponding handlers defined under the handlers package.
package routes

import (
	"github.com/LiskHQ/op-fault-detector/pkg/api/handlers"
	"github.com/LiskHQ/op-fault-detector/pkg/log"
	"github.com/gin-gonic/gin"
)

// RegisterHandlers is responsible to register all handlers for routes without any base path.
func RegisterHandlers(logger log.Logger, router *gin.Engine) {
	router.GET("/ping", handlers.GetPing)
}
