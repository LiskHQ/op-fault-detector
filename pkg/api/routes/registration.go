// Package routes implements the logic to register the API endpoints against their corresponding handlers defined under the handlers package.
package routes

import (
	"github.com/LiskHQ/op-fault-detector/pkg/api/handlers"
	v1 "github.com/LiskHQ/op-fault-detector/pkg/api/handlers/v1"
	"github.com/LiskHQ/op-fault-detector/pkg/log"
	"github.com/gin-gonic/gin"
)

// RegisterHandlers is responsible to register all handlers for routes without any base path.
func RegisterHandlers(logger log.Logger, router *gin.Engine) {
	router.GET("/ping", handlers.GetPing)
}

// RegisterHandlersByGroup is responsible to register all the handlers for routes that are prefixed under a specified base path as a routerGroup.
func RegisterHandlersByGroup(logger log.Logger, routerGroup *gin.RouterGroup, versions []string) {
	for _, version := range versions {
		RegisterHandlersForVersion(logger, routerGroup, version)
	}
}

// RegisterHandlersForVersion is responsible to register API version specific route handlers.
func RegisterHandlersForVersion(logger log.Logger, routerGroup *gin.RouterGroup, version string) {
	group := routerGroup.Group(version)

	switch version {
	case "v1":
		group.GET("/status", v1.GetStatus)

	default:
		logger.Warningf("No routes and handlers defined for version %s. Please verify the API config.", version)
	}
}
