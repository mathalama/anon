package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mathalama/nektokz/moderation-service/internal/domain"
)

type moderationUsecase struct {
	repo    domain.ReportRepository
	userCli domain.UserClient
	chatCli domain.ChatClient
	notif   domain.NotificationClient

	internalToken string
	toxicWords    []string
}

func New(repo domain.ReportRepository, userCli domain.UserClient, chatCli domain.ChatClient, notif domain.NotificationClient, toxicWords []string) domain.ModerationUsecase {
	return &moderationUsecase{
		repo:       repo,
		userCli:    userCli,
		chatCli:    chatCli,
		notif:      notif,
		toxicWords: toxicWords,
	}
}

func (u *moderationUsecase) CreateReport(ctx context.Context, reporterUserID, reportedUserID, roomID, reason string) (*domain.Report, *domain.ReportCounts, error) {
	if reporterUserID == "" {
		return nil, nil, errors.New("reporter_user_id is required")
	}
	if reportedUserID == "" {
		return nil, nil, errors.New("reported_user_id is required")
	}
	if roomID == "" {
		return nil, nil, errors.New("room_id is required")
	}

	now := time.Now()
	r := &domain.Report{
		ID:             uuid.NewString(),
		RoomID:         roomID,
		ReporterUserID: reporterUserID,
		ReportedUserID: reportedUserID,
		Reason:         reason,
		CreatedAt:      now,
	}

	if err := u.repo.CreateReport(ctx, r); err != nil {
		return nil, nil, err
	}

	counts, err := u.repo.CountReports(ctx, reportedUserID, now)
	if err != nil {
		return r, nil, err
	}

	// Auto-ban thresholds (spec):
	// 3 reports / 24h -> 24h ban
	// 5 reports / 7d -> 7d ban
	// 10 total -> permanent
	banHours := 0
	switch {
	case counts.Total >= 10:
		banHours = 24 * 365 * 100 // ~100 years "perm"
	case counts.Last7d >= 5:
		banHours = 24 * 7
	case counts.Last24h >= 3:
		banHours = 24
	}

	if banHours > 0 {
		_ = u.userCli.Ban(ctx, reportedUserID, "auto-ban: report threshold exceeded", "system", banHours)
		_ = u.chatCli.Disconnect(ctx, reportedUserID)
		_ = u.notif.Notify(ctx, "ban_applied", reportedUserID, map[string]any{"hours": banHours})
	}

	return r, &counts, nil
}

func (u *moderationUsecase) GetReport(ctx context.Context, id string) (*domain.Report, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	return u.repo.GetReport(ctx, id)
}

func (u *moderationUsecase) ListReports(ctx context.Context, limit int) ([]*domain.Report, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	return u.repo.ListReports(ctx, limit)
}

func (u *moderationUsecase) ModerateMessage(ctx context.Context, content string) (bool, error) {
	s := strings.ToLower(content)
	for _, w := range u.toxicWords {
		ww := strings.TrimSpace(strings.ToLower(w))
		if ww == "" {
			continue
		}
		if strings.Contains(s, ww) {
			return true, nil
		}
	}
	return false, nil
}

