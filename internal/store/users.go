package store

import (
	"context"
	"database/sql"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int64    `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Password  password `json:"-"`
	CreatedAt string   `json:"created_at"`
	IsActive  bool     `json:"is_active"`
	RoleID    int64    `json:"role_id"`
	Role      Role     `json:"role,omitempty"`
}

type password struct {
	text *string
	hash []byte
}

func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	p.text = &text
	p.hash = hash

	return nil
}

type UsersStore struct {
	db *sql.DB
}

func (s *UsersStore) GetByID(ctx context.Context, id int64) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT users.id, username, email, password, created_at, roles.id, roles.name, roles.level, roles.description
		FROM users
		JOIN roles ON users.role_id = roles.id
		WHERE users.id = $1
	`

	user := &User{}

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.CreatedAt,
		&user.Role.ID,
		&user.Role.Name,
		&user.Role.Level,
		&user.Role.Description,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return user, nil
}

func (s *UsersStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT users.id, username, email, password, created_at, roles.id, roles.name, roles.level, roles.description FROM users 
	JOIN roles ON users.role_id = roles.id
	WHERE email = $1 AND is_active = true`

	user := &User{}

	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.CreatedAt,
		&user.Role.ID,
		&user.Role.Name,
		&user.Role.Level,
		&user.Role.Description,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return user, nil
}

func (s *UsersStore) Create(ctx context.Context, tx *sql.Tx, user *User) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO users (username, email, password, role_id)
		VALUES ($1, $2, $3, (SELECT id FROM roles WHERE name = $4))
		RETURNING id, created_at
	`

	role := user.Role.Name
	if role == "" {
		role = "user"
	}

	err := tx.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.Password.hash,
		role,
	).Scan(
		&user.ID,
		&user.CreatedAt,
	)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case err.Error() == `pq: duplicate key value violates unique constraint "users_username_key"`:
			return ErrDuplicateUsername
		default:
			return err
		}
	}

	return nil
}

func (s *UsersStore) CreateAndInvite(ctx context.Context, user *User,
	token string, invitationExp time.Duration) error {

	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		if err := s.Create(ctx, tx, user); err != nil {
			return err
		}

		err := s.createUserInvitation(ctx, tx, token, invitationExp, user.ID)

		if err != nil {
			return err
		}

		return nil
	})

}

func (s *UsersStore) Delete(ctx context.Context, id int64) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		if err := s.delete(ctx, tx, id); err != nil {
			return err
		}

		if err := s.deleteInvitation(ctx, tx, id); err != nil {
			return err
		}

		return nil
	})
}

func (s *UsersStore) createUserInvitation(ctx context.Context, tx *sql.Tx, token string,
	invitationExp time.Duration, userID int64) error {

	query := `INSERT INTO user_invitations (token, user_id, expiry) VALUES ($1, $2, $3)`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, token, userID, time.Now().Add(invitationExp))

	return err
}

func (s *UsersStore) Activate(ctx context.Context, token string) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		user, err := s.getUserFromInvitation(ctx, tx, token)
		if err != nil {
			return err
		}

		user.IsActive = true
		if err := s.update(ctx, tx, user); err != nil {
			return err
		}

		if err := s.deleteInvitation(ctx, tx, user.ID); err != nil {
			return err
		}
		return nil
	})
}

func (s *UsersStore) getUserFromInvitation(ctx context.Context, tx *sql.Tx, token string) (*User, error) {
	query := `SELECT u.id, u.username, u.email, u.created_at, u.is_active
	FROM users u
	JOIN user_invitations ui ON ui.user_id = u.id
	WHERE ui.token = $1 AND ui.expiry > $2`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	user := &User{}

	err := tx.QueryRowContext(ctx, query, token, time.Now()).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.IsActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return user, nil
}

func (s *UsersStore) update(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `UPDATE users SET username = $1, email = $2, is_active = $3 WHERE id = $4`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, user.Username, user.Email, user.IsActive, user.ID)

	return err
}

func (s *UsersStore) deleteInvitation(ctx context.Context, tx *sql.Tx, userID int64) error {
	query := `DELETE FROM user_invitations WHERE user_id = $1`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, userID)
	return err
}

func (s *UsersStore) delete(ctx context.Context, tx *sql.Tx, id int64) error {
	query := `DELETE FROM users WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, id)
	return err
}

func (p *password) Compare(text string) error {
	return bcrypt.CompareHashAndPassword(p.hash, []byte(text))
}
