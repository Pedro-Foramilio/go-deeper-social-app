package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type Follower struct {
	UserID     int64  `json:"user_id"`
	FollowerID int64  `json:"follower_id"`
	CreatedAt  string `json:"created_at"`
}

type FollowerStore struct {
	db *sql.DB
}

func (s *FollowerStore) Follow(ctx context.Context, followerId int64, userID int64) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO followers (user_id, follower_id)
		VALUES ($1, $2)
	`
	_, err := s.db.ExecContext(ctx, query, userID, followerId)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrAlredyExists
		}
	}

	return err
}

func (s *FollowerStore) Unfollow(ctx context.Context, followerId int64, userID int64) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		DELETE FROM followers
		WHERE user_id = $1 AND follower_id = $2
	`
	_, err := s.db.ExecContext(ctx, query, userID, followerId)
	return err
}

func (s *FollowerStore) GetFollowers(ctx context.Context, userID int64) ([]User, error) {
	return nil, nil
}
