// Package formatter converts Azure DevOps data into human-readable responses.
package formatter

import (
	"fmt"
	"strings"

	"github.com/ashutosh/sprintgpt-backend/pkg/models"
)

// FormatWorkItem converts a work item into a human-readable string.
func FormatWorkItem(item *models.WorkItem) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("**%s #%d: %s**\n\n", item.WorkItemType, item.ID, item.Title))

	sb.WriteString(fmt.Sprintf("• **Status:** %s\n", item.State))

	if item.AssignedTo != "" {
		sb.WriteString(fmt.Sprintf("• **Assigned To:** %s\n", item.AssignedTo))
	} else {
		sb.WriteString("• **Assigned To:** Unassigned\n")
	}

	if item.Priority > 0 {
		sb.WriteString(fmt.Sprintf("• **Priority:** %d\n", item.Priority))
	}

	if item.IterationPath != "" {
		// Show only the last part of the iteration path (sprint name)
		parts := strings.Split(item.IterationPath, "\\")
		sprint := parts[len(parts)-1]
		sb.WriteString(fmt.Sprintf("• **Sprint:** %s\n", sprint))
	}

	if item.URL != "" {
		sb.WriteString(fmt.Sprintf("\n[View in Azure DevOps](%s)", item.URL))
	}

	return sb.String()
}

// FormatWorkItemList formats a list of work items.
func FormatWorkItemList(title string, items []models.WorkItem) string {
	if len(items) == 0 {
		return fmt.Sprintf("I couldn't find any %s.", strings.ToLower(title))
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**%s (%d found)**\n\n", title, len(items)))

	for _, item := range items {
		assignee := "Unassigned"
		if item.AssignedTo != "" {
			assignee = item.AssignedTo
		}
		sb.WriteString(fmt.Sprintf("• **[#%d] %s** — %s (%s)\n", item.ID, item.Title, item.State, assignee))
	}

	return sb.String()
}

// FormatUnknownIntent returns a helpful message when the intent can't be detected.
func FormatUnknownIntent(rawMessage string) string {
	return fmt.Sprintf(`I'm not sure what you're asking. Here's what I can help with:

• **Task Status** — "What is the status of task 123?" or "#123"
• **User Work** — "What is Rahul working on?"
• **Active Bugs** — "Show active bugs"
• **Sprint Summary** — "Sprint summary"

You asked: "%s"`, rawMessage)
}

// FormatHelp returns a help message listing supported queries.
func FormatHelp() string {
	return `👋 **Welcome to SprintGPT!**

I can help you query your Azure DevOps project. Try asking:

• **Task Status** — "What is the status of task 123?" or just "#123"
• **User Work** — "What is Rahul working on?" or "Show tasks for Rahul"
• **Active Bugs** — "Show active bugs" or just "bugs"
• **Sprint Summary** — "Sprint summary" or "Summarize the sprint"

Just type your question naturally and I'll fetch the data for you!`
}

// FormatError returns a formatted error message.
func FormatError(err error) string {
	return fmt.Sprintf("❌ **Error:** %s\n\nPlease check your Azure DevOps configuration and try again.", err.Error())
}
