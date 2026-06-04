# Technical Critique: Current Codebase Flaws

### 🔑 Security & Token Storage
* **Plaintext Storage:** Personal Access Tokens (PATs) are saved directly in database columns. If PostgreSQL is read, all user credentials are leaked.

### 🧠 RAG Search & Retrieval
* **Fixed Context Size:** Hardcoded `LIMIT 3` database retrieval fails to scale down for specific queries, or up for broad queries.
* **No Similarity Filter:** The query fetches chunks even if they are completely unrelated to the user's question.
* **Brittle Text Splitting:** Character-based splitting cuts Markdown tables, bullet points, and code blocks in half.

### 📡 Syncing & Ingestion
* **In-Memory Syncing:** Background ingestion runs in raw Go goroutines. A server crash or restart terminates the sync without any way to resume.
* **Full Rebuilds Only:** Saving configs triggers full downloads and re-embedding of all wiki pages, wasting time and API costs.

### 💬 Chat UX
* **Blocking Responses:** The user must wait for the full response to generate before receiving anything (no real-time output).

### 🖥️ Core Reliability
* **Non-Concurrent DB Connection:** Using a single `pgx.Conn` is not safe for concurrent API requests. Simultaneous user queries will block or crash the database driver.
* **Unstructured Logging:** The application prints raw text using standard `log.Printf`, making it hard to search and parse logs in production.
* **No Production Monitoring:** Missing application performance monitoring (APM) and centralized error logging (like Sentry).
