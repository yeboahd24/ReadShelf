package handler

import (
	"encoding/json"
	"net/http"

	"github.com/dominic/readshelf/internal/adapter/inbound/http/httputil"
	"github.com/dominic/readshelf/internal/core/port/inbound"
)

type AuthHandler struct {
	auth inbound.AuthService
}

func NewAuthHandler(auth inbound.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	User        interface{} `json:"user"`
	AccessToken string      `json:"access_token"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		httputil.Error(w, http.StatusBadRequest, "email and password are required")
		return
	}

	user, tokens, err := h.auth.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		httputil.HandleError(w, err)
		return
	}

	setRefreshCookie(w, tokens.RefreshToken)
	httputil.JSON(w, http.StatusCreated, authResponse{
		User:        user,
		AccessToken: tokens.AccessToken,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, tokens, err := h.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		httputil.HandleError(w, err)
		return
	}

	setRefreshCookie(w, tokens.RefreshToken)
	httputil.JSON(w, http.StatusOK, authResponse{
		User:        user,
		AccessToken: tokens.AccessToken,
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		httputil.Error(w, http.StatusUnauthorized, "no refresh token")
		return
	}

	tokens, err := h.auth.Refresh(r.Context(), cookie.Value)
	if err != nil {
		httputil.HandleError(w, err)
		return
	}

	setRefreshCookie(w, tokens.RefreshToken)
	httputil.JSON(w, http.StatusOK, map[string]string{
		"access_token": tokens.AccessToken,
	})
}

func setRefreshCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/api/auth",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   30 * 24 * 60 * 60, // 30 days
	})
}
