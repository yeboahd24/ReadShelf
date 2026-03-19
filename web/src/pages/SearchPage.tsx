import { useState, type FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { searchAnnotations } from '../api/annotations';
import type { Annotation } from '../types';

export default function SearchPage() {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<Annotation[]>([]);
  const [searched, setSearched] = useState(false);
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const handleSearch = async (e: FormEvent) => {
    e.preventDefault();
    if (!query.trim()) return;
    setLoading(true);
    try {
      const data = await searchAnnotations(query);
      setResults(data);
      setSearched(true);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="search-page">
      <h1>Search Annotations</h1>
      <form onSubmit={handleSearch} className="search-form">
        <input
          type="text"
          placeholder="Search across all your annotations..."
          value={query}
          onChange={(e) => setQuery(e.target.value)}
        />
        <button type="submit" className="btn-primary" disabled={loading}>
          {loading ? 'Searching...' : 'Search'}
        </button>
      </form>

      {searched && results.length === 0 && (
        <p className="empty-state">No annotations found for "{query}".</p>
      )}

      <div className="annotation-list">
        {results.map((a) => (
          <div
            key={a.id}
            className="annotation-card clickable"
            onClick={() => navigate(`/books/${a.book_id}?page=${a.page}`)}
          >
            <span className={`badge badge-${a.type}`}>{a.type}</span>
            <p className="annotation-content">{a.content}</p>
            <div className="annotation-meta">
              <span>{a.book_title}</span>
              <span>Page {a.page}</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
