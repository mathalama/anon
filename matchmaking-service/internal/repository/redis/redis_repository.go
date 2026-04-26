package redis

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/mathalama/nektokz/matchmaking-service/internal/domain"
)

const (
	queueKey       = "queue:searching"
	filtersKeyPref = "queue:filters:"
	roomKeyPref    = "room:"
	userRoomPref   = "user:room:"
)

type RedisMatchRepository struct {
	rdb        *goredis.Client
	filterTTL  time.Duration
	roomTTL    time.Duration
}

func NewRedisMatchRepository(redisURL string) (*RedisMatchRepository, error) {
	opts, err := parseRedisOptions(redisURL)
	if err != nil {
		return nil, err
	}
	rdb := goredis.NewClient(opts)
	return &RedisMatchRepository{
		rdb:       rdb,
		filterTTL: 60 * time.Second,
		roomTTL:   3 * time.Hour,
	}, nil
}

func (r *RedisMatchRepository) Close() error {
	return r.rdb.Close()
}

func (r *RedisMatchRepository) AddToQueue(ctx context.Context, entry *domain.QueueEntry) error {
	score := float64(entry.JoinedAt.Unix())
	if err := r.rdb.ZAdd(ctx, queueKey, goredis.Z{Score: score, Member: entry.UserID}).Err(); err != nil {
		return err
	}

	interestsJSON, _ := json.Marshal(entry.Filter.Interests)
	filterKey := filtersKeyPref + entry.UserID
	if err := r.rdb.HSet(ctx, filterKey,
		"gender", entry.Filter.Gender,
		"interests", string(interestsJSON),
	).Err(); err != nil {
		return err
	}
	return r.rdb.Expire(ctx, filterKey, r.filterTTL).Err()
}

func (r *RedisMatchRepository) RemoveFromQueue(ctx context.Context, userID string) error {
	if err := r.rdb.ZRem(ctx, queueKey, userID).Err(); err != nil {
		return err
	}
	return r.rdb.Del(ctx, filtersKeyPref+userID).Err()
}

func (r *RedisMatchRepository) GetQueue(ctx context.Context) ([]*domain.QueueEntry, error) {
	items, err := r.rdb.ZRangeWithScores(ctx, queueKey, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	out := make([]*domain.QueueEntry, 0, len(items))
	for _, it := range items {
		userID, ok := it.Member.(string)
		if !ok {
			continue
		}
		filter, _ := r.getFilter(ctx, userID)
		out = append(out, &domain.QueueEntry{
			UserID: userID,
			Filter: filter,
			JoinedAt: time.Unix(int64(it.Score), 0),
		})
	}
	return out, nil
}

func (r *RedisMatchRepository) CreateRoom(ctx context.Context, room *domain.Room) error {
	roomKey := roomKeyPref + room.ID
	createdAtUnix := room.CreatedAt.Unix()

	if err := r.rdb.HSet(ctx, roomKey,
		"user_a", room.UserA,
		"user_b", room.UserB,
		"created_at", strconv.FormatInt(createdAtUnix, 10),
	).Err(); err != nil {
		return err
	}
	if err := r.rdb.Expire(ctx, roomKey, r.roomTTL).Err(); err != nil {
		return err
	}

	if err := r.rdb.Set(ctx, userRoomPref+room.UserA, room.ID, r.roomTTL).Err(); err != nil {
		return err
	}
	return r.rdb.Set(ctx, userRoomPref+room.UserB, room.ID, r.roomTTL).Err()
}

func (r *RedisMatchRepository) GetRoom(ctx context.Context, userID string) (*domain.Room, error) {
	roomID, err := r.rdb.Get(ctx, userRoomPref+userID).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	m, err := r.rdb.HGetAll(ctx, roomKeyPref+roomID).Result()
	if err != nil {
		return nil, err
	}
	if len(m) == 0 {
		return nil, nil
	}

	createdAtUnix, _ := strconv.ParseInt(m["created_at"], 10, 64)
	return &domain.Room{
		ID:        roomID,
		UserA:     m["user_a"],
		UserB:     m["user_b"],
		CreatedAt: time.Unix(createdAtUnix, 0),
	}, nil
}

func (r *RedisMatchRepository) getFilter(ctx context.Context, userID string) (domain.Filter, error) {
	m, err := r.rdb.HGetAll(ctx, filtersKeyPref+userID).Result()
	if err != nil {
		return domain.Filter{}, err
	}
	if len(m) == 0 {
		return domain.Filter{Gender: "any"}, nil
	}
	var interests []string
	if s := m["interests"]; s != "" {
		_ = json.Unmarshal([]byte(s), &interests)
	}
	g := m["gender"]
	if g == "" {
		g = "any"
	}
	return domain.Filter{Gender: g, Interests: interests}, nil
}

func parseRedisOptions(redisURL string) (*goredis.Options, error) {
	if strings.HasPrefix(redisURL, "redis://") || strings.HasPrefix(redisURL, "rediss://") {
		return goredis.ParseURL(redisURL)
	}
	if redisURL == "" {
		return nil, errors.New("REDIS_URL is empty")
	}
	return &goredis.Options{Addr: redisURL}, nil
}

