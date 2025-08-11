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

## Front-end Sides

- Client Side
  - Accessed by scanning a barcode/QR that encodes a session_id.
  - Only shows the chatbot interface scoped to that session.
  - Must validate session via BE before loading chat.

- Admin Side (Restaurant Owner)
  - Auth required (login/register).
  - Separate steps on distinct routes:
    - Create Restaurant (e.g., /admin/restaurants/new)
    - Create Branch (e.g., /admin/branches/new)
    - Create/Manage Chatbots (e.g., /admin/chatbots)
  - Data is retained and editable; no duplicate creation when revisiting.

---

## Auth

- Admin auth via BE (Supabase-backed). FE calls:
  - POST /auth/register → { user }
  - POST /auth/login → { token }
  - GET /me → { user }
- FE stores JWT in localStorage and sends Authorization: Bearer <token>.

---

## Client Session Validation

- FE calls GET /sessions/:session_id/validate.
- If ok=true, render chatbot UI; otherwise show error.
- The barcode should contain a URL to the client page with the session_id.

---

## Chatbots and Vector Updates

- Create chatbot: POST /chatbots with { branch_id, content }.
- Update vectors: POST /chatbots with same payload to upsert/update by content hash.
- BE should keep metadata, vector DB, and sessions synchronized.

---

## Expected Backend Endpoints

- Auth:
  - POST /auth/register, POST /auth/login, GET /me
- Restaurants/Branches/Chatbots:
  - POST /restaurants, POST /branches, POST /chatbots (create/update)
  - Optional GET/PUT endpoints to fetch/update without duplicates
- Chat:
  - POST /branches/:branch_id/query-with-history
- Client session:
  - GET /sessions/:session_id/validate

---

## Navigation

- Header links include Admin Portal and Client Chat placeholders.
- Implement routes under FE/my-app/app as needed:
  - /admin/... pages for each step
  - /c/[sessionId] page for client chat that calls validate endpoint

---

## Front-end Routes (remade)

- Admin
  - /admin → redirects to /admin/restaurants/new
  - /admin/restaurants/new → create/view restaurants (owned by logged-in admin)
  - /admin/branches/new → create branches for a selected restaurant
  - /admin/chatbots → create/update chatbot content and vectors for a selected branch

- Client
  - /c → enter or scan a sessionId and go to /c/[sessionId]
  - /c/[sessionId] → validates session via GET /sessions/:session_id/validate and shows chat scoped to that session

Notes:
- The /demo page remains for local testing but is not part of the new flow.
- Ensure BE supports list endpoints used by the Admin pages (see lib/api.ts).

---

## Temporary UI (until dedicated routes exist)

- Admin side
  - Use /demo to login/register (top box) and run the 3-step wizard.
  - All admin form fields persist to localStorage so you can revisit without losing data.

- Client side
  - Use /demo?sessionId=YOUR_SESSION_ID
  - The page validates the session via GET /sessions/:session_id/validate and only then enables chat, scoped to that session.
  - Intended as a stopgap for /c/[sessionId].

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

## Spec-driven Frontend (JSON)

Quick Start: Open /demo, click Load Spec, click Apply to Forms, then use the Create buttons in steps 1–3 to persist.

- Place your spec file at:
  - FE/my-app/public/fronend-spec.json
  - The app also tries /frontend-spec.json or /spec.json as fallback.
- Go to http://localhost:3000/demo → "Spec Loader" → Load Spec → Apply to Forms.
- The demo pre-fills Restaurant, Branch, and Chatbot content from the spec; you can then click Create to persist via the backend.

Example spec:
```json
{
  "admin": {
    "restaurant": { "name": "Sakura Bistro", "description": "Modern Japanese dining" },
    "branch": { "name": "Downtown", "address": "123 Main St" },
    "chatbot": {
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
    }
  },
  "client": {
    "sessionParam": "sessionId"
  }
}
```

Notes:
- The spec loader is non-destructive; it only fills the form fields.
- Ensure the backend endpoints are reachable before creating entities.