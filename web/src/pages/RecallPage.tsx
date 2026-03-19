import { useState, type FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { query as recallQuery } from '../api/recall';
import type { RecallResult } from '../types';

export default function RecallPage() {
  const [input, setInput] = useState('');
  const [result, setResult] = useState<RecallResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const navigate = useNavigate();

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!input.trim()) return;
    setLoading(true);
    setError('');
    try {
      const data = await recallQuery(input);
      setResult(data);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Recall failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="recall-page">
      <h1>AI Recall</h1>
      <p className="subtitle">Ask questions about your annotations across all books.</p>

      <form onSubmit={handleSubmit} className="search-form">
        <input
          type="text"
          placeholder="What did I learn about..."
          value={input}
          onChange={(e) => setInput(e.target.value)}
        />
        <button type="submit" className="btn-primary" disabled={loading}>
          {loading ? 'Thinking...' : 'Ask'}
        </button>
      </form>

      {error && <div className="error-msg">{error}</div>}

      {result && (
        <div className="recall-result">
          <div className="recall-answer">
            <h3>Answer</h3>
            <p style={{ whiteSpace: 'pre-wrap' }}>{result.answer}</p>
          </div>

          {result.sources && result.sources.length > 0 && (
            <div className="recall-sources">
              <h3>Sources</h3>
              {result.sources.map((s, i) => (
                <div
                  key={i}
                  className="annotation-card clickable"
                  onClick={() => navigate(`/books/${s.book_id}?page=${s.page}`)}
                >
                  <p className="annotation-content">{s.content}</p>
                  <div className="annotation-meta">
                    <span>{s.book_title}</span>
                    <span>Page {s.page}</span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
