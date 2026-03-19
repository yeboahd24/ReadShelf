import { useState, useEffect, useRef, useCallback } from 'react';
import { useParams, useSearchParams } from 'react-router-dom';
import { pdfjsLib } from '../lib/pdfjs';
import { TextLayer } from 'pdfjs-dist';
import 'pdfjs-dist/web/pdf_viewer.css';
import type { PDFDocumentProxy, PDFPageProxy } from 'pdfjs-dist';
import * as booksApi from '../api/books';
import * as annotationsApi from '../api/annotations';
import type { Book, Annotation } from '../types';

const SCALE = 1.5;

export default function ReaderPage() {
  const { id } = useParams<{ id: string }>();
  const [searchParams] = useSearchParams();
  const targetPage = searchParams.get('page');
  const [book, setBook] = useState<Book | null>(null);
  const [pdfDoc, setPdfDoc] = useState<PDFDocumentProxy | null>(null);
  const [annotations, setAnnotations] = useState<Annotation[]>([]);
  const [selection, setSelection] = useState<{
    text: string;
    page: number;
    rects: Array<{ x: number; y: number; width: number; height: number }>;
    position: { top: number; left: number };
  } | null>(null);
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const viewerRef = useRef<HTMLDivElement>(null);

  // Load book metadata + PDF + annotations
  useEffect(() => {
    if (!id) return;
    booksApi.getBook(id).then(setBook);
    annotationsApi.listAnnotations(id).then(setAnnotations);

    booksApi.getSignedURL(id).then((url) => {
      pdfjsLib.getDocument(url).promise.then(setPdfDoc);
    });
  }, [id]);

  // Scroll to target page from URL (?page=N), or restore saved scroll position.
  useEffect(() => {
    const viewer = viewerRef.current;
    if (!viewer || !pdfDoc || !id) return;

    if (targetPage) {
      // Scroll to specific page from query param (e.g., from recall/search).
      requestAnimationFrame(() => {
        const el = viewer.querySelector(`[data-page-number="${targetPage}"]`) as HTMLElement | null;
        if (el) {
          const offset = el.offsetTop - viewer.offsetTop;
          viewer.scrollTo({ top: offset, behavior: 'auto' });
        }
      });
      return; // Don't restore saved position when navigating to a specific page.
    }

    // Restore saved scroll position.
    const saved = sessionStorage.getItem(`readshelf-scroll-${id}`);
    if (saved) {
      requestAnimationFrame(() => {
        viewer.scrollTop = parseInt(saved, 10);
      });
    }

    // Save on scroll (debounced).
    let timer: ReturnType<typeof setTimeout>;
    const handleScroll = () => {
      clearTimeout(timer);
      timer = setTimeout(() => {
        sessionStorage.setItem(`readshelf-scroll-${id}`, String(viewer.scrollTop));
      }, 200);
    };
    viewer.addEventListener('scroll', handleScroll);
    return () => {
      clearTimeout(timer);
      viewer.removeEventListener('scroll', handleScroll);
    };
  }, [pdfDoc, id]);

  // Handle text selection
  const handleMouseUp = useCallback(() => {
    const sel = window.getSelection();
    if (!sel || sel.isCollapsed || !sel.toString().trim()) {
      setSelection(null);
      return;
    }

    const text = sel.toString().trim();
    const range = sel.getRangeAt(0);

    // Find the page container
    let node: HTMLElement | null = range.startContainer.parentElement;
    let pageNum = 0;
    while (node && !node.dataset.pageNumber) {
      node = node.parentElement;
    }
    if (node) {
      pageNum = parseInt(node.dataset.pageNumber || '0', 10);
    }
    if (!pageNum) return;

    // Get selection rects in page-relative coordinates
    const pageEl = node!;
    const pageRect = pageEl.getBoundingClientRect();
    const clientRects = range.getClientRects();
    const rects: Array<{ x: number; y: number; width: number; height: number }> = [];

    for (let i = 0; i < clientRects.length; i++) {
      const r = clientRects[i];
      rects.push({
        x: (r.left - pageRect.left) / SCALE,
        y: (r.top - pageRect.top) / SCALE,
        width: r.width / SCALE,
        height: r.height / SCALE,
      });
    }

    // Position toolbar near selection
    const lastRect = clientRects[clientRects.length - 1];
    setSelection({
      text,
      page: pageNum,
      rects,
      position: {
        top: lastRect.bottom + window.scrollY + 5,
        left: lastRect.left + window.scrollX,
      },
    });
  }, []);

  const handleAnnotate = async (type: 'highlight' | 'strikethrough') => {
    if (!selection || !id) return;
    try {
      const annotation = await annotationsApi.createAnnotation(id, {
        type,
        content: selection.text,
        page: selection.page,
        rects: selection.rects,
      });
      setAnnotations((prev) => [...prev, annotation]);
    } catch (err: any) {
      if (err.response?.status === 409) {
        // Duplicate, ignore
      } else {
        console.error(err);
      }
    }
    window.getSelection()?.removeAllRanges();
    setSelection(null);
  };

  const handleDeleteAnnotation = async (annId: string) => {
    try {
      await annotationsApi.deleteAnnotation(annId);
      setAnnotations((prev) => prev.filter((a) => a.id !== annId));
    } catch (err) {
      console.error(err);
    }
  };

  const handleUpdateNote = async (annId: string, note: string) => {
    try {
      await annotationsApi.updateNote(annId, note);
      setAnnotations((prev) =>
        prev.map((a) => (a.id === annId ? { ...a, user_note: note } : a))
      );
    } catch (err) {
      console.error(err);
    }
  };

  return (
    <div className="reader-page">
      <div className="reader-toolbar">
        <h2>{book?.title || 'Loading...'}</h2>
        <button onClick={() => setSidebarOpen(!sidebarOpen)} className="btn-link">
          {sidebarOpen ? 'Hide' : 'Show'} Annotations
        </button>
      </div>

      <div className="reader-layout">
        <div
          className="pdf-viewer"
          ref={viewerRef}
          onMouseUp={handleMouseUp}
        >
          {pdfDoc &&
            Array.from({ length: pdfDoc.numPages }, (_, i) => (
              <PdfPage
                key={i + 1}
                pdfDoc={pdfDoc}
                pageNumber={i + 1}
                annotations={annotations.filter((a) => a.page === i + 1)}
                scrollRoot={viewerRef}
              />
            ))}
        </div>

        {sidebarOpen && (
          <div className="annotation-sidebar">
            <div className="sidebar-header">
              <h3>Annotations</h3>
              {annotations.length > 0 && (
                <div className="sidebar-stats">
                  <div className="sidebar-stat">
                    <span className="sidebar-stat-num stat-all">{annotations.length}</span>
                    <span className="sidebar-stat-label">Total</span>
                  </div>
                  <div className="sidebar-stat">
                    <span className="sidebar-stat-num stat-hl">{annotations.filter(a => a.type === 'highlight').length}</span>
                    <span className="sidebar-stat-label">Highlights</span>
                  </div>
                  <div className="sidebar-stat">
                    <span className="sidebar-stat-num stat-st">{annotations.filter(a => a.type === 'strikethrough').length}</span>
                    <span className="sidebar-stat-label">Strikes</span>
                  </div>
                </div>
              )}
            </div>
            {annotations.length === 0 ? (
              <div className="sidebar-empty">
                <div className="sidebar-empty-icon">&#9998;</div>
                <p>Select text in the PDF to create your first annotation.</p>
              </div>
            ) : (
              <div className="sidebar-list">
                {annotations.map((a) => (
                  <AnnotationCard
                    key={a.id}
                    annotation={a}
                    onDelete={() => handleDeleteAnnotation(a.id)}
                    onUpdateNote={(note) => handleUpdateNote(a.id, note)}
                    onScrollTo={() => {
                      const viewer = viewerRef.current;
                      const el = viewer?.querySelector(
                        `[data-page-number="${a.page}"]`
                      ) as HTMLElement | null;
                      if (el && viewer) {
                        const offset = el.offsetTop - viewer.offsetTop;
                        viewer.scrollTo({ top: offset, behavior: 'smooth' });
                      }
                    }}
                  />
                ))}
              </div>
            )}
          </div>
        )}
      </div>

      {selection && (
        <div
          className="annotation-toolbar"
          style={{ top: selection.position.top, left: selection.position.left }}
        >
          <button className="toolbar-highlight" onClick={() => handleAnnotate('highlight')}>
            <span className="toolbar-icon">&#9998;</span> Highlight
          </button>
          <button className="toolbar-strikethrough" onClick={() => handleAnnotate('strikethrough')}>
            <span className="toolbar-icon">&#9473;</span> Strikethrough
          </button>
        </div>
      )}
    </div>
  );
}

