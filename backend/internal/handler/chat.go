// Package handler contains HTTP request handlers for the API endpoints.
package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ashutosh/sprintgpt-backend/internal/azuredevops"
	"github.com/ashutosh/sprintgpt-backend/internal/formatter"
	"github.com/ashutosh/sprintgpt-backend/internal/intent"
	"github.com/ashutosh/sprintgpt-backend/pkg/models"
)

// ChatHandler handles incoming chat messages.
type ChatHandler struct {
	detector intent.Detector
}

// NewChatHandler creates a new chat handler with the given intent detector.
func NewChatHandler(detector intent.Detector) *ChatHandler {
	return &ChatHandler{
		detector: detector,
	}
}

// Handle processes a chat message: detect intent → call Azure DevOps → format response.
func (h *ChatHandler) Handle(c *gin.Context) {
	var req models.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Code:    http.StatusBadRequest,
			Details: err.Error(),
		})
		return
	}

	// Step 1: Detect intent from user message
	detected, err := h.detector.Detect(req.Message)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Failed to understand message",
			Code:    http.StatusBadRequest,
			Details: err.Error(),
		})
		return
	}

	// Step 2: Create Azure DevOps client
	adoClient := azuredevops.NewClient(
		req.Context.Organization,
		req.Context.Project,
		req.Context.PAT,
	)

	// Step 3: Execute based on detected intent
	var response models.ChatResponse
	response.Intent = detected.Intent
	response.Timestamp = time.Now()

	switch detected.Intent {
	case "task_status":
		response = h.handleTaskStatus(adoClient, detected, response)

	case "help":
		kd, ok := h.detector.(*intent.KeywordDetector)
		if ok {
			response.Response = kd.SupportedIntents()
		} else {
			response.Response = formatter.FormatHelp()
		}

	case "user_work":
		response = h.handleUserWork(adoClient, detected, response)

	case "active_bugs":
		response = h.handleActiveBugs(adoClient, detected, response)

	case "sprint_summary":
		response = h.handleSprintSummary(adoClient, detected, response)

	default:
		response.Response = formatter.FormatUnknownIntent(detected.Params["raw"])
	}

	c.JSON(http.StatusOK, response)
}

// handleTaskStatus fetches a work item by ID and formats the response.
func (h *ChatHandler) handleTaskStatus(client *azuredevops.Client, detected *models.DetectedIntent, response models.ChatResponse) models.ChatResponse {
	idStr, ok := detected.Params["id"]
	if !ok {
		response.Response = "I couldn't find a task ID in your message. Try: \"status of task 123\""
		return response
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Response = "Invalid task ID. Please provide a numeric ID like: \"status of task 123\""
		return response
	}

	workItem, err := client.GetWorkItem(id)
	if err != nil {
		response.Response = formatter.FormatError(err)
		return response
	}

	response.Response = formatter.FormatWorkItem(workItem)
	response.Data = workItem
	return response
}

func (h *ChatHandler) handleActiveBugs(client *azuredevops.Client, detected *models.DetectedIntent, response models.ChatResponse) models.ChatResponse {
	// WIQL query for active bugs
	wiql := `Select [System.Id], [System.Title], [System.State] From WorkItems Where [System.WorkItemType] = 'Bug' AND [System.State] NOT CONTAINS 'Closed' AND [System.State] NOT CONTAINS 'Done' AND [System.State] NOT CONTAINS 'Resolved' Order By [Microsoft.VSTS.Common.Priority] Asc`

	items, err := client.QueryWorkItems(wiql, 10)
	if err != nil {
		response.Response = formatter.FormatError(err)
		return response
	}

	response.Response = formatter.FormatWorkItemList("Active Bugs", items)
	response.Data = items
	return response
}

func (h *ChatHandler) handleUserWork(client *azuredevops.Client, detected *models.DetectedIntent, response models.ChatResponse) models.ChatResponse {
	userName, ok := detected.Params["user"]
	if !ok {
		response.Response = "I couldn't figure out who you are asking about."
		return response
	}

	// WIQL query for active items assigned to a user (fuzzy match)
	wiql := fmt.Sprintf(`Select [System.Id], [System.Title], [System.State] From WorkItems Where [System.AssignedTo] Contains '%s' AND [System.State] NOT CONTAINS 'Closed' AND [System.State] NOT CONTAINS 'Done' AND [System.State] NOT CONTAINS 'Resolved'`, userName)

	items, err := client.QueryWorkItems(wiql, 10)
	if err != nil {
		response.Response = formatter.FormatError(err)
		return response
	}

	title := fmt.Sprintf("Active Tasks for %s", userName)
	response.Response = formatter.FormatWorkItemList(title, items)
	response.Data = items
	return response
}

func (h *ChatHandler) handleSprintSummary(client *azuredevops.Client, detected *models.DetectedIntent, response models.ChatResponse) models.ChatResponse {
	// Query the 30 most recently modified items to figure out the active sprint and its stats
	wiql := `Select [System.Id], [System.Title], [System.State], [System.IterationPath] From WorkItems Order By [System.ChangedDate] Desc`

	items, err := client.QueryWorkItems(wiql, 30)
	if err != nil {
		response.Response = formatter.FormatError(err)
		return response
	}

	if len(items) == 0 {
		response.Response = "I couldn't find any recent tasks in this project to summarize."
		return response
	}

	// 1. Find the most common iteration path
	iterationCounts := make(map[string]int)
	for _, item := range items {
		iterationCounts[item.IterationPath]++
	}

	var currentSprint string
	maxCount := 0
	for path, count := range iterationCounts {
		if count > maxCount {
			maxCount = count
			currentSprint = path
		}
	}

	// 2. Calculate stats for this sprint
	var total, active, done, bugs int
	for _, item := range items {
		if item.IterationPath == currentSprint {
			total++

			state := item.State
			if state == "Closed" || state == "Done" || state == "Resolved" || state == "SIGNOFF" {
				done++
			} else {
				active++
			}

			if item.WorkItemType == "Bug" {
				bugs++
			}
		}
	}

	// 3. Format the summary
	var sb string
	sb += fmt.Sprintf("📊 **Sprint Summary: %s**\n\n", currentSprint)
	sb += fmt.Sprintf("Based on recent activity, here is the status of the current iteration:\n\n")
	sb += fmt.Sprintf("• **Total Items:** %d\n", total)
	sb += fmt.Sprintf("• **Completed:** %d\n", done)
	sb += fmt.Sprintf("• **Active/To Do:** %d\n", active)
	sb += fmt.Sprintf("• **Bugs:** %d\n", bugs)

	progress := 0
	if total > 0 {
		progress = (done * 100) / total
	}
	sb += fmt.Sprintf("\n**Progress:** %d%% completed.", progress)

	response.Response = sb
	response.Data = map[string]interface{}{
		"sprint": currentSprint,
		"total":  total,
		"done":   done,
		"active": active,
		"bugs":   bugs,
	}

	return response
}
