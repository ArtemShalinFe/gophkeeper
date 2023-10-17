package user

import (
	"context"
	"errors"
	"fmt"
)

type UserDTO struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type User struct {
	ID           string `json:"uuid"`
	Login        string `json:"login"`
	PasswordHash string `json:"password"`
}

var ErrLoginIsBusy = errors.New("login is busy")
var ErrUnknowUser = errors.New("unknow user")

type UserStorage interface {
	AddUser(ctx context.Context, us *UserDTO) (*User, error)
	GetUser(ctx context.Context, us *UserDTO) (*User, error)
}

func (u *UserDTO) AddUser(ctx context.Context, db UserStorage) (*User, error) {
	if u.Login == "" {
		return nil, ErrUnknowUser
	}

	us, err := db.AddUser(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("add user was failed err: %w", err)
	}
	return us, nil
}

func (u *UserDTO) GetUser(ctx context.Context, db UserStorage) (*User, error) {
	if u.Login == "" {
		return nil, ErrUnknowUser
	}

	us, err := db.GetUser(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("get user was failed err: %w", err)
	}
	return us, nil
}
