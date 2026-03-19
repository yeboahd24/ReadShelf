package inbound

import (
	"context"

	"github.com/dominic/readshelf/internal/core/domain"
)

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type AuthService interface {
	Register(ctx context.Context, email, password string) (*domain.User, *TokenPair, error)
	Login(ctx context.Context, email, password string) (*domain.User, *TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (*TokenPair, error)
}
