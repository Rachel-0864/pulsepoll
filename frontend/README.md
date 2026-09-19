# PulsePoll — Live Polling Frontend

React + Vite frontend for the Live Polling Tool architecture.

## Run

```bash
npm install
npm run dev
```

Open http://localhost:5173

## Current mode

The frontend includes a mock/demo mode so it can be tested before the Go backend is complete.

The API layer is already prepared for:

- POST /api/auth/signup
- POST /api/auth/login
- POST /api/polls
- GET /api/polls/:id
- GET /api/polls
- POST /api/polls/:id/vote
- WS /ws/poll/:id

Copy `.env.example` to `.env` when connecting to the real backend.

## Demo account

Any email/password can be used in demo mode.

## Main pages

- `/` — Home
- `/login` — Login
- `/signup` — Signup
- `/dashboard` — My polls
- `/create` — Create poll
- `/poll/:id` — Vote and view live results
