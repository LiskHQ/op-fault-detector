package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type statusResponse struct {
	Status string `json:"status"`
}

// GetStatus is the handler for the 'GET /api/v1/status' endpoint.
func GetStatus(c *gin.Context, b bool) {
	var s statusResponse
	if b {
		s.Status = "ok"
	} else {
		s.Status = "not"
	}
	c.IndentedJSON(http.StatusOK, s)
}
