package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type statusResponse struct {
	Ok bool `json:"ok"`
}

// GetStatus is the handler for the 'GET /api/v1/status' endpoint.
func GetStatus(c *gin.Context, isFaultDetected bool) {
	status := statusResponse{
		Ok: !isFaultDetected,
	}
	c.IndentedJSON(http.StatusOK, status)
}
