package handler

import (
	"encoding/json"
	"net/http"

	"github.com/dominic/readshelf/internal/adapter/inbound/http/httputil"
	"github.com/dominic/readshelf/internal/adapter/inbound/http/middleware"
	"github.com/dominic/readshelf/internal/core/domain"
	"github.com/dominic/readshelf/internal/core/port/inbound"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type AnnotationHandler struct {
	annotations inbound.AnnotationService
}

func NewAnnotationHandler(annotations inbound.AnnotationService) *AnnotationHandler {
	return &AnnotationHandler{annotations: annotations}
}

type createAnnotationRequest struct {
	Type      string           `json:"type"`
	Content   string           `json:"content"`
	Page      int              `json:"page"`
	Chapter   string           `json:"chapter,omitempty"`
	UserNote  string           `json:"user_note,omitempty"`
	CharStart *int             `json:"char_start,omitempty"`
	CharEnd   *int             `json:"char_end,omitempty"`
	Rects     json.RawMessage  `json:"rects,omitempty"`
}

func (h *AnnotationHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromCtx(r.Context())
	bookID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid book id")
		return
	}

	var req createAnnotationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Type == "" || req.Content == "" {
		httputil.Error(w, http.StatusBadRequest, "type and content are required")
		return
	}

	if req.Type != "highlight" && req.Type != "strikethrough" {
		httputil.Error(w, http.StatusBadRequest, "type must be 'highlight' or 'strikethrough'")
		return
	}

	annotation := &domain.Annotation{
		BookID:    bookID,
		UserID:    userID,
		Type:      req.Type,
		Content:   req.Content,
		Page:      req.Page,
		Chapter:   req.Chapter,
		UserNote:  req.UserNote,
		CharStart: req.CharStart,
		CharEnd:   req.CharEnd,
		Rects:     req.Rects,
	}

	result, err := h.annotations.Create(r.Context(), annotation)
	if err != nil {
		httputil.HandleError(w, err)
		return
	}

	httputil.JSON(w, http.StatusCreated, result)
}

func (h *AnnotationHandler) ListByBook(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromCtx(r.Context())
	bookID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid book id")
		return
	}

	annotations, err := h.annotations.ListByBook(r.Context(), userID, bookID)
	if err != nil {
		httputil.HandleError(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, annotations)
}

func (h *AnnotationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromCtx(r.Context())
	annotationID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid annotation id")
		return
	}

	if err := h.annotations.Delete(r.Context(), userID, annotationID); err != nil {
		httputil.HandleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type updateNoteRequest struct {
	Note string `json:"note"`
}

func (h *AnnotationHandler) UpdateNote(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromCtx(r.Context())
	annotationID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid annotation id")
		return
	}

	var req updateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.annotations.UpdateNote(r.Context(), userID, annotationID, req.Note); err != nil {
		httputil.HandleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AnnotationHandler) Search(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromCtx(r.Context())
	query := r.URL.Query().Get("q")
	if query == "" {
		httputil.Error(w, http.StatusBadRequest, "query parameter 'q' is required")
		return
	}

	results, err := h.annotations.Search(r.Context(), userID, query)
	if err != nil {
		httputil.HandleError(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, results)
}
