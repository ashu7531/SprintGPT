// SprintGPT Backend — AI-powered Azure DevOps Assistant
//
// This is the entry point for the Go backend server.
// It initializes the Gin router, registers middleware and routes,
// and starts the HTTP server.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/ashutosh/sprintgpt-backend/internal/config"
	"github.com/ashutosh/sprintgpt-backend/internal/handler"
	"github.com/ashutosh/sprintgpt-backend/internal/intent"
	"github.com/ashutosh/sprintgpt-backend/internal/middleware"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Set Gin to release mode in production
	// gin.SetMode(gin.ReleaseMode)

	// Create Gin router
	router := gin.New()

	// Apply middleware
	router.Use(middleware.Logger())
	router.Use(middleware.CORS(cfg.CORSOrigin))
	router.Use(gin.Recovery())

	// Initialize intent detector (Milestone 4: Gemini AI, fallback to Regex MVP)
	var detector intent.Detector
	geminiKey := os.Getenv("GEMINI_API_KEY")

	if geminiKey != "" {
		d, err := intent.NewGeminiDetector(context.Background(), geminiKey)
		if err != nil {
			log.Fatalf("Failed to initialize Gemini detector: %v", err)
		}
		defer d.Close()
		detector = d
		log.Printf("🧠 Intent detector: Google Gemini AI (Milestone 4)")
	} else {
		detector = intent.NewKeywordDetector()
		log.Printf("🧠 Intent detector: Keyword/Regex fallback (No GEMINI_API_KEY provided)")
	}

	// Initialize handlers
	chatHandler := handler.NewChatHandler(detector)

	// Register routes
	api := router.Group("/api/v1")
	{
		api.GET("/health", handler.Health())
		api.POST("/chat", chatHandler.Handle)
		api.POST("/config/validate", handler.ValidateConfig())
	}

	// Start server
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("🚀 SprintGPT backend starting on %s", addr)
	log.Printf("📡 CORS allowed origin: %s", cfg.CORSOrigin)
	log.Printf("🧠 Intent detector: keyword/regex (MVP)")

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
