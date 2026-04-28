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

func (r *RedisMatchRepository) getQueueKey(mode, gender string) string {
	return "queue:" + mode + ":" + gender
}

func (r *RedisMatchRepository) AddToQueue(ctx context.Context, entry *domain.QueueEntry) error {
	score := float64(entry.JoinedAt.Unix())
	key := r.getQueueKey(entry.Filter.Mode, entry.Filter.MyGender)
	
	if err := r.rdb.ZAdd(ctx, key, goredis.Z{Score: score, Member: entry.UserID}).Err(); err != nil {
		return err
	}

	interestsJSON, _ := json.Marshal(entry.Filter.Interests)
	filterKey := filtersKeyPref + entry.UserID
	if err := r.rdb.HSet(ctx, filterKey,
		"my_gender", entry.Filter.MyGender,
		"gender", entry.Filter.Gender,
		"interests", string(interestsJSON),
		"mode", entry.Filter.Mode,
	).Err(); err != nil {
		return err
	}
	return r.rdb.Expire(ctx, filterKey, r.filterTTL).Err()
}

func (r *RedisMatchRepository) RemoveFromQueue(ctx context.Context, userID string) error {
	filter, err := r.getFilter(ctx, userID)
	if err == nil && filter.Mode != "" {
		key := r.getQueueKey(filter.Mode, filter.MyGender)
		r.rdb.ZRem(ctx, key, userID)
	}
	// Also try global/fallback if exists (for migration or safety)
	r.rdb.ZRem(ctx, queueKey, userID) 
	
	return r.rdb.Del(ctx, filtersKeyPref+userID).Err()
}

func (r *RedisMatchRepository) GetQueue(ctx context.Context) ([]*domain.QueueEntry, error) {
	// Find all keys like queue:text:male, queue:voice:female, etc.
	keys, err := r.rdb.Keys(ctx, "queue:*:*").Result()
	if err != nil {
		return nil, err
	}

	var allItems []goredis.Z
	for _, key := range keys {
		items, _ := r.rdb.ZRangeWithScores(ctx, key, 0, -1).Result()
		allItems = append(allItems, items...)
	}

	if len(allItems) == 0 {
		return nil, nil
	}

	pipe := r.rdb.Pipeline()
	cmds := make([]*goredis.MapStringStringCmd, len(allItems))
	for i, it := range allItems {
		userID := it.Member.(string)
		cmds[i] = pipe.HGetAll(ctx, filtersKeyPref+userID)
	}

	_, _ = pipe.Exec(ctx)

	out := make([]*domain.QueueEntry, 0, len(allItems))
	for i, it := range allItems {
		userID := it.Member.(string)
		m, err := cmds[i].Result()
		
		filter := domain.Filter{Gender: "any", Mode: "text"}
		if err == nil && len(m) > 0 {
			if s := m["interests"]; s != "" {
				_ = json.Unmarshal([]byte(s), &filter.Interests)
			}
			if g := m["gender"]; g != "" {
				filter.Gender = g
			}
			filter.MyGender = m["my_gender"]
			filter.Mode = m["mode"]
		}

		out = append(out, &domain.QueueEntry{
			UserID:   userID,
			Filter:   filter,
			JoinedAt: time.Unix(int64(it.Score), 0),
		})
	}
	return out, nil
}

func (r *RedisMatchRepository) CreateRoom(ctx context.Context, room *domain.Room) error {
	roomKey := roomKeyPref + room.ID
	userAKey := userRoomPref + room.UserA
	userBKey := userRoomPref + room.UserB
	createdAtUnix := room.CreatedAt.Unix()
	ttlSec := int(r.roomTTL.Seconds())

	// Lua script to atomically:
	// 1. Check if users are already matched
	// 2. Create room HSET
	// 3. Remove users from mode-specific gender buckets
	script := `
		if redis.call("EXISTS", KEYS[1]) == 1 or redis.call("EXISTS", KEYS[2]) == 1 then
			return 0
		end
		redis.call("HSET", KEYS[3], "user_a", ARGV[1], "user_b", ARGV[2], "mode", ARGV[3], "created_at", ARGV[4])
		redis.call("EXPIRE", KEYS[3], ARGV[5])
		redis.call("SET", KEYS[1], ARGV[6], "EX", ARGV[5])
		redis.call("SET", KEYS[2], ARGV[6], "EX", ARGV[5])
		
		-- Remove from all possible gender buckets for this mode
		redis.call("ZREM", "queue:" .. ARGV[3] .. ":male", ARGV[1], ARGV[2])
		redis.call("ZREM", "queue:" .. ARGV[3] .. ":female", ARGV[1], ARGV[2])
		
		redis.call("DEL", ARGV[7] .. ARGV[1], ARGV[7] .. ARGV[2])
		return 1
	`
	
	res, err := r.rdb.Eval(ctx, script, 
		[]string{userAKey, userBKey, roomKey}, 
		room.UserA, room.UserB, room.Mode, strconv.FormatInt(createdAtUnix, 10), 
		ttlSec, room.ID, filtersKeyPref,
	).Int()

	if err != nil {
		return err
	}
	if res == 0 {
		return errors.New("one or both users already matched")
	}
	return nil
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
		Mode:      m["mode"],
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
	return domain.Filter{
		MyGender:  m["my_gender"],
		Gender:    g,
		Interests: interests,
		Mode:      m["mode"],
	}, nil
}
 
func (r *RedisMatchRepository) PublishMatch(ctx context.Context, userID string, match *domain.MatchFound) error {
	b, _ := json.Marshal(match)
	return r.rdb.Publish(ctx, "match:"+userID, b).Err()
}
 
func (r *RedisMatchRepository) SubscribeToMatch(ctx context.Context, userID string) (<-chan *domain.MatchFound, func(), error) {
	pubsub := r.rdb.Subscribe(ctx, "match:"+userID)
	ch := make(chan *domain.MatchFound)

	go func() {
		defer close(ch)
		for msg := range pubsub.Channel() {
			var m domain.MatchFound
			if err := json.Unmarshal([]byte(msg.Payload), &m); err == nil {
				ch <- &m
			}
		}
	}()

	cleanup := func() {
		pubsub.Close()
	}

	return ch, cleanup, nil
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


func (r *RedisMatchRepository) HealthCheck(ctx context.Context) error {
	return r.rdb.Ping(ctx).Err()
}
