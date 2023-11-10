package server

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/ArtemShalinFe/gophkeeper/internal/config"
	"github.com/ArtemShalinFe/gophkeeper/internal/models"
)

// GKClient - a grpc client that works with the gophkeeper server.
type GKClient struct {
	cc  grpc.ClientConnInterface
	log *zap.Logger
	// addr - an address of gophekeeper server.
	addr string
	// certpath - absolute path to cert.crt file
	certPath string
}

// NewGKClient - Object Constructor.
func NewGKClient(ctx context.Context, cfg *config.ClientCfg, log *zap.Logger) (*GKClient, error) {
	c := &GKClient{
		addr:     cfg.GKeeper,
		log:      log,
		certPath: cfg.CertFilePath,
	}

	if err := c.setupConn(ctx); err != nil {
		return nil, fmt.Errorf("an error occured while setup conn to server, err: %w", err)
	}

	return c, nil
}

func getClientCreds(certFilePath string) (credentials.TransportCredentials, error) {
	if certFilePath == "" {
		creds := insecure.NewCredentials()
		return creds, nil
	}

	creds, err := credentials.NewClientTLSFromFile(
		certFilePath,
		"")
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials: %w", err)
	}
	return creds, nil
}

func (c *GKClient) setupConn(ctx context.Context) error {
	opts := c.getDialOpts()

	creds, err := getClientCreds(c.certPath)
	if err != nil {
		return fmt.Errorf("an error occured when retrieving client credentials: %w", err)
	}

	opts = append(opts, grpc.WithTransportCredentials(creds))
	conn, err := grpc.DialContext(ctx, c.addr, opts...)
	if err != nil {
		return fmt.Errorf("server is not available at %s, err: %w", c.addr, err)
	}

	c.cc = conn

	return nil
}

func (c *GKClient) getDialOpts() []grpc.DialOption {
	var opts []grpc.DialOption
	const defaultBackoff = 2
	const defaultAttempt = 3

	retryopts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffLinear(defaultBackoff * time.Second)),
		grpc_retry.WithCodes(grpc_retry.DefaultRetriableCodes...),
		grpc_retry.WithMax(defaultAttempt),
	}

	chain := grpc.WithChainUnaryInterceptor(
		grpc_retry.UnaryClientInterceptor(retryopts...),
	)

	opts = append(opts, chain)

	return opts
}

