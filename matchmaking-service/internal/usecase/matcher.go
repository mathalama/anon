package usecase

import (
	"github.com/mathalama/nektokz/matchmaking-service/internal/domain"
)

// Match attempts to find a partner for a target user from a list of candidates.
// It returns the partner if found, otherwise nil.
func Match(target *domain.QueueEntry, candidates []*domain.QueueEntry) *domain.QueueEntry {
	for _, candidate := range candidates {
		if candidate.UserID == target.UserID {
			continue
		}

		if IsCompatible(target.Filter, candidate.Filter) {
			return candidate
		}
	}
	return nil
}

// IsCompatible checks if two filters match.
// Simplified: if gender is 'any' or matches, and at least one interest matches (if any).
func IsCompatible(f1, f2 domain.Filter) bool {
	if f1.Gender != "any" && f2.Gender != "any" && f1.Gender != f2.Gender {
		return false
	}

	// For MVP, if interests are empty, they match.
	// If both have interests, check for intersection.
	if len(f1.Interests) == 0 || len(f2.Interests) == 0 {
		return true
	}

	for _, i1 := range f1.Interests {
		for _, i2 := range f2.Interests {
			if i1 == i2 {
				return true
			}
		}
	}

	return false
}
