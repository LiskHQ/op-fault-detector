package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type statusResponse struct {
	Status string `json:"status"`
}

// GetStatus is the handler for the 'GET /api/v1/status' endpoint.
func GetStatus(c *gin.Context) {
	status := statusResponse{
		Status: "ok",
	}

	c.IndentedJSON(http.StatusOK, status)
}
