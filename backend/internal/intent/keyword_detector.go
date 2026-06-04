package intent

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/ashutosh/sprintgpt-backend/pkg/models"
)

// KeywordDetector implements the Detector interface using regex/keyword matching.
// This is the MVP approach — fast, free, no LLM dependency.
type KeywordDetector struct {
	patterns []intentPattern
}

// intentPattern maps a regex to an intent name.
type intentPattern struct {
	regex  *regexp.Regexp
	intent string
	// extractParams is a function that extracts parameters from regex matches.
	extractParams func(matches []string) map[string]string
}

// NewKeywordDetector creates a new keyword-based intent detector with predefined patterns.
func NewKeywordDetector() *KeywordDetector {
	return &KeywordDetector{
		patterns: []intentPattern{
			// P1: Task/work item status lookup by ID
			{
				regex:  regexp.MustCompile(`(?i)(?:status|details?|info|show|get|fetch|what(?:'s| is))(?: (?:of|for|about))?(?: (?:task|work item|item|bug|story|ticket))?[# ]*(\d+)`),
				intent: "task_status",
				extractParams: func(matches []string) map[string]string {
					return map[string]string{"id": matches[1]}
				},
			},
			// Alternate: just a number like "task 123" or "#123"
			{
				regex:  regexp.MustCompile(`(?i)^(?:task|work item|item|bug|story|ticket)[# ]*(\d+)$`),
				intent: "task_status",
				extractParams: func(matches []string) map[string]string {
					return map[string]string{"id": matches[1]}
				},
			},
			// Alternate: just "#123" or "123"
			{
				regex:  regexp.MustCompile(`(?i)^#?(\d+)$`),
				intent: "task_status",
				extractParams: func(matches []string) map[string]string {
					return map[string]string{"id": matches[1]}
				},
			},
			// P2: What is someone working on
			{
				regex:  regexp.MustCompile(`(?i)(?:what(?:'s| is))(?: .*)?\b(\w+)\b(?: working on| assigned| doing)`),
				intent: "user_work",
				extractParams: func(matches []string) map[string]string {
					return map[string]string{"user": matches[1]}
				},
			},
			// P2: Alternate — "show tasks assigned to <name>"
			{
				regex:  regexp.MustCompile(`(?i)(?:show|list|get|fetch)(?: (?:tasks?|items?|work))?(?: assigned)? (?:to|for|of) (\w+)`),
				intent: "user_work",
				extractParams: func(matches []string) map[string]string {
					return map[string]string{"user": matches[1]}
				},
			},
			// P3: Active bugs
			{
				regex:  regexp.MustCompile(`(?i)(?:show|list|get|active|open|current)\s+bugs?`),
				intent: "active_bugs",
				extractParams: func(matches []string) map[string]string {
					return map[string]string{}
				},
			},
			// P3: Alternate — "bugs" alone
			{
				regex:  regexp.MustCompile(`(?i)^bugs?$`),
				intent: "active_bugs",
				extractParams: func(matches []string) map[string]string {
					return map[string]string{}
				},
			},
			// P4: Sprint summary
			{
				regex:  regexp.MustCompile(`(?i)(?:sprint|iteration)(?: summary| status| progress| overview)?`),
				intent: "sprint_summary",
				extractParams: func(matches []string) map[string]string {
					return map[string]string{}
				},
			},
			// P4: Alternate — "current sprint" or "summarize sprint"
			{
				regex:  regexp.MustCompile(`(?i)(?:summarize|current|show)(?: the)? (?:sprint|iteration)`),
				intent: "sprint_summary",
				extractParams: func(matches []string) map[string]string {
					return map[string]string{}
				},
			},
			// P6: Task Explanation
			{
				regex:  regexp.MustCompile(`(?i)(?:explain|describe|details? of)(?: (?:task|work item|item|bug|story|ticket))?[# ]*(\d+)`),
				intent: "task_explanation",
				extractParams: func(matches []string) map[string]string {
					return map[string]string{"id": matches[1]}
				},
			},
			// P7: Workload Analysis (for specific user)
			{
				regex:  regexp.MustCompile(`(?i)is\s+(\w+)\s+(?:overloaded|busy)`),
				intent: "workload_analysis",
				extractParams: func(matches []string) map[string]string {
					return map[string]string{"user": matches[1]}
				},
			},
			// P7: Workload Analysis (overall)
			{
				regex:  regexp.MustCompile(`(?i)(?:workload|who has (?:the )?highest workload|team workload|task distribution)`),
				intent: "workload_analysis",
				extractParams: func(matches []string) map[string]string {
					return map[string]string{}
				},
			},
			// P8: Stale Tickets with days
			{
				regex:  regexp.MustCompile(`(?i)(?:show\s+)?stale(?:\s+tickets?)?\s+(?:for|older than|after)?\s*(\d+)\s*(?:days?)?`),
				intent: "stale_tickets",
				extractParams: func(matches []string) map[string]string {
					return map[string]string{"days": matches[1]}
				},
			},
			// P8: Stale Tickets (default)
			{
				regex:  regexp.MustCompile(`(?i)(?:show\s+)?stale(?:\s+tickets?)?`),
				intent: "stale_tickets",
				extractParams: func(matches []string) map[string]string {
					return map[string]string{}
				},
			},
			// P9: Sprint Health
			{
				regex:  regexp.MustCompile(`(?i)(?:sprint|iteration)\s+health|what(?:'s|\s+is)\s+blocking(?:\s+the)?\s+sprint|sprint\s+risks?`),
				intent: "sprint_health",
				extractParams: func(matches []string) map[string]string {
					return map[string]string{}
				},
			},
			// Help
			{
				regex:  regexp.MustCompile(`(?i)^(?:help|commands|what can you do|how to use|\?)$`),
				intent: "help",
				extractParams: func(matches []string) map[string]string {
					return map[string]string{}
				},
			},
		},
	}
}

