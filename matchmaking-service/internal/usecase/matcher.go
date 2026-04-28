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
	mode1 := f1.Mode
	if mode1 == "" {
		mode1 = "text"
	}
	mode2 := f2.Mode
	if mode2 == "" {
		mode2 = "text"
	}
	if mode1 != mode2 {
		return false
	}

	// 1. Check if f1 wants f2's gender
	if f1.Gender != "any" && f1.Gender != f2.MyGender {
		return false
	}
	// 2. Check if f2 wants f1's gender
	if f2.Gender != "any" && f2.Gender != f1.MyGender {
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
