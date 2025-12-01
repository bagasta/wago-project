# WAGO – WhatsApp Multi-Session Gateway

A full-stack gateway for managing multiple WhatsApp Web sessions with webhook delivery, QR onboarding, auto-reconnect, and realtime dashboard updates.

## Key Features
- Multi-user, multi-session WhatsApp management (QR pairing, connect/disconnect)
- Auto-reconnect on backend restart using stored JIDs/device info
- Webhook forwarding for incoming messages with mention-aware group handling
- Realtime status + QR updates via WebSocket
- Session metadata persistence (JID/phone, device info, last connected)

## Project Structure
- `backend/` – Go 1.24 service (REST + WebSocket + WhatsApp integration)
- `frontend/` – React 18 + Vite + Tailwind dashboard
- `migrations/` – SQL schema for Postgres

## Requirements
- Go 1.24+
- Node 18+ and npm
- PostgreSQL 15+

## Quick Start
1) Clone and configure:
```bash
cp backend/.env.example backend/.env
# update DATABASE_URL, APP_PORT, JWT_SECRET as needed
```

2) Backend:
```bash
cd backend
go mod tidy
go run cmd/server/main.go
# Server runs migrations on startup and auto-reconnects stored sessions
```

3) Frontend:
```bash
cd frontend
npm install
npm run dev
```

## Backend Notes
- Auto-reconnect: on startup, sessions with stored `phone_number` (full JID) are reconnected and logged (`Reconnecting session: ...`).
- Group mention logic: bot replies only when mentioned; checks both user JID and LID variants.
- Migrations run automatically at boot from `backend/migrations/`.
- Tests: `cd backend && GOCACHE=$(mktemp -d) go test ./...`

## API & Auth
- Base path: `/api/v1`
- PIN-based auth (see `backend/HOW-TO-USE.md` for flow).
- WebSocket: `/ws/sessions/{id}?token=...` for QR/status updates per session.

## Common Tasks
- Create session: `POST /api/v1/sessions`
- Start session (QR/connect): `POST /api/v1/sessions/{id}/start`
- Delete session: `DELETE /api/v1/sessions/{id}`
- Health check: `/health`

## Deployment Hints
- Persist Postgres and the WhatsApp SQL store (same DB) across restarts.
- Expose only the frontend and backend HTTP ports; keep DB private.
- Ensure time sync (NTP) for stable WhatsApp connections.
