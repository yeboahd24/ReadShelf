package service

import (
	"context"
	"time"

	"github.com/dominic/readshelf/internal/core/domain"
	"github.com/dominic/readshelf/internal/core/port/inbound"
	"github.com/dominic/readshelf/internal/core/port/outbound"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost          = 12
	accessTokenExpiry   = 15 * time.Minute
	refreshTokenExpiry  = 30 * 24 * time.Hour
)

type authService struct {
	users     outbound.UserRepository
	jwtSecret []byte
}

func NewAuthService(users outbound.UserRepository, jwtSecret string) inbound.AuthService {
	return &authService{
		users:     users,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *authService) Register(ctx context.Context, email, password string) (*domain.User, *inbound.TokenPair, error) {
	existing, _ := s.users.GetByEmail(ctx, email)
	if existing != nil {
		return nil, nil, domain.ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return nil, nil, err
	}

	user := &domain.User{
		ID:        uuid.New(),
		Email:     email,
		Password:  string(hash),
		Plan:      "free",
		CreatedAt: time.Now(),
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, nil, err
	}

	tokens, err := s.generateTokens(user.ID)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (*domain.User, *inbound.TokenPair, error) {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return nil, nil, domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, nil, domain.ErrInvalidCredentials
	}

	tokens, err := s.generateTokens(user.ID)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (*inbound.TokenPair, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(t *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, domain.ErrInvalidToken
	}

	sub, err := claims.GetSubject()
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	userID, err := uuid.Parse(sub)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	// Verify user still exists.
	if _, err := s.users.GetByID(ctx, userID); err != nil {
		return nil, domain.ErrInvalidToken
	}

	return s.generateTokens(userID)
}

func (s *authService) generateTokens(userID uuid.UUID) (*inbound.TokenPair, error) {
	now := time.Now()

	accessClaims := jwt.RegisteredClaims{
		Subject:   userID.String(),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(accessTokenExpiry)),
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	refreshClaims := jwt.RegisteredClaims{
		Subject:   userID.String(),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(refreshTokenExpiry)),
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &inbound.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
