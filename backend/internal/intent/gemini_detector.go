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
1. "intent": MUST be one of ["task_status", "user_work", "active_bugs", "sprint_summary", "help", "unknown"].
2. "params": A JSON object containing extracted entities.
- If intent is "task_status", extract the task ID into params as "id" (string, just the numbers).
- If intent is "user_work", extract the person's name into params as "user" (string).

Examples:
"What is the status of bug 12345?" -> {"intent": "task_status", "params": {"id": "12345", "raw": "..."}}
"Show me what Rahul is working on" -> {"intent": "user_work", "params": {"user": "Rahul", "raw": "..."}}
"Give me a sprint summary" -> {"intent": "sprint_summary", "params": {"raw": "..."}}
"Show active bugs" -> {"intent": "active_bugs", "params": {"raw": "..."}}`),
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
