# Repository Guidelines

## Project Structure & Modules
- Root: `backend/` (Go 1.24 API, WhatsApp integration, migrations), `frontend/` (React/Vite dashboard), `migrations/` shared SQL.
- Backend key dirs: `cmd/server` (entry), `internal/{handler,service,repository,whatsapp,websocket,middleware,config,database,utils}`, `internal/whatsapp` for WA client/events, `internal/database/migrations.go` runner, `migrations/*.sql`.
- Frontend: `src/pages`, `src/hooks`, `src/components/ui`, `src/services`, `src/context`.

## Build, Test, Run
- Backend run: `cd backend && go run cmd/server/main.go` (runs migrations, starts API/WS).
- Backend tests: `cd backend && GOCACHE=$(mktemp -d) go test ./...`.
- Frontend dev: `cd frontend && npm install && npm run dev`.
- Frontend build: `cd frontend && npm run build`.

## Coding Style & Conventions
- Go: `gofmt` (tabs, std formatting). Keep logs minimal in prod; set `LOG_LEVEL`.
- JS/React: follow existing patterns (functional components, hooks). Use meaningful names; keep files ASCII.
- Naming: snake_case in JSON fields; Go uses PascalCase for exported structs. Session IDs are UUIDs, phone_number stores full JID.

## Testing Guidelines
- Go tests via `go test ./...` (no custom framework). Add table-driven tests when adding logic.
- Place tests alongside code (`*_test.go`). Ensure migrations don’t break idempotency (schema_migrations table in place).

## Commit & PR Guidelines
- Commits: concise, imperative summaries (e.g., “Fix WS auth check”, “Add mention detector”). Group related changes.
- PRs: include purpose, key changes, testing done (`go test ./...`, front-end build), screenshots for UI tweaks, and note any config/env changes (e.g., `ALLOWED_ORIGINS`, `LOG_LEVEL`).

## Security & Configuration Tips
- Set `ALLOWED_ORIGINS` to your frontend origins (avoid `*` in prod).
- WS requires `token` query param (JWT) and session ownership validation.
- Persist Postgres data; migrations are tracked via `schema_migrations`.
- Use graceful shutdown (built-in) and avoid clearing session JIDs on shutdown.
