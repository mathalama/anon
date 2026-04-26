package domain

import (
	"context"
	"time"
)

type User struct {
	ID           string    `json:"id"`
	DeviceID     string    `json:"device_id,omitempty"`
	Email        string    `json:"email,omitempty"`
	PasswordHash string    `json:"-"`
	Gender       string    `json:"gender"`
	Interests    []string  `json:"interests"`
	IsAnonymous  bool      `json:"is_anonymous"`
	TelegramID   int64     `json:"telegram_id,omitempty"`
	FirstName    string    `json:"first_name,omitempty"`
	LastName     string    `json:"last_name,omitempty"`
	Username     string    `json:"username,omitempty"`
	PhotoURL     string    `json:"photo_url,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
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
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByTelegramID(ctx context.Context, telegramID int64) (*User, error)
	Update(ctx context.Context, user *User) error
	
	CreateBan(ctx context.Context, ban *Ban) error
	GetActiveBan(ctx context.Context, userID string) (*Ban, error)
}

type UserUsecase interface {
	CreateAnonymous(ctx context.Context, deviceID string) (string, string, error) // Returns access, refresh
	Register(ctx context.Context, email, password string) error
	Login(ctx context.Context, email, password string) (string, string, error)
	LoginTelegram(ctx context.Context, data map[string]string) (string, string, error)
	Refresh(ctx context.Context, refreshToken string) (string, string, error)
	GetMe(ctx context.Context, userID string) (*User, error)
	UpdateMe(ctx context.Context, userID string, gender string, interests []string) error
	GetBanStatus(ctx context.Context, userID string) (*Ban, bool, error)
	BanUser(ctx context.Context, userID, reason, bannedBy string, duration time.Duration) error
}
