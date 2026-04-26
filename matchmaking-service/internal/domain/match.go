package domain

import (
	"context"
	"time"
)

type Filter struct {
	Gender    string   `json:"gender"`
	Interests []string `json:"interests"`
}

type QueueEntry struct {
	UserID    string    `json:"user_id"`
	Filter    Filter    `json:"filter"`
	JoinedAt  time.Time `json:"joined_at"`
}

type Room struct {
	ID        string    `json:"id"`
	UserA     string    `json:"user_a"`
	UserB     string    `json:"user_b"`
	CreatedAt time.Time `json:"created_at"`
}

type MatchRepository interface {
	AddToQueue(ctx context.Context, entry *QueueEntry) error
	RemoveFromQueue(ctx context.Context, userID string) error
	GetQueue(ctx context.Context) ([]*QueueEntry, error)
	CreateRoom(ctx context.Context, room *Room) error
	GetRoom(ctx context.Context, userID string) (*Room, error)
}

type UserClient interface {
	IsBanned(ctx context.Context, userID string) (bool, error)
}

type ChatClient interface {
	CreateRoom(ctx context.Context, roomID string, userA, userB string) error
}

type MatchUsecase interface {
	Search(ctx context.Context, userID string, filter Filter) error
	Cancel(ctx context.Context, userID string) error
	GetStatus(ctx context.Context, userID string) (*Room, error)
	Next(ctx context.Context, userID string) error
}
