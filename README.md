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
- [frontend/src/context](/C:/Users/arthu_/Downloads/CampusCommons/frontend/src/context)
  Session-based auth state bootstrap and auth actions.
- [frontend/src/services/api](/C:/Users/arthu_/Downloads/CampusCommons/frontend/src/services/api)
  Centralized API client and forum API methods.
- [frontend/src/types](/C:/Users/arthu_/Downloads/CampusCommons/frontend/src/types)
  Standardized DTO and domain types.
- [frontend/src/config](/C:/Users/arthu_/Downloads/CampusCommons/frontend/src/config)
  Environment-based configuration such as the API base URL.

## API Overview

Core endpoints preserved:

- `POST /auth/signup`
- `POST /auth/login`
- `POST /auth/logout`
- `GET /auth/me`
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

Successful responses now use a consistent envelope:

```json
{ "data": ... }
```

Error responses use:

```json
{ "error": { "message": "...", "code": "..." } }
```

Protected mutations now authorize against the authenticated session user, not a client-supplied identity header.

- Topic create/edit/delete requires a valid session
- Post create/edit/delete requires a valid session
- Comment create requires a valid session
- Comments remain create/read only in this app version

## Auth Flow

ArtHub uses bcrypt password hashing plus server-side cookie sessions.

- `POST /auth/signup`
  Creates a password-backed account and starts a session.
- `POST /auth/login`
  Verifies credentials and sets an `HttpOnly` session cookie.
- `POST /auth/logout`
  Deletes the active session and clears the cookie.
- `GET /auth/me`
  Returns the current authenticated user from the session cookie.

Existing local databases may already contain users created before password auth existed. Those users can sign up again with the same username to attach a password, or you can delete the local SQLite file in `backend/data/` for a clean reset during development.

## Frontend Auth Flow

The React app now uses the backend session flow directly:

- session bootstrap from `GET /auth/me` on app load
- signup via `POST /auth/signup`
- login via `POST /auth/login`
- logout via `POST /auth/logout`
- protected create/edit routes redirect to `/login` when unauthenticated
- expired sessions clear local auth state and return the user to `/login` on protected flows

The frontend no longer treats `localStorage` as the source of truth for identity, and it no longer sends `X-User-ID` for authorization.

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

- Sign up with a username and password.
- Refresh the app and confirm the session is restored.
- Log out and confirm protected pages redirect to `/login`.
- Create a topic.
- Edit the topic.
- Create a post inside that topic.
- Edit the post.
- Add a comment to the post.
- Delete the post and confirm navigation returns to the topic.
- Delete the topic and confirm it disappears from the list.

## Notes For Future Reuse

The refactor keeps the forum domain generic. Topics, posts, and comments can later be attached to another product domain such as tickers, reports, or research workspaces without rebuilding the discussion layer from scratch.
