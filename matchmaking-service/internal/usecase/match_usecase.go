package usecase

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/mathalama/nektokz/matchmaking-service/internal/domain"
)

type matchUsecase struct {
	repo               domain.MatchRepository
	userClient         domain.UserClient
	chatClient         domain.ChatClient
	matchTimeoutSec    int
	matchFilterDropSec int
}

func NewMatchUsecase(repo domain.MatchRepository, user domain.UserClient, chat domain.ChatClient, matchTimeoutSec, matchFilterDropSec int) domain.MatchUsecase {
	return &matchUsecase{
		repo:               repo,
		userClient:         user,
		chatClient:         chat,
		matchTimeoutSec:    matchTimeoutSec,
		matchFilterDropSec: matchFilterDropSec,
	}
}

func (u *matchUsecase) Search(ctx context.Context, userID string, filter domain.Filter) error {
	log.Printf("MATCHMAKING: Search requested by user %s with mode=%s gender=%s", userID, filter.Mode, filter.Gender)
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

	// 2. Remove old queue entry first (clears stale mode/filter from previous search)
	if err := u.repo.RemoveFromQueue(ctx, userID); err != nil {
		log.Printf("MATCHMAKING: Warning - failed to remove old queue entry for %s: %v", userID, err)
	}

	// 3. Add fresh entry with new filter
	entry := &domain.QueueEntry{
		UserID:   userID,
		Filter:   filter,
		JoinedAt: time.Now(),
	}
	if err := u.repo.AddToQueue(ctx, entry); err != nil {
		return err
	}

	return nil
}

func (u *matchUsecase) StartWorker(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			u.RunMatching(ctx)
		}
	}
}

func (u *matchUsecase) RunMatching(ctx context.Context) {
	candidates, err := u.repo.GetQueue(ctx)
	if err != nil || len(candidates) < 2 {
		return
	}

	// Group by mode to reduce cross-checks
	byMode := make(map[string][]*domain.QueueEntry)
	for _, c := range candidates {
		byMode[c.Filter.Mode] = append(byMode[c.Filter.Mode], c)
	}

	matched := make(map[string]bool)

	for mode, modeCandidates := range byMode {
		// Limit candidates per mode to prevent O(N^2) explosion
		if len(modeCandidates) > 500 {
			modeCandidates = modeCandidates[:500]
		}

		for i := 0; i < len(modeCandidates); i++ {
			userA := modeCandidates[i]
			if matched[userA.UserID] {
				continue
			}

			for j := i + 1; j < len(modeCandidates); j++ {
				userB := modeCandidates[j]
				if matched[userB.UserID] {
					continue
				}

				if IsCompatible(userA.Filter, userB.Filter) {
					roomID := uuid.New().String()
					room := &domain.Room{
						ID:        roomID,
						UserA:     userA.UserID,
						UserB:     userB.UserID,
						Mode:      mode,
						CreatedAt: time.Now(),
					}

					if err := u.repo.CreateRoom(ctx, room); err == nil {
						_ = u.chatClient.CreateRoom(ctx, roomID, userA.UserID, userB.UserID)

						log.Printf("MATCHMAKING: Match found in %s! %s <-> %s", mode, userA.UserID, userB.UserID)
						_ = u.repo.PublishMatch(ctx, userA.UserID, &domain.MatchFound{
							RoomID:        roomID,
							Mode:          room.Mode,
							IsInitiator:   true,
							PartnerGender: userB.Filter.MyGender,
						})
						_ = u.repo.PublishMatch(ctx, userB.UserID, &domain.MatchFound{
							RoomID:        roomID,
							Mode:          room.Mode,
							IsInitiator:   false,
							PartnerGender: userA.Filter.MyGender,
						})

						matched[userA.UserID] = true
						matched[userB.UserID] = true
						break
					}
				}
			}
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

func (u *matchUsecase) SubscribeToMatch(ctx context.Context, userID string) (<-chan *domain.MatchFound, func(), error) {
	return u.repo.SubscribeToMatch(ctx, userID)
}

func (u *matchUsecase) HealthCheck(ctx context.Context) error {
	return u.repo.HealthCheck(ctx)
}