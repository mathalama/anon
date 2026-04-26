package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/mathalama/nektokz/matchmaking-service/internal/domain"
)

type matchUsecase struct {
	repo       domain.MatchRepository
	userClient domain.UserClient
	chatClient domain.ChatClient
}

func NewMatchUsecase(repo domain.MatchRepository, user domain.UserClient, chat domain.ChatClient) domain.MatchUsecase {
	return &matchUsecase{
		repo:       repo,
		userClient: user,
		chatClient: chat,
	}
}

func (u *matchUsecase) Search(ctx context.Context, userID string, filter domain.Filter) error {
	// 1. Check ban
	banned, err := u.userClient.IsBanned(ctx, userID)
	if err != nil {
		return err
	}
	if banned {
		return errors.New("user is banned")
	}

	// 2. Add to queue
	entry := &domain.QueueEntry{
		UserID:   userID,
		Filter:   filter,
		JoinedAt: time.Now(),
	}
	if err := u.repo.AddToQueue(ctx, entry); err != nil {
		return err
	}

	// 3. Try to find a match immediately
	go u.RunMatching(ctx, entry)

	return nil
}

func (u *matchUsecase) RunMatching(ctx context.Context, target *domain.QueueEntry) {
	// Simplified matching loop for MVP
	// In a real app, this would be a background process or triggered by events.
	// Here we just try once for demonstration or use a simple timer.
	
	// Wait a bit or loop
	for i := 0; i < 60; i++ { // Timeout after 60s
		time.Sleep(1 * time.Second)
		
		candidates, _ := u.repo.GetQueue(ctx)
		partner := Match(target, candidates)
		if partner != nil {
			// Found a match!
			roomID := uuid.New().String()
			room := &domain.Room{
				ID:        roomID,
				UserA:     target.UserID,
				UserB:     partner.UserID,
				CreatedAt: time.Now(),
			}
			
			u.repo.RemoveFromQueue(ctx, target.UserID)
			u.repo.RemoveFromQueue(ctx, partner.UserID)
			u.repo.CreateRoom(ctx, room)
			
			u.chatClient.CreateRoom(ctx, roomID, target.UserID, partner.UserID)
			return
		}

		// Filter drop logic after 30s
		if i == 30 {
			target.Filter.Gender = "any"
			target.Filter.Interests = nil
		}
	}
}

func (u *matchUsecase) Cancel(ctx context.Context, userID string) error {
	return u.repo.RemoveFromQueue(ctx, userID)
}

func (u *matchUsecase) GetStatus(ctx context.Context, userID string) (*domain.Room, error) {
	// This would typically return the room ID if matched, or 'searching' status
	// For simplicity, we just use the repo to check if the user is in a room
	// Note: I need to add GetRoom to the repo interface if I want this to work properly
	if repo, ok := u.repo.(interface {
		GetRoom(ctx context.Context, userID string) (*domain.Room, error)
	}); ok {
		return repo.GetRoom(ctx, userID)
	}
	return nil, nil
}

func (u *matchUsecase) Next(ctx context.Context, userID string) error {
	// End current room and start new search
	// For MVP, just start a new search
	return u.Search(ctx, userID, domain.Filter{Gender: "any"})
}
