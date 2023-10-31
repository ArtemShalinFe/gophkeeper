package models

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ArtemShalinFe/gophkeeper/internal/vectors"
	"github.com/google/uuid"
)

type UserDTO struct {
	Login    string `cbor:"login"`
	Password string `cbor:"password"`
}

type User struct {
	ID           string `cbor:"uuid"`
	Login        string `cbor:"login"`
	PasswordHash string `cbor:"password"`
}

var ErrLoginIsBusy = errors.New("login is busy")
var ErrUnknowUser = errors.New("unknow user")

const errSyncRecordTmp = "an error occure while update record (ID=%s), err: %w"
const DefaultLimit = 30

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
		if !errors.Is(err, ErrLoginIsBusy) {
			return nil, fmt.Errorf("an occured error while add user, err: %w", err)
		}
		return nil, fmt.Errorf("login is busy, err: %w", err)
	}
	return us, nil
}

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

func (u *User) GetRecords(ctx context.Context, db RecordStorage, offset int, limit int) ([]*Record, error) {
	rs, err := db.List(ctx, u.ID, offset, limit)
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
	if record.ID == "" {
		return nil, ErrRecordNotFound
	}
	rs, err := db.Update(ctx, u.ID, record)
	if err != nil {
		return nil, fmt.Errorf("an error occured while update record, err: %w", err)
	}

	return rs, nil
}

func (u *User) DeleteRecord(ctx context.Context, db RecordStorage, recordID string) error {
	if recordID == "" {
		return ErrRecordNotFound
	}
	if err := db.Delete(ctx, u.ID, recordID); err != nil {
		return fmt.Errorf("an error occured while delete record, err: %w", err)
	}

	return nil
}

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
