# Polling App — Backend (Go + Gin)

## Setup

1. Copy `.env.example` to `.env` and fill in real values (Mongo URI, Redis address, a real JWT secret).
2. Install dependencies and tidy the module (this machine had no network access to the Go module proxy, so this hasn't been run yet — run it locally):
   ```
   go mod tidy
   ```
3. Run the server:
   ```
   go run ./cmd/server
   ```
4. Check it's alive: `GET http://localhost:8080/health`

You'll need a MongoDB instance and a Redis instance reachable at the addresses in `.env` — either run them locally (`docker run -p 27017:27017 mongo`, `docker run -p 6379:6379 redis`) or point at MongoDB Atlas / Upstash for a deployed setup.

## What's here

- `cmd/server` — entrypoint, route wiring
- `internal/auth` — signup/login, JWT issue/verify, auth middleware
- `internal/polls` — create/get/list polls, all server-side validated
- `internal/votes` — vote casting: validates the option belongs to the poll, blocks an obvious duplicate vote via Redis, updates the live Redis counter, publishes to the poll's Redis channel, and writes a durable audit record to Mongo
- `internal/ws` — WebSocket hub: one Redis subscription per poll (not per client), fanning updates out to every connected viewer
- `internal/db` — Mongo and Redis connection setup
- `internal/models` — shared structs

## Key decisions (for the README/video)

- **Voting is anonymous, poll creation is not.** The brief only asks that poll creation have a real check — the audience shouldn't need an account to vote.
- **Redis does two real jobs**: atomic vote counters (`HINCRBY`) and pub/sub fan-out (`PUBLISH`/`SUBSCRIBE`) so every connected client gets pushed the new tally the moment a vote lands — no polling from the frontend.
- **Mongo stays the durable source of truth.** Redis holds the hot, frequently-mutated counters; Mongo holds polls, users, and an audit trail of individual votes.
- **Duplicate-vote guard is fingerprint-based** (hashed IP + User-Agent), not account-based, since voters aren't logged in. It's an honest "basic" deterrent, not a bulletproof one — worth saying that plainly if asked.

## Not yet done (frontend still to come)

This is the backend skeleton only. Next: the React frontend, then deployment.
