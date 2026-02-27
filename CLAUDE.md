# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make build      # Run templ generate + go build (outputs binary at ./main)
make run        # Run the application (go run cmd/api/main.go)
make test       # Run all tests with verbose output (go test ./... -v)
make watch      # Live reload with air
make coverage   # Coverage report excluding generated _templ.go files
make clean      # Remove binary and tmp files
```

Run a single test:

```bash
go test ./cmd/web/ -run TestArticlesPageHandler -v
```

**Important**: `templ generate` must be run before building or testing when `.templ` files have changed. `make build` handles this automatically; running `go test` or `go build` directly will use stale generated files.

## Architecture

This is a personal blog and content management system built with Go, Templ, and HTMX.

### Request Flow

```
cmd/api/main.go → internal/server/server.go (NewServer)
                → internal/server/routes.go (RegisterRoutes)
                → cmd/web/*.go (handlers)
                → internal/storage/ (YAML files or S3)
```

### Key Packages

- **`cmd/web/`** — HTTP handlers and Templ templates. Each content type has a `*.go` (handler logic), `*.templ` (template source), and generated `*_templ.go`. **Never edit `_templ.go` files directly.**
- **`internal/server/`** — HTTP server setup and route registration. Routes pass `storage.Storage` and `auth.Auth` into handlers.
- **`internal/storage/`** — Dual-mode storage abstraction (S3 or local filesystem). Content files are YAML; the `body` field is markdown converted to HTML at read time. Also manages the SQLite database connection.
- **`internal/model/`** — Shared `Document` struct (`title`, `subtitle`, `body`, `tags`) embedded by all content types via `yaml:",inline"`.
- **`internal/auth/`** — Cookie-based sessions (gorilla/sessions) with bcrypt password hashing. SQLite stores user records.
- **`internal/ai/`** — OpenAI GPT-4o integration for the writer's AI suggestion feature. Prompt instruction files live in `prompts/`.

### Content Types

Four content types all embed `model.Document`: `Article` (has `date`), `Project`, `ReadingList` (has `author`, `link`), `Letter`. Each is stored as a YAML file under `storage/{articles,projects,reading-list,letters}/`.

### HTMX Pattern

Handlers check `r.Header.Get("Hx-Request") == "true"` to return either a full page component or a partial fragment. This is the primary mechanism for partial page updates.

### Storage Modes

Controlled by the `USE_S3` env var. In local mode, files are read directly from `storage/`. In S3 mode, files are downloaded to `storage/` on demand before being served. The `storage/` directory always acts as a local cache.

### Required Environment Variables

| Variable                         | Purpose                                           |
| -------------------------------- | ------------------------------------------------- |
| `PORT`                           | HTTP server port                                  |
| `SESSION_NAME`                   | Cookie session key name                           |
| `USE_S3`                         | Set to `"true"` to use S3; otherwise local        |
| `AWS_BUCKET_NAME`                | S3 bucket (required if `USE_S3=true`)             |
| `AWS_REGION`                     | AWS region (required if `USE_S3=true`)            |
| `OPENAI_API_KEY`                 | Required for AI writer suggestions                |
| `SSL_CERT_FILE` / `SSL_KEY_FILE` | Optional TLS; server falls back to HTTP if absent |

Load via `.env` file — `godotenv/autoload` is imported in `internal/server/server.go`.
