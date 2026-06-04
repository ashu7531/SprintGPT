package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ashutosh/sprintgpt-backend/internal/rag"
)

type IngestRequest struct {
	Organization string `json:"organization" binding:"required"`
	Project      string `json:"project" binding:"required"`
	Title        string `json:"title" binding:"required"`
	Content      string `json:"content" binding:"required"`
	URL          string `json:"url"`
}

func HandleIngest() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req IngestRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := rag.IngestDocument(c.Request.Context(), req.Organization, req.Project, req.Title, req.Content, req.URL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":      "Successfully ingested document",
			"title":        req.Title,
			"organization": req.Organization,
			"project":      req.Project,
		})
	}
}