function PdfPage({
  pdfDoc,
  pageNumber,
  annotations,
  scrollRoot,
}: {
  pdfDoc: PDFDocumentProxy;
  pageNumber: number;
  annotations: Annotation[];
  scrollRoot: React.RefObject<HTMLDivElement | null>;
}) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const textLayerRef = useRef<HTMLDivElement>(null);
  const [visible, setVisible] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  // Lazy rendering: only render when page is near the viewport.
  useEffect(() => {
    const el = containerRef.current;
    if (!el) return;

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setVisible(true);
          observer.disconnect();
        }
      },
      {
        root: scrollRoot.current,
        rootMargin: '1200px',
      }
    );
    observer.observe(el);
    return () => observer.disconnect();
  }, [scrollRoot]);

  // Set placeholder size from page dimensions (fast, no rendering).
  useEffect(() => {
    let cancelled = false;
    pdfDoc.getPage(pageNumber).then((page: PDFPageProxy) => {
      if (cancelled) return;
      const viewport = page.getViewport({ scale: SCALE });
      if (containerRef.current) {
        containerRef.current.style.width = `${viewport.width}px`;
        containerRef.current.style.height = `${viewport.height}px`;
      }
    });
    return () => { cancelled = true; };
  }, [pdfDoc, pageNumber]);

  // Render canvas + text layer only when visible.
  useEffect(() => {
    if (!visible) return;
    let cancelled = false;

    pdfDoc.getPage(pageNumber).then((page: PDFPageProxy) => {
      if (cancelled) return;
      const viewport = page.getViewport({ scale: SCALE });

      const canvas = canvasRef.current!;
      canvas.height = viewport.height;
      canvas.width = viewport.width;
      const ctx = canvas.getContext('2d')!;
      page.render({ canvasContext: ctx, viewport });

      page.getTextContent().then((textContent) => {
        if (cancelled) return;
        const textLayerEl = textLayerRef.current!;
        textLayerEl.innerHTML = '';

        const tl = new TextLayer({
          textContentSource: textContent,
          container: textLayerEl,
          viewport,
        });
        tl.render();
      });
    });

    return () => { cancelled = true; };
  }, [pdfDoc, pageNumber, visible]);

  return (
    <div
      className="pdf-page"
      data-page-number={pageNumber}
      ref={containerRef}
    >
      {visible ? (
        <>
          <canvas ref={canvasRef} />
          <div ref={textLayerRef} className="textLayer" />
        </>
      ) : (
        <div className="page-placeholder">Loading page {pageNumber}...</div>
      )}
      {/* Annotation overlay — renders highlight/strikethrough rects */}
      <div
        style={{
          position: 'absolute',
          top: 0,
          left: 0,
          width: '100%',
          height: '100%',
          pointerEvents: 'none',
          zIndex: 3,
        }}
      >
        {annotations.map((a) =>
          a.rects?.map((r, i) => (
            <div
              key={`${a.id}-${i}`}
              style={{
                position: 'absolute',
                left: `${r.x * SCALE}px`,
                top: `${r.y * SCALE}px`,
                width: `${r.width * SCALE}px`,
                height: `${r.height * SCALE}px`,
                background: a.type === 'highlight'
                  ? 'rgba(250, 204, 21, 0.35)'
                  : 'rgba(239, 68, 68, 0.12)',
              }}
            >
              {a.type === 'strikethrough' && (
                <div
                  style={{
                    position: 'absolute',
                    left: 0,
                    right: 0,
                    top: '50%',
                    height: '2px',
                    background: '#ef4444',
                  }}
                />
              )}
            </div>
          ))
        )}
      </div>
      <div className="page-number">Page {pageNumber}</div>
    </div>
  );
}

