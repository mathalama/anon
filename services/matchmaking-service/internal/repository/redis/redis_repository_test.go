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

func TestRedisMatchRepository_GetQueue_LazyDeletion(t *testing.T) {
	s := miniredis.RunT(t)
	rdb := goredis.NewClient(&goredis.Options{
		Addr: s.Addr(),
	})
	
	repo := &RedisMatchRepository{
		rdb:       rdb,
		filterTTL: 1 * time.Hour,
	}

	ctx := context.Background()

	// Add user1 (active)
	err := repo.AddToQueue(ctx, &domain.QueueEntry{
		UserID: "user1",
		Filter: domain.Filter{
			MyGender: "male",
			Gender:   "female",
			Mode:     "text",
		},
		JoinedAt: time.Now(),
	})
	assert.NoError(t, err)

	// Add user2 (will become ghost)
	err = repo.AddToQueue(ctx, &domain.QueueEntry{
		UserID: "user2",
		Filter: domain.Filter{
			MyGender: "female",
			Gender:   "male",
			Mode:     "text",
		},
		JoinedAt: time.Now(),
	})
	assert.NoError(t, err)

	// Simulate expiration of user2's filter
	rdb.Del(ctx, filtersKeyPref+"user2")

	// Call GetQueue
	queue, err := repo.GetQueue(ctx)
	assert.NoError(t, err)

	// Verify only user1 is in the queue output
	assert.Len(t, queue, 1)
	assert.Equal(t, "user1", queue[0].UserID)

	// Verify user2 is removed from the Redis ZSET
	exists, _ := rdb.ZScore(ctx, "queue:text:female:male", "user2").Result()
	assert.Zero(t, exists)
}

func TestRedisMatchRepository_CreateRoom_DynamicZREM(t *testing.T) {
	s := miniredis.RunT(t)
	rdb := goredis.NewClient(&goredis.Options{
		Addr: s.Addr(),
	})
	
	repo := &RedisMatchRepository{
		rdb:     rdb,
		roomTTL: 1 * time.Hour,
	}

	ctx := context.Background()

	// Queue user1 (male searching female)
	err := repo.AddToQueue(ctx, &domain.QueueEntry{
		UserID: "user1",
		Filter: domain.Filter{
			MyGender: "male",
			Gender:   "female",
			Mode:     "text",
		},
		JoinedAt: time.Now(),
	})
	assert.NoError(t, err)

	// Queue user2 (female searching male)
	err = repo.AddToQueue(ctx, &domain.QueueEntry{
		UserID: "user2",
		Filter: domain.Filter{
			MyGender: "female",
			Gender:   "male",
			Mode:     "text",
		},
		JoinedAt: time.Now(),
	})
	assert.NoError(t, err)

	// Verify they are in their queues
	score1, err := rdb.ZScore(ctx, "queue:text:male:female", "user1").Result()
	assert.NoError(t, err)
	assert.True(t, score1 > 0)

	score2, err := rdb.ZScore(ctx, "queue:text:female:male", "user2").Result()
	assert.NoError(t, err)
	assert.True(t, score2 > 0)

	// Match them (CreateRoom)
	room := &domain.Room{
		ID:        "room123",
		UserA:     "user1",
		UserB:     "user2",
		Mode:      "text",
		CreatedAt: time.Now(),
	}
	err = repo.CreateRoom(ctx, room)
	assert.NoError(t, err)

	// Verify they were dynamically ZREMed from their specific queues
	exists1, _ := rdb.ZScore(ctx, "queue:text:male:female", "user1").Result()
	assert.Zero(t, exists1)

	exists2, _ := rdb.ZScore(ctx, "queue:text:female:male", "user2").Result()
	assert.Zero(t, exists2)
}
