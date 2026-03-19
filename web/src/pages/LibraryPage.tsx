import { useState, useEffect, useRef, type FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import * as booksApi from '../api/books';
import type { Book } from '../types';

const COLORS = ['#4A90D9', '#E74C3C', '#2ECC71', '#F39C12', '#9B59B6', '#1ABC9C'];

export default function LibraryPage() {
  const [books, setBooks] = useState<Book[]>([]);
  const [showUpload, setShowUpload] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState('');
  const navigate = useNavigate();

  useEffect(() => {
    booksApi.listBooks().then(setBooks).catch(console.error);
  }, []);

  const handleDelete = async (id: string, e: React.MouseEvent) => {
    e.stopPropagation();
    if (!confirm('Delete this book and all its annotations?')) return;
    try {
      await booksApi.deleteBook(id);
      setBooks((prev) => prev.filter((b) => b.id !== id));
    } catch (err: any) {
      alert(err.response?.data?.error || 'Delete failed');
    }
  };

  return (
    <div className="library-page">
      <div className="library-header">
        <h1>My Library</h1>
        <button onClick={() => setShowUpload(true)} className="btn-primary">
          + Upload PDF
        </button>
      </div>

      {error && <div className="error-msg">{error}</div>}

      {showUpload && (
        <UploadDialog
          onClose={() => setShowUpload(false)}
          onUploaded={(book) => {
            setBooks((prev) => [book, ...prev]);
            setShowUpload(false);
          }}
          uploading={uploading}
          setUploading={setUploading}
          setError={setError}
        />
      )}

      {books.length === 0 ? (
        <p className="empty-state">No books yet. Upload a PDF to get started.</p>
      ) : (
        <div className="book-grid">
          {books.map((book, i) => (
            <div
              key={book.id}
              className="book-card"
              onClick={() => navigate(`/books/${book.id}`)}
            >
              <div
                className="book-spine"
                style={{ backgroundColor: book.color || COLORS[i % COLORS.length] }}
              />
              <div className="book-info">
                <h3>{book.title}</h3>
                {book.author && <p className="book-author">{book.author}</p>}
                <p className="book-date">
                  {new Date(book.created_at).toLocaleDateString()}
                </p>
              </div>
              <button
                className="btn-delete"
                onClick={(e) => handleDelete(book.id, e)}
                title="Delete book"
              >
                ×
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

function UploadDialog({
  onClose,
  onUploaded,
  uploading,
  setUploading,
  setError,
}: {
  onClose: () => void;
  onUploaded: (book: Book) => void;
  uploading: boolean;
  setUploading: (v: boolean) => void;
  setError: (v: string) => void;
}) {
  const [title, setTitle] = useState('');
  const [author, setAuthor] = useState('');
  const fileRef = useRef<HTMLInputElement>(null);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    const file = fileRef.current?.files?.[0];
    if (!file) return;
    setUploading(true);
    setError('');
    try {
      const book = await booksApi.uploadBook(file, title || undefined, author || undefined);
      onUploaded(book);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Upload failed');
    } finally {
      setUploading(false);
    }
  };

  return (
    <div className="dialog-overlay" onClick={onClose}>
      <form className="dialog" onClick={(e) => e.stopPropagation()} onSubmit={handleSubmit}>
        <h2>Upload PDF</h2>
        <input
          ref={fileRef}
          type="file"
          accept=".pdf"
          required
          onChange={() => {
            const name = fileRef.current?.files?.[0]?.name;
            if (name && !title) {
              setTitle(name.replace(/\.pdf$/i, '').replace(/[._-]/g, ' ').trim());
            }
          }}
        />
        <input
          type="text"
          placeholder="Title (optional, defaults to filename)"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
        />
        <input
          type="text"
          placeholder="Author (optional)"
          value={author}
          onChange={(e) => setAuthor(e.target.value)}
        />
        <div className="dialog-actions">
          <button type="button" onClick={onClose} disabled={uploading}>Cancel</button>
          <button type="submit" className="btn-primary" disabled={uploading}>
            {uploading ? 'Uploading...' : 'Upload'}
          </button>
        </div>
      </form>
    </div>
  );
}
