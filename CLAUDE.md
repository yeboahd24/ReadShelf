# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ReadShelf is a web-based PDF reading and annotation management platform. Users upload PDFs, read them in-browser via PDF.js, annotate text (highlights/strikethroughs), and query their annotations using AI-powered natural language recall.

## Tech Stack

- **Frontend:** React + Vite, PDF.js for rendering
- **Backend:** Go (net/http or Chi router), hexagonal architecture
- **Database:** PostgreSQL 16 with pgvector for embedding search
- **File storage:** Cloudflare R2 (S3-compatible)
- **Auth:** JWT (access + refresh tokens, bcrypt password hashing)
- **AI:** Claude API (claude-sonnet-4-6) for embeddings and recall

## Architecture

The Go backend uses **hexagonal architecture** (ports and adapters):

- `cmd/server/main.go` — wiring and startup (only place concrete adapters are instantiated)
- `internal/core/domain/` — domain structs (User, Book, Annotation), domain errors
- `internal/core/port/inbound/` — driving port interfaces (what the app exposes)
- `internal/core/port/outbound/` — driven port interfaces (what the app depends on)
- `internal/core/service/` — use case implementations (implement inbound ports)
- `internal/adapter/inbound/http/` — HTTP handlers and JWT middleware
- `internal/adapter/outbound/postgres/` — PostgreSQL repository implementations
- `internal/adapter/outbound/r2/` — Cloudflare R2 file storage
- `internal/adapter/outbound/claude/` — Claude API embedder and AI client
- `migrations/` — SQL migration files

**Dependency rule:** `core` never imports from `adapter`. Adapters implement interfaces from `core/port`. Dependency injection happens only in `main.go`.

## Domain Rules

Key business logic enforced in the domain layer (`internal/core/domain/`):

- **Duplicate debounce:** Same `content + page + type` within 30 seconds → reject (HTTP 409)
- **Conflicting annotation types:** A highlight and strikethrough on the same `char_start/char_end/page` cannot coexist — the later one wins and deletes the earlier
- **Embed text composition:** Always `"{book_title} | {chapter} | {content} | {user_note}"` (missing fields omitted), defined via `Annotation.EmbedText()`
- **Note updates trigger re-embedding:** `PATCH /api/annotations/:id/note` clears and regenerates the embedding

## API Routes

- Auth: `POST /api/auth/{register,login,refresh}`
- Books: `GET|POST /api/books`, `GET|DELETE /api/books/:id`
- Annotations: `GET|POST /api/books/:id/annotations`, `DELETE /api/annotations/:id`, `GET /api/annotations/search?q=`
- AI Recall: `POST /api/recall`

## Key Design Decisions

- PDFs are streamed directly to R2 on upload; the Go server never buffers the full file
- Frontend fetches signed URLs from the API; PDF.js loads files directly from R2
- Embeddings are generated asynchronously via background goroutines on annotation save
- MVP scoped to clean single-column PDFs; multi-column/scanned PDF support is post-MVP
- Annotation `rects` (JSONB) store selection bounding boxes for precise re-rendering across zoom levels
- Cross-book search uses PostgreSQL `tsvector`; AI recall uses pgvector cosine similarity (IVFFlat index)
