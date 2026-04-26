package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/mathalama/nektokz/user-service/internal/domain"
	"golang.org/x/crypto/bcrypt"
	"fmt"
)

type userUsecase struct {
	repo         domain.UserRepository
	tokenManager *TokenManager
	botToken     string
}

func NewUserUsecase(repo domain.UserRepository, tm *TokenManager, botToken string) domain.UserUsecase {
	return &userUsecase{
		repo:         repo,
		tokenManager: tm,
		botToken:     botToken,
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

func (u *userUsecase) LoginTelegram(ctx context.Context, data map[string]string) (string, string, error) {
	if err := VerifyTelegramHash(data, u.botToken); err != nil {
		return "", "", errors.New("invalid telegram signature")
	}

	tgIDStr := data["id"]
	var tgID int64
	fmt.Sscanf(tgIDStr, "%d", &tgID)

	user, err := u.repo.GetByTelegramID(ctx, tgID)
	if err != nil {
		if err != nil && err.Error() != "user not found" {
			return "", "", err
		}

		// Create new user linked to Telegram
		user = &domain.User{
			ID:          uuid.New().String(),
			TelegramID:  tgID,
			FirstName:   data["first_name"],
			LastName:    data["last_name"],
			Username:    data["username"],
			PhotoURL:    data["photo_url"],
			IsAnonymous: false,
			CreatedAt:   time.Now(),
		}
		if err := u.repo.Create(ctx, user); err != nil {
			return "", "", err
		}
	} else {
		// Update user info from Telegram
		user.FirstName = data["first_name"]
		user.LastName = data["last_name"]
		user.Username = data["username"]
		user.PhotoURL = data["photo_url"]
		u.repo.Update(ctx, user)
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
