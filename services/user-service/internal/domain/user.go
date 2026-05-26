package domain

import (
	"context"
	"errors"
	"time"
)

// Sentinel errors for the user domain.
var (
	ErrUserNotFound = errors.New("user not found")
)

type User struct {
	ID          string    `json:"id"`
	DeviceID    string    `json:"device_id,omitempty"`
	Gender      string    `json:"gender"`
	Interests   []string  `json:"interests"`
	IsAnonymous bool      `json:"is_anonymous"`
	CreatedAt   time.Time `json:"created_at"`
}

type Ban struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Reason    string    `json:"reason"`
	BannedBy  string    `json:"banned_by"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByDeviceID(ctx context.Context, deviceID string) (*User, error)
	Update(ctx context.Context, user *User) error

	CreateBan(ctx context.Context, ban *Ban) error
	GetActiveBan(ctx context.Context, userID string) (*Ban, error)
}
type UserUsecase interface {
	CreateAnonymous(ctx context.Context, deviceID string) (string, string, error) // Returns access, refresh
	Refresh(ctx context.Context, refreshToken string) (string, string, error)
	GetMe(ctx context.Context, userID string) (*User, error)
	UpdateMe(ctx context.Context, userID string, gender string, interests []string) error
	GetBanStatus(ctx context.Context, userID string) (*Ban, bool, error)
	BanUser(ctx context.Context, userID, reason, bannedBy string, duration time.Duration) error
}
