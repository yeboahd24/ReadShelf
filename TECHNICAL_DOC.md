# ReadShelf — Technical Document

**Version:** 0.1 (MVP)
**Last updated:** March 2026
**Author:** Dominic

---

## 1. Overview

ReadShelf is a web-based PDF reading and annotation management platform. Users upload PDF books, read them in the browser, and annotate text via highlights or strikethroughs. Every annotation is automatically tagged with the book title, page number, and timestamp, and organised into a searchable personal knowledge base. An AI recall layer allows users to query their annotations using natural language across all their books.

---

## 2. Goals

**Primary goal:** Give developers and avid readers a single place to read PDF books and capture what they learn, with zero manual organisation.

**MVP goals:**
- PDF upload and in-browser reading
- Highlight and strikethrough annotation capture
- Annotations auto-tagged with book, page, and date
- Per-book annotation view
- Cross-book annotation search

**Post-MVP goals:**
- AI recall ("ask questions over your highlights")
- Export to Markdown, Notion, Obsidian
- Browser extension for web articles
- Mobile-responsive reader

---

## 3. System Architecture

```
┌─────────────────────────────────────────┐
│           React Frontend                │
│  PDF.js reader · annotation events      │
│  highlight/strikethrough capture        │
└────────────────┬────────────────────────┘
                 │ HTTPS (REST)
┌────────────────▼────────────────────────┐
│           Go HTTP API                   │
│  Upload · annotation CRUD · search      │
│  AI query handler · auth (JWT)          │
└──────┬─────────────┬────────────────────┘
       │             │                │
┌──────▼──────┐ ┌────▼──────┐ ┌──────▼──────┐
│ PostgreSQL  │ │  Object   │ │  pgvector   │
│ books,      │ │  Storage  │ │  annotation │
│ annotations │ │  (PDF     │ │  embeddings │
│ users       │ │  files)   │ │             │
└─────────────┘ └───────────┘ └──────┬──────┘
                                     │
                              ┌──────▼──────┐
                              │ Claude API  │
                              │ embeddings  │
                              │ + AI recall │
                              └─────────────┘
```

---

## 4. Tech Stack

| Layer | Technology | Reason |
|---|---|---|
| Frontend | React + Vite | Fast dev experience, component model fits the reader UI |
| PDF rendering | PDF.js | Open source, runs in browser, exposes text layer for annotation |
| Backend | Go (net/http or Chi router) | Performance, clean concurrency, compiled binary |
| Database | PostgreSQL 16 | Relational, reliable, pgvector extension for AI search |
| Vector search | pgvector | Avoids a separate vector DB, co-located with app data |
| File storage | Cloudflare R2 | S3-compatible, cheap egress, good latency globally |
| Auth | JWT (access + refresh tokens) | Stateless, simple to implement in Go |
| AI layer | Claude API (claude-sonnet-4-6) | Annotation embedding + natural language recall |
| Hosting | Fly.io or Railway | Simple Go deployment, close to zero ops overhead |

---

## 5. Database Schema

### users
```sql
CREATE TABLE users (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email       TEXT UNIQUE NOT NULL,
  password    TEXT NOT NULL,          -- bcrypt hash
  plan        TEXT DEFAULT 'free',    -- 'free' | 'pro'
  created_at  TIMESTAMPTZ DEFAULT now()
);
```

### books
```sql
CREATE TABLE books (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id      UUID REFERENCES users(id) ON DELETE CASCADE,
  title        TEXT NOT NULL,
  author       TEXT,
  file_key     TEXT NOT NULL,         -- object storage key
  page_count   INT,
  color        TEXT,                  -- spine colour in UI
  created_at   TIMESTAMPTZ DEFAULT now()
);
```

### annotations
```sql
CREATE TABLE annotations (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  book_id       UUID REFERENCES books(id) ON DELETE CASCADE,
  user_id       UUID REFERENCES users(id) ON DELETE CASCADE,
  type          TEXT NOT NULL,        -- 'highlight' | 'strikethrough'
  content       TEXT NOT NULL,        -- selected text
  page          INT NOT NULL,
  char_offset   INT,                  -- position in page text layer
  embedding     vector(1536),         -- pgvector: for AI recall
  created_at    TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX ON annotations USING ivfflat (embedding vector_cosine_ops);
```

---

## 6. API Endpoints

### Auth
```
POST   /api/auth/register
POST   /api/auth/login
POST   /api/auth/refresh
```

### Books
```
GET    /api/books              — list all books for user
POST   /api/books              — upload new PDF
GET    /api/books/:id          — get book metadata
DELETE /api/books/:id          — delete book + annotations
```

### Annotations
```
GET    /api/books/:id/annotations   — list annotations for a book
POST   /api/books/:id/annotations   — save new annotation
DELETE /api/annotations/:id         — delete annotation
GET    /api/annotations/search?q=   — cross-book full-text search
```

### AI Recall
```
POST   /api/recall             — natural language query over user's highlights
```

---

## 7. PDF Annotation Capture (Frontend)

This is the most technically nuanced part of the MVP.

PDF.js renders each page as a canvas layer with a transparent text layer on top. The text layer is made of `<span>` elements that map to the actual characters in the PDF.

**Capture flow:**
1. User selects text in the PDF viewer — the browser fires a `selectionchange` event.
2. A floating toolbar appears (highlight / strikethrough buttons).
3. On action, the frontend reads `window.getSelection()` to get the selected text and determines the current page number from the PDF.js page viewer state.
4. The annotation is sent to the Go API with: `{ type, content, page, bookId }`.
5. The API saves it to PostgreSQL and asynchronously generates an embedding via the Claude API, storing the vector in pgvector.

