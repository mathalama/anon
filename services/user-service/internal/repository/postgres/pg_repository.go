package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mathalama/nektokz/user-service/internal/domain"
)

type PGUserRepository struct {
	pool *pgxpool.Pool
}

func NewPGUserRepository(pool *pgxpool.Pool) *PGUserRepository {
	return &PGUserRepository{pool: pool}
}

func (r *PGUserRepository) Create(ctx context.Context, user *domain.User) error {
	if user.Interests == nil {
		user.Interests = []string{}
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO users (id, device_id, gender, interests, is_anonymous, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)
	`, user.ID, nullIfEmpty(user.DeviceID),
		user.Gender, user.Interests, user.IsAnonymous, user.CreatedAt)
	return err
}

func (r *PGUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return r.getOne(ctx, `SELECT id, device_id, gender, interests, is_anonymous, created_at FROM users WHERE id=$1`, id)
}

func (r *PGUserRepository) GetByDeviceID(ctx context.Context, deviceID string) (*domain.User, error) {
	return r.getOne(ctx, `SELECT id, device_id, gender, interests, is_anonymous, created_at FROM users WHERE device_id=$1`, deviceID)
}

func (r *PGUserRepository) Update(ctx context.Context, user *domain.User) error {
	if user.Interests == nil {
		user.Interests = []string{}
	}
	ct, err := r.pool.Exec(ctx, `
		UPDATE users
		SET device_id=$2, gender=$3, interests=$4, is_anonymous=$5
		WHERE id=$1
	`, user.ID, nullIfEmpty(user.DeviceID),
		user.Gender, user.Interests, user.IsAnonymous)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}
func (r *PGUserRepository) CreateBan(ctx context.Context, ban *domain.Ban) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO bans (id, user_id, reason, banned_by, expires_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)
	`, ban.ID, ban.UserID, ban.Reason, ban.BannedBy, ban.ExpiresAt, ban.CreatedAt)
	return err
}

func (r *PGUserRepository) GetActiveBan(ctx context.Context, userID string) (*domain.Ban, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, reason, banned_by, expires_at, created_at
		FROM bans
		WHERE user_id=$1 AND expires_at > $2
		ORDER BY expires_at DESC
		LIMIT 1
	`, userID, time.Now())

	var b domain.Ban
	if err := row.Scan(&b.ID, &b.UserID, &b.Reason, &b.BannedBy, &b.ExpiresAt, &b.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &b, nil
}

func (r *PGUserRepository) getOne(ctx context.Context, q string, arg any) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, q, arg)

	var u domain.User
	var deviceID *string

	if err := row.Scan(&u.ID, &deviceID, &u.Gender, &u.Interests, &u.IsAnonymous, &u.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	if deviceID != nil {
		u.DeviceID = *deviceID
	}

	return &u, nil
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullIfZero(i int64) any {
	if i == 0 {
		return nil
	}
	return i
}
