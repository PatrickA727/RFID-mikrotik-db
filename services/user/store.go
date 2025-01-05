package user

import (
	"database/sql"
	"context"
	"github.com/PatrickA727/mikrotik-db-sys/types"
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
	err := s.db.QueryRow("SELECT id,username, email, role, password FROM users WHERE id = $1", id).Scan(
		&user.ID, &user.Username, &user.Email, &user.Role, &user.Password)
	if err != nil {
		return nil, err
	}

	return &user, nil
}