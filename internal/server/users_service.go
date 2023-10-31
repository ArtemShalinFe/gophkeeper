package server

import (
	context "context"
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

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

	hp, err := hashPassword(udto.Password)
	if err != nil {
		return &resp, fmt.Errorf("an error occured during registration, err: %w", err)
	}
	udto.Password = hp

	u, err := udto.AddUser(ctx, us.userStorage)
	if err != nil {
		return &resp, fmt.Errorf("registration, err: %w", err)
	}

	resp.User = &User{Id: u.ID}
	return &resp, nil
}

func (us *UsersService) Login(ctx context.Context, request *LoginRequest) (*LoginResponse, error) {
	var resp LoginResponse

	u := getUserDTOFromRequest(request)

	user, err := u.GetUser(ctx, us.userStorage)
	if err != nil {
		return &resp, fmt.Errorf("logged in, err: %w", err)
	}

	if !checkPasswordHash(user.PasswordHash, u.Password) {
		return &resp, models.ErrUnknowUser
	}

	resp.User = &User{Id: user.ID}
	return &resp, nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("an error occured while retrieving hash from password err: %w", err)
	}
	return string(bytes), nil
}

func checkPasswordHash(hash string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func getUserDTOFromRequest(request userRequest) *models.UserDTO {
	var udto models.UserDTO
	udto.Login = request.GetLogin()
	udto.Password = request.GetPassword()

	return &udto
}
