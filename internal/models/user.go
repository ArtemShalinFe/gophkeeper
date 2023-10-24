package models

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
		return nil, fmt.Errorf("an occured error while add user, err: %w", err)
	}
	return us, nil
}

func (u *UserDTO) GetUser(ctx context.Context, db UserStorage) (*User, error) {
	if u.Login == "" {
		return nil, ErrUnknowUser
	}

	us, err := db.GetUser(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("an occured error while retrivieng user, err: %w", err)
	}
	return us, nil
}

func (u *User) GetRecords(ctx context.Context, db RecordStorage) ([]*Record, error) {
	rs, err := db.List(ctx, u.ID)
	if err != nil {
		return nil, fmt.Errorf("an error occured while retrieving records, err: %w", err)
	}

	return rs, nil
}

func (u *User) GetRecord(ctx context.Context, db RecordStorage, recordID string) (*Record, error) {
	rs, err := db.Get(ctx, u.ID, recordID)
	if err != nil {
		return nil, fmt.Errorf("an error occured while retrieving record, err: %w", err)
	}

	return rs, nil
}

func (u *User) AddRecord(ctx context.Context, db RecordStorage, record *RecordDTO) (*Record, error) {
	rs, err := db.Add(ctx, u.ID, record)
	if err != nil {
		return nil, fmt.Errorf("an error occured while add record, err: %w", err)
	}

	return rs, nil
}

func (u *User) UpdateRecord(ctx context.Context, db RecordStorage, record *Record) (*Record, error) {
	rs, err := db.Update(ctx, u.ID, record)
	if err != nil {
		return nil, fmt.Errorf("an error occured while update record, err: %w", err)
	}

	return rs, nil
}

func (u *User) DeleteRecord(ctx context.Context, db RecordStorage, recordID string) error {
	if err := db.Delete(ctx, u.ID, recordID); err != nil {
		return fmt.Errorf("an error occured while delete record, err: %w", err)
	}

	return nil
}
