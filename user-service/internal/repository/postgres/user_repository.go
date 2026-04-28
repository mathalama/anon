package postgres

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/mathalama/nektokz/user-service/internal/domain"
)

// InMemoryUserRepository is a stub for testing and initial development
type InMemoryUserRepository struct {
	mu    sync.RWMutex
	users map[string]*domain.User
	bans  map[string][]*domain.Ban
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users: make(map[string]*domain.User),
		bans:  make(map[string][]*domain.Ban),
	}
}

func (r *InMemoryUserRepository) Create(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[user.ID] = user
	return nil
}

func (r *InMemoryUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, ok := r.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (r *InMemoryUserRepository) GetByDeviceID(ctx context.Context, deviceID string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if u.DeviceID == deviceID {
			return u, nil
		}
	}
	return nil, errors.New("user not found")
}

func (r *InMemoryUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, errors.New("user not found")
}

func (r *InMemoryUserRepository) Update(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[user.ID] = user
	return nil
}

func (r *InMemoryUserRepository) CreateBan(ctx context.Context, ban *domain.Ban) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.bans[ban.UserID] = append(r.bans[ban.UserID], ban)
	return nil
}

func (r *InMemoryUserRepository) GetActiveBan(ctx context.Context, userID string) (*domain.Ban, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	bans, ok := r.bans[userID]
	if !ok || len(bans) == 0 {
		return nil, nil
	}
	ban := bans[len(bans)-1]
	if ban.ExpiresAt.Before(time.Now()) {
		return nil, nil
	}
	return ban, nil
}
