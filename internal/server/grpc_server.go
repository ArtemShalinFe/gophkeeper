package server

import (
	"context"
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/ArtemShalinFe/gophkeeper/internal/config"
	"github.com/ArtemShalinFe/gophkeeper/internal/models"
)

// GKServer - the gophkeeper server object.
type GKServer struct {
	grpcServer     *grpc.Server
	log            *zap.Logger
	UsersService   *UsersService
	RecordsService *RecordsService
	addr           string
}

// InitServer - Initiates the gophkeeper server object.
func InitServer(rs models.RecordStorage,
	us models.UserStorage,
	log *zap.Logger,
	cfg *config.ServerCfg) (*GKServer, error) {
	srv := &GKServer{
		addr:           cfg.Addr,
		log:            log,
		UsersService:   NewUsersService(log, us),
		RecordsService: NewRecordsService(log, rs),
	}

	creds, err := serverCreds(cfg)
	if err != nil {
		return nil, err
	}
	opt := grpc.ChainUnaryInterceptor(
		srv.requestLogger(),
	)

	srv.grpcServer = grpc.NewServer(grpc.Creds(creds), opt)

	return srv, nil
}

func serverCreds(cfg *config.ServerCfg) (credentials.TransportCredentials, error) {
	if cfg.CertFilePath != "" && cfg.PrivateCryptoKey != "" {
		creds, err := credentials.NewServerTLSFromFile(cfg.CertFilePath, cfg.PrivateCryptoKey)
		if err != nil {
			return nil, fmt.Errorf("an occured error when loading TLS keys: %w", err)
		}
		return creds, nil
	} else {
		creds := insecure.NewCredentials()
		return creds, nil
	}
}

func (s *GKServer) requestLogger() grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)
		if err != nil {
			s.log.Error("an error occurred while processing RPC request", zap.Error(err))
			return nil, fmt.Errorf("an error occured while handle request, err: %w", err)
		} else {
			md, _ := metadata.FromIncomingContext(ctx)
			s.log.Info("incomming request",
				zap.String("RPC method", info.FullMethod),
				zap.Any("headers", md),
				zap.Duration("duration", duration),
				zap.Any("body", req),
			)
		}
		return resp, nil
	}
}

// Serve - starts servicing incoming requests. Used for tests.
func (s *GKServer) Serve(lis net.Listener) error {
	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("an occured error when server serve request, err: %w", err)
	}
	return nil
}

// ListenAndServe - starts servicing incoming requests.
func (s *GKServer) ListenAndServe() error {
	listen, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("an occured error when trying listen address %s, err: %w", s.addr, err)
	}

	RegisterUsersServer(s.grpcServer, s.UsersService)
	RegisterRecordsServer(s.grpcServer, s.RecordsService)

	if err := s.Serve(listen); err != nil {
		return fmt.Errorf("an occured error when grpc server serve, err: %w", err)
	}

	return nil
}

// Shutdown - Implements a server graceful shutdown.
func (s *GKServer) Shutdown(ctx context.Context) {
	s.grpcServer.GracefulStop()
}
