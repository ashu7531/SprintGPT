// Package intent provides intent detection from natural language user messages.
package intent

import "github.com/ashutosh/sprintgpt-backend/pkg/models"

// Detector is the interface for detecting user intent from a message.
// This allows swapping keyword-based detection with LLM-based detection later.
type Detector interface {
	// Detect analyzes a user message and returns the detected intent with parameters.
	Detect(message string) (*models.DetectedIntent, error)
}
