package server

import (
	"context"
	"net"
	"reflect"
	"testing"

	gomock "go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	"github.com/ArtemShalinFe/gophkeeper/internal/config"
	"github.com/ArtemShalinFe/gophkeeper/internal/models"
	"github.com/google/uuid"
)

func NewUserSrvListener(srv UsersServer) func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)

	server := grpc.NewServer()
	RegisterUsersServer(server, srv)

	go func() {
		if err := server.Serve(listener); err != nil {
			zap.S().Errorf("grpc serve failed, err: %v", err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

func NewRecordsSrvListener(srv RecordsServer) func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)

	server := grpc.NewServer()
	RegisterRecordsServer(server, srv)

	go func() {
		if err := server.Serve(listener); err != nil {
			zap.S().Errorf("grpc serve failed, err: %v", err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

func TestGKClient_AddUser(t *testing.T) {
	ctx := context.Background()
	log := zap.L()

	ctrl := gomock.NewController(t)
	mock := NewMockUsersServer(ctrl)

	cfg := config.NewClientCfg()
	c := &GKClient{
		addr:     cfg.GKeeper,
		log:      log,
		certPath: cfg.CertFilePath,
	}
	opts := c.getDialOpts()
	lis := NewUserSrvListener(mock)
	creds, err := getClientCreds("")
	if err != nil {
		t.Errorf("an error occured while get client gredentials, err: %v", err)
	}
	opts = append(opts, grpc.WithContextDialer(lis), grpc.WithTransportCredentials(creds))

	conn, err := grpc.DialContext(ctx, "", opts...)
	if err != nil {
		t.Errorf("an occured error when getting conn grpc client, err: %v", err)
	}
	defer conn.Close()

	c.cc = conn

	uuid := uuid.NewString()
	hp, err := hashPassword(uuid)
	if err != nil {
		t.Errorf("an error occured whil retrieving hash password, err: %v", err)
	}

	tests := []struct {
		name    string
		us      *models.UserDTO
		want    *models.User
		wantErr bool
	}{
		{
			name: "case positive list",
			us: &models.UserDTO{
				Login:    uuid,
				Password: uuid,
			},
			want: &models.User{
				ID:           uuid,
				Login:        uuid,
				PasswordHash: hp,
			},
			wantErr: false,
		},
		{
			name: "case list, but server return error",
			us: &models.UserDTO{
				Login:    uuid,
				Password: uuid,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				mock.EXPECT().Register(gomock.Any(), gomock.Any()).Return(nil, errSomethingWentWrong)
			} else {
				mock.EXPECT().Register(gomock.Any(), gomock.Any()).Return(&RegisterResponse{
					User: &User{
						Id: tt.want.ID,
					},
				}, nil)
			}
			got, err := c.AddUser(ctx, tt.us)
			if (err != nil) != tt.wantErr {
				t.Errorf("GKClient.AddUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got.ID, tt.want.ID) {
				t.Errorf("GKClient.AddUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGKClient_GetUser(t *testing.T) {
	ctx := context.Background()
	log := zap.L()

	ctrl := gomock.NewController(t)
	mock := NewMockUsersServer(ctrl)

	cfg := config.NewClientCfg()
	c := &GKClient{
		addr:     cfg.GKeeper,
		log:      log,
		certPath: cfg.CertFilePath,
	}
	opts := c.getDialOpts()
	lis := NewUserSrvListener(mock)
	creds, err := getClientCreds("")
	if err != nil {
		t.Errorf("an error occured while get client gredentials, err: %v", err)
	}
	opts = append(opts, grpc.WithContextDialer(lis), grpc.WithTransportCredentials(creds))

	conn, err := grpc.DialContext(ctx, "", opts...)
	if err != nil {
		t.Errorf("an occured error when getting conn grpc client, err: %v", err)
	}
	defer conn.Close()

	c.cc = conn

	uuid := uuid.NewString()
	hp, err := hashPassword(uuid)
	if err != nil {
		t.Errorf("an error occured whil retrieving hash password, err: %v", err)
	}

	tests := []struct {
		name    string
		us      *models.UserDTO
		want    *models.User
		wantErr bool
	}{
		{
			name: "case login register",
			us: &models.UserDTO{
				Login:    uuid,
				Password: uuid,
			},
			want: &models.User{
				ID:           uuid,
				Login:        uuid,
				PasswordHash: hp,
			},
			wantErr: false,
		},
		{
			name: "case login, but server return error",
			us: &models.UserDTO{
				Login:    uuid,
				Password: uuid,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				mock.EXPECT().Login(gomock.Any(), gomock.Any()).Return(nil, errSomethingWentWrong)
			} else {
				mock.EXPECT().Login(gomock.Any(), gomock.Any()).Return(&LoginResponse{
					User: &User{
						Id: tt.want.ID,
					},
				}, nil)
			}
			got, err := c.GetUser(ctx, tt.us)
			if (err != nil) != tt.wantErr {
				t.Errorf("GKClient.GetUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got.ID, tt.want.ID) {
				t.Errorf("GKClient.GetUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGKClient_List(t *testing.T) {
	ctx := context.Background()
	log := zap.L()

	ctrl := gomock.NewController(t)
	mock := NewMockRecordsServer(ctrl)

	cfg := config.NewClientCfg()
	c := &GKClient{
		addr:     cfg.GKeeper,
		log:      log,
		certPath: cfg.CertFilePath,
	}
	opts := c.getDialOpts()
	lis := NewRecordsSrvListener(mock)
	creds, err := getClientCreds("")
	if err != nil {
		t.Errorf("an error occured while get client gredentials, err: %v", err)
	}
	opts = append(opts, grpc.WithContextDialer(lis), grpc.WithTransportCredentials(creds))

	conn, err := grpc.DialContext(ctx, "", opts...)
	if err != nil {
		t.Errorf("an occured error when getting conn grpc client, err: %v", err)
	}
	defer conn.Close()

	c.cc = conn

	uuid := uuid.NewString()
	hp, err := hashPassword(uuid)
	if err != nil {
		t.Errorf("an error occured whil retrieving hash password, err: %v", err)
	}

	tests := []struct {
		name    string
		us      *models.User
		offset  int
		limit   int
		want    []*models.Record
		wantErr bool
	}{
		{
			name: "case list records",
			us: &models.User{
				ID:           uuid,
				Login:        uuid,
				PasswordHash: hp,
			},
			offset:  0,
			limit:   models.DefaultLimit,
			want:    generateRecords(t, models.DefaultLimit),
			wantErr: false,
		},
		{
			name: "case list records, but server return error",
			us: &models.User{
				ID:           uuid,
				Login:        uuid,
				PasswordHash: hp,
			},
			offset:  0,
			limit:   models.DefaultLimit,
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				mock.EXPECT().List(gomock.Any(), gomock.Any()).Return(nil, errSomethingWentWrong)
			} else {
				var rspb []*Record
				for _, r := range tt.want {
					rpb, err := convRecordToProtobuff(r)
					if err != nil {
						t.Errorf("an error occured while convert record to protobuff, err: %v", err)
					}
					rspb = append(rspb, rpb)
				}
				mock.EXPECT().List(gomock.Any(), gomock.Any()).Return(&ListRecordResponse{
					Records: rspb,
				}, nil)
			}
			got, err := c.List(ctx, tt.us.ID, tt.offset, tt.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("GKClient.List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != len(tt.want) {
				t.Errorf("GKClient.List() = %v, want %v", len(got), tt.want)
			}
		})
	}
}

func TestGKClient_Get(t *testing.T) {
	ctx := context.Background()
	log := zap.L()

	ctrl := gomock.NewController(t)
	mock := NewMockRecordsServer(ctrl)

	cfg := config.NewClientCfg()
	c := &GKClient{
		addr:     cfg.GKeeper,
		log:      log,
		certPath: cfg.CertFilePath,
	}
	opts := c.getDialOpts()
	lis := NewRecordsSrvListener(mock)
	creds, err := getClientCreds("")
	if err != nil {
		t.Errorf("an error occured while get client gredentials, err: %v", err)
	}
	opts = append(opts, grpc.WithContextDialer(lis), grpc.WithTransportCredentials(creds))

	conn, err := grpc.DialContext(ctx, "", opts...)
	if err != nil {
		t.Errorf("an occured error when getting conn grpc client, err: %v", err)
	}
	defer conn.Close()

	c.cc = conn

	uuid := uuid.NewString()
	hp, err := hashPassword(uuid)
	if err != nil {
		t.Errorf("an error occured whil retrieving hash password, err: %v", err)
	}

	tests := []struct {
		name     string
		us       *models.User
		recordID string
		want     *models.Record
		wantErr  bool
	}{
		{
			name: "case get record",
			us: &models.User{
				ID:           uuid,
				Login:        uuid,
				PasswordHash: hp,
			},
			recordID: uuid,
			want:     generateAuthRecord(t),
			wantErr:  false,
		},
		{
			name: "case get record, but server return error",
			us: &models.User{
				ID:           uuid,
				Login:        uuid,
				PasswordHash: hp,
			},
			recordID: "",
			want:     nil,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				mock.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, errSomethingWentWrong)
			} else {
				rpb, err := convRecordToProtobuff(tt.want)
				if err != nil {
					t.Errorf("an error occured while convert record to protobuff, err: %v", err)
				}
				mock.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&GetRecordResponse{
					Record: rpb,
				}, nil)
			}
			got, err := c.Get(ctx, tt.us.ID, tt.recordID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GKClient.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.ID != tt.want.ID {
				t.Errorf("GKClient.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGKClient_Delete(t *testing.T) {
	ctx := context.Background()
	log := zap.L()

	ctrl := gomock.NewController(t)
	mock := NewMockRecordsServer(ctrl)

	cfg := config.NewClientCfg()
	c := &GKClient{
		addr:     cfg.GKeeper,
		log:      log,
		certPath: cfg.CertFilePath,
	}
	opts := c.getDialOpts()
	lis := NewRecordsSrvListener(mock)
	creds, err := getClientCreds("")
	if err != nil {
		t.Errorf("an error occured while get client gredentials, err: %v", err)
	}
	opts = append(opts, grpc.WithContextDialer(lis), grpc.WithTransportCredentials(creds))

	conn, err := grpc.DialContext(ctx, "", opts...)
	if err != nil {
		t.Errorf("an occured error when getting conn grpc client, err: %v", err)
	}
	defer conn.Close()

	c.cc = conn

	uuid := uuid.NewString()
	hp, err := hashPassword(uuid)
	if err != nil {
		t.Errorf("an error occured whil retrieving hash password, err: %v", err)
	}

	tests := []struct {
		name     string
		us       *models.User
		recordID string
		want     *models.Record
		wantErr  bool
	}{
		{
			name: "case delete record",
			us: &models.User{
				ID:           uuid,
				Login:        uuid,
				PasswordHash: hp,
			},
			recordID: uuid,
			want:     generateAuthRecord(t),
			wantErr:  false,
		},
		{
			name: "case delete record, but server return error",
			us: &models.User{
				ID:           uuid,
				Login:        uuid,
				PasswordHash: hp,
			},
			recordID: "",
			want:     nil,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				mock.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil, errSomethingWentWrong)
			} else {
				mock.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(&DeleteRecordResponse{}, nil)
			}
			err := c.Delete(ctx, tt.us.ID, tt.recordID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GKClient.Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestGKClient_Add(t *testing.T) {
	ctx := context.Background()
	log := zap.L()

	ctrl := gomock.NewController(t)
	mock := NewMockRecordsServer(ctrl)

	cfg := config.NewClientCfg()
	c := &GKClient{
		addr:     cfg.GKeeper,
		log:      log,
		certPath: cfg.CertFilePath,
	}
	opts := c.getDialOpts()
	lis := NewRecordsSrvListener(mock)
	creds, err := getClientCreds("")
	if err != nil {
		t.Errorf("an error occured while get client gredentials, err: %v", err)
	}
	opts = append(opts, grpc.WithContextDialer(lis), grpc.WithTransportCredentials(creds))

	conn, err := grpc.DialContext(ctx, "", opts...)
	if err != nil {
		t.Errorf("an occured error when getting conn grpc client, err: %v", err)
	}
	defer conn.Close()

	c.setupConn(ctx)
	c.cc = conn

	uuid := uuid.NewString()
	hp, err := hashPassword(uuid)
	if err != nil {
		t.Errorf("an error occured whil retrieving hash password, err: %v", err)
	}
	r := generateAuthRecord(t)

	tests := []struct {
		name    string
		us      *models.User
		record  *models.RecordDTO
		want    *models.Record
		wantErr bool
	}{
		{
			name: "case add record",
			us: &models.User{
				ID:           uuid,
				Login:        uuid,
				PasswordHash: hp,
			},
			record: &models.RecordDTO{
				Description: r.Description,
				Type:        r.Type,
				Data:        r.Data,
				Hashsum:     r.Hashsum,
				Metadata:    r.Metadata,
			},
			want:    r,
			wantErr: false,
		},
		{
			name: "case add record, but server return error",
			us: &models.User{
				ID:           uuid,
				Login:        uuid,
				PasswordHash: hp,
			},
			record: &models.RecordDTO{
				Description: r.Description,
				Type:        r.Type,
				Data:        r.Data,
				Hashsum:     r.Hashsum,
				Metadata:    r.Metadata,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				mock.EXPECT().Add(gomock.Any(), gomock.Any()).Return(nil, errSomethingWentWrong)
			} else {
				mock.EXPECT().Add(gomock.Any(), gomock.Any()).Return(&AddRecordResponse{
					Id: r.ID,
				}, nil)
			}
			got, err := c.Add(ctx, tt.us.ID, tt.record)
			if (err != nil) != tt.wantErr {
				t.Errorf("GKClient.Add() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Description != tt.want.Description {
				t.Errorf("GKClient.Add() = %v, want %v", got.Description, tt.want.Description)
			}
		})
	}
}

func TestGKClient_Update(t *testing.T) {
	ctx := context.Background()
	log := zap.L()

	ctrl := gomock.NewController(t)
	mock := NewMockRecordsServer(ctrl)

	cfg := config.NewClientCfg()
	c := &GKClient{
		addr:     cfg.GKeeper,
		log:      log,
		certPath: cfg.CertFilePath,
	}
	opts := c.getDialOpts()
	lis := NewRecordsSrvListener(mock)
	creds, err := getClientCreds("")
	if err != nil {
		t.Errorf("an error occured while get client gredentials, err: %v", err)
	}
	opts = append(opts, grpc.WithContextDialer(lis), grpc.WithTransportCredentials(creds))

	conn, err := grpc.DialContext(ctx, "", opts...)
	if err != nil {
		t.Errorf("an occured error when getting conn grpc client, err: %v", err)
	}
	defer conn.Close()

	c.cc = conn

	uuid := uuid.NewString()
	hp, err := hashPassword(uuid)
	if err != nil {
		t.Errorf("an error occured whil retrieving hash password, err: %v", err)
	}
	r := generateAuthRecord(t)

	tests := []struct {
		name    string
		us      *models.User
		record  *models.Record
		want    *models.Record
		wantErr bool
	}{
		{
			name: "case update record",
			us: &models.User{
				ID:           uuid,
				Login:        uuid,
				PasswordHash: hp,
			},
			record:  r,
			want:    r,
			wantErr: false,
		},
		{
			name: "case update record, but server return error",
			us: &models.User{
				ID:           uuid,
				Login:        uuid,
				PasswordHash: hp,
			},
			record:  r,
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				mock.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, errSomethingWentWrong)
			} else {
				mock.EXPECT().Update(gomock.Any(), gomock.Any()).Return(&UpdateRecordResponse{
					Id: r.ID,
				}, nil)
			}
			got, err := c.Update(ctx, tt.us.ID, tt.record)
			if (err != nil) != tt.wantErr {
				t.Errorf("GKClient.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.ID != tt.want.ID {
				t.Errorf("GKClient.Update() = %v, want %v", got.Description, tt.want.Description)
			}
		})
	}
}
