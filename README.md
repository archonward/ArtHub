# ArtHub

ArtHub is a local discussion app for ART Analytics.

The domain model is:

`Companies -> Posts -> Comments`

Each discussion stream is anchored to a company record with a unique ticker. Users browse companies, open a company page, and discuss the company through posts and comments.

## Stack

- Backend: Go, `net/http`, SQLite
- Frontend: Vue 3, TypeScript, Vite
- Routing: `vue-router`
- Testing: Go `testing`, Vitest

## Local Setup

Run the backend and frontend in two terminals.

### Backend

```powershell
cd backend
go mod download
go run main.go
```

Default backend URL: `http://localhost:8080`

The backend creates a local SQLite database at `backend/data/arthub.db` by default.

Optional backend environment variables:

- `PORT`: backend port. Default: `8080`
- `DB_PATH`: SQLite database path. Default: `data/arthub.db`
- `ARTHUB_ALLOWED_ORIGINS`: comma-separated frontend origins allowed by CORS. Default: `http://localhost:3000,http://127.0.0.1:3000`
- `ARTHUB_COOKIE_SECURE`: set to `true` only when serving over HTTPS. Default: `false`
- `ARTHUB_COOKIE_SAMESITE`: `lax`, `strict`, or `none`. Default: `lax`

### Frontend

```powershell
cd frontend
npm install
npm run dev
```

Default frontend URL: `http://localhost:3000`

The checked-in local example and local `.env` point the frontend at:

```text
VITE_API_BASE_URL=http://localhost:8080
```

## Tests

### Backend

```powershell
cd backend
go test ./...
```

### Frontend

```powershell
cd frontend
npm run build
npm test
```

## API Overview

- `POST /auth/signup`
- `POST /auth/login`
- `POST /auth/logout`
- `GET /auth/me`
- `GET /companies`
- `POST /companies`
- `GET /companies/:id`
- `PUT /companies/:id`
- `DELETE /companies/:id`
- `GET /companies/:id/posts`
- `POST /companies/:id/posts`
- `GET /posts/:id`
- `PUT /posts/:id`
- `DELETE /posts/:id`
- `GET /posts/:id/comments`
- `POST /posts/:id/comments`
- `POST /posts/:id/vote`
- `DELETE /posts/:id/vote`

Successful responses use:

```json
{ "data": "..." }
```

Error responses use:

```json
{ "error": { "message": "...", "code": "..." } }
```

## Manual Smoke Test

- Sign up and confirm the session is restored on refresh.
- Create a company with a ticker and name.
- Edit the company.
- Create a post under that company.
- Sort company posts by `Top` and `New`.
- Page through company posts.
- Vote on a post.
- Add a comment to the post.
- Delete the post and confirm navigation returns to the company page.
- Delete the company and confirm it disappears from the company list.
