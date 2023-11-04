package models

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/ArtemShalinFe/gophkeeper/internal/vectors"
)

// UserDTO - Data transfer object for User.
type UserDTO struct {
	Login    string `cbor:"login"`
	Password string `cbor:"password"`
}

// User - An object that identifies the user.
type User struct {
	ID           string `cbor:"uuid"`
	Login        string `cbor:"login"`
	PasswordHash string `cbor:"password"`
}

// ErrLoginIsBusy - The error is returned if the username is already occupied.
var ErrLoginIsBusy = errors.New("login is busy")

// ErrUnknowUser - The error is returned if the password is not correct.
var ErrUnknowUser = errors.New("unknow user")

const (
	errSyncRecordTmp = "an error occure while update record (ID=%s), err: %w"
	// DefaultLimit - Limit of records that are returned from storage.
	DefaultLimit = 30
)

// UserStorage - The interface that the repository should implement for user registration and authorization.
type UserStorage interface {
	// AddUser - The method is used when registering a user.
	AddUser(ctx context.Context, us *UserDTO) (*User, error)
	// GetUser - This method is used when the user logs in.
	GetUser(ctx context.Context, us *UserDTO) (*User, error)
}

// AddUser - The method is used when registering a user.
// The method checks that the Login field is not empty.
func (u *UserDTO) AddUser(ctx context.Context, db UserStorage) (*User, error) {
	if u.Login == "" {
		return nil, ErrUnknowUser
	}

	us, err := db.AddUser(ctx, u)
	if err != nil {
		if !errors.Is(err, ErrLoginIsBusy) {
			return nil, fmt.Errorf("an occured error while add user, err: %w", err)
		}
		return nil, fmt.Errorf("login is busy, err: %w", err)
	}
	return us, nil
}

// GetUser - This method is used when the user logs in.
// The method checks that the Login field is not empty.
func (u *UserDTO) GetUser(ctx context.Context, db UserStorage) (*User, error) {
	if u.Login == "" {
		return nil, ErrUnknowUser
	}

	us, err := db.GetUser(ctx, u)
	if err != nil {
		if !errors.Is(err, ErrUnknowUser) {
			return nil, fmt.Errorf("an occured error while retrivieng user, err: %w", err)
		}
		return nil, fmt.Errorf("User not found, err: %w", err)
	}
	return us, nil
}

// GetRecords - The method is used to get a list of user records from the storage.
func (u *User) GetRecords(ctx context.Context, db RecordStorage, offset int, limit int) ([]*Record, error) {
	rs, err := db.List(ctx, u.ID, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("an error occured while retrieving records, err: %w", err)
	}

	return rs, nil
}

// GetRecord - The method is used to get a record by the user's recordID from the storage.
func (u *User) GetRecord(ctx context.Context, db RecordStorage, recordID string) (*Record, error) {
	rs, err := db.Get(ctx, u.ID, recordID)
	if err != nil {
		return nil, fmt.Errorf("an error occured while retrieving record, err: %w", err)
	}

	return rs, nil
}

// AddRecord - This method is used to add a user record to the repository.
func (u *User) AddRecord(ctx context.Context, db RecordStorage, record *RecordDTO) (*Record, error) {
	rs, err := db.Add(ctx, u.ID, record)
	if err != nil {
		return nil, fmt.Errorf("an error occured while add record, err: %w", err)
	}

	return rs, nil
}

// UpdateRecord - This method is used to update a user record to the repository.
func (u *User) UpdateRecord(ctx context.Context, db RecordStorage, record *Record) (*Record, error) {
	if record.ID == "" {
		return nil, ErrRecordNotFound
	}
	rs, err := db.Update(ctx, u.ID, record)
	if err != nil {
		return nil, fmt.Errorf("an error occured while update record, err: %w", err)
	}

	return rs, nil
}

// DeleteRecord - This method is used to delete a user record to the repository.
func (u *User) DeleteRecord(ctx context.Context, db RecordStorage, recordID string) error {
	if recordID == "" {
		return ErrRecordNotFound
	}
	if err := db.Delete(ctx, u.ID, recordID); err != nil {
		return fmt.Errorf("an error occured while delete record, err: %w", err)
	}

	return nil
}

// SyncRecords - This method is used to synchronize user records between repositories.
func (u *User) SyncRecords(ctx context.Context, stg1 RecordStorage, stg2 RecordStorage, tick int) error {
	const t = "an error occured while sync stg1 (%T) with stg2 (%T), err: %w"

	ticker := time.NewTicker(time.Second * time.Duration(tick))

syncloop:
	for {
		if err := u.syncStorages(ctx, stg1, stg2); err != nil {
			return fmt.Errorf(t, stg1, stg2, err)
		}

		if err := u.syncStorages(ctx, stg2, stg1); err != nil {
			return fmt.Errorf(t, stg2, stg1, err)
		}

		select {
		case <-ctx.Done():
			break syncloop
		case <-ticker.C:
		}
	}

	return nil
}

func (u *User) syncStorages(ctx context.Context, stg1 RecordStorage, stg2 RecordStorage) error {
	offset := 0
	for {
		stg2rs, err := stg2.List(ctx, u.ID, offset, DefaultLimit)
		if err != nil {
			return fmt.Errorf("an error occured while retrieving list records from server, err: %w", err)
		}

		if len(stg2rs) == 0 {
			break
		}

		for _, r2 := range stg2rs {
			r1, err := stg1.Get(ctx, u.ID, r2.ID)
			if err != nil {
				if !errors.Is(err, ErrRecordNotFound) {
					return fmt.Errorf("an error occured while trying update record(ID=%s), err: %w", r2.ID, err)
				}
				_, err = stg1.Update(ctx, u.ID, r2)
				if err != nil {
					return fmt.Errorf(errSyncRecordTmp, r2.ID, err)
				}
				continue
			}

			switch vectors.NewComparison(r2, r1).Compare() {
			case vectors.VectorAIsEqualsVectorB:
			case vectors.VectorAIsHigherVectorB:
				if _, err := stg1.Update(ctx, u.ID, r2); err != nil {
					return fmt.Errorf(errSyncRecordTmp, r2.ID, err)
				}
			case vectors.VectorAIsLowerVectorB:
				if _, err := stg2.Update(ctx, u.ID, r1); err != nil {
					return fmt.Errorf(errSyncRecordTmp, r1.ID, err)
				}
			default:
				if _, err := stg1.Update(ctx, u.ID, r2); err != nil {
					return fmt.Errorf(errSyncRecordTmp, r2.ID, err)
				}

				r1.ID = uuid.NewString()
				r1.Description = fmt.Sprintf("(COPY) %s", r1.Description)
				if _, err := stg2.Update(ctx, u.ID, r1); err != nil {
					return fmt.Errorf(errSyncRecordTmp, r1.ID, err)
				}
			}
		}
		offset += DefaultLimit
	}

	return nil
}
