package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nodeimage/api/internal/models"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, user models.User) error {
	const query = `
		INSERT INTO users (
			id, email, password_hash, display_name, role, status, avatar_url, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, NOW(), NOW()
		)
	`

	_, err := r.pool.Exec(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.DisplayName,
		user.Role,
		user.Status,
		user.AvatarURL,
	)
	return err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (models.User, error) {
	const query = `
		SELECT id, email, password_hash, display_name, role, status, avatar_url, created_at, updated_at
		FROM users WHERE email = $1
	`

	row := r.pool.QueryRow(ctx, query, email)
	var user models.User
	if err := row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.DisplayName,
		&user.Role,
		&user.Status,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, err
	}
	return user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (models.User, error) {
	const query = `
		SELECT id, email, password_hash, display_name, role, status, avatar_url, created_at, updated_at
		FROM users WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	var user models.User
	if err := row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.DisplayName,
		&user.Role,
		&user.Status,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, err
	}
	return user, nil
}

func (r *UserRepository) UpdateStatus(ctx context.Context, id string, status models.UserStatus) error {
	const query = `
		UPDATE users SET status = $2, updated_at = NOW() WHERE id = $1
	`
	cmd, err := r.pool.Exec(ctx, query, id, status)
	if err != nil {
		return err
}
	if cmd.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}
