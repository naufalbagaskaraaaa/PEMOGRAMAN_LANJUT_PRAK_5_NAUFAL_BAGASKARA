package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"latihan-fiber/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound         = errors.New("pengguna tidak ditemukan")
	ErrUsernameOrEmailTaken = errors.New("username atau email sudah terdaftar")
	ErrRefreshTokenInvalid  = errors.New("refresh token tidak valid")
)

type AuthRepository interface {
	CreateUser(ctx context.Context, user model.User) (model.User, error)
	FindUserByUsername(ctx context.Context, username string) (model.User, error)
	FindUserByID(ctx context.Context, id int) (model.User, error)
	FindRefreshToken(ctx context.Context, tokenHash string) (model.RefreshToken, error)
	CreateRefreshToken(ctx context.Context, token model.RefreshToken) error
	RotateRefreshToken(ctx context.Context, oldHash string, token model.RefreshToken) error
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
}

type authPostgresRepository struct{ pool *pgxpool.Pool }

func NewAuthRepository(pool *pgxpool.Pool) AuthRepository {
	return &authPostgresRepository{pool: pool}
}

func (r *authPostgresRepository) CreateUser(ctx context.Context, user model.User) (model.User, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO users (username, email, password_hash, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, username, email, password_hash, role, created_at`,
		user.Username, user.Email, user.PasswordHash, user.Role,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return model.User{}, ErrUsernameOrEmailTaken
		}
		return model.User{}, fmt.Errorf("membuat pengguna: %w", err)
	}
	return user, nil
}

func (r *authPostgresRepository) FindUserByUsername(ctx context.Context, username string) (model.User, error) {
	return r.findUser(ctx, `WHERE username = $1`, username)
}

func (r *authPostgresRepository) FindUserByID(ctx context.Context, id int) (model.User, error) {
	return r.findUser(ctx, `WHERE id = $1`, id)
}

func (r *authPostgresRepository) findUser(ctx context.Context, condition string, arg any) (model.User, error) {
	var user model.User
	err := r.pool.QueryRow(ctx, `SELECT id, username, email, password_hash, role, created_at FROM users `+condition, arg).
		Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, ErrUserNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("mengambil pengguna: %w", err)
	}
	return user, nil
}

func (r *authPostgresRepository) FindRefreshToken(ctx context.Context, tokenHash string) (model.RefreshToken, error) {
	var token model.RefreshToken
	err := r.pool.QueryRow(ctx, `SELECT id, user_id, token_hash, expires_at, revoked_at, created_at FROM refresh_tokens WHERE token_hash = $1`, tokenHash).
		Scan(&token.ID, &token.UserID, &token.TokenHash, &token.ExpiresAt, &token.RevokedAt, &token.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.RefreshToken{}, ErrRefreshTokenInvalid
	}
	if err != nil {
		return model.RefreshToken{}, fmt.Errorf("mengambil refresh token: %w", err)
	}
	return token, nil
}

func (r *authPostgresRepository) CreateRefreshToken(ctx context.Context, token model.RefreshToken) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`, token.UserID, token.TokenHash, token.ExpiresAt)
	return err
}

func (r *authPostgresRepository) RotateRefreshToken(ctx context.Context, oldHash string, token model.RefreshToken) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	result, err := tx.Exec(ctx, `UPDATE refresh_tokens SET revoked_at = NOW() WHERE token_hash = $1 AND revoked_at IS NULL`, oldHash)
	if err != nil || result.RowsAffected() != 1 {
		return ErrRefreshTokenInvalid
	}
	if _, err = tx.Exec(ctx, `INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`, token.UserID, token.TokenHash, token.ExpiresAt); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *authPostgresRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := r.pool.Exec(ctx, `UPDATE refresh_tokens SET revoked_at = NOW() WHERE token_hash = $1 AND revoked_at IS NULL`, tokenHash)
	return err
}

func refreshTokenUsable(token model.RefreshToken, now time.Time) bool {
	return token.RevokedAt == nil && token.ExpiresAt.After(now)
}
