// Package handlers implements various route handlers.
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetPing is the handler for the 'GET /ping' endpoint.
func GetPing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
