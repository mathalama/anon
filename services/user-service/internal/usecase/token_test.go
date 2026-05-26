package usecase

import (
	"testing"
	"time"
)

func TestTokenManager_GeneratePair(t *testing.T) {
	tm := NewTokenManager("test-secret", 15*time.Minute, 24*time.Hour)
	access, refresh, err := tm.GeneratePair("user-123")
	if err != nil {
		t.Fatalf("GeneratePair failed: %v", err)
	}
	if access == "" || refresh == "" {
		t.Error("expected non-empty tokens")
	}
	if access == refresh {
		t.Error("access and refresh tokens should be different")
	}
}

func TestTokenManager_ValidateAccess(t *testing.T) {
	tm := NewTokenManager("test-secret", 15*time.Minute, 24*time.Hour)
	access, _, err := tm.GeneratePair("user-456")
	if err != nil {
		t.Fatalf("GeneratePair failed: %v", err)
	}

	userID, err := tm.ValidateAndGetSubject(access, "access")
	if err != nil {
		t.Fatalf("ValidateAndGetSubject failed: %v", err)
	}
	if userID != "user-456" {
		t.Errorf("expected user-456, got %s", userID)
	}
}

func TestTokenManager_ValidateRefresh(t *testing.T) {
	tm := NewTokenManager("test-secret", 15*time.Minute, 24*time.Hour)
	_, refresh, err := tm.GeneratePair("user-789")
	if err != nil {
		t.Fatalf("GeneratePair failed: %v", err)
	}

	userID, err := tm.ValidateAndGetSubject(refresh, "refresh")
	if err != nil {
		t.Fatalf("ValidateAndGetSubject failed: %v", err)
	}
	if userID != "user-789" {
		t.Errorf("expected user-789, got %s", userID)
	}
}

func TestTokenManager_WrongType(t *testing.T) {
	tm := NewTokenManager("test-secret", 15*time.Minute, 24*time.Hour)
	access, _, err := tm.GeneratePair("user-wrong")
	if err != nil {
		t.Fatalf("GeneratePair failed: %v", err)
	}

	// Try to validate access token as refresh — should fail
	_, err = tm.ValidateAndGetSubject(access, "refresh")
	if err == nil {
		t.Error("expected error when validating access token as refresh")
	}
}

func TestTokenManager_InvalidSignature(t *testing.T) {
	tm1 := NewTokenManager("secret-1", 15*time.Minute, 24*time.Hour)
	tm2 := NewTokenManager("secret-2", 15*time.Minute, 24*time.Hour)

	access, _, err := tm1.GeneratePair("user-sig")
	if err != nil {
		t.Fatalf("GeneratePair failed: %v", err)
	}

	// Token signed with secret-1 should not validate with secret-2
	_, err = tm2.ValidateAndGetSubject(access, "access")
	if err == nil {
		t.Error("expected error for wrong signing key")
	}
}

func TestTokenManager_ExpiredToken(t *testing.T) {
	// Create token manager with 0 TTL (immediately expired)
	tm := NewTokenManager("test-secret", 0, 0)
	access, _, err := tm.GeneratePair("user-expired")
	if err != nil {
		t.Fatalf("GeneratePair failed: %v", err)
	}

	// Wait a moment to ensure it's expired
	time.Sleep(1 * time.Second)

	_, err = tm.ValidateAndGetSubject(access, "access")
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestTokenManager_GarbageToken(t *testing.T) {
	tm := NewTokenManager("test-secret", 15*time.Minute, 24*time.Hour)
	_, err := tm.ValidateAndGetSubject("not.a.valid.jwt", "access")
	if err == nil {
		t.Error("expected error for garbage token")
	}
}