function AnnotationCard({
  annotation,
  onDelete,
  onUpdateNote,
  onScrollTo,
}: {
  annotation: Annotation;
  onDelete: () => void;
  onScrollTo?: () => void;
  onUpdateNote: (note: string) => void;
}) {
  const [editing, setEditing] = useState(false);
  const [note, setNote] = useState(annotation.user_note || '');

  const saveNote = () => {
    onUpdateNote(note);
    setEditing(false);
  };

  return (
    <div
      className={`ann-card ann-card-${annotation.type}`}
      onClick={onScrollTo}
    >
      <div className="ann-card-top">
        <span className={`ann-badge ann-badge-${annotation.type}`}>
          {annotation.type === 'highlight' ? '✦ Highlight' : '— Strikethrough'}
        </span>
        <span className="ann-card-page">p. {annotation.page}</span>
        <button
          className="ann-card-delete"
          onClick={(e) => { e.stopPropagation(); onDelete(); }}
          title="Delete"
        >
          ×
        </button>
      </div>
      <p className="ann-card-text">{annotation.content}</p>
      {editing ? (
        <div className="ann-note-editor" onClick={(e) => e.stopPropagation()}>
          <textarea
            value={note}
            onChange={(e) => setNote(e.target.value)}
            placeholder="Why does this matter to you?"
            rows={2}
          />
          <div className="ann-note-actions">
            <button className="ann-note-save" onClick={saveNote}>Save</button>
            <button onClick={() => setEditing(false)}>Cancel</button>
          </div>
        </div>
      ) : (
        <div className="ann-note-display" onClick={(e) => { e.stopPropagation(); setEditing(true); }}>
          {annotation.user_note ? (
            <p className="ann-note-text">"{annotation.user_note}"</p>
          ) : (
            <p className="ann-note-add">+ Add note</p>
          )}
        </div>
      )}
    </div>
  );
}
