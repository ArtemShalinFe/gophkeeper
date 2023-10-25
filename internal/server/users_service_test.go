package server

import (
	context "context"
	"errors"
	"net"
	"reflect"
	"testing"

	"github.com/ArtemShalinFe/gophkeeper/internal/config"
	"github.com/ArtemShalinFe/gophkeeper/internal/models"
	"github.com/google/uuid"
	gomock "go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const gophkeeper = "gophkeeper"

var errSomethingWentWrong = errors.New("something went wrong")
var randomUUID = uuid.NewString()

type userDialer struct {
	lis *bufconn.Listener
}

func (d *userDialer) bufDialer(context.Context, string) (net.Conn, error) {
	return d.lis.Dial()
}

func NewUserServiceDialer(t *testing.T, us models.UserStorage) (*userDialer, error) {
	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)

	log := zap.L()
	s, err := InitServer(nil, us, log, config.NewServerCfg())
	if err != nil {
		t.Fatalf("an occured error when initial grpc server, err: %v", err)
	}
	usrvc := NewUsersService(log, us)

	RegisterUsersServer(s.grpcServer, usrvc)

	go func() {
		if err := s.Serve(lis); err != nil {
			t.Errorf("server exited with error: %v", err)
		}
	}()

	return &userDialer{
		lis: lis,
	}, nil
}

func userDTO() *models.UserDTO {
	return &models.UserDTO{
		Login:    gophkeeper,
		Password: gophkeeper,
	}
}

func user() *models.User {
	return &models.User{
		ID:           randomUUID,
		Login:        gophkeeper,
		PasswordHash: gophkeeper,
	}
}

func TestUsersService_Register(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	stg := NewMockUserStorage(ctrl)
	stg.EXPECT().AddUser(gomock.Any(), userDTO()).Return(user(), nil)
	stg.EXPECT().AddUser(gomock.Any(), userDTO()).Return(nil, models.ErrLoginIsBusy)
	stg.EXPECT().AddUser(gomock.Any(), userDTO()).Return(nil, errSomethingWentWrong)

	d, err := NewUserServiceDialer(t, stg)
	if err != nil {
		t.Errorf("an occured error when creating a new dialer, err: %v", err)
	}

	tests := []struct {
		name    string
		request *RegisterRequest
		want    *RegisterResponse
		wantErr bool
	}{
		{
			name: "login is empty case",
			request: &RegisterRequest{
				Login:    "",
				Password: userDTO().Password,
			},
			want:    &RegisterResponse{},
			wantErr: true,
		},
		{
			name: "positive case",
			request: &RegisterRequest{
				Login:    userDTO().Login,
				Password: userDTO().Password,
			},
			want: &RegisterResponse{
				User: &User{
					Id: randomUUID,
				},
			},
			wantErr: false,
		},
		{
			name: "login is busy case",
			request: &RegisterRequest{
				Login:    userDTO().Login,
				Password: userDTO().Password,
			},
			want:    &RegisterResponse{},
			wantErr: true,
		},
		{
			name: "error case",
			request: &RegisterRequest{
				Login:    userDTO().Login,
				Password: userDTO().Password,
			},
			want:    &RegisterResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			conn, err := grpc.DialContext(ctx, "bufnet",
				grpc.WithContextDialer(d.bufDialer),
				grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				t.Errorf("failed to dial bufnet: %v", err)
			}
			defer conn.Close()

			client := NewUsersClient(conn)

			got, err := client.Register(ctx, tt.request)
			if (got.Error != "") != tt.wantErr {
				t.Errorf("UsersService.Register() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got.User, tt.want.User) {
				t.Errorf("UsersService.Register() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUsersService_Login(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	stg := NewMockUserStorage(ctrl)
	stg.EXPECT().GetUser(gomock.Any(), userDTO()).Return(user(), nil)
	stg.EXPECT().GetUser(gomock.Any(), userDTO()).Return(nil, models.ErrLoginIsBusy)
	stg.EXPECT().GetUser(gomock.Any(), userDTO()).Return(nil, errSomethingWentWrong)

	d, err := NewUserServiceDialer(t, stg)
	if err != nil {
		t.Errorf("an occured error when creating a new dialer, err: %v", err)
	}

	tests := []struct {
		name    string
		request *LoginRequest
		want    *LoginResponse
		wantErr bool
	}{
		{
			name: "login is empty case",
			request: &LoginRequest{
				Login:    "",
				Password: userDTO().Password,
			},
			want:    &LoginResponse{},
			wantErr: true,
		},
		{
			name: "positive case",
			request: &LoginRequest{
				Login:    userDTO().Login,
				Password: userDTO().Password,
			},
			want: &LoginResponse{
				User: &User{
					Id: randomUUID,
				},
			},
			wantErr: false,
		},
		{
			name: "login is busy case",
			request: &LoginRequest{
				Login:    userDTO().Login,
				Password: userDTO().Password,
			},
			want:    &LoginResponse{},
			wantErr: true,
		},
		{
			name: "error case",
			request: &LoginRequest{
				Login:    userDTO().Login,
				Password: userDTO().Password,
			},
			want:    &LoginResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			conn, err := grpc.DialContext(ctx, "bufnet",
				grpc.WithContextDialer(d.bufDialer),
				grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				t.Errorf("failed to dial bufnet: %v", err)
			}
			defer conn.Close()

			client := NewUsersClient(conn)

			got, err := client.Login(ctx, tt.request)
			if (got.Error != "") != tt.wantErr {
				t.Errorf("UsersService.Login() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got.User, tt.want.User) {
				t.Errorf("UsersService.Login() = %v, want %v", got, tt.want)
			}
		})
	}
}
