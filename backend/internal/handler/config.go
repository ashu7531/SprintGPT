package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ashutosh/sprintgpt-backend/internal/azuredevops"
	"github.com/ashutosh/sprintgpt-backend/pkg/models"
)

// ValidateConfig handles Azure DevOps credential validation.
func ValidateConfig() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.ConfigValidation
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "Invalid request",
				Code:    http.StatusBadRequest,
				Details: err.Error(),
			})
			return
		}

		client := azuredevops.NewClient(req.Organization, req.Project, req.PAT)
		result, err := client.ValidateConnection()
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "Validation failed",
				Code:    http.StatusInternalServerError,
				Details: err.Error(),
			})
			return
		}

		if !result.Valid {
			c.JSON(http.StatusUnauthorized, result)
			return
		}

		c.JSON(http.StatusOK, result)
	}
}
