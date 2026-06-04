# Gap to Production Grade

### 🔑 Security & Token Storage
* **PAT Encryption:** Implement AES-256-GCM encryption/decryption functions for PAT strings.
* **Secret Configuration:** Add master encryption key to Hugging Face secrets.

### 🧠 RAG Search & Retrieval
* **Similarity Filtering:** Implement Cosine Similarity threshold filtering in `rag.go`.
* **Dynamic Budgeting:** Replace `LIMIT 3` with dynamic K based on token budgeting.
* **Markdown Parsing:** Implement Markdown-aware semantic chunking logic.

### 📡 Syncing & Ingestion
* **Task Queue:** Integrate background task queue worker to handle retries and survive server restarts.

### 🖥️ Core Reliability
* **Database Connection Pool:** Replace `pgx.Conn` with `pgxpool.Pool` to support safe, concurrent database queries under multi-user load.
* **Structured Logging:** Implement structured JSON logging (using a library like `zap` or `zerolog`) to output query performance and request metadata.
* **Error Monitoring:** Set up Sentry error monitoring middleware for production tracking.