// Detect analyzes the user message and returns the detected intent.
func (kd *KeywordDetector) Detect(message string) (*models.DetectedIntent, error) {
	message = strings.TrimSpace(message)

	if message == "" {
		return nil, fmt.Errorf("empty message")
	}

	for _, pattern := range kd.patterns {
		matches := pattern.regex.FindStringSubmatch(message)
		if matches != nil {
			return &models.DetectedIntent{
				Intent: pattern.intent,
				Params: pattern.extractParams(matches),
			}, nil
		}
	}

	return &models.DetectedIntent{
		Intent: "unknown",
		Params: map[string]string{"raw": message},
	}, nil
}

// SupportedIntents returns a human-readable list of supported queries.
func (kd *KeywordDetector) SupportedIntents() string {
	return `Here's what I can help you with:

• **Task Status** — "What is the status of task 123?" or just "#123"
• **User Work** — "What is Rahul working on?" or "Show tasks assigned to Rahul"
• **Active Bugs** — "Show active bugs" or just "bugs"
• **Sprint Summary** — "Sprint summary" or "Summarize the current sprint"
• **Task Explanation** — "Explain task 123"
• **Workload Analysis** — "Who has the highest workload?" or "Is Rahul overloaded?"
• **Stale Tickets** — "Show stale tickets" or "stale tickets for 7 days"
• **Sprint Health** — "Sprint health" or "What is blocking the sprint?"
• **Help** — "help" or "?"

Just type your question naturally!`
}

// ExplainTask implements the Detector interface by returning a formatted version of the task's properties.
func (kd *KeywordDetector) ExplainTask(ctx context.Context, task *models.WorkItem) (string, error) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📋 **Task %d: %s**\n\n", task.ID, task.Title))
	sb.WriteString(fmt.Sprintf("• **Type:** %s\n", task.WorkItemType))
	sb.WriteString(fmt.Sprintf("• **State:** %s\n", task.State))
	sb.WriteString(fmt.Sprintf("• **Priority:** %d\n", task.Priority))
	sb.WriteString(fmt.Sprintf("• **Assignee:** %s\n\n", task.AssignedTo))
	
	if task.Description != "" {
		sb.WriteString("### Description:\n")
		// Strip simple HTML tags if present (e.g. from ADO description rich text)
		desc := stripHTML(task.Description)
		sb.WriteString(desc)
	} else {
		sb.WriteString("*(No description provided)*")
	}
	
	return sb.String(), nil
}

// Helper to strip simple HTML tags for clean Markdown formatting
func stripHTML(s string) string {
	// A simple regex to strip HTML tags
	r := regexp.MustCompile("<[^>]*>")
	return r.ReplaceAllString(s, "")
}
