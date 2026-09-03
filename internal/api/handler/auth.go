package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/dbplatform/api/internal/api/middleware"
	"github.com/dbplatform/api/internal/api/response"
	"github.com/dbplatform/api/internal/audit"
	"github.com/dbplatform/api/internal/auth"
	apperrors "github.com/dbplatform/api/internal/errors"
	"github.com/dbplatform/api/internal/user"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
)

type AuthHandler struct {
	authSvc  *auth.Service
	auditSvc *audit.Service
	validate *validator.Validate
	log      zerolog.Logger
}

func NewAuthHandler(authSvc *auth.Service, auditSvc *audit.Service, log zerolog.Logger) *AuthHandler {
	return &AuthHandler{
		authSvc:  authSvc,
		auditSvc: auditSvc,
		validate: validator.New(),
		log:      log,
	}
}


type registerRequest struct {
	Email    string `json:"email"     validate:"required,email,max=255"`
	Password string `json:"password"  validate:"required,min=8,max=72"`
	FullName string `json:"full_name" validate:"required,min=1,max=255"`
}

type loginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type userResponse struct {
	ID          string     `json:"id"`
	Email       string     `json:"email"`
	FullName    string     `json:"full_name"`
	Role        string     `json:"role"`
	IsActive    bool       `json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type tokenResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
	User         userResponse `json:"user"`
}

func toUserResponse(u *user.User) userResponse {
	return userResponse{
		ID:          u.ID.String(),
		Email:       u.Email,
		FullName:    u.FullName,
		Role:        string(u.Role),
		IsActive:    u.IsActive,
		LastLoginAt: u.LastLoginAt,
		CreatedAt:   u.CreatedAt,
	}
}


func decodeAndValidate(r *http.Request, v *validator.Validate, dst interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return apperrors.NewValidation("invalid request body")
	}
	if err := v.Struct(dst); err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok && len(ve) > 0 {
			return apperrors.NewValidation(ve[0].Translate(nil))
		}
		return apperrors.NewValidation("validation failed")
	}
	return nil
}

func clientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return strings.SplitN(forwarded, ",", 2)[0]
	}
	host, _, err := splitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func splitHostPort(hostport string) (host, port string, err error) {
	idx := strings.LastIndex(hostport, ":")
	if idx < 0 {
		return hostport, "", nil
	}
	return hostport[:idx], hostport[idx+1:], nil
}


func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := decodeAndValidate(r, h.validate, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	u, err := h.authSvc.Register(r.Context(), auth.RegisterInput{
		Email:    strings.ToLower(strings.TrimSpace(req.Email)),
		Password: req.Password,
		FullName: strings.TrimSpace(req.FullName),
	})
	if err != nil {
		response.Error(w, r, err)
		return
	}

	h.auditSvc.LogAsync(&audit.AuditLog{
		UserID:    &u.ID,
		UserEmail: u.Email,
		Action:    audit.ActionRegister,
		IPAddress: clientIP(r),
		UserAgent: r.UserAgent(),
		Status:    "success",
	})

	response.Created(w, r, toUserResponse(u))
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeAndValidate(r, h.validate, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	u, tokens, err := h.authSvc.Login(r.Context(), auth.LoginInput{
		Email:     strings.ToLower(strings.TrimSpace(req.Email)),
		Password:  req.Password,
		UserAgent: r.UserAgent(),
		IPAddress: clientIP(r),
	})
	if err != nil {
		h.auditSvc.LogAsync(&audit.AuditLog{
			UserEmail: req.Email,
			Action:    audit.ActionLogin,
			IPAddress: clientIP(r),
			UserAgent: r.UserAgent(),
			Status:    "failure",
		})
		response.Error(w, r, err)
		return
	}

	h.auditSvc.LogAsync(&audit.AuditLog{
		UserID:    &u.ID,
		UserEmail: u.Email,
		Action:    audit.ActionLogin,
		IPAddress: clientIP(r),
		UserAgent: r.UserAgent(),
		Status:    "success",
	})

	response.Success(w, r, tokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
		User:         toUserResponse(u),
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := decodeAndValidate(r, h.validate, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	u, tokens, err := h.authSvc.Refresh(r.Context(), req.RefreshToken, r.UserAgent(), clientIP(r))
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.Success(w, r, tokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
		User:         toUserResponse(u),
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req logoutRequest
	if err := decodeAndValidate(r, h.validate, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		response.Error(w, r, apperrors.NewUnauthorized())
		return
	}

	_ = h.authSvc.Logout(r.Context(), req.RefreshToken)

	h.auditSvc.LogAsync(&audit.AuditLog{
		UserEmail: claims.Email,
		Action:    audit.ActionLogout,
		IPAddress: clientIP(r),
		UserAgent: r.UserAgent(),
		Status:    "success",
	})

	response.NoContent(w)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		response.Error(w, r, apperrors.NewUnauthorized())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	u, err := h.authSvc.GetUserByID(ctx, mustParseUUID(claims.UserID))
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.Success(w, r, toUserResponse(u))
}

