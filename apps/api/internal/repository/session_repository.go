package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nodeimage/api/internal/models"
)

var ErrSessionNotFound = errors.New("session not found")

type SessionRepository struct {
	pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

func (r *SessionRepository) Create(ctx context.Context, session models.Session) error {
	const query = `
		INSERT INTO user_sessions (
			id, user_id, device_id, device_name, refresh_token_hash, ip_address, user_agent, created_at, last_seen_at, expires_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, NOW(), NOW(), $8
		)
	ON CONFLICT (user_id, device_id)
	DO UPDATE SET
			id = EXCLUDED.id,
			refresh_token_hash = EXCLUDED.refresh_token_hash,
			ip_address = EXCLUDED.ip_address,
			user_agent = EXCLUDED.user_agent,
			last_seen_at = NOW(),
			expires_at = EXCLUDED.expires_at
	`

	_, err := r.pool.Exec(ctx, query,
		session.ID,
		session.UserID,
		session.DeviceID,
		session.DeviceName,
		session.RefreshTokenHash,
		session.IPAddress,
		session.UserAgent,
		session.ExpiresAt,
	)
	return err
}

func (r *SessionRepository) CountByUser(ctx context.Context, userID string) (int, error) {
	const query = `SELECT COUNT(*) FROM user_sessions WHERE user_id = $1`
	row := r.pool.QueryRow(ctx, query, userID)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *SessionRepository) DeleteOldestSessions(ctx context.Context, userID string, keepLatest int) error {
	const query = `
		DELETE FROM user_sessions
		WHERE id IN (
			SELECT id FROM user_sessions
			WHERE user_id = $1
			ORDER BY last_seen_at DESC
			OFFSET $2
		)
	`
	_, err := r.pool.Exec(ctx, query, userID, keepLatest)
	return err
}

func (r *SessionRepository) GetByID(ctx context.Context, id string) (models.Session, error) {
	const query = `
		SELECT id, user_id, device_id, device_name, refresh_token_hash, ip_address, user_agent, created_at, last_seen_at, expires_at
		FROM user_sessions
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	var session models.Session
	if err := row.Scan(
		&session.ID,
		&session.UserID,
		&session.DeviceID,
		&session.DeviceName,
		&session.RefreshTokenHash,
		&session.IPAddress,
		&session.UserAgent,
		&session.CreatedAt,
		&session.LastSeenAt,
		&session.ExpiresAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Session{}, ErrSessionNotFound
		}
		return models.Session{}, err
	}
	return session, nil
}

func (r *SessionRepository) DeleteByID(ctx context.Context, id string) error {
	const query = `DELETE FROM user_sessions WHERE id = $1`
	cmd, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrSessionNotFound
	}
	return nil
}

func (r *SessionRepository) DeleteByDevice(ctx context.Context, userID string, deviceID string) error {
	const query = `DELETE FROM user_sessions WHERE user_id = $1 AND device_id = $2`
	_, err := r.pool.Exec(ctx, query, userID, deviceID)
	return err
}

func (r *SessionRepository) FindByRefreshHash(ctx context.Context, userID string, refreshHash []byte) (models.Session, error) {
	const query = `
		SELECT id, user_id, device_id, device_name, refresh_token_hash, ip_address, user_agent, created_at, last_seen_at, expires_at
		FROM user_sessions
		WHERE user_id = $1 AND refresh_token_hash = $2
	`
	row := r.pool.QueryRow(ctx, query, userID, refreshHash)
	var session models.Session
	if err := row.Scan(
		&session.ID,
		&session.UserID,
		&session.DeviceID,
		&session.DeviceName,
		&session.RefreshTokenHash,
		&session.IPAddress,
		&session.UserAgent,
		&session.CreatedAt,
		&session.LastSeenAt,
		&session.ExpiresAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Session{}, ErrSessionNotFound
		}
		return models.Session{}, err
	}
	return session, nil
}

func (r *SessionRepository) ListByUser(ctx context.Context, userID string) ([]models.Session, error) {
	const query = `
		SELECT id, user_id, device_id, device_name, refresh_token_hash, ip_address, user_agent, created_at, last_seen_at, expires_at
		FROM user_sessions
		WHERE user_id = $1
		ORDER BY last_seen_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []models.Session
	for rows.Next() {
		var session models.Session
		if err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.DeviceID,
			&session.DeviceName,
			&session.RefreshTokenHash,
			&session.IPAddress,
			&session.UserAgent,
			&session.CreatedAt,
			&session.LastSeenAt,
			&session.ExpiresAt,
		); err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, rows.Err()
}

func (r *SessionRepository) Touch(ctx context.Context, sessionID string, ip string, userAgent string) error {
	const query = `
		UPDATE user_sessions
		SET last_seen_at = NOW(),
		    ip_address = COALESCE(NULLIF($2, ''), ip_address),
		    user_agent = COALESCE(NULLIF($3, ''), user_agent)
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, sessionID, ip, userAgent)
	return err
}
