package redis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/mathalama/nektokz/matchmaking-service/internal/domain"
)

func TestRedisMatchRepository_CreateRoom_Atomicity(t *testing.T) {
	s := miniredis.RunT(t)
	rdb := goredis.NewClient(&goredis.Options{
		Addr: s.Addr(),
	})
	
	repo := &RedisMatchRepository{
		rdb:      rdb,
		roomTTL:  1 * time.Hour,
	}

	ctx := context.Background()
	room := &domain.Room{
		ID:        "room1",
		UserA:     "user1",
		UserB:     "user2",
		Mode:      "text",
		CreatedAt: time.Now(),
	}

	// 1. First creation should succeed
	err := repo.CreateRoom(ctx, room)
	assert.NoError(t, err)

	// 2. Second creation with same users should fail (atomic check)
	err = repo.CreateRoom(ctx, room)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already matched")

	// 3. User mapping should exist
	roomID, _ := rdb.Get(ctx, userRoomPref+"user1").Result()
	assert.Equal(t, "room1", roomID)
	
	// 4. Users should be removed from queue
	exists, _ := rdb.ZScore(ctx, queueKey, "user1").Result()
	assert.Zero(t, exists)
}
