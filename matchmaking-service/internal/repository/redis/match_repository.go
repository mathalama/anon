package redis

import (
	"context"
	"sync"

	"github.com/mathalama/nektokz/matchmaking-service/internal/domain"
)

type InMemoryMatchRepository struct {
	mu          sync.RWMutex
	queue       []*domain.QueueEntry
	rooms       map[string]*domain.Room
	subscribers map[string]chan *domain.MatchFound
}

func NewInMemoryMatchRepository() *InMemoryMatchRepository {
	return &InMemoryMatchRepository{
		rooms:       make(map[string]*domain.Room),
		subscribers: make(map[string]chan *domain.MatchFound),
	}
}

func (r *InMemoryMatchRepository) AddToQueue(ctx context.Context, entry *domain.QueueEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.queue = append(r.queue, entry)
	return nil
}

func (r *InMemoryMatchRepository) RemoveFromQueue(ctx context.Context, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, e := range r.queue {
		if e.UserID == userID {
			r.queue = append(r.queue[:i], r.queue[i+1:]...)
			return nil
		}
	}
	return nil
}

func (r *InMemoryMatchRepository) GetQueue(ctx context.Context) ([]*domain.QueueEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.queue, nil
}

func (r *InMemoryMatchRepository) CreateRoom(ctx context.Context, room *domain.Room) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rooms[room.UserA] = room
	r.rooms[room.UserB] = room
	return nil
}

func (r *InMemoryMatchRepository) GetRoom(ctx context.Context, userID string) (*domain.Room, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	room, ok := r.rooms[userID]
	if !ok {
		return nil, nil
	}
	return room, nil
}

func (r *InMemoryMatchRepository) PublishMatch(ctx context.Context, userID string, match *domain.MatchFound) error {
	r.mu.RLock()
	ch, ok := r.subscribers[userID]
	r.mu.RUnlock()
	if ok {
		select {
		case ch <- match:
		default:
		}
	}
	return nil
}

func (r *InMemoryMatchRepository) SubscribeToMatch(ctx context.Context, userID string) (<-chan *domain.MatchFound, func(), error) {
	ch := make(chan *domain.MatchFound, 1)
	r.mu.Lock()
	r.subscribers[userID] = ch
	r.mu.Unlock()

	cleanup := func() {
		r.mu.Lock()
		delete(r.subscribers, userID)
		r.mu.Unlock()
		close(ch)
	}
	return ch, cleanup, nil
}

func (r *InMemoryMatchRepository) HealthCheck(ctx context.Context) error {
	return nil
}
