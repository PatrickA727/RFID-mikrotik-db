package types

import (
	"context"
	"time"
)

type UserStore interface {
	RegisterNewUser(user User) error
	GetUserByEmail(email string) (*User, error)
	GetUserById(id int) (*User, error)
	DeleteUserById(id int, ctx context.Context) error
	CreateSession(ctx context.Context, session Session) error
	RevokeSession(session Session) error
	CheckSession(tokenString string) (bool, int, error)
	RevokeSessionBulk(id int) error
}

type User struct {
	ID			int			`json:"id"`
	CreatedAt	time.Time	`json:"createdat"`
	Username	string		`json:"username"`
	Email		string		`json:"email"`
	Password	string		`json:"-"`	// - is to ignore this field for the response(obvious reasons)
}

type UserPayload struct {
	Username	string	`json:"username" validate:"required"`
	Email		string 	`json:"email" validate:"required,email"`
	Password 	string 	`json:"password" validate:"required,min=3,max=130"`
}

type LoginPayload struct {
	Email	string	`json:"email" validate:"required,email"`
	Password 	string 	`json:"password" validate:"required"`
}

type DeletePayload struct {
	ID	int 
}

type Session struct {
	Userid 			int		`json:"userid" validate:"required"`
	RefreshToken	string	`json:"refresh_token" validate:"required"`
}
