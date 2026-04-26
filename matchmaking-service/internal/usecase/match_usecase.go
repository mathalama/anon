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
	matchTimeoutSec    int
	matchFilterDropSec int
}

func NewMatchUsecase(repo domain.MatchRepository, user domain.UserClient, chat domain.ChatClient, matchTimeoutSec, matchFilterDropSec int) domain.MatchUsecase {
	return &matchUsecase{
		repo:       repo,
		userClient: user,
		chatClient: chat,
		matchTimeoutSec:    matchTimeoutSec,
		matchFilterDropSec: matchFilterDropSec,
	}
}

func (u *matchUsecase) Search(ctx context.Context, userID string, filter domain.Filter) error {
	if filter.Gender == "" {
		filter.Gender = "any"
	}

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
	go u.RunMatching(context.Background(), entry)

	return nil
}

func (u *matchUsecase) RunMatching(ctx context.Context, target *domain.QueueEntry) {
	// Simplified matching loop for MVP
	// In a real app, this would be a background process or triggered by events.
	// Here we just try once for demonstration or use a simple timer.
	
	// Wait a bit or loop
	timeout := u.matchTimeoutSec
	if timeout <= 0 {
		timeout = 60
	}
	dropAfter := u.matchFilterDropSec
	if dropAfter <= 0 {
		dropAfter = 30
	}

	for i := 0; i < timeout; i++ { // Timeout after N seconds
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

		// Filter drop logic after N seconds
		if i == dropAfter {
			target.Filter.Gender = "any"
			target.Filter.Interests = nil
			_ = u.repo.AddToQueue(ctx, target)
		}
	}
}

func (u *matchUsecase) Cancel(ctx context.Context, userID string) error {
	return u.repo.RemoveFromQueue(ctx, userID)
}

func (u *matchUsecase) GetStatus(ctx context.Context, userID string) (*domain.Room, error) {
	return u.repo.GetRoom(ctx, userID)
}

func (u *matchUsecase) Next(ctx context.Context, userID string) error {
	// End current room and start new search
	// For MVP, just start a new search
	return u.Search(ctx, userID, domain.Filter{Gender: "any"})
}
