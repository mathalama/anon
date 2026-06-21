package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/mathalama/nektokz/user-service/internal/domain"
	goredis "github.com/redis/go-redis/v9"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	TotalRegisteredUsers = promauto.NewCounter(prometheus.CounterOpts{
		Name: "nektokz_users_registered_total",
		Help: "Total number of registered users",
	})
)

type userUsecase struct {
	repo         domain.UserRepository
	tokenManager *TokenManager
	rdb          *goredis.Client
}

func NewUserUsecase(repo domain.UserRepository, tm *TokenManager, rdb *goredis.Client) domain.UserUsecase {
	return &userUsecase{
		repo:         repo,
		tokenManager: tm,
		rdb:          rdb,
	}
}

func (u *userUsecase) CreateAnonymous(ctx context.Context, deviceID string) (string, string, error) {
	user, err := u.repo.GetByDeviceID(ctx, deviceID)
	if err != nil {
		if !errors.Is(err, domain.ErrUserNotFound) {
			return "", "", err
		}
		user = &domain.User{
			ID:          uuid.New().String(),
			DeviceID:    deviceID,
			IsAnonymous: true,
			CreatedAt:   time.Now(),
		}
		if err := u.repo.Create(ctx, user); err != nil {
			return "", "", err
		}
		TotalRegisteredUsers.Inc()
	}

	return u.tokenManager.GeneratePair(user.ID)
}

func (u *userUsecase) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	userID, err := u.tokenManager.ValidateAndGetSubject(refreshToken, "refresh")
	if err != nil {
		return "", "", err
	}
	return u.tokenManager.GeneratePair(userID)
}

func (u *userUsecase) GetMe(ctx context.Context, userID string) (*domain.User, error) {
	if u.rdb != nil {
		key := "user:" + userID
		b, err := u.rdb.Get(ctx, key).Bytes()
		if err == nil {
			var user domain.User
			if err := json.Unmarshal(b, &user); err == nil {
				return &user, nil
			}
		} else if err != goredis.Nil {
			// Cache errors should not fail the request
		}
	}

	user, err := u.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if u.rdb != nil {
		key := "user:" + userID
		if b, err := json.Marshal(user); err == nil {
			_ = u.rdb.Set(ctx, key, b, 5*time.Minute).Err()
		}
	}

	return user, nil
}

func (u *userUsecase) UpdateMe(ctx context.Context, userID string, gender string, interests []string) error {
	user, err := u.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	user.Gender = gender
	user.Interests = interests

	if err := u.repo.Update(ctx, user); err != nil {
		return err
	}

	if u.rdb != nil {
		_ = u.rdb.Del(ctx, "user:"+userID).Err()
	}
	return nil
}

func (u *userUsecase) GetBanStatus(ctx context.Context, userID string) (*domain.Ban, bool, error) {
	ban, err := u.repo.GetActiveBan(ctx, userID)
	if err != nil {
		return nil, false, err
	}
	if ban == nil {
		return nil, false, nil
	}
	return ban, true, nil
}

func (u *userUsecase) BanUser(ctx context.Context, userID, reason, bannedBy string, duration time.Duration) error {
	ban := &domain.Ban{
		ID:        uuid.New().String(),
		UserID:    userID,
		Reason:    reason,
		BannedBy:  bannedBy,
		ExpiresAt: time.Now().Add(duration),
		CreatedAt: time.Now(),
	}
	return u.repo.CreateBan(ctx, ban)
}
