package models

import "time"

type ImageStatus string

const (
	ImageStatusProcessing ImageStatus = "processing"
	ImageStatusReady      ImageStatus = "ready"
	ImageStatusBlocked    ImageStatus = "blocked"
	ImageStatusDeleted    ImageStatus = "deleted"
)

type Image struct {
	ID         string
	UserID     string
	Bucket     string
	ObjectKey  string
	Format     string
	Width      int
	Height     int
	Frames     int
	SizeBytes  int64
	NSFWScore  *float32
	Visibility string
	Status     ImageStatus
	Checksum   []byte
	Signature  []byte
	ExpireAt   *time.Time
	DeletedAt  *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
