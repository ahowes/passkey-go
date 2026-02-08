# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

passkey-go is a Go web application demonstrating passwordless authentication using FIDO2/WebAuthn passkeys with discoverable credentials (usernameless login). Built with Gin, Bun ORM (PostgreSQL), and the go-webauthn library.

## Commands

```bash
# Run the server (requires PostgreSQL with a "passkey_go" database)
go run .

# Run tests
go test ./...

# Run a single test
go test ./handlers -run TestBeginRegistration

# Build
go build -o passkey-go .

# Format code
gofmt -w .

# Vet
go vet ./...
```

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `LISTEN_ADDR` | `:8080` | Server listen address |
| `DATABASE_DSN` | `postgres://localhost:5432/passkey_go?sslmode=disable` | PostgreSQL connection string |
| `WEBAUTHN_RP_DISPLAY_NAME` | `Passkey Go` | WebAuthn Relying Party display name |
| `WEBAUTHN_RP_ID` | `localhost` | WebAuthn Relying Party ID (domain) |
| `WEBAUTHN_RP_ORIGINS` | `http://localhost:8080` | Comma-separated allowed origins |
| `SESSION_SECRET` | `super-secret-key-change-me-in-prod` | Cookie signing key |

## Architecture

### Package Layout

- **main.go** — Entrypoint: loads config, connects DB, creates tables, configures WebAuthn and session middleware, defines routes
- **config/** — Environment-based configuration with defaults
- **db/** — PostgreSQL connection via Bun ORM and auto table creation
- **models/** — `User` (implements `webauthn.User` interface) and `WebAuthnCredential` with conversion helpers between DB and go-webauthn library types
- **handlers/** — `AuthHandler` (registration/login WebAuthn flows, logout) and page rendering handlers
- **middleware/** — `RequireAuth` session-checking middleware
- **templates/** — HTML templates with embedded JavaScript for WebAuthn browser API calls

### WebAuthn Authentication Flow

**Registration:** Frontend POSTs to `/api/register/begin` → backend generates challenge, stores `SessionData` as JSON in cookie session → frontend calls `navigator.credentials.create()` → POSTs attestation to `/api/register/finish` → backend verifies and stores credential.

**Login (Discoverable):** Frontend POSTs to `/api/login/begin` (no username) → backend generates challenge → frontend calls `navigator.credentials.get()` → POSTs assertion to `/api/login/finish` → backend looks up user by `userHandle` from the assertion, verifies signature, updates sign count and clone warning flags.

### Key Design Decisions

- **Cookie-based sessions** (gin-contrib/sessions) rather than JWTs — session data includes serialized WebAuthn `SessionData` during ceremony flows
- **Discoverable credentials** — login requires no username; the authenticator provides the `userHandle` (user UUID)
- **Clone detection** — sign count is tracked per credential and `CloneWarning` flag is set if the count decreases
- **Auto-migration** — tables are created on startup via `db.CreateTables()`
- **Dependency injection** — `AuthHandler` receives `*bun.DB` and `*webauthn.WebAuthn` instances

### Database Schema

Two tables with a foreign key relationship:
- `users` — UUID primary key, unique username, display name, timestamps
- `webauthn_credentials` — stores credential ID, public key, AAGUID, sign count, attestation type, transport types, backup/verification flags; FK to users
