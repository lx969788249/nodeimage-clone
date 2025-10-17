package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"nodeimage/api/internal/config"
	"nodeimage/api/internal/ids"
	"nodeimage/api/internal/media/sniffer"
	"nodeimage/api/internal/media/svg"
	"nodeimage/api/internal/models"
	"nodeimage/api/internal/repository"
	"nodeimage/api/internal/security"
	"nodeimage/api/internal/storage"
)

type UploadInput struct {
	User       models.User
	DeviceID   string
	File       multipart.File
	Header     *multipart.FileHeader
	Visibility string
	ExpireAt   *time.Time
}

type UploadResult struct {
	Image models.Image
	URL   string
}

type UploadService struct {
	images *repository.ImageRepository
	store  *storage.ObjectStore
	queue  *redis.Client
	cfg    *config.AppConfig
	log    zerolog.Logger
}

func NewUploadService(images *repository.ImageRepository, store *storage.ObjectStore, queue *redis.Client, cfg *config.AppConfig, log zerolog.Logger) *UploadService {
	return &UploadService{
		images: images,
		store:  store,
		queue:  queue,
		cfg:    cfg,
		log:    log,
	}
}

func (s *UploadService) Upload(ctx context.Context, input UploadInput) (UploadResult, error) {
	if input.File == nil || input.Header == nil {
		return UploadResult{}, errors.New("invalid file payload")
	}

	head := make([]byte, 512)
	n, err := input.File.Read(head)
	if err != nil && !errors.Is(err, io.EOF) {
		return UploadResult{}, fmt.Errorf("read head: %w", err)
	}
	head = head[:n]

	var data []byte
	if seeker, ok := input.File.(io.ReadSeeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return UploadResult{}, fmt.Errorf("rewind: %w", err)
		}
		data, err = io.ReadAll(seeker)
		if err != nil {
			return UploadResult{}, fmt.Errorf("read file: %w", err)
		}
	} else {
		rest, err := io.ReadAll(input.File)
		if err != nil {
			return UploadResult{}, fmt.Errorf("read file: %w", err)
		}
		data = append(head, rest...)
	}

	result, err := sniffer.DetectHead(head)
	if err != nil {
		return UploadResult{}, fmt.Errorf("detect type: %w", err)
	}

	declared := sniffer.MimeTypeFromHTTP(input.Header.Header)
	if declared != "" && declared != result.MIME {
		return UploadResult{}, fmt.Errorf("content type mismatch: declared %s, actual %s", declared, result.MIME)
	}

	if len(data) == 0 {
		return UploadResult{}, errors.New("empty file")
	}

	if result.Type == sniffer.TypeSVG {
		clean, err := svg.Sanitize(data)
		if err != nil {
			return UploadResult{}, fmt.Errorf("sanitize svg: %w", err)
		}
		data = clean
	}

	imageID := ids.New()
	objectKey := s.buildObjectKey(imageID, string(result.Type))

	reader := bytes.NewReader(data)

	options := minio.PutObjectOptions{
		ContentType: result.MIME,
	}

	uploadInfo, err := s.store.Client().PutObject(ctx, s.cfg.Storage.BucketOriginals, objectKey, reader, int64(len(data)), options)
	if err != nil {
		return UploadResult{}, fmt.Errorf("put object: %w", err)
	}

	sum := sha256.Sum256(data)
	checksum := make([]byte, len(sum))
	copy(checksum, sum[:])
	signature := security.SignResource(s.cfg.Security.SignatureSecret, imageID, objectKey)

	image := models.Image{
		ID:        imageID,
		UserID:    input.User.ID,
		Bucket:    s.cfg.Storage.BucketOriginals,
		ObjectKey: objectKey,
		Format:    string(result.Type),
		Width:     0,
		Height:    0,
		Frames:    1,
		SizeBytes: uploadInfo.Size,
		Status:    models.ImageStatusProcessing,
		Visibility: func() string {
			if input.Visibility != "" {
				return input.Visibility
			}
			return "public"
		}(),
		Checksum:  checksum,
		Signature: signature,
		ExpireAt:  input.ExpireAt,
	}
	now := time.Now().UTC()
	image.CreatedAt = now
	image.UpdatedAt = now

	if err := s.images.Create(ctx, image); err != nil {
		return UploadResult{}, fmt.Errorf("save metadata: %w", err)
	}

	if err := s.enqueueProcessing(ctx, image); err != nil {
		s.log.Warn().Err(err).Str("image_id", image.ID).Msg("enqueue processing failed")
	}

	url := s.buildPublicURL(s.cfg.Storage.BucketOriginals, objectKey)

	return UploadResult{
		Image: image,
		URL:   url,
	}, nil
}

func (s *UploadService) buildObjectKey(imageID string, ext string) string {
	datePrefix := time.Now().UTC().Format("2006/01/02")
	return path.Join(datePrefix, fmt.Sprintf("%s.%s", imageID, ext))
}

func (s *UploadService) buildPublicURL(bucket, objectKey string) string {
	base := strings.TrimSuffix(s.cfg.Storage.Endpoint, "/")
	if !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		base = "https://" + base
	}
	return fmt.Sprintf("%s/%s/%s", base, bucket, objectKey)
}

func (s *UploadService) enqueueProcessing(ctx context.Context, image models.Image) error {
	if s.queue == nil {
		return nil
	}

	payload := map[string]any{
		"type":    "ingest",
		"imageId": image.ID,
		"bucket":  image.Bucket,
		"object":  image.ObjectKey,
		"format":  image.Format,
	}
	_, err := s.queue.XAdd(ctx, &redis.XAddArgs{
		Stream: "media:ingest",
		Values: payload,
	}).Result()
	return err
}
