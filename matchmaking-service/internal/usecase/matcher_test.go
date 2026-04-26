package usecase

import (
	"testing"
	"github.com/mathalama/nektokz/matchmaking-service/internal/domain"
)

func TestIsCompatible(t *testing.T) {
	tests := []struct {
		name     string
		f1       domain.Filter
		f2       domain.Filter
		want     bool
	}{
		{
			name: "Any gender match",
			f1:   domain.Filter{Gender: "any"},
			f2:   domain.Filter{Gender: "male"},
			want: true,
		},
		{
			name: "Specific gender match",
			f1:   domain.Filter{Gender: "male"},
			f2:   domain.Filter{Gender: "male"},
			want: true,
		},
		{
			name: "Specific gender mismatch",
			f1:   domain.Filter{Gender: "male"},
			f2:   domain.Filter{Gender: "female"},
			want: false,
		},
		{
			name: "Interest match",
			f1:   domain.Filter{Gender: "any", Interests: []string{"tech"}},
			f2:   domain.Filter{Gender: "any", Interests: []string{"tech", "music"}},
			want: true,
		},
		{
			name: "Interest mismatch",
			f1:   domain.Filter{Gender: "any", Interests: []string{"sports"}},
			f2:   domain.Filter{Gender: "any", Interests: []string{"tech"}},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsCompatible(tt.f1, tt.f2); got != tt.want {
				t.Errorf("IsCompatible() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatch(t *testing.T) {
	target := &domain.QueueEntry{
		UserID: "user1",
		Filter: domain.Filter{Gender: "male", Interests: []string{"tech"}},
	}
	candidates := []*domain.QueueEntry{
		{UserID: "user2", Filter: domain.Filter{Gender: "female", Interests: []string{"music"}}},
		{UserID: "user3", Filter: domain.Filter{Gender: "male", Interests: []string{"tech"}}},
	}

	got := Match(target, candidates)
	if got == nil || got.UserID != "user3" {
		t.Errorf("Match() = %v, want user3", got)
	}
}
