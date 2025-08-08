# MindMenu

## Overview

MindMenu is a full-stack project structured as a monorepo with separate backend (`BE/`) and frontend (`FE/`) folders.

---

## Backend (`BE/`)

- **Language:** Go
- **Integrations:**
  - [Supabase](https://supabase.com/) (Go SDK)
  - [Pinecone](https://www.pinecone.io/) (Go SDK)
  - [Google Gemini (Generative Language)](https://ai.google.dev/) (Go SDK)
- **Structure:**
  - `main.go`: Initializes services and starts the server
  - `handlers.go`: HTTP handlers
  - `models.go`: Data models
  - `routes.go`: Route definitions
  - `go.mod`, `go.sum`: Dependency management
- **Deployment:**  
  Dockerfile provided for containerization. Deploy to your preferred cloud provider (e.g., Render, Railway, AWS, GCP, etc.).

---

## Frontend (`FE/my-app/`)

- **Framework:** Next.js (App Router, TypeScript)
- **Styling:** Tailwind CSS, custom fonts
- **Structure:**
  - `app/`: Main app directory
  - `public/`: Static assets
  - `package.json`: Project scripts and dependencies
- **Deployment:**  
  Recommended to deploy directly to [Vercel](https://vercel.com/).  
  Set the project root to `FE/my-app` when connecting your repo.

---

## Deployment Workflow

1. **Frontend:**  
   - Deploy `FE/my-app` to Vercel.  
   - Set environment variables (e.g., backend API URL) in the Vercel dashboard.

2. **Backend:**  
   - Build and deploy the Go backend (`BE/`) using Docker or your preferred method.
   - Ensure the backend is accessible via a public URL.

3. **Connecting Frontend & Backend:**  
   - Configure the frontend to use the backend’s public API URL via environment variables.

---

## Local Development

**Backend:**
```sh
cd BE
go run main.go
```

**Frontend:**
```sh
cd FE/my-app
npm install
npm run dev
```

- Demo UI: open http://localhost:3000/demo
- Configure FE to talk to BE by setting:
  ```
  NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
  ```
  in `FE/my-app/.env.local` (copy from `.env.example`).
- Verify connection:
  - Ensure the backend is running and http://localhost:8080/health returns `{"status":"ok"}`.
  - The /demo page shows a connectivity indicator and a Retry button.
  - If you see a message referencing a Supabase domain in the health check, your `NEXT_PUBLIC_API_BASE_URL` is misconfigured. It must point to the Go backend (e.g., `http://localhost:8080`), not your Supabase project URL.

### Example: Create Chatbot via cURL
```sh
curl -X POST http://localhost:8080/chatbots \
  -H "Content-Type: application/json" \
  -d '{
    "branch_id": "ad18ad2b-2d2a-4b59-a76b-044e1cab690a",
    "content": {
      "menu": {
        "appetizers": ["Spring Rolls - $8", "Chicken Wings - $12"],
        "mains": ["Grilled Salmon - $24", "Beef Burger - $16"],
        "desserts": ["Cheesecake - $8", "Chocolate Brownie - $7"],
        "non_alcoholic_drinks": ["Iced Tea - $4", "Orange Juice - $5"],
        "alcoholic_drinks": ["House Red Wine - $9", "Draft Beer - $7"]
      },
      "hours": "Monday–Sunday: 9AM–11PM"
    }
  }'
```

---