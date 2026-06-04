package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
)

var Conn *pgx.Conn

// InitPostgres connects to Supabase PostgreSQL and runs migrations/table setups.
func InitPostgres() error {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Println("⚠️ DATABASE_URL not set, skipping database connection (RAG features will be disabled)")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Connect to the database
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}
	Conn = conn

	log.Println("✅ Connected to PostgreSQL (Supabase)")

	// 2. Enable pgvector extension and create table
	initSQL := `
	CREATE EXTENSION IF NOT EXISTS vector;
	
	CREATE TABLE IF NOT EXISTS document_chunks (
		id SERIAL PRIMARY KEY,
		organization TEXT NOT NULL,
		project TEXT NOT NULL,
		title TEXT NOT NULL,
		url TEXT,
		content TEXT NOT NULL,
		embedding vector(3072),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS user_configs (
		user_id UUID PRIMARY KEY,
		organization TEXT NOT NULL,
		project TEXT NOT NULL,
		pat TEXT NOT NULL,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE OR REPLACE FUNCTION clean_orphaned_chunks()
	RETURNS TRIGGER AS $func$
	BEGIN
		IF NOT EXISTS (
			SELECT 1 FROM user_configs 
			WHERE organization = OLD.organization AND project = OLD.project
		) THEN
			DELETE FROM document_chunks 
			WHERE organization = OLD.organization AND project = OLD.project;
		END IF;
		RETURN OLD;
	END;
	$func$ LANGUAGE plpgsql;

	DROP TRIGGER IF EXISTS trg_clean_orphaned_chunks ON user_configs;
	CREATE TRIGGER trg_clean_orphaned_chunks
	AFTER DELETE ON user_configs
	FOR EACH ROW
	EXECUTE FUNCTION clean_orphaned_chunks();
	`
	
	_, err = Conn.Exec(ctx, initSQL)
	if err != nil {
		return fmt.Errorf("failed to run database migrations: %w", err)
	}

	log.Println("✅ Database tables and pgvector extension initialized")
	return nil
}

// Close closes the database connection.
func Close() {
	if Conn != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = Conn.Close(ctx)
		log.Println("🔌 Closed PostgreSQL connection")
	}
}
