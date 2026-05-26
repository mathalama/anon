package usecase

import (
	"context"
	"testing"
)

func TestModerateMessage_NoToxicWords(t *testing.T) {
	uc := New(nil, nil, nil, nil, nil)
	toxic, err := uc.ModerateMessage(context.Background(), "hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if toxic {
		t.Error("expected non-toxic message")
	}
}

func TestModerateMessage_EmptyToxicList(t *testing.T) {
	uc := New(nil, nil, nil, nil, []string{})
	toxic, err := uc.ModerateMessage(context.Background(), "hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if toxic {
		t.Error("expected non-toxic message with empty list")
	}
}

func TestModerateMessage_ToxicWordDetected(t *testing.T) {
	uc := New(nil, nil, nil, nil, []string{"bad", "toxic"})
	toxic, err := uc.ModerateMessage(context.Background(), "this is a bad message")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !toxic {
		t.Error("expected toxic message to be detected")
	}
}

func TestModerateMessage_CaseInsensitive(t *testing.T) {
	uc := New(nil, nil, nil, nil, []string{"bad"})
	toxic, err := uc.ModerateMessage(context.Background(), "This is BAD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !toxic {
		t.Error("expected case-insensitive toxic detection")
	}
}

func TestModerateMessage_ObfuscatedWord(t *testing.T) {
	uc := New(nil, nil, nil, nil, []string{"fuck"})

	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{"plain", "fuck you", true},
		{"with dots", "f.u.c.k", true},
		{"with spaces", "f u c k", true},
		{"with underscores", "f_u_c_k", true},
		{"with dashes", "f-u-c-k", true},
		{"clean message", "hello friend", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toxic, err := uc.ModerateMessage(context.Background(), tt.content)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if toxic != tt.want {
				t.Errorf("content=%q: got toxic=%v, want %v", tt.content, toxic, tt.want)
			}
		})
	}
}

func TestModerateMessage_CyrillicToxicWords(t *testing.T) {
	uc := New(nil, nil, nil, nil, []string{"сука", "блять"})

	tests := []struct {
		content string
		want    bool
	}{
		{"ты сука", true},
		{"привет", false},
		{"БЛЯТЬ", true},
	}

	for _, tt := range tests {
		t.Run(tt.content, func(t *testing.T) {
			toxic, err := uc.ModerateMessage(context.Background(), tt.content)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if toxic != tt.want {
				t.Errorf("got toxic=%v, want %v", toxic, tt.want)
			}
		})
	}
}

func TestModerateMessage_EmptyContent(t *testing.T) {
	uc := New(nil, nil, nil, nil, []string{"bad"})
	toxic, err := uc.ModerateMessage(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if toxic {
		t.Error("empty content should not be toxic")
	}
}
