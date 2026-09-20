package repository

import (
	"context"
	"errors"
	"fmt"

	"latihan-fiber/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userPostgresRepository struct{ pool *pgxpool.Pool }

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userPostgresRepository{pool: pool}
}

func (r *userPostgresRepository) FindAll(ctx context.Context) ([]model.User, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, username, email, role, created_at FROM users ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar user: %w", err)
	}
	defer rows.Close()
	users := []model.User{}
	for rows.Next() {
		var user model.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.CreatedAt); err != nil {
			return nil, fmt.Errorf("membaca user: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("membaca daftar user: %w", err)
	}
	return users, nil
}

func (r *userPostgresRepository) FindByID(ctx context.Context, id int) (model.User, error) {
	return r.find(ctx, "WHERE id = $1", id)
}

func (r *userPostgresRepository) Create(ctx context.Context, user model.User) (model.User, error) {
	err := r.pool.QueryRow(ctx, `INSERT INTO users (username, email, password_hash, role) VALUES ($1, $2, $3, $4) RETURNING id, username, email, role, created_at`, user.Username, user.Email, user.PasswordHash, user.Role).
		Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return model.User{}, ErrUsernameOrEmailTaken
		}
		return model.User{}, fmt.Errorf("membuat user: %w", err)
	}
	return user, nil
}

func (r *userPostgresRepository) Update(ctx context.Context, user model.User) (model.User, error) {
	err := r.pool.QueryRow(ctx, `UPDATE users SET username = $1, email = $2 WHERE id = $3 RETURNING id, username, email, role, created_at`, user.Username, user.Email, user.ID).
		Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, ErrUserNotFound
	}
	if err != nil {
		if isUniqueViolation(err) {
			return model.User{}, ErrUsernameOrEmailTaken
		}
		return model.User{}, fmt.Errorf("memperbarui user: %w", err)
	}
	return user, nil
}

func (r *userPostgresRepository) UpdateRole(ctx context.Context, id int, role string) (model.User, error) {
	return r.updateRole(ctx, id, role)
}

func (r *userPostgresRepository) updateRole(ctx context.Context, id int, role string) (model.User, error) {
	var user model.User
	err := r.pool.QueryRow(ctx, `UPDATE users SET role = $1 WHERE id = $2 RETURNING id, username, email, role, created_at`, role, id).
		Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, ErrUserNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("mengubah role user: %w", err)
	}
	return user, nil
}

func (r *userPostgresRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("menghapus user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *userPostgresRepository) find(ctx context.Context, condition string, arg any) (model.User, error) {
	var user model.User
	err := r.pool.QueryRow(ctx, `SELECT id, username, email, role, created_at FROM users `+condition, arg).
		Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, ErrUserNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("mengambil user: %w", err)
	}
	return user, nil
}