// AddUser - The method is used when registering a user.
func (c *GKClient) AddUser(ctx context.Context, us *models.UserDTO) (*models.User, error) {
	resp, err := NewUsersClient(c.cc).Register(ctx, &RegisterRequest{
		Login:    us.Login,
		Password: us.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("an error occured while register user at server, err: %w", err)
	}

	return &models.User{
		ID:           resp.GetUser().GetId(),
		Login:        us.Login,
		PasswordHash: us.Password,
	}, nil
}

// GetUser - This method is used when the user logs in.
func (c *GKClient) GetUser(ctx context.Context, us *models.UserDTO) (*models.User, error) {
	resp, err := NewUsersClient(c.cc).Login(ctx, &LoginRequest{
		Login:    us.Login,
		Password: us.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("an error occured while logged in user, err: %w", err)
	}

	return &models.User{
		ID:           resp.GetUser().GetId(),
		Login:        us.Login,
		PasswordHash: us.Password,
	}, nil
}

// ListRecords - used to retrieving user records.
func (c *GKClient) ListRecords(ctx context.Context, userID string, offset int, limit int) ([]*models.Record, error) {
	serverStorage := NewRecordsClient(c.cc)

	headers := map[string]string{
		userIDHeader: userID,
	}

	mctx := metadata.NewOutgoingContext(ctx, metadata.New(headers))
	req := &ListRecordRequest{}
	req.Offset = int32(offset)
	req.Limit = int32(limit)

	lr, err := serverStorage.ListRecords(mctx, req)
	if err != nil {
		return nil, fmt.Errorf("an error occured while retrieving list records, err: %w", err)
	}

	rs := make([]*models.Record, len(lr.Records))
	for i := 0; i < len(lr.Records); i++ {
		r := lr.Records[i]

		rc, err := convRecordFromProtobuff(r)
		if err != nil {
			return nil, fmt.Errorf("an error occured while converting record from protobuff, err: %w", err)
		}

		rs[i] = rc
	}

	return rs, nil
}

// GetRecord - used to retrieving record.
func (c *GKClient) GetRecord(ctx context.Context, userID string, recordID string) (*models.Record, error) {
	serverStorage := NewRecordsClient(c.cc)

	headers := map[string]string{
		userIDHeader: userID,
	}

	mctx := metadata.NewOutgoingContext(ctx, metadata.New(headers))
	req := &GetRecordRequest{Id: recordID}
	rr, err := serverStorage.GetRecord(mctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == codes.NotFound {
				return nil, models.ErrRecordNotFound
			}
			return nil, fmt.Errorf("an error occured while retrieving record, err: %w", err)
		}
		return nil, fmt.Errorf("an error occured while retrieving status for record, err: %w", err)
	}

	r, err := convRecordFromProtobuff(rr.Record)
	if err != nil {
		return nil, fmt.Errorf("an error occured while encode record to protobuff, err: %w", err)
	}

	return r, nil
}

// DeleteRecord - mark records as deleted.
func (c *GKClient) DeleteRecord(ctx context.Context, userID string, recordID string) error {
	serverStorage := NewRecordsClient(c.cc)

	headers := map[string]string{
		userIDHeader: userID,
	}

	mctx := metadata.NewOutgoingContext(ctx, metadata.New(headers))
	req := &DeleteRecordRequest{Id: recordID}
	_, err := serverStorage.DeleteRecord(mctx, req)
	if err != nil {
		return fmt.Errorf("an error occured while removing record, err: %w", err)
	}

	return nil
}

// AddRecord - add new record to the storage.
func (c *GKClient) AddRecord(ctx context.Context, userID string, record *models.RecordDTO) (*models.Record, error) {
	if len(record.Data) > models.MaxFileSize {
		return nil, errors.New(models.ErrLargeFile)
	}
	serverStorage := NewRecordsClient(c.cc)

	headers := map[string]string{
		userIDHeader: userID,
	}

	mctx := metadata.NewOutgoingContext(ctx, metadata.New(headers))

	id := uuid.NewString()
	now := time.Now()
	r := &models.Record{
		ID:          id,
		Owner:       userID,
		Description: record.Description,
		Type:        record.Type,
		Created:     now,
		Modified:    now,
		Data:        record.Data,
		Hashsum:     record.Hashsum,
		Metadata:    record.Metadata,
		Version:     1,
	}

	rpb, err := convRecordToProtobuff(r)
	if err != nil {
		return nil, fmt.Errorf("an error add occured while convert record to protobuff, err: %w", err)
	}
	_, err = serverStorage.AddRecord(mctx, &AddRecordRequest{
		Record: rpb,
	})
	if err != nil {
		return nil, fmt.Errorf("an error occured while adding record, err: %w", err)
	}

	return r, nil
}

// UpdateRecord - Update record to the storage.
func (c *GKClient) UpdateRecord(ctx context.Context, userID string, record *models.Record) (*models.Record, error) {
	serverStorage := NewRecordsClient(c.cc)

	headers := map[string]string{
		userIDHeader: userID,
	}

	mctx := metadata.NewOutgoingContext(ctx, metadata.New(headers))

	rpb, err := convRecordToProtobuff(record)
	if err != nil {
		return nil, fmt.Errorf("an error update occured while convert record to protobuff, err: %w", err)
	}
	_, err = serverStorage.UpdateRecord(mctx, &UpdateRecordRequest{
		Record: rpb,
	})
	if err != nil {
		return nil, fmt.Errorf("an error occured while updating record, err: %w", err)
	}

	return record, nil
}
