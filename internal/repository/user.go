package repository

import (
	"context"

	"gitee.com/meinongyihe/travel-rpc/ent"
)

// UserRepository is the persistence boundary for end-user (customer) accounts.
type UserRepository interface {
	Create(ctx context.Context, input CreateUserInput) (*ent.User, error)
	GetByID(ctx context.Context, id int64) (*ent.User, error)
	GetByUsername(ctx context.Context, username string) (*ent.User, error)
	GetByMobile(ctx context.Context, mobile string) (*ent.User, error)
	GetByToken(ctx context.Context, token string) (*ent.User, error)
	Update(ctx context.Context, id int64, input UpdateUserInput) (*ent.User, error)
	UpdatePassword(ctx context.Context, id int64, hashedPassword, salt string) error
	UpdateToken(ctx context.Context, id int64, token string) error
	UpdateLoginInfo(ctx context.Context, id int64, loginIP string, loginTime int64) error
}

type CreateUserInput struct {
	Username string
	Password string
	Salt     string
	Email    string
	Mobile   string
	Nickname string
	JoinIP   string
}

type UpdateUserInput struct {
	Nickname *string
	Avatar   *string
	Email    *string
	Mobile   *string
	Bio      *string
	Birthday *string
	Gender   *int8
}
