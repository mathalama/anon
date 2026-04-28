package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/mathalama/nektokz/user-service/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type userUsecase struct {
	repo         domain.UserRepository
	tokenManager *TokenManager
}

func NewUserUsecase(repo domain.UserRepository, tm *TokenManager) domain.UserUsecase {
	return &userUsecase{
		repo:         repo,
		tokenManager: tm,
	}
}

func (u *userUsecase) CreateAnonymous(ctx context.Context, deviceID string) (string, string, error) {
	user, err := u.repo.GetByDeviceID(ctx, deviceID)
	if err != nil {
		if err.Error() != "user not found" {
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
	}

	return u.tokenManager.GeneratePair(user.ID)
}

func (u *userUsecase) Register(ctx context.Context, email, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &domain.User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: string(hash),
		IsAnonymous:  false,
		CreatedAt:    time.Now(),
	}

	return u.repo.Create(ctx, user)
}

func (u *userUsecase) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := u.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", "", errors.New("invalid credentials")
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
	return u.repo.GetByID(ctx, userID)
}

func (u *userUsecase) UpdateMe(ctx context.Context, userID string, gender string, interests []string) error {
	user, err := u.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	user.Gender = gender
	user.Interests = interests

	return u.repo.Update(ctx, user)
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
