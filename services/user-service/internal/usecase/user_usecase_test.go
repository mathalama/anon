package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mathalama/nektokz/user-service/internal/domain"
	"github.com/mathalama/nektokz/user-service/internal/repository/postgres"
)

func newTestUsecase() domain.UserUsecase {
	repo := postgres.NewInMemoryUserRepository()
	tm := NewTokenManager("test-secret-key-for-unit-tests", 15*time.Minute, 24*time.Hour)
	return NewUserUsecase(repo, tm, nil)
}

func TestCreateAnonymous_NewUser(t *testing.T) {
	uc := newTestUsecase()
	access, refresh, err := uc.CreateAnonymous(context.Background(), "device-123")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if access == "" {
		t.Error("expected non-empty access token")
	}
	if refresh == "" {
		t.Error("expected non-empty refresh token")
	}
}

func TestCreateAnonymous_ExistingUser(t *testing.T) {
	uc := newTestUsecase()

	// First call creates the user
	access1, _, err := uc.CreateAnonymous(context.Background(), "device-456")
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	// Second call should find the existing user
	access2, _, err := uc.CreateAnonymous(context.Background(), "device-456")
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}

	// Both should produce valid tokens (different because of iat)
	if access1 == "" || access2 == "" {
		t.Error("expected non-empty tokens")
	}
}

func TestRefresh_Valid(t *testing.T) {
	uc := newTestUsecase()
	_, refresh, err := uc.CreateAnonymous(context.Background(), "device-789")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	newAccess, newRefresh, err := uc.Refresh(context.Background(), refresh)
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	if newAccess == "" || newRefresh == "" {
		t.Error("expected non-empty tokens from refresh")
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	uc := newTestUsecase()
	_, _, err := uc.Refresh(context.Background(), "invalid-token")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestRefresh_AccessTokenAsRefresh(t *testing.T) {
	uc := newTestUsecase()
	access, _, err := uc.CreateAnonymous(context.Background(), "device-xxx")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// Using access token as refresh should fail (wrong type)
	_, _, err = uc.Refresh(context.Background(), access)
	if err == nil {
		t.Error("expected error when using access token as refresh")
	}
}

func TestGetMe_ExistingUser(t *testing.T) {
	uc := newTestUsecase()
	_, _, err := uc.CreateAnonymous(context.Background(), "device-getme")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// We need the user ID — extract from token
	tm := NewTokenManager("test-secret-key-for-unit-tests", 15*time.Minute, 24*time.Hour)
	access, _, _ := uc.CreateAnonymous(context.Background(), "device-getme")
	userID, err := tm.ValidateAndGetSubject(access, "access")
	if err != nil {
		t.Fatalf("failed to extract user id: %v", err)
	}

	user, err := uc.GetMe(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetMe failed: %v", err)
	}
	if user.ID != userID {
		t.Errorf("expected user ID %s, got %s", userID, user.ID)
	}
	if user.DeviceID != "device-getme" {
		t.Errorf("expected device_id 'device-getme', got '%s'", user.DeviceID)
	}
}

func TestGetMe_NonExistentUser(t *testing.T) {
	uc := newTestUsecase()
	_, err := uc.GetMe(context.Background(), "non-existent-id")
	if err == nil {
		t.Error("expected error for non-existent user")
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got: %v", err)
	}
}

func TestUpdateMe(t *testing.T) {
	uc := newTestUsecase()
	access, _, err := uc.CreateAnonymous(context.Background(), "device-update")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	tm := NewTokenManager("test-secret-key-for-unit-tests", 15*time.Minute, 24*time.Hour)
	userID, _ := tm.ValidateAndGetSubject(access, "access")

	err = uc.UpdateMe(context.Background(), userID, "male", []string{"music", "gaming"})
	if err != nil {
		t.Fatalf("UpdateMe failed: %v", err)
	}

	user, err := uc.GetMe(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetMe after update failed: %v", err)
	}
	if user.Gender != "male" {
		t.Errorf("expected gender 'male', got '%s'", user.Gender)
	}
	if len(user.Interests) != 2 {
		t.Errorf("expected 2 interests, got %d", len(user.Interests))
	}
}

func TestBanUser_And_GetBanStatus(t *testing.T) {
	uc := newTestUsecase()
	access, _, err := uc.CreateAnonymous(context.Background(), "device-ban")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	tm := NewTokenManager("test-secret-key-for-unit-tests", 15*time.Minute, 24*time.Hour)
	userID, _ := tm.ValidateAndGetSubject(access, "access")

	// No ban initially
	_, isBanned, err := uc.GetBanStatus(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetBanStatus failed: %v", err)
	}
	if isBanned {
		t.Error("expected user not to be banned initially")
	}

	// Ban the user
	err = uc.BanUser(context.Background(), userID, "test ban", "system", 24*time.Hour)
	if err != nil {
		t.Fatalf("BanUser failed: %v", err)
	}

	// Should be banned now
	ban, isBanned, err := uc.GetBanStatus(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetBanStatus after ban failed: %v", err)
	}
	if !isBanned {
		t.Error("expected user to be banned")
	}
	if ban.Reason != "test ban" {
		t.Errorf("expected reason 'test ban', got '%s'", ban.Reason)
	}
}
