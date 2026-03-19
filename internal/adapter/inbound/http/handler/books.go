package handler

import (
	"net/http"
	"strings"

	"github.com/dominic/readshelf/internal/adapter/inbound/http/httputil"
	"github.com/dominic/readshelf/internal/adapter/inbound/http/middleware"
	"github.com/dominic/readshelf/internal/core/port/inbound"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type BookHandler struct {
	books inbound.BookService
}

func NewBookHandler(books inbound.BookService) *BookHandler {
	return &BookHandler{books: books}
}

func (h *BookHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromCtx(r.Context())

	books, err := h.books.List(r.Context(), userID)
	if err != nil {
		httputil.HandleError(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, books)
}

func (h *BookHandler) Upload(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromCtx(r.Context())

	// 50MB max for free tier.
	r.Body = http.MaxBytesReader(w, r.Body, 50<<20)

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		httputil.Error(w, http.StatusBadRequest, "file too large or invalid form data")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".pdf") {
		httputil.Error(w, http.StatusBadRequest, "only PDF files are allowed")
		return
	}

	title := r.FormValue("title")
	if title == "" {
		// Clean filename: remove extension, replace dots/underscores/hyphens with spaces.
		name := header.Filename
		name = strings.TrimSuffix(strings.TrimSuffix(name, ".pdf"), ".PDF")
		name = strings.NewReplacer(".", " ", "_", " ", "-", " ").Replace(name)
		title = strings.TrimSpace(name)
	}
	author := r.FormValue("author")

	book, err := h.books.Upload(r.Context(), userID, title, author, file, header.Filename)
	if err != nil {
		httputil.HandleError(w, err)
		return
	}

	httputil.JSON(w, http.StatusCreated, book)
}

func (h *BookHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromCtx(r.Context())
	bookID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid book id")
		return
	}

	book, err := h.books.Get(r.Context(), userID, bookID)
	if err != nil {
		httputil.HandleError(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, book)
}

func (h *BookHandler) GetURL(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromCtx(r.Context())
	bookID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid book id")
		return
	}

	url, err := h.books.GetSignedURL(r.Context(), userID, bookID)
	if err != nil {
		httputil.HandleError(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]string{"url": url})
}

func (h *BookHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromCtx(r.Context())
	bookID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid book id")
		return
	}

	if err := h.books.Delete(r.Context(), userID, bookID); err != nil {
		httputil.HandleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
