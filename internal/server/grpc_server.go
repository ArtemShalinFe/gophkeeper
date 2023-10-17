package server

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/ArtemShalinFe/gophkeeper/internal/config"
)

type GRPCServer struct {
	grpcServer *grpc.Server
	log        *zap.Logger
	addr       string
}

func InitServer(cfg *config.ServerCfg) (*GRPCServer, error) {
	srv := &GRPCServer{
		addr: cfg.Addr,
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

func (s *GRPCServer) requestLogger() grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)
		if err != nil {
			s.log.Error("an error occurred while processing RPC request", zap.Error(err))
		} else {
			md, _ := metadata.FromIncomingContext(ctx)
			size, err := responseSize(resp)
			if err != nil {
				s.log.Error("an error occurred while calculate response size", zap.Error(err))
			}
			s.log.Info("incomming request",
				zap.String("RPC method", info.FullMethod),
				zap.Any("headers", md),
				zap.Duration("duration", duration),
				zap.Any("body", req),
				zap.Int("size", size),
			)
		}
		return resp, nil
	}
}

func responseSize(val any) (int, error) {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(val)
	if err != nil {
		return 0, fmt.Errorf("an occured error when convert val to bytes, err: %w", err)
	}
	b := buff.Bytes()
	return binary.Size(b), nil
}
