# ArtHub

ArtHub is the discussion layer for ART Analytics.

The domain model is now:

`Companies -> Posts -> Comments`

Each discussion stream is anchored to a company record with a unique ticker. Users browse companies, open a company page, and discuss the company through posts and comments.

## Stack

- Backend: Go, `net/http`, SQLite
- Frontend: React, TypeScript, Create React App
- Routing: `react-router-dom`
- Testing: Go `testing`, React Testing Library, Jest

## Architecture

### Backend

- [backend/main.go](/C:/Users/arthu_/Downloads/CampusCommons/backend/main.go)
  Starts the API server and initializes SQLite.
- [backend/server/router.go](/C:/Users/arthu_/Downloads/CampusCommons/backend/server/router.go)
  Registers HTTP routes and CORS.
- [backend/database/db.go](/C:/Users/arthu_/Downloads/CampusCommons/backend/database/db.go)
  Creates the schema and migrates legacy topic-based databases into the company model.
- [backend/handlers](/C:/Users/arthu_/Downloads/CampusCommons/backend/handlers)
  Session auth, company, post, comment, and vote handlers.

### Frontend

- [frontend/src/pages](/C:/Users/arthu_/Downloads/CampusCommons/frontend/src/pages)
  Route-level pages for company browsing, company detail, post detail, and create/edit flows.
- [frontend/src/services/api](/C:/Users/arthu_/Downloads/CampusCommons/frontend/src/services/api)
  API client and domain-specific request helpers.
- [frontend/src/types](/C:/Users/arthu_/Downloads/CampusCommons/frontend/src/types)
  DTO and app domain types.
- [frontend/src/context](/C:/Users/arthu_/Downloads/CampusCommons/frontend/src/context)
  Session bootstrap and auth state.

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
{ "data": ... }
```

Error responses use:

```json
{ "error": { "message": "...", "code": "..." } }
```

## Company Model

Companies are the root entity in the discussion tree.

- `id`
- `ticker`
  Required, unique, normalized to uppercase.
- `name`
  Required company name.
- `description`
  Optional summary used on company list/detail pages.

Posts belong to a company via `company_id`. Comments still belong to posts.

## Auth

ArtHub uses bcrypt password hashing plus server-side cookie sessions.

- Company create/edit/delete requires a valid session.
- Post create/edit/delete requires a valid session.
- Comment create requires a valid session.
- Voting requires a valid session.

## Local Setup

### Backend

```bash
cd backend
go run main.go
```

Default API host: `http://localhost:8080`

Optional environment variables:

- `CAMPUSCOMMONS_ALLOWED_ORIGIN`

### Frontend

```bash
cd frontend
npm install
cp .env.example .env
npm start
```

Default frontend host: `http://localhost:3000`

The frontend uses `REACT_APP_API_BASE_URL` and falls back to `http://localhost:8080`.

## Tests

### Backend

```bash
cd backend
go test ./...
```

### Frontend

```bash
cd frontend
npm test -- --watchAll=false
npm run build
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
