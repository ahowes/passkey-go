# passkey-go

A web application demonstrating passwordless authentication using passkeys (WebAuthn) built with Go.

## Tech Stack

- **Web framework:** [Gin](https://github.com/gin-gonic/gin)
- **WebAuthn:** [go-webauthn](https://github.com/go-webauthn/webauthn) (discoverable login / passkeys)
- **Database:** PostgreSQL via [Bun ORM](https://github.com/uptrace/bun)
- **Sessions:** Cookie-based via [gin-contrib/sessions](https://github.com/gin-contrib/sessions)

## Prerequisites

- Go 1.25.6+
- PostgreSQL

## Configuration

All settings are read from environment variables with sensible defaults for local development:

| Variable | Default | Description |
|---|---|---|
| `LISTEN_ADDR` | `:8080` | Address and port the server listens on |
| `DATABASE_DSN` | `postgres://localhost:5432/passkey_go?sslmode=disable` | PostgreSQL connection string |
| `WEBAUTHN_RP_DISPLAY_NAME` | `Passkey Go` | Relying Party display name shown to users |
| `WEBAUTHN_RP_ID` | `localhost` | Relying Party ID (typically the domain) |
| `WEBAUTHN_RP_ORIGINS` | `http://localhost:8080` | Comma-separated list of allowed origins |
| `SESSION_SECRET` | `super-secret-key-change-me-in-prod` | Secret used to sign session cookies |

## Getting Started

1. **Create the database:**

   ```sh
   createdb passkey_go
   ```

   Tables are created automatically on startup.

2. **Run the server:**

   ```sh
   go run .
   ```

3. **Open the app:** visit http://localhost:8080

## Routes

| Method | Path | Description |
|---|---|---|
| `GET` | `/` | Home page |
| `GET` | `/register` | Registration page |
| `GET` | `/login` | Login page |
| `GET` | `/dashboard` | Dashboard (requires auth) |
| `POST` | `/api/register/begin` | Start passkey registration |
| `POST` | `/api/register/finish` | Complete passkey registration |
| `POST` | `/api/login/begin` | Start passkey login |
| `POST` | `/api/login/finish` | Complete passkey login |
| `POST` | `/logout` | Log out and clear session |

## Project Structure

```
.
├── config/         # Environment-based configuration
├── db/             # Database connection and table creation
├── handlers/       # HTTP handlers (auth + pages)
├── middleware/      # Auth middleware
├── models/         # User and WebAuthn credential models
├── templates/      # HTML templates
└── main.go         # Entrypoint and route setup
```
