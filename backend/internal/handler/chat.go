// Package handler contains HTTP request handlers for the API endpoints.
package handler

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ashutosh/sprintgpt-backend/internal/azuredevops"
	"github.com/ashutosh/sprintgpt-backend/internal/formatter"
	"github.com/ashutosh/sprintgpt-backend/internal/intent"
	"github.com/ashutosh/sprintgpt-backend/internal/rag"
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

	org := req.Context.Organization
	if org == "" {
		org = os.Getenv("AZURE_ORG")
	}
	proj := req.Context.Project
	if proj == "" {
		proj = os.Getenv("AZURE_PROJECT")
	}
	pat := req.Context.PAT
	if pat == "" {
		pat = os.Getenv("AZURE_PAT")
	}

	// Step 2: Create Azure DevOps client
	adoClient := azuredevops.NewClient(org, proj, pat)

	// Step 3: Execute based on detected intent
	var response models.ChatResponse
	response.Intent = detected.Intent
	response.Timestamp = time.Now()

	switch detected.Intent {
	case "task_status":
		response = h.handleTaskStatus(adoClient, detected, response)

	case "task_explanation":
		response = h.handleTaskExplanation(adoClient, detected, response)

	case "workload_analysis":
		response = h.handleWorkloadAnalysis(adoClient, detected, response)

	case "stale_tickets":
		response = h.handleStaleTickets(adoClient, detected, response)

	case "sprint_health":
		response = h.handleSprintHealth(adoClient, detected, response)

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

	case "knowledge_search", "unknown":
		response = h.handleKnowledgeSearch(org, proj, detected, response)

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
	wiql := fmt.Sprintf(`Select [System.Id], [System.Title], [System.State] From WorkItems Where [System.AssignedTo] Contains '%s' AND [System.State] NOT CONTAINS 'Closed' AND [System.State] NOT CONTAINS 'Done' AND [System.State] NOT CONTAINS 'Resolved'`, escapeWIQL(userName))

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

// handleTaskExplanation fetches details of a task and asks the detector to explain it using the LLM.
func (h *ChatHandler) handleTaskExplanation(client *azuredevops.Client, detected *models.DetectedIntent, response models.ChatResponse) models.ChatResponse {
	idStr, ok := detected.Params["id"]
	if !ok {
		response.Response = "I couldn't find a task ID in your message. Try: \"Explain task 123\""
		return response
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Response = "Invalid task ID. Please provide a numeric ID like: \"Explain task 123\""
		return response
	}

	workItem, err := client.GetWorkItem(id)
	if err != nil {
		response.Response = formatter.FormatError(err)
		return response
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	explanation, err := h.detector.ExplainTask(ctx, workItem)
	if err != nil {
		response.Response = fmt.Sprintf("I retrieved the task details but couldn't generate an AI explanation: %s\n\nFallback Details:\n\n%s", err.Error(), formatter.FormatWorkItem(workItem))
		return response
	}

	response.Response = explanation
	response.Data = workItem
	return response
}

// handleWorkloadAnalysis analyzes workload across the team or for a specific user.
func (h *ChatHandler) handleWorkloadAnalysis(client *azuredevops.Client, detected *models.DetectedIntent, response models.ChatResponse) models.ChatResponse {
	// Query active items (not closed/done/resolved)
	wiql := `Select [System.Id], [System.Title], [System.State], [System.AssignedTo] From WorkItems Where [System.State] NOT CONTAINS 'Closed' AND [System.State] NOT CONTAINS 'Done' AND [System.State] NOT CONTAINS 'Resolved'`

	items, err := client.QueryWorkItems(wiql, 100)
	if err != nil {
		response.Response = formatter.FormatError(err)
		return response
	}

	// Group by user
	userWorkloads := make(map[string]int)
	userStates := make(map[string]map[string]int) // user -> state -> count
	var unassignedCount int

	for _, item := range items {
		assignee := item.AssignedTo
		if assignee == "" {
			unassignedCount++
			continue
		}
		userWorkloads[assignee]++
		if userStates[assignee] == nil {
			userStates[assignee] = make(map[string]int)
		}
		userStates[assignee][item.State]++
	}

	targetUser, hasTargetUser := detected.Params["user"]
	if hasTargetUser && targetUser != "" {
		// Specific user workload analysis
		// Find matching user (fuzzy match)
		var actualName string
		for name := range userWorkloads {
			if strings.Contains(strings.ToLower(name), strings.ToLower(targetUser)) {
				actualName = name
				break
			}
		}

		if actualName == "" {
			response.Response = fmt.Sprintf("I couldn't find any active tasks assigned to '%s'. They might have 0 active tasks!", targetUser)
			return response
		}

		taskCount := userWorkloads[actualName]
		rating := "Low"
		if taskCount >= 6 {
			rating = "⚠️ High"
		} else if taskCount >= 3 {
			rating = "Moderate"
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("📊 **Workload Analysis for %s**\n\n", actualName))
		sb.WriteString(fmt.Sprintf("• **Active Tasks:** %d\n", taskCount))
		sb.WriteString(fmt.Sprintf("• **Workload Level:** %s\n\n", rating))
		sb.WriteString("**Task States:**\n")
		for state, count := range userStates[actualName] {
			sb.WriteString(fmt.Sprintf("- %s: %d\n", state, count))
		}

		response.Response = sb.String()
		response.Data = map[string]interface{}{
			"user":       actualName,
			"task_count": taskCount,
			"level":      rating,
			"states":     userStates[actualName],
		}
		return response
	}

	// Overall team workload analysis
	var sb strings.Builder
	sb.WriteString("📊 **Team Workload Distribution**\n\n")
	if len(userWorkloads) == 0 {
		sb.WriteString("No active tasks found in the project. Everyone's queue is empty!")
		response.Response = sb.String()
		return response
	}

	var highestUser string
	maxCount := -1
	for name, count := range userWorkloads {
		sb.WriteString(fmt.Sprintf("• **%s:** %d active tasks\n", name, count))
		if count > maxCount {
			maxCount = count
			highestUser = name
		}
	}

	if unassignedCount > 0 {
		sb.WriteString(fmt.Sprintf("• **Unassigned:** %d active tasks\n", unassignedCount))
	}

	if highestUser != "" {
		sb.WriteString(fmt.Sprintf("\n🔥 **Highest Workload:** %s with %d active tasks.", highestUser, maxCount))
	}

	response.Response = sb.String()
	response.Data = map[string]interface{}{
		"distribution":     userWorkloads,
		"unassigned_count": unassignedCount,
		"highest_workload": map[string]interface{}{
			"user":  highestUser,
			"count": maxCount,
		},
	}
	return response
}

// handleStaleTickets finds tickets that haven't been updated for a given number of days.
func (h *ChatHandler) handleStaleTickets(client *azuredevops.Client, detected *models.DetectedIntent, response models.ChatResponse) models.ChatResponse {
	days := 5 // default
	if daysStr, ok := detected.Params["days"]; ok && daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	// Calculate cutoff date in YYYY-MM-DD
	cutoffDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	// Query active items older than cutoff
	wiql := fmt.Sprintf(`Select [System.Id], [System.Title], [System.State], [System.AssignedTo], [System.ChangedDate] From WorkItems Where [System.State] NOT CONTAINS 'Closed' AND [System.State] NOT CONTAINS 'Done' AND [System.State] NOT CONTAINS 'Resolved' AND [System.ChangedDate] < '%s' Order By [System.ChangedDate] Asc`, cutoffDate)

	items, err := client.QueryWorkItems(wiql, 30)
	if err != nil {
		response.Response = formatter.FormatError(err)
		return response
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("⏳ **Stale Tickets (No updates for %d+ days)**\n\n", days))

	if len(items) == 0 {
		sb.WriteString(fmt.Sprintf("Great news! There are no active tickets that haven't been updated in the last %d days.", days))
		response.Response = sb.String()
		return response
	}

	for _, item := range items {
		var changedTime time.Time
		var parseErr error

		// Try standard formats returned by Azure DevOps API
		if strings.Contains(item.ChangedDate, ".") {
			// e.g. "2026-06-04T15:54:56.1234567Z"
			parts := strings.Split(item.ChangedDate, ".")
			if len(parts) > 0 {
				base := parts[0] + "Z"
				changedTime, parseErr = time.Parse("2006-01-02T15:04:05Z", base)
			}
		} else {
			changedTime, parseErr = time.Parse(time.RFC3339, item.ChangedDate)
		}

		daysStale := days
		if parseErr == nil {
			daysStale = int(time.Since(changedTime).Hours() / 24)
		}

		assignee := item.AssignedTo
		if assignee == "" {
			assignee = "Unassigned"
		}

		sb.WriteString(fmt.Sprintf("• **#%d:** %s\n", item.ID, item.Title))
		sb.WriteString(fmt.Sprintf("  *State: %s | Assignee: %s | Stale for %d days*\n", item.State, assignee, daysStale))
	}

	response.Response = sb.String()
	response.Data = items
	return response
}

// handleSprintHealth analyzes sprint risks, blocked items, and bottlenecks.
func (h *ChatHandler) handleSprintHealth(client *azuredevops.Client, detected *models.DetectedIntent, response models.ChatResponse) models.ChatResponse {
	// 1. First, query recent items to identify the active sprint (same logic as summary)
	recentWiql := `Select [System.Id], [System.Title], [System.State], [System.IterationPath] From WorkItems Order By [System.ChangedDate] Desc`
	recentItems, err := client.QueryWorkItems(recentWiql, 30)
	if err != nil {
		response.Response = formatter.FormatError(err)
		return response
	}

	if len(recentItems) == 0 {
		response.Response = "I couldn't find any recent items to determine the active sprint."
		return response
	}

	// Find the current iteration
	iterationCounts := make(map[string]int)
	for _, item := range recentItems {
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

	// 2. Fetch all items in the current sprint (up to 100)
	sprintWiql := fmt.Sprintf(`Select [System.Id], [System.Title], [System.State], [System.AssignedTo], [System.WorkItemType], [Microsoft.VSTS.Common.Priority], [System.Description] From WorkItems Where [System.IterationPath] = '%s'`, escapeWIQL(currentSprint))
	sprintItems, err := client.QueryWorkItems(sprintWiql, 100)
	if err != nil {
		response.Response = formatter.FormatError(err)
		return response
	}

	// 3. Analyze sprint health metrics
	var totalItems, doneItems, activeItems, blockedItems int
	var blockedList []models.WorkItem
	var highPriorityTodo []models.WorkItem
	userTaskCounts := make(map[string]int)

	for _, item := range sprintItems {
		totalItems++
		stateLower := strings.ToLower(item.State)
		
		isDone := stateLower == "closed" || stateLower == "done" || stateLower == "resolved" || stateLower == "signoff"
		if isDone {
			doneItems++
			continue
		}
		
		activeItems++
		if item.AssignedTo != "" {
			userTaskCounts[item.AssignedTo]++
		}

		// Detect if blocked
		isBlockedState := stateLower == "blocked" || strings.Contains(strings.ToLower(item.Title), "blocked")
		if isBlockedState {
			blockedItems++
			blockedList = append(blockedList, item)
		}

		// Detect high priority risks (Priority 1 or 2 and still New/To Do/Proposed)
		isHighPriority := item.Priority == 1 || item.Priority == 2
		isUnstartedActive := stateLower == "new" || stateLower == "to do" || stateLower == "proposed"
		if isHighPriority && isUnstartedActive {
			highPriorityTodo = append(highPriorityTodo, item)
		}
	}

	// Find bottlenecks (anyone with more than 4 active tasks)
	var bottlenecks []string
	for name, count := range userTaskCounts {
		if count > 4 {
			bottlenecks = append(bottlenecks, fmt.Sprintf("%s (%d active tasks)", name, count))
		}
	}

	// Calculate a health score
	healthScore := 100
	if totalItems > 0 {
		// Deduct 20 points per blocked item (max 60)
		blockedDeduction := blockedItems * 20
		if blockedDeduction > 60 {
			blockedDeduction = 60
		}
		healthScore -= blockedDeduction

		// Deduct 10 points per high-priority unstarted item (max 30)
		unstartedDeduction := len(highPriorityTodo) * 10
		if unstartedDeduction > 30 {
			unstartedDeduction = 30
		}
		healthScore -= unstartedDeduction
	}

	// Determine status emoji
	statusEmoji := "🟢 Healthy"
	if healthScore < 50 {
		statusEmoji = "🔴 Critical Risk"
	} else if healthScore < 80 {
		statusEmoji = "🟡 Warning"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🏥 **Sprint Health Report: %s**\n\n", currentSprint))
	sb.WriteString(fmt.Sprintf("• **Overall Health:** %s (%d/100)\n", statusEmoji, healthScore))
	sb.WriteString(fmt.Sprintf("• **Total Items:** %d | **Done:** %d | **Active:** %d\n\n", totalItems, doneItems, activeItems))

	// Section 1: Blockers
	sb.WriteString("🚫 **Blocked Items:** ")
	if blockedItems == 0 {
		sb.WriteString("None! No active blockers reported.\n")
	} else {
		sb.WriteString(fmt.Sprintf("%d item(s) currently blocked:\n", blockedItems))
		for _, item := range blockedList {
			sb.WriteString(fmt.Sprintf("  * **#%d:** %s (Assigned to: %s)\n", item.ID, item.Title, item.AssignedTo))
		}
	}
	sb.WriteString("\n")

	// Section 2: Bottlenecks
	sb.WriteString("👥 **Resource Bottlenecks:** ")
	if len(bottlenecks) == 0 {
		sb.WriteString("None. Task assignment is well distributed.\n")
	} else {
		sb.WriteString(fmt.Sprintf("%d team member(s) with high task queues:\n", len(bottlenecks)))
		for _, b := range bottlenecks {
			sb.WriteString(fmt.Sprintf("  * %s\n", b))
		}
	}
	sb.WriteString("\n")

	// Section 3: High Priority Risks
	sb.WriteString("⚠️ **High Priority Unstarted Risks:** ")
	if len(highPriorityTodo) == 0 {
		sb.WriteString("None. High-priority items are underway.\n")
	} else {
		sb.WriteString(fmt.Sprintf("%d high-priority items still in To Do:\n", len(highPriorityTodo)))
		for _, item := range highPriorityTodo {
			sb.WriteString(fmt.Sprintf("  * **#%d:** %s (Priority %d | Assignee: %s)\n", item.ID, item.Title, item.Priority, item.AssignedTo))
		}
	}

	response.Response = sb.String()
	response.Data = map[string]interface{}{
		"score":         healthScore,
		"status":        statusEmoji,
		"blocked_count": blockedItems,
		"blocked_items": blockedList,
		"risks_count":   len(highPriorityTodo),
		"risks":         highPriorityTodo,
		"bottlenecks":   bottlenecks,
	}

	return response
}

// escapeWIQL escapes special characters like backslashes and single quotes for safe injection in WIQL queries.
func escapeWIQL(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "''")
	return s
}

// handleKnowledgeSearch searches the local pgvector document database and uses Gemini to answer.
func (h *ChatHandler) handleKnowledgeSearch(org string, proj string, detected *models.DetectedIntent, response models.ChatResponse) models.ChatResponse {
	query, ok := detected.Params["raw"]
	if !ok || strings.TrimSpace(query) == "" {
		response.Response = "Please ask a question about your project documentation."
		return response
	}

	answer, err := rag.SearchAndAnswer(context.Background(), org, proj, query)
	if err != nil {
		response.Response = fmt.Sprintf("❌ **Error searching knowledge base:** %v", err)
		return response
	}

	response.Response = answer
	return response
}
