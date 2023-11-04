package models

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	gomock "go.uber.org/mock/gomock"
)

var errSomethingWentWrong = errors.New("something went wrong")

func TestUserDTO_AddUser(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	stg := NewMockUserStorage(ctrl)

	type fields struct {
		Login    string
		Password string
	}
	tests := []struct {
		name    string
		fields  fields
		want    *User
		wantErr bool
		err     error
	}{
		{
			name: "case add user",
			fields: fields{
				Login:    uuid.NewString(),
				Password: uuid.NewString(),
			},
			want: &User{
				ID:           uuid.NewString(),
				Login:        uuid.NewString(),
				PasswordHash: uuid.NewString(),
			},
			err: nil,
		},
		{
			name: "case add user, but storage return error",
			fields: fields{
				Login:    uuid.NewString(),
				Password: uuid.NewString(),
			},
			want: &User{
				ID:           uuid.NewString(),
				Login:        uuid.NewString(),
				PasswordHash: uuid.NewString(),
			},
			err: errSomethingWentWrong,
		},
		{
			name: "case add user, but login is busy",
			fields: fields{
				Login:    uuid.NewString(),
				Password: uuid.NewString(),
			},
			want: &User{
				ID:           uuid.NewString(),
				Login:        uuid.NewString(),
				PasswordHash: uuid.NewString(),
			},
			err: ErrLoginIsBusy,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserDTO{
				Login:    tt.fields.Login,
				Password: tt.fields.Password,
			}
			wantErr := (tt.err != nil)
			if wantErr {
				stg.EXPECT().AddUser(ctx, u).Return(nil, tt.err)
			} else {
				stg.EXPECT().AddUser(ctx, u).Return(tt.want, nil)
			}

			got, err := u.AddUser(ctx, stg)
			if (err != nil) != wantErr {
				t.Errorf("UserDTO.AddUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserDTO.AddUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserDTO_GetUser(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	stg := NewMockUserStorage(ctrl)

	type fields struct {
		Login    string
		Password string
	}
	tests := []struct {
		name    string
		fields  fields
		want    *User
		wantErr bool
		err     error
	}{
		{
			name: "case get user",
			fields: fields{
				Login:    uuid.NewString(),
				Password: uuid.NewString(),
			},
			want: &User{
				ID:           uuid.NewString(),
				Login:        uuid.NewString(),
				PasswordHash: uuid.NewString(),
			},
			err: nil,
		},
		{
			name: "case get user, but storage return error",
			fields: fields{
				Login:    uuid.NewString(),
				Password: uuid.NewString(),
			},
			want: &User{
				ID:           uuid.NewString(),
				Login:        uuid.NewString(),
				PasswordHash: uuid.NewString(),
			},
			err: errSomethingWentWrong,
		},
		{
			name: "case get user, but login is busy",
			fields: fields{
				Login:    uuid.NewString(),
				Password: uuid.NewString(),
			},
			want: &User{
				ID:           uuid.NewString(),
				Login:        uuid.NewString(),
				PasswordHash: uuid.NewString(),
			},
			err: ErrUnknowUser,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserDTO{
				Login:    tt.fields.Login,
				Password: tt.fields.Password,
			}
			wantErr := (tt.err != nil)
			if wantErr {
				stg.EXPECT().GetUser(ctx, u).Return(nil, tt.err)
			} else {
				stg.EXPECT().GetUser(ctx, u).Return(tt.want, nil)
			}

			got, err := u.GetUser(ctx, stg)
			if (err != nil) != wantErr {
				t.Errorf("UserDTO.GetUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserDTO.GetUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_GetRecords(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	stg := NewMockRecordStorage(ctrl)

	type fields struct {
		ID           string
		Login        string
		PasswordHash string
	}
	type args struct {
		offset int
		limit  int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*Record
		wantErr bool
	}{
		{
			name: "positive get list record",
			fields: fields{
				ID:           uuid.NewString(),
				Login:        uuid.NewString(),
				PasswordHash: uuid.NewString(),
			},
			args: args{
				offset: 0,
				limit:  DefaultLimit,
			},
			want:    generateRecords(t, DefaultLimit),
			wantErr: false,
		},
		{
			name: "case get list record, but storage return error",
			fields: fields{
				ID:           uuid.NewString(),
				Login:        uuid.NewString(),
				PasswordHash: uuid.NewString(),
			},
			args: args{
				offset: 0,
				limit:  DefaultLimit,
			},
			want:    generateRecords(t, DefaultLimit),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{
				ID:           tt.fields.ID,
				Login:        tt.fields.Login,
				PasswordHash: tt.fields.PasswordHash,
			}

			if tt.wantErr {
				stg.EXPECT().ListRecords(gomock.Any(), tt.fields.ID, tt.args.offset, tt.args.limit).
					Return(nil, errSomethingWentWrong)
			} else {
				stg.EXPECT().ListRecords(gomock.Any(), tt.fields.ID, tt.args.offset, tt.args.limit).
					Return(tt.want, nil)
			}

			got, err := u.GetRecords(ctx, stg, tt.args.offset, tt.args.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("User.GetRecords() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(len(got), len(tt.want)) {
				t.Errorf("User.GetRecords() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_GetRecord(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	stg := NewMockRecordStorage(ctrl)

	type fields struct {
		ID           string
		Login        string
		PasswordHash string
	}
	type args struct {
		recordID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Record
		wantErr bool
	}{
		{
			name: "positive get record",
			fields: fields{
				ID:           uuid.NewString(),
				Login:        uuid.NewString(),
				PasswordHash: uuid.NewString(),
			},
			args: args{
				recordID: uuid.NewString(),
			},
			want:    generateRecords(t, 1)[0],
			wantErr: false,
		},
		{
			name: "case get record, but storage return error",
			fields: fields{
				ID:           uuid.NewString(),
				Login:        uuid.NewString(),
				PasswordHash: uuid.NewString(),
			},
			args: args{
				recordID: uuid.NewString(),
			},
			want:    generateRecords(t, 1)[0],
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{
				ID:           tt.fields.ID,
				Login:        tt.fields.Login,
				PasswordHash: tt.fields.PasswordHash,
			}

			if tt.wantErr {
				stg.EXPECT().GetRecord(gomock.Any(), tt.fields.ID, tt.args.recordID).
					Return(nil, errSomethingWentWrong)
			} else {
				stg.EXPECT().GetRecord(gomock.Any(), tt.fields.ID, tt.args.recordID).
					Return(tt.want, nil)
			}

			got, err := u.GetRecord(ctx, stg, tt.args.recordID)
			if (err != nil) != tt.wantErr {
				t.Errorf("User.GetRecord() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("User.GetRecord() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_AddRecord(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	stg := NewMockRecordStorage(ctrl)

	type fields struct {
		ID           string
		Login        string
		PasswordHash string
	}
	type args struct {
		record *RecordDTO
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Record
		wantErr bool
	}{
		{
			name: "positive add record",
			fields: fields{
				ID:           uuid.NewString(),
				Login:        uuid.NewString(),
				PasswordHash: uuid.NewString(),
			},
			args: args{
				record: &RecordDTO{
					Description: uuid.NewString(),
					Type:        string(AuthType),
					Data:        []byte(uuid.NewString()),
					Hashsum:     uuid.NewString(),
				},
			},
			want:    generateRecords(t, 1)[0],
			wantErr: false,
		},
		{
			name: "case add record, but storage return error",
			fields: fields{
				ID:           uuid.NewString(),
				Login:        uuid.NewString(),
				PasswordHash: uuid.NewString(),
			},
			args: args{
				record: &RecordDTO{
					Description: uuid.NewString(),
					Type:        string(AuthType),
					Data:        []byte(uuid.NewString()),
					Hashsum:     uuid.NewString(),
				},
			},
			want:    generateRecords(t, 1)[0],
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{
				ID:           tt.fields.ID,
				Login:        tt.fields.Login,
				PasswordHash: tt.fields.PasswordHash,
			}

			if tt.wantErr {
				stg.EXPECT().AddRecord(gomock.Any(), tt.fields.ID, gomock.Any()).
					Return(nil, errSomethingWentWrong)
			} else {
				stg.EXPECT().AddRecord(gomock.Any(), tt.fields.ID, gomock.Any()).
					Return(tt.want, nil)
			}

			got, err := u.AddRecord(ctx, stg, tt.args.record)
			if (err != nil) != tt.wantErr {
				t.Errorf("User.AddRecords() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("User.AddRecords() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_UpdateRecord(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	stg := NewMockRecordStorage(ctrl)

	type fields struct {
		ID           string
		Login        string
		PasswordHash string
	}
	type args struct {
		record *Record
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Record
		wantErr bool
	}{
		{
			name: "positive update record",
			fields: fields{
				ID:           uuid.NewString(),
				Login:        uuid.NewString(),
				PasswordHash: uuid.NewString(),
			},
			args: args{
				record: generateRecords(t, 1)[0],
			},
			want:    generateRecords(t, 1)[0],
			wantErr: false,
		},
		{
			name: "case update record, but storage return error",
			fields: fields{
				ID:           uuid.NewString(),
				Login:        uuid.NewString(),
				PasswordHash: uuid.NewString(),
			},
			args: args{
				record: generateRecords(t, 1)[0],
			},
			want:    generateRecords(t, 1)[0],
			wantErr: true,
		},
		{
			name: "case update record, but storage return error",
			fields: fields{
				ID:           uuid.NewString(),
				Login:        uuid.NewString(),
				PasswordHash: uuid.NewString(),
			},
			args: args{
				record: &Record{ID: ""},
			},
			want:    generateRecords(t, 1)[0],
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{
				ID:           tt.fields.ID,
				Login:        tt.fields.Login,
				PasswordHash: tt.fields.PasswordHash,
			}

			if tt.args.record.ID != "" {
				if tt.wantErr {
					stg.EXPECT().UpdateRecord(gomock.Any(), tt.fields.ID, gomock.Any()).
						Return(nil, errSomethingWentWrong)
				} else {
					stg.EXPECT().UpdateRecord(gomock.Any(), tt.fields.ID, gomock.Any()).
						Return(tt.want, nil)
				}
			}

			got, err := u.UpdateRecord(ctx, stg, tt.args.record)
			if (err != nil) != tt.wantErr {
				t.Errorf("User.UpdateRecord() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("User.UpdateRecord() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_DeleteRecord(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	stg := NewMockRecordStorage(ctrl)

	type fields struct {
		ID           string
		Login        string
		PasswordHash string
	}
	type args struct {
		recordID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "positive delete record",
			fields: fields{
				ID:           uuid.NewString(),
				Login:        uuid.NewString(),
				PasswordHash: uuid.NewString(),
			},
			args: args{
				recordID: generateRecords(t, 1)[0].ID,
			},
			wantErr: false,
		},
		{
			name: "case delete record, but storage return error",
			fields: fields{
				ID:           uuid.NewString(),
				Login:        uuid.NewString(),
				PasswordHash: uuid.NewString(),
			},
			args: args{
				recordID: generateRecords(t, 1)[0].ID,
			},
			wantErr: true,
		},
		{
			name: "case delete record, but record ID is empty",
			fields: fields{
				ID:           uuid.NewString(),
				Login:        uuid.NewString(),
				PasswordHash: uuid.NewString(),
			},
			args: args{
				recordID: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{
				ID:           tt.fields.ID,
				Login:        tt.fields.Login,
				PasswordHash: tt.fields.PasswordHash,
			}

			if tt.args.recordID != "" {
				if tt.wantErr {
					stg.EXPECT().DeleteRecord(gomock.Any(), tt.fields.ID, gomock.Any()).
						Return(errSomethingWentWrong)
				} else {
					stg.EXPECT().DeleteRecord(gomock.Any(), tt.fields.ID, gomock.Any()).
						Return(nil)
				}
			}

			err := u.DeleteRecord(ctx, stg, tt.args.recordID)
			if (err != nil) != tt.wantErr {
				t.Errorf("User.DeleteRecord() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func generateMetadata(mic int) []string {
	ss := make([]string, mic)
	for i := 0; i < mic; i++ {
		ss[i] = fmt.Sprintf("%s:%s", uuid.NewString(), uuid.NewString())
	}
	return ss
}

func generateRecord(t *testing.T, dt DataType, a RecordData) *Record {
	t.Helper()

	mic := 10
	mis, err := NewMetadataFromStringArray(generateMetadata(mic))
	if err != nil {
		t.Errorf("an error occured while generating metadata, err: %v", err)
	}

	r, err := NewRecord(uuid.NewString(), uuid.NewString(), dt,
		time.Now(), time.Now(), a, mis, false, 1)
	if err != nil {
		t.Errorf("an error occured while generating record, err: %v", err)
	}
	return r
}

func generateAuthRecord(t *testing.T) *Record {
	return generateRecord(t, AuthType, &Auth{
		Login:    uuid.NewString(),
		Password: uuid.NewString(),
	})
}

func generateTextRecord(t *testing.T) *Record {
	return generateRecord(t, TextType, &Text{
		Data: uuid.NewString(),
	})
}

func generateBinaryRecord(t *testing.T) *Record {
	return generateRecord(t, BinaryType, &Binary{
		Data: []byte(uuid.NewString()),
	})
}

func generateCardRecord(t *testing.T) *Record {
	return generateRecord(t, CardType, &Card{
		Number: uuid.NewString(),
		Term:   time.Now(),
		Owner:  uuid.NewString(),
	})
}

func generateRecords(t *testing.T, limit int) []*Record {
	var records []*Record
	for i := 0; i < limit; i++ {
		switch i % 2 {
		case 0:
			records = append(records, generateAuthRecord(t))
		case 1:
			records = append(records, generateTextRecord(t))
		case 3:
			records = append(records, generateBinaryRecord(t))
		default:
			records = append(records, generateCardRecord(t))
		}
		if len(records) >= DefaultLimit {
			break
		}
	}
	return records
}
