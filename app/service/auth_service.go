package service

import (
	"errors"
	"strings"
	"sync"
	"time"

	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/helper"

	"github.com/gofiber/fiber/v2"
)

const invalidCredentialsMessage = "username atau password salah"

type AuthService struct {
	repo        repository.AuthRepository
	permissions *helper.PermissionSet
	jwtManager  *helper.JWTManager
	accessTTL   time.Duration
	refreshTTL  time.Duration
	loginMu     sync.Mutex
	loginFails  map[string]loginFailure
}

type loginFailure struct {
	count     int
	retryTill time.Time
}

const (
	maxLoginFailures = 5
	loginBlockPeriod = 5 * time.Minute
)

func NewAuthService(repo repository.AuthRepository, jwtManager *helper.JWTManager, permissions *helper.PermissionSet, accessTTL, refreshTTL time.Duration) *AuthService {
	return &AuthService{
		repo: repo, permissions: permissions, jwtManager: jwtManager, accessTTL: accessTTL, refreshTTL: refreshTTL,
		loginFails: make(map[string]loginFailure),
	}
}

func (s *AuthService) Register(c *fiber.Ctx) error {
	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body JSON tidak valid")
	}
	req.Username, req.Email = strings.TrimSpace(req.Username), strings.TrimSpace(req.Email)
	if errs := ValidateRegister(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}
	hash, err := helper.HashPassword(req.Password)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memproses password")
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	user, err := s.repo.CreateUser(ctx, model.User{
		Username: req.Username, Email: req.Email, PasswordHash: hash, Role: "user",
	})
	if err != nil {
		if errors.Is(err, repository.ErrUsernameOrEmailTaken) {
			return helper.Fail(c, fiber.StatusConflict, err.Error())
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mendaftarkan pengguna")
	}
	return helper.Created(c, "pengguna berhasil didaftarkan", user, "/api/v1/auth/me")
}

func (s *AuthService) Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body JSON tidak valid")
	}
	if errs := ValidateLogin(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}
	loginKey := strings.ToLower(strings.TrimSpace(req.Username)) + "|" + c.IP()
	if retryAfter, blocked := s.loginBlocked(loginKey, time.Now()); blocked {
		c.Set("Retry-After", retryAfter.String())
		return helper.Fail(c, fiber.StatusTooManyRequests, "terlalu banyak percobaan login")
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	user, err := s.repo.FindUserByUsername(ctx, strings.TrimSpace(req.Username))
	if err != nil {
		helper.VerifyDummyPassword(req.Password)
		if errors.Is(err, repository.ErrUserNotFound) {
			return s.loginFailureResponse(c, loginKey)
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal masuk")
	}
	if !helper.VerifyPassword(user.PasswordHash, req.Password) {
		return s.loginFailureResponse(c, loginKey)
	}
	s.resetLoginFailures(loginKey)
	pair, refresh, err := s.issueTokenPair(user)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal membuat token")
	}
	if err := s.repo.CreateRefreshToken(ctx, refresh); err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal menyimpan refresh token")
	}
	return helper.OK(c, "berhasil masuk", pair)
}

func (s *AuthService) loginFailureResponse(c *fiber.Ctx, key string) error {
	if retryAfter, blocked := s.recordLoginFailure(key, time.Now()); blocked {
		c.Set("Retry-After", retryAfter.String())
		return helper.Fail(c, fiber.StatusTooManyRequests, "terlalu banyak percobaan login")
	}
	return helper.Fail(c, fiber.StatusUnauthorized, invalidCredentialsMessage)
}

func (s *AuthService) loginBlocked(key string, now time.Time) (time.Duration, bool) {
	s.loginMu.Lock()
	defer s.loginMu.Unlock()
	failure, ok := s.loginFails[key]
	if !ok || !now.Before(failure.retryTill) {
		if ok {
			delete(s.loginFails, key)
		}
		return 0, false
	}
	return time.Until(failure.retryTill).Round(time.Second), true
}

func (s *AuthService) recordLoginFailure(key string, now time.Time) (time.Duration, bool) {
	s.loginMu.Lock()
	defer s.loginMu.Unlock()
	failure := s.loginFails[key]
	if !failure.retryTill.IsZero() && !now.Before(failure.retryTill) {
		failure = loginFailure{}
	}
	failure.count++
	if failure.count > maxLoginFailures {
		failure.retryTill = now.Add(loginBlockPeriod)
		s.loginFails[key] = failure
		return loginBlockPeriod, true
	}
	s.loginFails[key] = failure
	return 0, false
}

func (s *AuthService) resetLoginFailures(key string) {
	s.loginMu.Lock()
	delete(s.loginFails, key)
	s.loginMu.Unlock()
}

func (s *AuthService) Refresh(c *fiber.Ctx) error {
	var req model.RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body JSON tidak valid")
	}
	if strings.TrimSpace(req.RefreshToken) == "" {
		return helper.Fail(c, fiber.StatusUnauthorized, "refresh token tidak valid")
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	oldHash := helper.SHA256Hex(req.RefreshToken)
	oldToken, err := s.repo.FindRefreshToken(ctx, oldHash)
	if err != nil || oldToken.RevokedAt != nil || !oldToken.ExpiresAt.After(time.Now()) {
		return helper.Fail(c, fiber.StatusUnauthorized, "refresh token tidak valid")
	}
	user, err := s.repo.FindUserByID(ctx, oldToken.UserID)
	if err != nil {
		return helper.Fail(c, fiber.StatusUnauthorized, "refresh token tidak valid")
	}
	pair, newToken, err := s.issueTokenPair(user)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal membuat token")
	}
	if err := s.repo.RotateRefreshToken(ctx, oldHash, newToken); err != nil {
		return helper.Fail(c, fiber.StatusUnauthorized, "refresh token tidak valid")
	}
	return helper.OK(c, "token berhasil diperbarui", pair)
}

func (s *AuthService) Logout(c *fiber.Ctx) error {
	var req model.RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body JSON tidak valid")
	}
	if strings.TrimSpace(req.RefreshToken) == "" {
		return helper.Fail(c, fiber.StatusBadRequest, "refresh token wajib diisi")
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	if err := s.repo.RevokeRefreshToken(ctx, helper.SHA256Hex(req.RefreshToken)); err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal keluar")
	}
	return helper.NoContent(c)
}

func (s *AuthService) Me(c *fiber.Ctx) error {
	user, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "token tidak valid")
	}
	return helper.Success(c, fiber.StatusOK, "profil berhasil diambil", fiber.Map{
		"user":        user,
		"permissions": s.permissions.PermissionsOf(user.Role),
	})
}

func (s *AuthService) issueTokenPair(user model.User) (model.TokenPair, model.RefreshToken, error) {
	authUser := model.AuthUser{UserID: user.ID, Username: user.Username, Role: user.Role}
	accessToken, err := s.jwtManager.Issue(authUser)
	if err != nil {
		return model.TokenPair{}, model.RefreshToken{}, err
	}
	refreshToken, err := helper.RandomToken(32)
	if err != nil {
		return model.TokenPair{}, model.RefreshToken{}, err
	}
	return model.TokenPair{
			AccessToken: accessToken, RefreshToken: refreshToken,
			TokenType: "Bearer", ExpiresIn: int(s.accessTTL.Seconds()),
		}, model.RefreshToken{
			UserID: user.ID, TokenHash: helper.SHA256Hex(refreshToken),
			ExpiresAt: time.Now().Add(s.refreshTTL),
		}, nil
}
