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

---