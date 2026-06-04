package rag

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ashutosh/sprintgpt-backend/internal/database"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// IngestDocument chunks the content, generates embeddings, and saves them to Postgres.
func IngestDocument(ctx context.Context, organization string, project string, title string, content string, url string) error {
	if database.Conn == nil {
		return fmt.Errorf("database connection not initialized")
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("GEMINI_API_KEY not set")
	}

	// 1. Chunk the document
	chunks := chunkText(content, 800)
	if len(chunks) == 0 {
		return nil
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return fmt.Errorf("failed to create GenAI client: %w", err)
	}
	defer client.Close()

	model := client.EmbeddingModel("gemini-embedding-001")

	log.Printf("Ingesting '%s' for project '%s/%s': generated %d chunks", title, organization, project, len(chunks))

	// 2. Process each chunk
	for i, chunk := range chunks {
		// Generate embedding
		res, err := model.EmbedContent(ctx, genai.Text(chunk))
		if err != nil {
			return fmt.Errorf("failed to embed chunk %d: %w", i, err)
		}

		vec := res.Embedding.Values
		vecStr := formatVector(vec)

		// Save to Postgres with organization and project tags
		_, err = database.Conn.Exec(ctx,
			"INSERT INTO document_chunks (organization, project, title, url, content, embedding) VALUES ($1, $2, $3, $4, $5, $6)",
			organization, project, title, url, chunk, vecStr,
		)
		if err != nil {
			return fmt.Errorf("failed to insert chunk %d: %w", i, err)
		}
	}

	log.Printf("✅ Successfully ingested document: %s", title)
	return nil
}

// SearchAndAnswer retrieves context from Postgres and uses Gemini to answer the user query.
func SearchAndAnswer(ctx context.Context, organization string, project string, query string) (string, error) {
	if database.Conn == nil {
		return "Sorry, the database is not configured so I cannot search the project documentation.", nil
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY not set")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return "", fmt.Errorf("failed to create GenAI client: %w", err)
	}
	defer client.Close()

	// 1. Embed user query
	embModel := client.EmbeddingModel("gemini-embedding-001")
	res, err := embModel.EmbedContent(ctx, genai.Text(query))
	if err != nil {
		return "", fmt.Errorf("failed to embed query: %w", err)
	}

	vecStr := formatVector(res.Embedding.Values)

	// 2. Query top 3 matching chunks filtered by organization and project
	rows, err := database.Conn.Query(ctx,
		"SELECT title, url, content FROM document_chunks WHERE organization = $1 AND project = $2 ORDER BY embedding <=> $3 ASC LIMIT 3",
		organization, project, vecStr,
	)
	if err != nil {
		return "", fmt.Errorf("failed to query document chunks: %w", err)
	}
	defer rows.Close()

	var contextParts []string
	for rows.Next() {
		var title, url, content string
		if err := rows.Scan(&title, &url, &content); err != nil {
			return "", fmt.Errorf("failed to scan chunk row: %w", err)
		}
		
		source := title
		if url != "" {
			source = fmt.Sprintf("%s (%s)", title, url)
		}
		contextParts = append(contextParts, fmt.Sprintf("Source: %s\nContent:\n%s", source, content))
	}

	if len(contextParts) == 0 {
		return "I searched the project documentation but couldn't find any relevant information to answer your question.", nil
	}

	contextText := strings.Join(contextParts, "\n\n---\n\n")

	// 3. Generate answer using Gemini
	genModel := client.GenerativeModel("gemini-3.5-flash")
	
	systemPrompt := `You are an expert project assistant for SprintGPT. 
Your task is to answer the user's question using the provided context from the project documentation or wikis.
Ensure your response is clear, accurate, and structured in clean Markdown.
If the answer cannot be found in the provided context, state that you couldn't find it in the project documentation. Do not invent details.`

	prompt := fmt.Sprintf(`System Instruction: %s

Context from project documentation:
----------------------------------
%s
----------------------------------

User Question: %s
Answer:`, systemPrompt, contextText, query)

	resp, err := genModel.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate AI response: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return "I encountered an error generating the response.", nil
	}

	answerText := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		if textPart, ok := part.(genai.Text); ok {
			answerText += string(textPart)
		}
	}

	return strings.TrimSpace(answerText), nil
}

// chunkText splits text into meaningful chunks, combining adjacent small paragraphs 
// to preserve context and skipping noise lines (like markdown horizontal rules).
func chunkText(text string, chunkSize int) []string {
	var chunks []string
	paragraphs := strings.Split(text, "\n\n")
	
	var currentBlock []string
	currentBlockLen := 0

	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		// Skip formatting-only noise like horizontal rules or divider symbols (e.g. ***, ---, ===, * * *)
		cleaned := strings.ReplaceAll(strings.ReplaceAll(p, " ", ""), "*", "")
		cleaned = strings.ReplaceAll(cleaned, "-", "")
		cleaned = strings.ReplaceAll(cleaned, "=", "")
		cleaned = strings.ReplaceAll(cleaned, "_", "")
		if len(cleaned) == 0 {
			continue
		}

		// If a single paragraph is too large, flush the current block, and split the large paragraph
		if len(p) > chunkSize {
			if len(currentBlock) > 0 {
				chunks = append(chunks, strings.Join(currentBlock, "\n\n"))
				currentBlock = nil
				currentBlockLen = 0
			}

			// Split the large paragraph into word blocks
			words := strings.Fields(p)
			var tempChunk []string
			tempLen := 0
			for _, w := range words {
				tempChunk = append(tempChunk, w)
				tempLen += len(w) + 1
				if tempLen >= chunkSize {
					chunks = append(chunks, strings.Join(tempChunk, " "))
					tempChunk = nil
					tempLen = 0
				}
			}
			if len(tempChunk) > 0 {
				chunks = append(chunks, strings.Join(tempChunk, " "))
			}
		} else {
			// If adding this paragraph exceeds the chunk size, flush the current block first
			if currentBlockLen+len(p)+2 > chunkSize && len(currentBlock) > 0 {
				chunks = append(chunks, strings.Join(currentBlock, "\n\n"))
				currentBlock = nil
				currentBlockLen = 0
			}
			
			currentBlock = append(currentBlock, p)
			currentBlockLen += len(p) + 2
		}
	}

	// Flush any remaining text in the block
	if len(currentBlock) > 0 {
		chunks = append(chunks, strings.Join(currentBlock, "\n\n"))
	}

	return chunks
}

// formatVector converts a []float32 slice to the pgvector '[v1,v2,v3]' format.
func formatVector(vec []float32) string {
	var sb strings.Builder
	sb.WriteString("[")
	for i, val := range vec {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf("%f", val))
	}
	sb.WriteString("]")
	return sb.String()
}
