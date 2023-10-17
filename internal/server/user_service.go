package server

import (
	context "context"

	"go.uber.org/zap"
)

type UserService struct {
	UnimplementedUsersServer
	log *zap.Logger
}

func NewMetricService(log *zap.Logger) *UserService {
	return &UserService{
		log: log,
	}
}

func (us *UserService) Register(ctx context.Context, request *RegisterRequest) (*RegisterResponse, error) {
	var resp RegisterResponse

	return &resp, nil
}

func (us *UserService) Login(ctx context.Context, request *LoginRequest) (*LoginResponse, error) {
	var resp LoginResponse

	return &resp, nil
}
