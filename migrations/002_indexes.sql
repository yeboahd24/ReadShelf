-- Full-text search index on annotation content
ALTER TABLE annotations ADD COLUMN IF NOT EXISTS tsv tsvector
    GENERATED ALWAYS AS (to_tsvector('english', content)) STORED;

CREATE INDEX idx_annotations_tsv ON annotations USING GIN (tsv);

-- Vector similarity index for AI recall
CREATE INDEX idx_annotations_embedding ON annotations
    USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);

-- Foreign key lookups
CREATE INDEX idx_books_user_id ON books (user_id);
CREATE INDEX idx_annotations_book_id ON annotations (book_id);
CREATE INDEX idx_annotations_user_id ON annotations (user_id);

-- Duplicate detection: same content + page + type within time window
CREATE INDEX idx_annotations_dedup ON annotations (user_id, book_id, page, type, content, created_at);
