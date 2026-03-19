import { useState } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export default function LandingPage() {
  const { user } = useAuth();
  const [activeTab, setActiveTab] = useState<'annotations' | 'askai'>('annotations');

  return (
    <div className="landing">
      {/* Nav */}
      <nav className="landing-nav">
        <div className="logo">Read<span>Shelf</span></div>
        <div className="landing-nav-links">
          <a href="#features">Features</a>
          <a href="#how">How it works</a>
          {user ? (
            <Link to="/library" className="btn btn-primary">Go to Library</Link>
          ) : (
            <>
              <Link to="/login" className="landing-link">Sign In</Link>
              <Link to="/register" className="btn btn-primary">Get Started Free</Link>
            </>
          )}
        </div>
      </nav>

      {/* Hero */}
      <section className="l-hero">
        <div className="l-hero-glow" />
        <div className="l-badge"><span className="l-badge-dot" /> Now in beta</div>
        <h1>Remember <span className="l-highlight">everything</span> you read</h1>
        <p>Upload your PDFs. Highlight what matters. Ask questions later. ReadShelf turns your reading into a searchable, AI-powered knowledge base.</p>
        <div className="l-hero-cta">
          <Link to="/register" className="btn btn-primary">Get started free</Link>
          <a href="#how" className="btn btn-secondary">See how it works</a>
        </div>
      </section>

      {/* App Mockup */}
      <section className="mockup-section">
        <div className="mockup">
          <div className="mockup-topbar">
            <div className="mockup-logo">Read<span>Shelf</span></div>
            <div className="mockup-book-title"><em>Concurrency in Go</em> &mdash; Katherine Cox-Buday</div>
            <div className="mockup-page-nav">
              <div className="mockup-page-btn">&lsaquo;</div>
              <div className="mockup-page-num">84</div>
              <div className="mockup-page-btn">&rsaquo;</div>
            </div>
          </div>
          <div className="mockup-body">
            <div className="mockup-reader">
              <div className="mock-text">
                has at least one goroutine: the main goroutine.
              </div>
              <div className="mock-text">
                Every Go program has at least one goroutine.
              </div>
              <div className="mock-text" style={{ position: 'relative' }}>
                <span className="mock-selection">In fact, on most architectures, the Go scheduler uses an M:N scheduling model, multiplying M goroutines onto N OS threads.</span> This gives the runtime a lot of flexibility.
                <div className="mock-toolbar">
                  <span className="mock-toolbar-btn hl">Highlight</span>
                  <span className="mock-toolbar-btn st">Strikethrough</span>
                </div>
              </div>
              <div className="mock-text">
                Goroutines are not OS threads, and they&rsquo;re not exactly green threads &mdash; threads managed by a language&rsquo;s runtime. They are a higher level of abstraction known as <span className="mock-strike">processes</span> <span className="mock-highlight">coroutines</span>.
              </div>
              <div className="mock-text">
                Coroutines are simply concurrent subroutines that are nonpreemptive &mdash; that is, they cannot be interrupted. Instead, <span className="mock-highlight">coroutines have multiple points throughout which allow for suspension or reentry</span>.
              </div>
              <div className="mock-text">
                Goroutines are not garbage collected automatically when the function they were invoked from returns. The Go runtime handles multiplexing goroutines onto OS threads for you.
              </div>
            </div>

            {/* Annotations Panel */}
            <div className="mockup-annotations">
              <div className="mockup-tabs">
                <div
                  className={`mockup-tab ${activeTab === 'annotations' ? 'active' : ''}`}
                  onClick={() => setActiveTab('annotations')}
                >
                  Annotations
                </div>
                <div
                  className={`mockup-tab ${activeTab === 'askai' ? 'active' : ''}`}
                  onClick={() => setActiveTab('askai')}
                >
                  Ask AI
                </div>
              </div>

              {activeTab === 'annotations' ? (
                <div className="mockup-ann-content">
                  <div className="mockup-ann-header">
                    <div className="mockup-ann-title">Annotations</div>
                    <div className="mockup-stats">
                      <div className="mockup-stat"><div className="mockup-stat-num all">12</div><div className="mockup-stat-label">Total</div></div>
                      <div className="mockup-stat"><div className="mockup-stat-num hl">9</div><div className="mockup-stat-label">Highlights</div></div>
                      <div className="mockup-stat"><div className="mockup-stat-num st">3</div><div className="mockup-stat-label">Strikes</div></div>
                    </div>
                  </div>
                  <div className="mockup-ann-label">Notes</div>
                  <div className="mockup-ann-list">
                    <div className="mockup-ann-card">
                      <div className="mockup-ann-meta"><span className="mockup-ann-page">p. 84</span><span className="mockup-ann-type hl">Highlight</span></div>
                      <div className="mockup-ann-text">the Go scheduler uses an M:N scheduling model, multiplying M goroutines onto N OS threads</div>
                      <div className="mockup-ann-note">"Key insight &mdash; explains why goroutines are so cheap"</div>
                    </div>
                    <div className="mockup-ann-card">
                      <div className="mockup-ann-meta"><span className="mockup-ann-page">p. 84</span><span className="mockup-ann-type hl">Highlight</span></div>
                      <div className="mockup-ann-text">coroutines have multiple points throughout which allow for suspension or reentry</div>
                    </div>
                    <div className="mockup-ann-card strike">
                      <div className="mockup-ann-meta"><span className="mockup-ann-page">p. 82</span><span className="mockup-ann-type st">Strikethrough</span></div>
                      <div className="mockup-ann-text">imposing goroutines to others</div>
                    </div>
                  </div>
                </div>
              ) : (
                <div className="mockup-ai-content">
                  <div className="mockup-ai-input">
                    <input readOnly value="What did I learn about goroutine scheduling?" placeholder="Ask about your highlights..." />
                    <button>Ask</button>
                  </div>
                  <div className="mockup-ai-q">What did I learn about goroutine scheduling?</div>
                  <div className="mockup-ai-a">
                    Based on your highlights, Go uses an <strong>M:N scheduling model</strong> that multiplexes M goroutines onto N OS threads. This is what makes goroutines lightweight &mdash; they&rsquo;re not 1:1 with OS threads. You noted this as a key insight explaining why goroutines are so cheap to create.
                    <div className="mockup-ai-cite">Concurrency in Go &mdash; p. 84</div>
                  </div>
                  <div className="mockup-ai-q">How are goroutines different from threads?</div>
                  <div className="mockup-ai-a">
                    Goroutines are neither OS threads nor green threads. Your highlights describe them as <strong>coroutines</strong> &mdash; concurrent subroutines with multiple suspension/reentry points that are nonpreemptive. They consume only 2&ndash;8 KB of stack space.
                    <div className="mockup-ai-cite">Concurrency in Go &mdash; p. 84&ndash;85</div>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </section>

      {/* Features */}
      <section className="l-features" id="features">
        <div className="l-section-label">Features</div>
        <h2>Your reading, finally organized</h2>
        <div className="l-features-grid">
          <div className="l-feature-card">
            <div className="l-feature-icon">&#128214;</div>
            <h3>In-browser PDF reader</h3>
            <p>Upload any PDF and read it right in your browser. No app install, no file format fuss. Your library lives in the cloud.</p>
          </div>
          <div className="l-feature-card">
            <div className="l-feature-icon">&#9998;</div>
            <h3>Highlight &amp; strikethrough</h3>
            <p>Select text and annotate instantly. Add a personal note to any highlight explaining <em>why</em> it matters to you.</p>
          </div>
          <div className="l-feature-card">
            <div className="l-feature-icon">&#128278;</div>
            <h3>Auto-tagged annotations</h3>
            <p>Every annotation is automatically tagged with the book title, chapter, page number, and date. Zero manual work.</p>
          </div>
          <div className="l-feature-card">
            <div className="l-feature-icon">&#128269;</div>
            <h3>Cross-book search</h3>
            <p>Search across all your highlights from every book at once. Find that quote you half-remember in seconds.</p>
          </div>
          <div className="l-feature-card">
            <div className="l-feature-icon">&#129302;</div>
            <h3>AI recall</h3>
            <p>"What did I learn about replication?" Ask questions in plain English across your entire library. Answers cite the book and page.</p>
          </div>
          <div className="l-feature-card">
            <div className="l-feature-icon">&#128230;</div>
            <h3>Export anywhere</h3>
            <p>Export your annotations to Markdown, Notion, or Obsidian. Your knowledge, your format. <span className="coming-soon">(Coming soon)</span></p>
          </div>
        </div>
      </section>

      {/* How it works */}
      <section className="l-how" id="how">
        <div className="l-section-label">How it works</div>
        <h2>Three steps to total recall</h2>
        <div className="l-steps">
          <div className="l-step">
            <div className="l-step-num">1</div>
            <div>
              <h3>Upload your PDFs</h3>
              <p>Drag and drop your books. ReadShelf stores them securely and builds your personal library.</p>
            </div>
          </div>
          <div className="l-step">
            <div className="l-step-num">2</div>
            <div>
              <h3>Read and annotate</h3>
              <p>Highlight passages, strikethrough what's wrong, add notes about why something matters. Every annotation is captured with full context.</p>
            </div>
          </div>
          <div className="l-step">
            <div className="l-step-num">3</div>
            <div>
              <h3>Search and ask</h3>
              <p>Find anything across your entire library with full-text search or ask the AI in plain language. It answers with citations.</p>
            </div>
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="l-cta">
        <div className="l-cta-inner">
          <div className="l-cta-glow" />
          <h2>Your reading, organized at last</h2>
          <p className="l-cta-sub">Free during beta. 3 books, unlimited annotations, full-text search.</p>
          <Link to="/register" className="btn btn-primary">Create free account</Link>
        </div>
      </section>

      {/* Footer */}
      <footer className="l-footer">
        &copy; 2026 ReadShelf. Built for readers who want to remember.
      </footer>
    </div>
  );
}
