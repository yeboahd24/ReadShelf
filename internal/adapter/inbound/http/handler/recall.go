package handler

import (
	"encoding/json"
	"net/http"

	"github.com/dominic/readshelf/internal/adapter/inbound/http/httputil"
	"github.com/dominic/readshelf/internal/adapter/inbound/http/middleware"
	"github.com/dominic/readshelf/internal/core/port/inbound"
)

type RecallHandler struct {
	recall inbound.RecallService
}

func NewRecallHandler(recall inbound.RecallService) *RecallHandler {
	return &RecallHandler{recall: recall}
}

type recallRequest struct {
	Query string `json:"query"`
}

func (h *RecallHandler) Query(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromCtx(r.Context())

	var req recallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Query == "" {
		httputil.Error(w, http.StatusBadRequest, "query is required")
		return
	}

	result, err := h.recall.Query(r.Context(), userID, req.Query)
	if err != nil {
		httputil.HandleError(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, result)
}
