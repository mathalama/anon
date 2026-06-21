package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	delivhttp "github.com/mathalama/nektokz/user-service/internal/delivery/http"
	"github.com/mathalama/nektokz/user-service/internal/domain"
)

type mockUserUsecase struct {
	CreateAnonymousFunc func(ctx context.Context, deviceID string) (string, string, error)
	GetMeFunc           func(ctx context.Context, userID string) (*domain.User, error)
}

func (m *mockUserUsecase) CreateAnonymous(ctx context.Context, deviceID string) (string, string, error) {
	return m.CreateAnonymousFunc(ctx, deviceID)
}

func (m *mockUserUsecase) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	return "", "", nil
}

func (m *mockUserUsecase) GetMe(ctx context.Context, userID string) (*domain.User, error) {
	return m.GetMeFunc(ctx, userID)
}

func (m *mockUserUsecase) UpdateMe(ctx context.Context, userID string, gender string, interests []string) error {
	return nil
}

func (m *mockUserUsecase) GetBanStatus(ctx context.Context, userID string) (*domain.Ban, bool, error) {
	return nil, false, nil
}

func (m *mockUserUsecase) BanUser(ctx context.Context, userID, reason, bannedBy string, duration time.Duration) error {
	return nil
}

func TestUserHandler_CreateAnonymous(t *testing.T) {
	usecase := &mockUserUsecase{
		CreateAnonymousFunc: func(ctx context.Context, deviceID string) (string, string, error) {
			if deviceID == "test-dev-1" {
				return "access123", "refresh123", nil
			}
			return "", "", domain.ErrUserNotFound
		},
	}

	r := chi.NewRouter()
	delivhttp.NewUserHandler(r, usecase, "internal-token")

	reqBody := `{"device_id":"test-dev-1"}`
	req := httptest.NewRequest(http.MethodPost, "/users/anonymous", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["access_token"] != "access123" {
		t.Errorf("expected access_token access123, got %s", resp["access_token"])
	}
}

func TestUserHandler_GetMe(t *testing.T) {
	usecase := &mockUserUsecase{
		GetMeFunc: func(ctx context.Context, userID string) (*domain.User, error) {
			if userID == "user1" {
				return &domain.User{ID: "user1", Gender: "male"}, nil
			}
			return nil, domain.ErrUserNotFound
		},
	}

	r := chi.NewRouter()
	delivhttp.NewUserHandler(r, usecase, "internal-token")

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req.Header.Set("X-User-ID", "user1")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var user domain.User
	json.NewDecoder(w.Body).Decode(&user)

	if user.ID != "user1" || user.Gender != "male" {
		t.Errorf("unexpected user data: %+v", user)
	}
}
