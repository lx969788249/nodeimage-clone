package models

import "time"

type UserRole string

const (
	UserRoleUser       UserRole = "user"
	UserRoleAdmin      UserRole = "admin"
	UserRoleSuperAdmin UserRole = "superadmin"
)

type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusPending   UserStatus = "pending"
)

type User struct {
	ID           string
	Email        string
	PasswordHash []byte
	DisplayName  string
	Role         UserRole
	Status       UserStatus
	AvatarURL    *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Session struct {
	ID               string
	UserID           string
	DeviceID         string
	DeviceName       string
	RefreshTokenHash []byte
	IPAddress        string
	UserAgent        string
	CreatedAt        time.Time
	LastSeenAt       time.Time
	ExpiresAt        time.Time
}
