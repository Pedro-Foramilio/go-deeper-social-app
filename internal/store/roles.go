package store

import (
	"context"
	"database/sql"
	"time"
)

type Role struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Level       int    `json:"level"`
}

type RoleStore struct {
	db *sql.DB
}

func (s *RoleStore) GetByName(ctx context.Context, name string) (*Role, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	query := `
		SELECT id, name, description, level
		FROM roles
		WHERE name = $1
	`

	role := &Role{}

	err := s.db.QueryRowContext(ctx, query, name).Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		&role.Level,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return role, nil
}
