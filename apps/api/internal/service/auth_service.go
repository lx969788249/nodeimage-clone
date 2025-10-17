package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"nodeimage/api/internal/config"
	"nodeimage/api/internal/ids"
	"nodeimage/api/internal/models"
	"nodeimage/api/internal/repository"
	"nodeimage/api/internal/security"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserSuspended      = errors.New("user suspended")
)

type AuthService struct {
	users    *repository.UserRepository
	sessions *repository.SessionRepository
	cache    *redis.Client
	cfg      *config.AppConfig
	log      zerolog.Logger
}

func NewAuthService(
	users *repository.UserRepository,
	sessions *repository.SessionRepository,
	cache *redis.Client,
	cfg *config.AppConfig,
	log zerolog.Logger,
) *AuthService {
	return &AuthService{
		users:    users,
		sessions: sessions,
		cache:    cache,
		cfg:      cfg,
		log:      log,
	}
}

type RegisterInput struct {
	Email       string
	Password    string
	DisplayName string
}

type AuthResult struct {
	AccessToken  string
	RefreshToken string
	User         models.User
	DeviceID     string
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (AuthResult, error) {
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	if input.Email == "" || input.Password == "" {
		return AuthResult{}, fmt.Errorf("email and password required")
	}

	if _, err := s.users.FindByEmail(ctx, input.Email); err == nil {
		return AuthResult{}, fmt.Errorf("email already registered")
	} else if !errors.Is(err, repository.ErrUserNotFound) {
		return AuthResult{}, err
	}

	passwordHash, err := security.HashPassword(input.Password)
	if err != nil {
		return AuthResult{}, err
	}

	user := models.User{
		ID:           ids.New(),
		Email:        input.Email,
		PasswordHash: passwordHash,
		DisplayName:  input.DisplayName,
		Role:         models.UserRoleUser,
		Status:       models.UserStatusActive,
	}

	if err := s.users.Create(ctx, user); err != nil {
		return AuthResult{}, err
	}

	deviceID := ids.New()
	session, tokens, err := s.createSession(ctx, user, deviceID, "New Device", "", "")
	if err != nil {
		return AuthResult{}, err
	}
	_ = session

	return tokens, nil
}

type LoginInput struct {
	Email      string
	Password   string
	DeviceID   string
	DeviceName string
	IPAddress  string
	UserAgent  string
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (AuthResult, error) {
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	user, err := s.users.FindByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return AuthResult{}, ErrInvalidCredentials
		}
		return AuthResult{}, err
	}

	if user.Status != models.UserStatusActive {
		return AuthResult{}, ErrUserSuspended
	}

	ok, err := security.VerifyPassword(input.Password, user.PasswordHash)
	if err != nil || !ok {
		return AuthResult{}, ErrInvalidCredentials
	}

	deviceID := input.DeviceID
	if deviceID == "" {
		deviceID = ids.New()
	}
	deviceName := input.DeviceName
	if deviceName == "" {
		deviceName = "Unknown Device"
	}

	_, tokens, err := s.createSession(ctx, user, deviceID, deviceName, input.IPAddress, input.UserAgent)
	if err != nil {
		return AuthResult{}, err
	}
	return tokens, nil
}

func (s *AuthService) createSession(
	ctx context.Context,
	user models.User,
	deviceID string,
	deviceName string,
	ipAddress string,
	userAgent string,
) (models.Session, AuthResult, error) {
	refreshToken, refreshHash, err := security.GenerateRefreshToken(64)
	if err != nil {
		return models.Session{}, AuthResult{}, err
	}

	session := models.Session{
		ID:               ids.New(),
		UserID:           user.ID,
		DeviceID:         deviceID,
		DeviceName:       deviceName,
		RefreshTokenHash: refreshHash,
		IPAddress:        ipAddress,
		UserAgent:        userAgent,
		ExpiresAt:        time.Now().Add(s.cfg.Security.JWTRefreshTTL),
	}

	accessToken, err := security.GenerateAccessToken(
		s.cfg.Security.JWTAccessSecret,
		user.ID,
		session.ID,
		deviceID,
		string(user.Role),
		nil,
		s.cfg.Security.JWTAccessTTL,
	)
	if err != nil {
		return models.Session{}, AuthResult{}, err
	}

	if err := s.sessions.Create(ctx, session); err != nil {
		return models.Session{}, AuthResult{}, err
	}

	if err := s.enforceSessionLimit(ctx, user.ID); err != nil {
		s.log.Warn().Err(err).Str("user_id", user.ID).Msg("enforce session limit failed")
	}

	return session, AuthResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
		DeviceID:     deviceID,
	}, nil
}

func (s *AuthService) enforceSessionLimit(ctx context.Context, userID string) error {
	count, err := s.sessions.CountByUser(ctx, userID)
	if err != nil {
		return err
	}
	if count <= s.cfg.Security.MaxSessions {
		return nil
	}

	return s.sessions.DeleteOldestSessions(ctx, userID, s.cfg.Security.MaxSessions)
}

type RefreshInput struct {
	UserID       string
	RefreshToken string
	DeviceID     string
}

func (s *AuthService) Refresh(ctx context.Context, input RefreshInput) (AuthResult, error) {
	user, err := s.users.GetByID(ctx, input.UserID)
	if err != nil {
		return AuthResult{}, err
	}
	if user.Status != models.UserStatusActive {
		return AuthResult{}, ErrUserSuspended
	}

	refreshHash := security.HashRefreshToken(input.RefreshToken)
	session, err := s.sessions.FindByRefreshHash(ctx, input.UserID, refreshHash)
	if err != nil {
		return AuthResult{}, ErrInvalidCredentials
	}

	if session.DeviceID != input.DeviceID {
		return AuthResult{}, ErrInvalidCredentials
	}

	if session.ExpiresAt.Before(time.Now()) {
		_ = s.sessions.DeleteByID(ctx, session.ID)
		return AuthResult{}, ErrInvalidCredentials
}

	refreshToken, newHash, err := security.GenerateRefreshToken(64)
	if err != nil {
		return AuthResult{}, err
	}

	session.RefreshTokenHash = newHash
	session.ExpiresAt = time.Now().Add(s.cfg.Security.JWTRefreshTTL)

	if err := s.sessions.Create(ctx, session); err != nil {
		return AuthResult{}, err
	}

	accessToken, err := security.GenerateAccessToken(
		s.cfg.Security.JWTAccessSecret,
		user.ID,
		session.ID,
		session.DeviceID,
		string(user.Role),
		nil,
		s.cfg.Security.JWTAccessTTL,
	)
	if err != nil {
		return AuthResult{}, err
	}

	return AuthResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
		DeviceID:     session.DeviceID,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, userID string, deviceID string) error {
	return s.sessions.DeleteByDevice(ctx, userID, deviceID)
}
