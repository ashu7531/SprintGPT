package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ashutosh/sprintgpt-backend/internal/azuredevops"
	"github.com/ashutosh/sprintgpt-backend/internal/database"
	"github.com/ashutosh/sprintgpt-backend/internal/rag"
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

// UserConfigReq represents configuration request payload
type UserConfigReq struct {
	Organization string `json:"organization" binding:"required"`
	Project      string `json:"project" binding:"required"`
	PAT          string `json:"pat" binding:"required"`
}

// SaveUserConfig saves user-specific Azure DevOps credentials in the database.
func SaveUserConfig() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get("userID")
		userID := "00000000-0000-0000-0000-000000000000"
		if exists {
			if s, ok := userIDVal.(string); ok && s != "" {
				userID = s
			}
		}

		if database.Conn == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection not available"})
			return
		}

		var req UserConfigReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		query := `
			INSERT INTO user_configs (user_id, organization, project, pat, updated_at)
			VALUES ($1, $2, $3, $4, NOW())
			ON CONFLICT (user_id) DO UPDATE
			SET organization = EXCLUDED.organization,
			    project = EXCLUDED.project,
			    pat = EXCLUDED.pat,
			    updated_at = NOW()
		`
		_, err := database.Conn.Exec(context.Background(), query, userID, req.Organization, req.Project, req.PAT)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save configuration: " + err.Error()})
			return
		}

		// Check if document chunks already exist for this organization and project
		var wikiExists bool
		checkQuery := `SELECT EXISTS(SELECT 1 FROM document_chunks WHERE organization = $1 AND project = $2)`
		err = database.Conn.QueryRow(context.Background(), checkQuery, req.Organization, req.Project).Scan(&wikiExists)
		if err == nil && !wikiExists {
			// Trigger background wiki ingestion for this new project onboarding
			go func(org, proj, pat string) {
				log.Printf("📥 [Auto-Ingest] Fetching wiki pages for new project %s/%s in background...", org, proj)
				client := azuredevops.NewClient(org, proj, pat)
				pages, err := client.FetchAllWikiPages()
				if err != nil {
					log.Printf("❌ [Auto-Ingest] Failed to fetch wiki pages for %s/%s: %v", org, proj, err)
					return
				}
				if len(pages) == 0 {
					log.Printf("📥 [Auto-Ingest] No wiki pages found for %s/%s", org, proj)
					return
				}
				log.Printf("📥 [Auto-Ingest] Found %d wiki pages for %s/%s. Chunking and embedding...", len(pages), org, proj)
				for _, page := range pages {
					err = rag.IngestDocument(context.Background(), org, proj, page.Title, page.Content, page.URL)
					if err != nil {
						log.Printf("❌ [Auto-Ingest] Failed to ingest page '%s': %v", page.Title, err)
					}
				}
				log.Printf("✅ [Auto-Ingest] Finished onboarding wiki for project %s/%s!", org, proj)
			}(req.Organization, req.Project, req.PAT)
		}

		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Configuration saved successfully"})
	}
}

// GetUserConfig fetches the saved user-specific Azure DevOps credentials from the database.
func GetUserConfig() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get("userID")
		userID := "00000000-0000-0000-0000-000000000000"
		if exists {
			if s, ok := userIDVal.(string); ok && s != "" {
				userID = s
			}
		}

		if database.Conn == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection not available"})
			return
		}

		query := `SELECT organization, project, pat FROM user_configs WHERE user_id = $1`
		var org, proj, pat string
		err := database.Conn.QueryRow(context.Background(), query, userID).Scan(&org, &proj, &pat)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"organization": "",
				"project":      "",
				"pat":          "",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"organization": org,
			"project":      proj,
			"pat":          pat,
		})
	}
}
