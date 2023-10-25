package server

import (
	context "context"
	"fmt"

	"go.uber.org/zap"

	"github.com/ArtemShalinFe/gophkeeper/internal/models"
)

type UsersService struct {
	UnimplementedUsersServer
	log         *zap.Logger
	userStorage models.UserStorage
}

type userRequest interface {
	GetLogin() string
	GetPassword() string
}

func NewUsersService(log *zap.Logger, userStorage models.UserStorage) *UsersService {
	return &UsersService{
		log:         log,
		userStorage: userStorage,
	}
}

func (us *UsersService) Register(ctx context.Context, request *RegisterRequest) (*RegisterResponse, error) {
	var resp RegisterResponse

	udto := getUserDTOFromRequest(request)

	u, err := udto.AddUser(ctx, us.userStorage)
	if err != nil {
		er := fmt.Errorf("an error occurred during user registration, err: %w", err)
		resp.Error = er.Error()
		return &resp, nil
	}

	resp.User = &User{Id: u.ID}
	return &resp, nil
}

func (us *UsersService) Login(ctx context.Context, request *LoginRequest) (*LoginResponse, error) {
	var resp LoginResponse

	udto := getUserDTOFromRequest(request)

	u, err := udto.GetUser(ctx, us.userStorage)
	if err != nil {
		er := fmt.Errorf("an error occurred during user logged in, err: %w", err)
		resp.Error = er.Error()
		return &resp, nil
	}

	resp.User = &User{Id: u.ID}
	return &resp, nil
}

func getUserDTOFromRequest(request userRequest) *models.UserDTO {
	var udto models.UserDTO
	udto.Login = request.GetLogin()
	udto.Password = request.GetPassword()

	return &udto
}
