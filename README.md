# SprintGPT: AI-Powered Azure DevOps Assistant

SprintGPT is a modern conversational assistant designed to bridge the gap between development teams and project workflows in Azure DevOps. By leveraging Google Gemini and vector search, it allows developers, scrum masters, and stakeholders to query task status, summarize sprint progress, and search internal wikis using natural language.

🌐 **Live Application:** [https://sprint-gpt.vercel.app/](https://sprint-gpt.vercel.app/)

---

## 📖 User Guide (How to Use)

### Step 1: Sign Up & Log In
1. Open the [Live Application](https://sprint-gpt.vercel.app/).
2. Sign up with your email or log in if you already have an account.

### Step 2: Create your Azure DevOps PAT (Personal Access Token)
SprintGPT needs secure read access to fetch your board items and wikis.
1. Go to Azure DevOps (`https://dev.azure.com/{your-organization}`).
2. In the top-right corner, click **User Settings** (gear icon) ➡️ **Personal Access Tokens**.
3. Click **+ New Token**.
4. Set the expiration date and enable these scopes:
   * **Work Items:** `Read` (for tasks, bugs, and backlog checks)
   * **Project and Team:** `Read` (for sprint and board configurations)
   * **Wiki:** `Read` (for documentation chunking and RAG indexing)
   * **Graph:** `Read` (for looking up team members)
5. Click **Create** and copy your token (keep it secret!).

### Step 3: Connect Your Project
1. In the SprintGPT interface, click the **Settings (⚙️)** icon in the sidebar.
2. Fill in:
   * **Organization Name:** The organization name from your DevOps URL (`dev.azure.com/{organization}`).
   * **Project Name:** The exact name of your Azure DevOps project.
   * **Personal Access Token:** The token you copied in Step 2.
3. Click **Save & Connect**. The backend will save your config and start indexing your wiki pages in the background.

---

## 💬 Conversation Guidelines (What to Ask)

SprintGPT automatically routes your questions using natural language. Here are examples of queries you can ask:

### 📊 Sprint Summaries
* *"Summarize our current active sprint"*
* *"What is the progress of iteration 4?"*
* *"Show all active tasks on the board"*

### 📝 Task Status & Details
* *"What is the status of task 1045?"*
* *"Show details for bug 2049"*
* *"Who is assigned to user story 982?"*

### 📚 Wiki & RAG Knowledge Search
* *"What are the details of the IVR API?"*
* *"Explain our codebase setup guide from the wiki"*
* *"How does the DB backup system work?"*

---

## 🚀 Key Features

* **Intelligent Query Routing:** Automatically detects user intent (sprint summaries, task status updates, or wiki queries) and fetches the relevant data.
* **Retrieval-Augmented Generation (RAG):** Automatically parses and embeds Azure DevOps Wikis using Gemini Embeddings and PostgreSQL `pgvector`, allowing users to query team documentation inside the chat.
* **JWT Security:** Modern JWKS-based verification securing backend APIs.
* **Redis Caching:** Minimizes API response times and prevents Azure DevOps rate limits by caching HTTP payloads.
* **Unified UI:** A highly polished, responsive dashboard built with React and Tailwind CSS.
* **Docker Native:** Fully containerized setup for simplified local testing.

---

## 🛠️ Tech Stack

* **Frontend:** React, Vite, CSS / Tailwind CSS
* **Backend:** Go (Golang), Gin Gonic HTTP router, pgx/v5 PostgreSQL driver
* **Databases:** PostgreSQL (Supabase with `pgvector` extension)
* **Cache:** Redis (Upstash or local Redis instance)
* **AI Engine:** Google Gemini (Gemini Flash & Gemini Embeddings)

---

## 📦 Local Setup Guide

### Prerequisites
* [Go 1.21+](https://go.dev/doc/install)
* [Node.js 18+](https://nodejs.org/)
* [Docker & Docker Compose](https://docs.docker.com/get-docker/)
* A Google AI Studio API Key (for Gemini)
* A Supabase project (or local PostgreSQL instance with pgvector support)

---

### Step 1: Clone the Repository
```bash
git clone https://github.com/ashu7531/SprintGPT.git
cd SprintGPT
```

---

### Step 2: Environment Configuration

#### 1. Backend Config
Create a `.env` file inside the `backend` directory:
```bash
# backend/.env
PORT=8080
DATABASE_URL=postgres://postgres.[proj-id]:[password]@aws-0-[region].pooler.supabase.com:6543/postgres?pgbouncer=true
SUPABASE_URL=https://[your-supabase-project].supabase.co
GEMINI_API_KEY=your_gemini_api_key_here
REDIS_URL=redis://default:password@your-redis-endpoint:port  # (Optional: If omitted, memory cache is used)
```

#### 2. Frontend Config
Create a `.env` file inside the `frontend` directory:
```bash
# frontend/.env
VITE_SUPABASE_URL=https://[your-supabase-project].supabase.co
VITE_SUPABASE_ANON_KEY=your_supabase_public_anon_key_here
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

---

### Step 3: Run Locally (Using Docker Compose - Recommended)
You can build and spin up the database, cache, backend, and frontend containers automatically:
```bash
docker compose up --build
```
* **Frontend:** http://localhost:5173
* **Backend:** http://localhost:8080

---

### Step 4: Run Locally (Manual Mode)

#### 1. Start Backend
The backend automatically runs migrations to create tables and enable the `pgvector` extension on startup.
```bash
cd backend
go run cmd/server/main.go
```

#### 2. Start Frontend
```bash
cd frontend
npm install
npm run dev
```

---

## 📖 Database Initialization Details
SprintGPT handles migrations automatically when the backend starts. It sets up the following schema:
* **`document_chunks`**: Stores vectorized segments of Wiki pages with Gemini embeddings.
* **`user_configs`**: Stores encrypted credentials for Azure DevOps sync.

*Note: If you run into RLS (Row Level Security) warnings in Supabase, make sure to enable RLS on these tables in the SQL editor:*
```sql
ALTER TABLE public.user_configs ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.document_chunks ENABLE ROW LEVEL SECURITY;
```

---

## 📂 Documentation & Flow Charts
For detailed diagrams and architectural blueprints, see:
* **System Topology & Request Flow:** See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)
* **Codebase Critique:** See [docs/critique.md](docs/critique.md)
* **Gaps to Production Readiness:** See [docs/gap_to_production_grade.md](docs/gap_to_production_grade.md)
* **Future Enhancement Roadmap:** See [docs/enhancements.md](docs/enhancements.md)
