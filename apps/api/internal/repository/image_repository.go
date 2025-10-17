package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nodeimage/api/internal/models"
)

type ImageRepository struct {
	pool *pgxpool.Pool
}

func NewImageRepository(pool *pgxpool.Pool) *ImageRepository {
	return &ImageRepository{pool: pool}
}

func (r *ImageRepository) Create(ctx context.Context, image models.Image) error {
	const query = `
		INSERT INTO images (
			id, user_id, bucket, object_key, format, width, height, frames, size_bytes,
			nsfw_score, visibility, status, checksum, signature, expire_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9,
			$10, $11, $12, $13, $14, $15, NOW(), NOW()
		)
	`

	_, err := r.pool.Exec(ctx, query,
		image.ID,
		image.UserID,
		image.Bucket,
		image.ObjectKey,
		image.Format,
		image.Width,
		image.Height,
		image.Frames,
		image.SizeBytes,
		image.NSFWScore,
		image.Visibility,
		image.Status,
		image.Checksum,
		image.Signature,
		image.ExpireAt,
	)
	return err
}

func (r *ImageRepository) UpdateStatus(ctx context.Context, id string, status models.ImageStatus, nsfwScore *float32) error {
	const query = `
		UPDATE images
		SET status = $2,
		    nsfw_score = COALESCE($3, nsfw_score),
		    updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, id, status, nsfwScore)
	return err
}

func (r *ImageRepository) GetByID(ctx context.Context, id string) (models.Image, error) {
	const query = `
		SELECT id, user_id, bucket, object_key, format, width, height, frames, size_bytes,
		       nsfw_score, visibility, status, checksum, signature, expire_at, deleted_at,
		       created_at, updated_at
		FROM images WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	var image models.Image
	if err := row.Scan(
		&image.ID,
		&image.UserID,
		&image.Bucket,
		&image.ObjectKey,
		&image.Format,
		&image.Width,
		&image.Height,
		&image.Frames,
		&image.SizeBytes,
		&image.NSFWScore,
		&image.Visibility,
		&image.Status,
		&image.Checksum,
		&image.Signature,
		&image.ExpireAt,
		&image.DeletedAt,
		&image.CreatedAt,
		&image.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Image{}, ErrImageNotFound
		}
		return models.Image{}, err
	}
	return image, nil
}

func (r *ImageRepository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]models.Image, error) {
	const query = `
		SELECT id, user_id, bucket, object_key, format, width, height, frames, size_bytes,
		       nsfw_score, visibility, status, checksum, signature, expire_at, deleted_at,
		       created_at, updated_at
		FROM images
		WHERE user_id = $1 AND status != 'deleted'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []models.Image
	for rows.Next() {
		var image models.Image
		if err := rows.Scan(
			&image.ID,
			&image.UserID,
			&image.Bucket,
			&image.ObjectKey,
			&image.Format,
			&image.Width,
			&image.Height,
			&image.Frames,
			&image.SizeBytes,
			&image.NSFWScore,
			&image.Visibility,
			&image.Status,
			&image.Checksum,
			&image.Signature,
			&image.ExpireAt,
			&image.DeletedAt,
			&image.CreatedAt,
			&image.UpdatedAt,
		); err != nil {
			return nil, err
		}
		images = append(images, image)
	}
	return images, rows.Err()
}

var ErrImageNotFound = errors.New("image not found")

func (r *ImageRepository) List(ctx context.Context, limit, offset int) ([]models.Image, error) {
	const query = `
		SELECT id, user_id, bucket, object_key, format, width, height, frames, size_bytes,
		       nsfw_score, visibility, status, checksum, signature, expire_at, deleted_at,
		       created_at, updated_at
		FROM images
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []models.Image
	for rows.Next() {
		var image models.Image
		if err := rows.Scan(
			&image.ID,
			&image.UserID,
			&image.Bucket,
			&image.ObjectKey,
			&image.Format,
			&image.Width,
			&image.Height,
			&image.Frames,
			&image.SizeBytes,
			&image.NSFWScore,
			&image.Visibility,
			&image.Status,
			&image.Checksum,
			&image.Signature,
			&image.ExpireAt,
			&image.DeletedAt,
			&image.CreatedAt,
			&image.UpdatedAt,
		); err != nil {
			return nil, err
		}
		images = append(images, image)
	}
	return images, rows.Err()
}
