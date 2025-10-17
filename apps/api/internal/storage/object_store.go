package storage

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"nodeimage/api/internal/config"
)

type ObjectStore struct {
	client *minio.Client
	cfg    config.StorageConfig
}

func NewObjectStore(cfg config.StorageConfig) (*ObjectStore, error) {
	endpoint := cfg.Endpoint
	useSSL := cfg.UseSSL

	if strings.HasPrefix(endpoint, "http") {
		u, err := url.Parse(endpoint)
		if err != nil {
			return nil, fmt.Errorf("parse endpoint: %w", err)
		}
		endpoint = u.Host
		useSSL = u.Scheme == "https"
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: useSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("init minio: %w", err)
	}

	return &ObjectStore{
		client: client,
		cfg:    cfg,
	}, nil
}

func (s *ObjectStore) EnsureBuckets(ctx context.Context) error {
	for _, bucket := range []string{s.cfg.BucketOriginals, s.cfg.BucketVariants} {
		exists, err := s.client.BucketExists(ctx, bucket)
		if err != nil {
			return fmt.Errorf("bucket exists %s: %w", bucket, err)
		}
		if !exists {
			if err := s.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: s.cfg.Region}); err != nil {
				return fmt.Errorf("create bucket %s: %w", bucket, err)
			}
		}
	}
	return nil
}

func (s *ObjectStore) Client() *minio.Client {
	return s.client
}
