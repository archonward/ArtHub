# ArtHub

ArtHub is an early-stage full-stack discussion app built around a simple forum hierarchy:

`Topics -> Posts -> Comments`

The current refactor keeps the existing Go + SQLite backend and React + TypeScript frontend intact while making the codebase easier to extend into a reusable community module later.

## Stack

- Backend: Go, `net/http`, SQLite
- Frontend: React, TypeScript, Create React App
- Routing: `react-router-dom`
- Testing: Go `testing`, React Testing Library + Jest

## Architecture

### Backend

- [backend/main.go](/C:/Users/arthu_/Downloads/CampusCommons/backend/main.go)
  Server entry point. Initializes SQLite and starts the HTTP server.
- [backend/server/router.go](/C:/Users/arthu_/Downloads/CampusCommons/backend/server/router.go)
  Central route registration and CORS setup.
- [backend/database/db.go](/C:/Users/arthu_/Downloads/CampusCommons/backend/database/db.go)
  Database initialization and schema creation.
- [backend/handlers](/C:/Users/arthu_/Downloads/CampusCommons/backend/handlers)
  Resource-oriented handlers plus shared HTTP/validation helpers.

### Frontend

- [frontend/src/pages](/C:/Users/arthu_/Downloads/CampusCommons/frontend/src/pages)
  Route-level pages for login, topic, post, and edit/create flows.
- [frontend/src/components](/C:/Users/arthu_/Downloads/CampusCommons/frontend/src/components)
  Shared layout and feedback primitives.
- [frontend/src/services/api](/C:/Users/arthu_/Downloads/CampusCommons/frontend/src/services/api)
  Centralized API client and forum API methods.
- [frontend/src/types](/C:/Users/arthu_/Downloads/CampusCommons/frontend/src/types)
  Standardized DTO and domain types.
- [frontend/src/config](/C:/Users/arthu_/Downloads/CampusCommons/frontend/src/config)
  Environment-based configuration such as the API base URL.

## API Overview

Core endpoints preserved:

- `POST /login`
- `GET /topics`
- `POST /topics`
- `GET /topics/:id`
- `PUT /topics/:id`
- `DELETE /topics/:id`
- `GET /topics/:id/posts`
- `POST /topics/:id/posts`
- `GET /posts/:id`
- `PUT /posts/:id`
- `DELETE /posts/:id`
- `GET /posts/:id/comments`
- `POST /posts/:id/comments`

Authorization scaffolding is intentionally lightweight for now. Update and delete handlers accept an optional `X-User-ID` header and reject mismatched owners. This keeps the current app usable while making a future auth layer easier to slot in.

## Local Setup

### Backend

1. Install Go.
2. Start the API:

```bash
cd backend
go run main.go
```

By default the backend runs on `http://localhost:8080`.

Optional environment variables:

- `CAMPUSCOMMONS_ALLOWED_ORIGIN`

### Frontend

1. Install dependencies:

```bash
cd frontend
npm install
```

2. Create an env file if you want a non-default API host:

```bash
cp .env.example .env
```

3. Start the frontend:

```bash
npm start
```

The frontend defaults to `http://localhost:3000` and uses `REACT_APP_API_BASE_URL`, falling back to `http://localhost:8080`.

## Tests

### Frontend

```bash
cd frontend
npm test -- --watchAll=false
npm run build
```

### Backend

```bash
cd backend
go test ./...
```

## CRA vs Vite

The frontend still uses Create React App on purpose. Migrating to Vite is reasonable later, but this refactor keeps the current tooling so structural cleanup and behavior changes stay separate. Once the API layer, routes, and shared UI are stable, a Vite migration becomes a smaller and safer follow-up change.

## Manual Smoke Test

- Log in with a new username.
- Create a topic.
- Edit the topic.
- Create a post inside that topic.
- Edit the post.
- Add a comment to the post.
- Delete the post and confirm navigation returns to the topic.
- Delete the topic and confirm it disappears from the list.

## Notes For Future Reuse

The refactor keeps the forum domain generic. Topics, posts, and comments can later be attached to another product domain such as tickers, reports, or research workspaces without rebuilding the discussion layer from scratch.
