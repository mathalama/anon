package domain

import (
	"context"
	"time"
)

type Filter struct {
	MyGender  string   `json:"my_gender" validate:"required,oneof=male female"`
	Gender    string   `json:"gender" validate:"oneof=male female any ''"`
	Interests []string `json:"interests" validate:"max=10"`
	Mode      string   `json:"mode" validate:"required,oneof=text voice"` // "text" or "voice"
}

type QueueEntry struct {
	UserID    string             `json:"user_id"`
	Filter    Filter             `json:"filter"`
	JoinedAt  time.Time          `json:"joined_at"`
	Cancel    context.CancelFunc `json:"-"`
}

type Room struct {
	ID        string    `json:"id"`
	UserA     string    `json:"user_a"`
	UserB     string    `json:"user_b"`
	Mode      string    `json:"mode"`
	CreatedAt time.Time `json:"created_at"`
}

type MatchFound struct {
	RoomID        string `json:"room_id"`
	Mode          string `json:"mode"`
	IsInitiator   bool   `json:"is_initiator"`
	PartnerGender string `json:"partner_gender"`
	PartnerUserID string `json:"partner_user_id"`
}

type MatchRepository interface {
	AddToQueue(ctx context.Context, entry *QueueEntry) error
	RemoveFromQueue(ctx context.Context, userID string) error
	GetQueue(ctx context.Context) ([]*QueueEntry, error)
	CreateRoom(ctx context.Context, room *Room) error
	GetRoom(ctx context.Context, userID string) (*Room, error)
	DeleteRoom(ctx context.Context, userID string) error
	PublishMatch(ctx context.Context, userID string, match *MatchFound) error
	SubscribeToMatch(ctx context.Context, userID string) (<-chan *MatchFound, func(), error)
	PopSegment(ctx context.Context, key string, count int) ([]string, error)
	HealthCheck(ctx context.Context) error
}

type UserClient interface {
	IsBanned(ctx context.Context, userID string) (bool, error)
}

type ChatClient interface {
	CreateRoom(ctx context.Context, roomID string, userA, userB string) error
}

type MQPublisher interface {
	Publish(ctx context.Context, subject string, data interface{}) error
}

type MatchUsecase interface {
	Search(ctx context.Context, userID string, filter Filter) error
	Cancel(ctx context.Context, userID string) error
	GetStatus(ctx context.Context, userID string) (*Room, error)
	Next(ctx context.Context, userID string) error
	SubscribeToMatch(ctx context.Context, userID string) (<-chan *MatchFound, func(), error)
	StartWorker(ctx context.Context)
	HealthCheck(ctx context.Context) error
}