**Key challenge:** PDF.js text layer spans don't always map 1:1 to visible text — especially in multi-column or scanned PDFs. For MVP, handle clean single-column PDFs first and add robustness later.

---

## 8. AI Recall Flow

```
User types: "What did I learn about goroutine scheduling?"
        │
        ▼
Go API embeds the query via Claude API → float32[1536]
        │
        ▼
pgvector cosine similarity search over user's annotation embeddings
        │
        ▼
Top 5 most similar annotations returned (with book title + page)
        │
        ▼
Claude API generates a summary answer grounded in those passages
        │
        ▼
Response: answer + source citations (book, page, original text)
```

Embeddings are generated once per annotation on save (background goroutine). Recall queries are fast because pgvector's IVFFlat index handles similarity search efficiently even at tens of thousands of annotations.

---

## 9. Go Backend Structure

The backend follows **hexagonal architecture** (ports and adapters). The domain core has zero knowledge of HTTP, PostgreSQL, or any external service — it only speaks through interfaces (ports). Adapters implement those interfaces for each concrete technology.

```
/cmd
  /server
    main.go                  — wire everything together, start HTTP server

/internal

  /core                      — domain layer (no imports from adapters)
    /domain
      user.go                — User, Book, Annotation structs
      errors.go              — domain-level error types
    /port
      /inbound               — driving ports (what the app exposes)
        annotation_service.go
        book_service.go
        auth_service.go
        recall_service.go
      /outbound              — driven ports (what the app depends on)
        annotation_repo.go   — interface: SaveAnnotation, ListByBook, Search...
        book_repo.go         — interface: SaveBook, GetByID, ListByUser...
        file_store.go        — interface: Upload, SignedURL, Delete
        embedder.go          — interface: Embed(text) ([]float32, error)
        ai_client.go         — interface: Recall(query, passages) (string, error)
    /service                 — use case implementations (implements inbound ports)
      annotation_service.go
      book_service.go
      auth_service.go
      recall_service.go

  /adapter
    /inbound
      /http                  — HTTP adapter (driving)
        server.go
        handler/
          auth.go
          books.go
          annotations.go
          recall.go
        middleware/
          auth.go            — JWT validation
    /outbound
      /postgres              — DB adapter (driven)
        annotation_repo.go
        book_repo.go
        user_repo.go
      /r2                    — Cloudflare R2 adapter (driven)
        file_store.go
      /claude                — Claude API adapter (driven)
        embedder.go
        ai_client.go

/migrations
  001_initial.sql
  002_pgvector.sql
```

**The dependency rule:** `core` has no knowledge of `adapter`. Adapters import from `core/port` to implement interfaces. `cmd/server/main.go` is the only place where concrete adapters are instantiated and injected into services.

**Example — saving an annotation:**
```
HTTP handler (adapter/inbound/http)
  → calls AnnotationService.Save() (core/service)
    → calls AnnotationRepo.Save() (outbound port interface)
      → implemented by postgres.AnnotationRepo (adapter/outbound/postgres)
    → calls Embedder.Embed() (outbound port interface)
      → implemented by claude.Embedder (adapter/outbound/claude)
```

This structure makes it straightforward to swap PostgreSQL for another DB, or replace the Claude embedder with OpenAI, without touching any business logic.

---

## 10. File Upload Flow

1. Frontend sends `multipart/form-data` POST to `/api/books`
2. Go handler validates file type (PDF only) and size limit (50MB free / 200MB pro)
3. File streamed directly to Cloudflare R2 using the S3-compatible SDK
4. File key (R2 object path) saved to `books.file_key` in PostgreSQL
5. PDF page count extracted using a Go PDF library (`pdfcpu` or `ledongthuc/pdfutil`)
6. Book record returned to frontend

For reading, the frontend fetches a short-lived signed URL from `/api/books/:id/url` and PDF.js loads the file directly from R2 — the Go server never proxies the file bytes.

---

## 11. Auth & Security

- Passwords hashed with bcrypt (cost factor 12)
- JWT access tokens (15 min expiry) + refresh tokens (30 days, stored in httpOnly cookie)
- All API routes protected by JWT middleware except `/auth/register` and `/auth/login`
- PDF files in R2 are private — accessed only via signed URLs generated server-side
- User data is fully isolated at the query level (`WHERE user_id = $1` on all queries)

---

## 12. Business Model

| Tier | Price | Limits |
|---|---|---|
| Free | $0 | 3 books, unlimited annotations, basic search |
| Pro | $7/month | Unlimited books, AI recall, export (MD/Notion/Obsidian) |

Monetisation milestone: 200 pro users = $1,400 MRR. Target: developer audience globally.

---

## 13. MVP Build Order

1. Go project scaffold + PostgreSQL setup + migrations
2. Auth (register, login, JWT middleware)
3. PDF upload to R2 + book CRUD
4. React frontend — library view + PDF.js reader
5. Annotation capture (selection → toolbar → API call)
6. Annotation panel — per-book list with badges
7. Cross-book search (full-text via PostgreSQL `tsvector`)
8. pgvector setup + embedding on annotation save
9. AI recall endpoint + frontend "Ask AI" tab
10. Export feature (Markdown, Notion API)

---

## 14. Key Technical Risks

| Risk | Mitigation |
|---|---|
| PDF.js text selection unreliable on some PDFs | Scope MVP to clean single-column PDFs; add fallback manual note input |
| Embedding cost at scale | Batch embed on save, cache results, only re-embed on content change |
| Large PDF upload performance | Stream directly to R2, never buffer full file in Go server memory |
| pgvector slow on large annotation sets | IVFFlat index with appropriate `lists` value; re-index at 10k+ rows |
