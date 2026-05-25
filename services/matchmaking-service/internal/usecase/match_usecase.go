package usecase

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mathalama/nektokz/matchmaking-service/internal/domain"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	queueUsersGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nektokz_matchmaking_queue_users",
			Help: "Number of users currently in the matchmaking queue.",
		},
		[]string{"mode", "my_gender", "target_gender"},
	)

	roomsCreatedCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "nektokz_matchmaking_rooms_created_total",
			Help: "Total number of chat rooms created by matchmaking.",
		},
		[]string{"mode"},
	)
)

type matchUsecase struct {
	repo               domain.MatchRepository
	userClient         domain.UserClient
	chatClient         domain.ChatClient
	mq                 domain.MQPublisher
	matchTimeoutSec    int
	matchFilterDropSec int
}

func NewMatchUsecase(repo domain.MatchRepository, user domain.UserClient, chat domain.ChatClient, mq domain.MQPublisher, matchTimeoutSec, matchFilterDropSec int) domain.MatchUsecase {
	return &matchUsecase{
		repo:               repo,
		userClient:         user,
		chatClient:         chat,
		mq:                 mq,
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

	if err := u.repo.DeleteRoom(ctx, userID); err != nil {
		log.Printf("Warning: failed to delete old room for %s: %v", userID, err)
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

	// Scrape queue sizes for metrics every 2 seconds in the background
	metricsTicker := time.NewTicker(2 * time.Second)
	defer metricsTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			u.RunMatching(ctx)
		case <-metricsTicker.C:
			u.collectQueueMetrics(ctx)
		}
	}
}

func (u *matchUsecase) collectQueueMetrics(ctx context.Context) {
	queue, err := u.repo.GetQueue(ctx)
	if err != nil {
		return
	}

	// Reset gauge to ensure stale tags from departed users are cleared
	queueUsersGauge.Reset()

	counts := make(map[string]int)
	for _, entry := range queue {
		mode := entry.Filter.Mode
		if mode == "" {
			mode = "text"
		}
		g := entry.Filter.Gender
		if g == "" {
			g = "any"
		}
		labelKey := mode + ":" + entry.Filter.MyGender + ":" + g
		counts[labelKey]++
	}

	for key, count := range counts {
		parts := strings.Split(key, ":")
		if len(parts) == 3 {
			queueUsersGauge.WithLabelValues(parts[0], parts[1], parts[2]).Set(float64(count))
		}
	}
}

func (u *matchUsecase) RunMatching(ctx context.Context) {
	modes := []string{"text", "voice"}

	for _, mode := range modes {
		// 1. Cross-gender matching (Male <-> Female)
		u.matchQueues(ctx, mode, "male", "female", "female", "male")
		u.matchQueues(ctx, mode, "male", "female", "female", "any")
		u.matchQueues(ctx, mode, "male", "any", "female", "male")
		u.matchQueues(ctx, mode, "male", "any", "female", "any")

		// 2. Same-gender matching (Male <-> Male)
		u.matchQueues(ctx, mode, "male", "male", "male", "male")
		u.matchQueues(ctx, mode, "male", "male", "male", "any")
		u.matchQueues(ctx, mode, "male", "any", "male", "any")

		// 3. Same-gender matching (Female <-> Female)
		u.matchQueues(ctx, mode, "female", "female", "female", "female")
		u.matchQueues(ctx, mode, "female", "female", "female", "any")
		u.matchQueues(ctx, mode, "female", "any", "female", "any")
	}
}

func (u *matchUsecase) matchQueues(ctx context.Context, mode, g1, t1, g2, t2 string) {
	q1 := "queue:" + mode + ":" + g1 + ":" + t1
	q2 := "queue:" + mode + ":" + g2 + ":" + t2

	// Limit number of pairs per tick to prevent blocking
	for i := 0; i < 50; i++ {
		var userA, userB string

		if q1 == q2 {
			users, err := u.repo.PopSegment(ctx, q1, 2)
			if err != nil || len(users) < 2 {
				return
			}
			userA, userB = users[0], users[1]
		} else {
			usersA, errA := u.repo.PopSegment(ctx, q1, 1)
			usersB, errB := u.repo.PopSegment(ctx, q2, 1)
			if errA != nil || errB != nil || len(usersA) < 1 || len(usersB) < 1 {
				return
			}
			userA, userB = usersA[0], usersB[0]
		}

		// Double check compatibility (optional but good for interests)
		// For now, since we segment by gender/mode, we just create the room
		u.createRoomForPair(ctx, userA, userB, mode, g1, g2)
	}
}

func (u *matchUsecase) createRoomForPair(ctx context.Context, userA, userB, mode, gA, gB string) {
	roomID := uuid.New().String()
	room := &domain.Room{
		ID:        roomID,
		UserA:     userA,
		UserB:     userB,
		Mode:      mode,
		CreatedAt: time.Now(),
	}

	if err := u.repo.CreateRoom(ctx, room); err != nil {
		// One or both users might have been matched already or canceled
		return
	}

	// Increment matchmaking room creations counter
	roomsCreatedCounter.WithLabelValues(mode).Inc()

	if err := u.chatClient.CreateRoom(ctx, roomID, userA, userB); err != nil {
		log.Printf("MATCHMAKING: failed to create chat room %s: %v", roomID, err)
		_ = u.repo.DeleteRoom(ctx, userA)
		return
	}

	// Publish to MQ for asynchronous processing (notifications, etc.)
	if u.mq != nil {
		_ = u.mq.Publish(ctx, "match.found", room)
	}

	log.Printf("MATCHMAKING: Match found! %s <-> %s (mode=%s)", userA, userB, mode)

	_ = u.repo.PublishMatch(ctx, userA, &domain.MatchFound{
		RoomID:        roomID,
		Mode:          mode,
		IsInitiator:   true,
		PartnerGender: gB,
	})
	_ = u.repo.PublishMatch(ctx, userB, &domain.MatchFound{
		RoomID:        roomID,
		Mode:          mode,
		IsInitiator:   false,
		PartnerGender: gA,
	})
}

func (u *matchUsecase) Cancel(ctx context.Context, userID string) error {
	return u.repo.RemoveFromQueue(ctx, userID)
}

func (u *matchUsecase) GetStatus(ctx context.Context, userID string) (*domain.Room, error) {
	return u.repo.GetRoom(ctx, userID)
}

func (u *matchUsecase) Next(ctx context.Context, userID string) error {
	return u.repo.RemoveFromQueue(ctx, userID) // просто убираем из очереди
}
func (u *matchUsecase) SubscribeToMatch(ctx context.Context, userID string) (<-chan *domain.MatchFound, func(), error) {
	return u.repo.SubscribeToMatch(ctx, userID)
}

func (u *matchUsecase) HealthCheck(ctx context.Context) error {
	return u.repo.HealthCheck(ctx)
}
