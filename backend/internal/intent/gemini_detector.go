package intent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ashutosh/sprintgpt-backend/pkg/models"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiDetector struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

func NewGeminiDetector(ctx context.Context, apiKey string) (*GeminiDetector, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}

	// Use gemini-3.5-flash (the 2026 flagship model)
	model := client.GenerativeModel("gemini-3.5-flash")
	model.ResponseMIMEType = "application/json"
	
	// Define the strict schema we want back
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{
			genai.Text(`You are an AI router for an Azure DevOps assistant.
Your job is to read the user's message and return a JSON object with two fields:
1. "intent": MUST be one of ["task_status", "user_work", "active_bugs", "sprint_summary", "task_explanation", "workload_analysis", "stale_tickets", "sprint_health", "help", "unknown"].
2. "params": A JSON object containing extracted entities.
- If intent is "task_status" or "task_explanation", extract the task ID into params as "id" (string, just the numbers).
- If intent is "user_work" or the user asks if a specific user is overloaded, extract the person's name into params as "user" (string).
- If intent is "stale_tickets", extract the number of days into params as "days" (string, default to "5" if not specified).

Examples:
"What is the status of bug 12345?" -> {"intent": "task_status", "params": {"id": "12345", "raw": "..."}}
"Explain user story 999" -> {"intent": "task_explanation", "params": {"id": "999", "raw": "..."}}
"Show me what Rahul is working on" -> {"intent": "user_work", "params": {"user": "Rahul", "raw": "..."}}
"Is Priya overloaded?" -> {"intent": "workload_analysis", "params": {"user": "Priya", "raw": "..."}}
"Who has the highest workload?" -> {"intent": "workload_analysis", "params": {"raw": "..."}}
"Show stale tickets older than 7 days" -> {"intent": "stale_tickets", "params": {"days": "7", "raw": "..."}}
"What is blocking the sprint?" -> {"intent": "sprint_health", "params": {"raw": "..."}}`),
		},
	}

	return &GeminiDetector{
		client: client,
		model:  model,
	}, nil
}

func (d *GeminiDetector) Detect(message string) (*models.DetectedIntent, error) {
	ctx := context.Background()
	resp, err := d.model.GenerateContent(ctx, genai.Text(message))
	if err != nil {
		return nil, fmt.Errorf("gemini API error: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from gemini")
	}

	part := resp.Candidates[0].Content.Parts[0]
	text, ok := part.(genai.Text)
	if !ok {
		return nil, fmt.Errorf("unexpected response type from gemini")
	}

	var result models.DetectedIntent
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("failed to parse gemini JSON: %w\nResponse was: %s", err, string(text))
	}

	// Ensure params map exists
	if result.Params == nil {
		result.Params = make(map[string]string)
	}
	result.Params["raw"] = message

	return &result, nil
}

func (d *GeminiDetector) Close() {
	if d.client != nil {
		d.client.Close()
	}
}

// ExplainTask implements the Detector interface using the Gemini API.
func (d *GeminiDetector) ExplainTask(ctx context.Context, task *models.WorkItem) (string, error) {
	// Create a dedicated model instance for general content generation (no strict JSON schema)
	model := d.client.GenerativeModel("gemini-3.5-flash")
	
	// Create a clean markdown representation of the description
	cleanDesc := stripHTML(task.Description)

	prompt := fmt.Sprintf(`You are an expert technical manager. 
Explain this Azure DevOps work item in simple, clear terms for the team. 
Focus on:
1. What the task is trying to achieve.
2. Key requirements/Acceptance criteria.
3. Current state/priority and assignee.

Work Item Details:
- ID: %d
- Title: %s
- Type: %s
- State: %s
- Priority: %d
- Assignee: %s
- Description: %s

Provide your explanation in well-formatted Markdown. Keep it concise.`, 
		task.ID, task.Title, task.WorkItemType, task.State, task.Priority, task.AssignedTo, cleanDesc)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate explanation: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response from Gemini")
	}

	part := resp.Candidates[0].Content.Parts[0]
	text, ok := part.(genai.Text)
	if !ok {
		return "", fmt.Errorf("unexpected response type from Gemini")
	}

	return string(text), nil
}
