// Package middlewares implements all the necessary HTTP middlewares.
package middlewares

import (
	"github.com/gin-gonic/gin"
)

func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Run authentication checks
		// Add logic to be executed BEFORE the request is processed

		// Do not forget to include Next inside all the middlewares to ensure execution of the pending handlers in the chain inside the calling handler
		c.Next()

		// Add logic to be executed AFTER the request is processed
	}
}
