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
  chapter       TEXT,                 -- e.g. "Chapter 3: Go's Concurrency Model"
  user_note     TEXT,                 -- optional personal memo ("why I flagged this")
  char_start    INT,                  -- start offset in the PDF.js text layer
  char_end      INT,                  -- end offset in the PDF.js text layer
  rects         JSONB,                -- serialized quadrilaterals for precise re-rendering
  embedding     vector(1536),         -- pgvector: for AI recall
  created_at    TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX ON annotations USING ivfflat (embedding vector_cosine_ops);
```

> **Why `rects`:** `window.getSelection()` returns reconstructed text that can be dirty (broken lines, extra spaces). Storing the serialized bounding boxes of the selection allows highlights to re-render precisely at any zoom level or screen resolution, independent of the text layer reconstruction.

### collections
```sql
CREATE TABLE collections (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID REFERENCES users(id) ON DELETE CASCADE,
  name        TEXT NOT NULL,           -- e.g. "Concurrency Patterns"
  description TEXT,
  is_smart    BOOLEAN DEFAULT false,   -- true = auto-populated via vector search
  query       TEXT,                    -- seed query for smart collections (V2)
  created_at  TIMESTAMPTZ DEFAULT now()
);
```

### collection_annotations
```sql
CREATE TABLE collection_annotations (
  collection_id UUID REFERENCES collections(id) ON DELETE CASCADE,
  annotation_id UUID REFERENCES annotations(id) ON DELETE CASCADE,
  added_at      TIMESTAMPTZ DEFAULT now(),
  PRIMARY KEY (collection_id, annotation_id)
);
```

### book_chapters
Populated once at upload time by parsing the PDF table of contents.
```sql
CREATE TABLE book_chapters (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  book_id    UUID REFERENCES books(id) ON DELETE CASCADE,
  title      TEXT NOT NULL,           -- "Chapter 3: Go's Concurrency Model"
  start_page INT NOT NULL,
  end_page   INT
);
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

```
POST   /api/collections                      — create collection
GET    /api/collections                      — list all user collections
GET    /api/collections/:id                  — get collection + annotations
DELETE /api/collections/:id                  — delete collection
POST   /api/collections/:id/annotations      — add annotation to collection
DELETE /api/collections/:id/annotations/:aid — remove annotation from collection
POST   /api/collections/:id/populate         — V2: auto-populate smart collection via vector search
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
2. A floating toolbar appears (highlight / strikethrough buttons) plus an optional one-line text input: *"Add a note — why did you flag this?"*
3. On action, the frontend reads `window.getSelection()` to get the selected text, determines the current page number from the PDF.js page viewer state, and looks up the chapter heading by matching the page number against the book's parsed chapter map (loaded into memory when the book opens).
4. The annotation is sent to the Go API with: `{ type, content, page, chapter, userNote, bookId }`.
5. The API saves it to PostgreSQL and asynchronously generates an embedding via the embeddings API. The embedded text includes the chapter and user note alongside the highlight content — making the vector richer and recall answers more precise.

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

## 10. Collections

Collections allow users to organise annotations across books into named folders or topics — breaking out of the per-book view into a knowledge-first view.

**Two modes:**

Manual collections are user-curated. The user creates a collection (e.g. "Concurrency Patterns") and explicitly adds highlights to it from any book. This is a simple many-to-many relationship between annotations and collections.

Smart collections are vector-powered (V2). The user names a topic and the system runs a similarity search across all their annotation embeddings to auto-populate the collection with the most semantically relevant highlights. The user can then curate from the result.

**Subdomain placement:** Collections live within the Annotation core subdomain. A `CollectionService` in `core/service/` depends only on the `AnnotationRepo` and `Embedder` outbound ports already defined — no new external dependencies.

**MVP vs V2:**
- V1 — manual collections (tagging UI, add/remove annotations)
- V2 — smart collections (auto-populate via vector search, alongside layered annotations)

---

## 11. File Upload Flow

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

## 14. Annotation Edge Cases

These rules live in the domain layer (`internal/core/domain/annotation.go`) and are enforced by `AnnotationList` before any persistence occurs.

### Edge case 1 — Duplicate annotation (debounce)

Two annotations on the same passage are allowed only if they carry different `user_note` values. An exact duplicate — same `content`, same `page`, same `type`, same `book_id` — created within a 30-second window is rejected as an accidental double-submit.

**Rule:** if `content + page + type` match an existing annotation and `createdAt` delta < 30s → return `ErrDuplicateAnnotation` (HTTP 409).

### Edge case 2 — Conflicting annotation type at the same offset

A highlight and a strikethrough cannot coexist on the same `char_start`/`char_end` range and `page`. If a user strikethroughs text they previously highlighted (or vice versa), the later annotation wins and the earlier one is automatically deleted before the new one is saved.

**Rule:** if `char_start + char_end + page` match an existing annotation of the opposite type → delete the existing annotation, then save the incoming one.

> **V2 note:** This "winner takes all" rule is intentionally simple for MVP. A common real-world case is highlighting a paragraph then striking a single sentence within it — partial overlaps. In V2, annotations will be treated as independent layers allowing nested and overlapping ranges without conflict resolution.

### Edge case 3 — User note added after the fact

A user may return to their annotation list and add or edit a `user_note` long after the annotation was created. This is a valid and expected flow. When a note is updated, the annotation's embedding is cleared and regenerated to reflect the richer context.

**Rule:** `PATCH /api/annotations/:id/note` → update `user_note`, clear `embedding`, trigger background re-embed using `chapter + content + user_note` as the composite embed text.

### Embed text composition

Regardless of whether it is a new annotation or a re-embed after a note update, the text sent to the embeddings API is always composed as:

```
"{book_title} | {chapter} | {content} | {user_note}"
```

Missing fields are omitted. The book title is included so cross-book recall queries like *"What did Designing Data-Intensive Applications say about replication?"* match correctly by title, not just by semantic content. This is the single source of truth defined on the `Annotation` domain struct via `EmbedText()`.

---

## 15. Key Technical Risks

| Risk | Mitigation |
|---|---|
| PDF.js text selection unreliable on some PDFs | Scope MVP to clean single-column PDFs; add fallback manual note input |
| Embedding cost at scale | Batch embed on save, cache results, only re-embed on content change |
| Large PDF upload performance | Stream directly to R2, never buffer full file in Go server memory |
| pgvector slow on large annotation sets | IVFFlat index with appropriate `lists` value; re-index at 10k+ rows |
