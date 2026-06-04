// Package models contains shared data structures used across the application.
package models

import "time"

// ChatRequest represents the incoming chat message from the user.
type ChatRequest struct {
	Message string      `json:"message" binding:"required"`
	Context UserContext `json:"context"`
}

// UserContext holds the Azure DevOps connection details provided by the user.
type UserContext struct {
	Organization string `json:"organization"`
	Project      string `json:"project"`
	PAT          string `json:"pat"`
}

// ChatResponse represents the response sent back to the user.
type ChatResponse struct {
	Response  string      `json:"response"`
	Intent    string      `json:"intent"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// DetectedIntent represents the result of intent detection from a user message.
type DetectedIntent struct {
	Intent string            `json:"intent"`
	Params map[string]string `json:"params"`
}

// WorkItem represents a simplified Azure DevOps work item.
type WorkItem struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	State       string `json:"state"`
	AssignedTo  string `json:"assignedTo"`
	WorkItemType string `json:"workItemType"`
	Priority    int    `json:"priority"`
	IterationPath string `json:"iterationPath"`
	CreatedDate string `json:"createdDate"`
	ChangedDate string `json:"changedDate"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url"`
}

// ConfigValidation represents a request to validate Azure DevOps credentials.
type ConfigValidation struct {
	Organization string `json:"organization" binding:"required"`
	Project      string `json:"project" binding:"required"`
	PAT          string `json:"pat" binding:"required"`
}

// ValidationResponse represents the result of credential validation.
type ValidationResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
	Project string `json:"project,omitempty"`
}

// ErrorResponse represents an API error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Details string `json:"details,omitempty"`
}
