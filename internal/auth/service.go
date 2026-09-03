package auth

import (
	"context"
	"time"

	apperrors "github.com/dbplatform/api/internal/errors"
	"github.com/dbplatform/api/internal/user"
	"github.com/google/uuid"
)

type Service struct {
	userRepo    user.Repository
	sessionRepo user.SessionRepository
	jwtSvc      *JWTService
	refreshTTL  time.Duration
}

func NewService(u user.Repository, s user.SessionRepository, j *JWTService, refreshTTL time.Duration) *Service {
	return &Service{
		userRepo:    u,
		sessionRepo: s,
		jwtSvc:      j,
		refreshTTL:  refreshTTL,
	}
}

type RegisterInput struct {
	Email    string
	Password string
	FullName string
}

type LoginInput struct {
	Email     string
	Password  string
	UserAgent string
	IPAddress string
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64 // seconds until access token expires
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (*user.User, error) {
	existing, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil && !apperrors.IsNotFound(err) {
		return nil, apperrors.NewInternal(err)
	}
	if existing != nil {
		return nil, apperrors.NewConflict("email already in use")
	}

	hash, err := HashPassword(input.Password)
	if err != nil {
		return nil, apperrors.NewInternal(err)
	}

	u := &user.User{
		Email:        input.Email,
		PasswordHash: hash,
		FullName:     input.FullName,
		Role:         user.RoleViewer,
		IsActive:     true,
	}

	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, apperrors.NewInternal(err)
	}

	return u, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*user.User, *TokenPair, error) {
	u, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil, nil, apperrors.NewInvalidCredentials()
		}
		return nil, nil, apperrors.NewInternal(err)
	}

	if !u.IsActive {
		return nil, nil, apperrors.NewInvalidCredentials()
	}

	if err := CheckPassword(input.Password, u.PasswordHash); err != nil {
		return nil, nil, apperrors.NewInvalidCredentials()
	}

	tokens, err := s.generateTokenPair(ctx, u, input.UserAgent, input.IPAddress)
	if err != nil {
		return nil, nil, err
	}

	now := time.Now()
	u.LastLoginAt = &now
	_ = s.userRepo.Update(ctx, u)

	return u, tokens, nil
}

func (s *Service) Refresh(ctx context.Context, rawRefreshToken, userAgent, ipAddress string) (*user.User, *TokenPair, error) {
	hash := hashToken(rawRefreshToken)

	session, err := s.sessionRepo.GetByTokenHash(ctx, hash)
	if err != nil {
		return nil, nil, apperrors.NewUnauthorized()
	}

	if session.RevokedAt != nil || time.Now().After(session.ExpiresAt) {
		return nil, nil, apperrors.NewUnauthorized()
	}

	if err := s.sessionRepo.RevokeByTokenHash(ctx, hash); err != nil {
		return nil, nil, apperrors.NewInternal(err)
	}

	u, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, nil, apperrors.NewUnauthorized()
	}

	if !u.IsActive {
		return nil, nil, apperrors.NewUnauthorized()
	}

	tokens, err := s.generateTokenPair(ctx, u, userAgent, ipAddress)
	if err != nil {
		return nil, nil, err
	}

	return u, tokens, nil
}

func (s *Service) Logout(ctx context.Context, rawRefreshToken string) error {
	hash := hashToken(rawRefreshToken)
	return s.sessionRepo.RevokeByTokenHash(ctx, hash)
}

func (s *Service) GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	u, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Service) generateTokenPair(ctx context.Context, u *user.User, userAgent, ipAddress string) (*TokenPair, error) {
	accessToken, err := s.jwtSvc.GenerateAccessToken(u)
	if err != nil {
		return nil, apperrors.NewInternal(err)
	}

	rawRefresh, refreshHash, err := s.jwtSvc.GenerateRefreshToken()
	if err != nil {
		return nil, apperrors.NewInternal(err)
	}

	session := &user.Session{
		UserID:           u.ID,
		RefreshTokenHash: refreshHash,
		UserAgent:        userAgent,
		IPAddress:        ipAddress,
		ExpiresAt:        time.Now().Add(s.refreshTTL),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, apperrors.NewInternal(err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresIn:    int64(s.jwtSvc.cfg.AccessTokenTTL.Seconds()),
	}, nil
}
