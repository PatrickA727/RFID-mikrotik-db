package user

import (
	"database/sql"
	"context"
	"github.com/PatrickA727/mikrotik-db-sys/types"
	"fmt"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) BeginTransaction(ctx context.Context) (*sql.Tx, error) {
	tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, err
    }
    return tx, nil
}

func (s *Store) RegisterNewUser(user types.User) error {
	_, err := s.db.Exec("INSERT INTO users (username, email, password, role) VALUES ($1, $2, $3, $4)", 
				user.Username, user.Email, user.Password, user.Role,
			)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetUserByEmail(email string) (*types.User, error) {
	var user types.User
	err := s.db.QueryRow("SELECT id,username, email, role, password FROM users WHERE email = $1", email).Scan(
		&user.ID, &user.Username, &user.Email, &user.Role, &user.Password)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *Store) GetUserById(id int) (*types.User, error) {
	var user types.User
	err := s.db.QueryRow("SELECT id,username, email, role FROM users WHERE id = $1", id).Scan(
		&user.ID, &user.Username, &user.Email, &user.Role)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *Store) DeleteUserById(id int, ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) CreateSession(ctx context.Context, session types.Session) error {
	_, err := s.db.ExecContext(ctx, "INSERT INTO sessions (userid, refresh_token) VALUES ($1, $2)", 
			session.Userid, session.RefreshToken,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) RevokeSession(session types.Session) error {
	res, err := s.db.Exec("UPDATE sessions SET is_revoked = TRUE WHERE userid = $1 AND refresh_token = $2", session.Userid, session.RefreshToken)
	if err != nil {
		return err
	}

	// Check if any rows were updated
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking affected rows: %v", err)
	}

	// If no rows were affected, return an error indicating the session wasn't found
	if rowsAffected == 0 {
		return fmt.Errorf("session not found or already revoked")
	}

	return nil
}
func (s *Store) RevokeSessionBulk(id int) error {
	res, err := s.db.Exec("UPDATE sessions SET is_revoked = TRUE where userid = $1", id)
	if err != nil {
		return err
	}

	// Check if any rows were updated
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking affected rows: %v", err)
	}

	// If no rows were affected, return an error indicating the session wasn't found
	if rowsAffected == 0 {
		return fmt.Errorf("session not found or already revoked")
	}

	return nil
}

func (s *Store) CheckSession(tokenString string) (bool, int, error) {
	var session types.Session
	err := s.db.QueryRow("SELECT userid FROM sessions WHERE refresh_token = $1 AND is_revoked = FALSE", 
	tokenString).Scan(&session.Userid)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, 0, nil
		}
		return false, 0, err
	}

	return true, session.Userid, nil
}