package client

import (
	"context"
	"fmt"
	"net/url"

	"github.com/ArtemShalinFe/gophkeeper/internal/models"
	"github.com/go-resty/resty/v2"
)

const authHeaderName = "Authorization"
const recordsPath = "/api/v1/records"

type GAgentClient struct {
	client *resty.Client
	Addr   string
	token  string
}

func NewGAgentClient() *GAgentClient {
	return &GAgentClient{
		client: resty.New(),
	}
}

func (c *GAgentClient) Register(ctx context.Context, u *models.UserDTO) (string, error) {
	p, err := url.JoinPath(c.Addr, "/api/v1/register")
	if err != nil {
		return "", fmt.Errorf("an error occured while join register path, err: %w", err)
	}

	resp, err := c.client.R().
		SetBody(u).
		SetContext(ctx).
		Post(p)
	if err != nil {
		return "", fmt.Errorf("an error occured while register user, err: %w", err)
	}
	return resp.Header().Get(authHeaderName), nil
}

func (c *GAgentClient) Login(ctx context.Context, u *models.UserDTO) (string, error) {
	p, err := url.JoinPath(c.Addr, "/api/v1/login")
	if err != nil {
		return "", fmt.Errorf("an error occured while join login path, err: %w", err)
	}

	resp, err := c.client.R().
		SetBody(u).
		SetContext(ctx).
		Post(p)
	if err != nil {
		return "", fmt.Errorf("an error occured while log in user, err: %w", err)
	}
	return resp.Header().Get(authHeaderName), nil
}

func (c *GAgentClient) GetRecords(ctx context.Context) ([]*models.Record, error) {
	var rs []*models.Record

	p, err := url.JoinPath(c.Addr, recordsPath)
	if err != nil {
		return nil, fmt.Errorf("an error occured while join records path, err: %w", err)
	}

	_, err = c.client.R().
		SetHeader(authHeaderName, c.token).
		SetResult(rs).
		SetContext(ctx).
		Get(p)
	if err != nil {
		return nil, fmt.Errorf("an error occured while retrieving user records, err: %w", err)
	}

	return rs, nil
}

func (c *GAgentClient) GetRecord(ctx context.Context, recordID string) (*models.Record, error) {
	var r *models.Record

	p, err := url.JoinPath(c.Addr, recordsPath, recordID)
	if err != nil {
		return nil, fmt.Errorf("an error occured while join record path, err: %w", err)
	}

	_, err = c.client.R().
		SetHeader(authHeaderName, c.token).
		SetResult(r).
		SetContext(ctx).
		Get(p)
	if err != nil {
		return nil, fmt.Errorf("an error occured while retrieving user record, err: %w", err)
	}

	return r, nil
}

func (c *GAgentClient) AddRecord(ctx context.Context, r *models.RecordDTO) (*models.Record, error) {
	record, err := c.addOrUpdateRecord(ctx, r, false)
	if err != nil {
		return nil, fmt.Errorf("an error occured while add user record, err: %w", err)
	}

	return record, nil
}

func (c *GAgentClient) UpdateRecord(ctx context.Context, r *models.Record) (*models.Record, error) {
	record, err := c.addOrUpdateRecord(ctx, r, true)
	if err != nil {
		return nil, fmt.Errorf("an error occured while updating user record, err: %w", err)
	}

	return record, nil
}

func (c *GAgentClient) addOrUpdateRecord(ctx context.Context, body any, update bool) (*models.Record, error) {
	p, err := url.JoinPath(c.Addr, recordsPath)
	if err != nil {
		return nil, fmt.Errorf("an error occured while join update path, err: %w", err)
	}

	record := &models.Record{}
	req := c.client.R().
		SetHeader(authHeaderName, c.token).
		SetBody(body).
		SetResult(record).
		SetContext(ctx)

	if update {
		if _, err := req.Put(p); err != nil {
			return nil, fmt.Errorf("an error occured while update record, err: %w", err)
		}
		return record, nil
	}

	if _, err := req.Post(p); err != nil {
		return nil, fmt.Errorf("an error occured while add record, err: %w", err)
	}

	return record, nil
}

func (c *GAgentClient) DeleteRecord(ctx context.Context, recordID string) error {
	p, err := url.JoinPath(c.Addr, recordsPath, recordID)
	if err != nil {
		return fmt.Errorf("an error occured while join delete path, err: %w", err)
	}

	_, err = c.client.R().
		SetHeader(authHeaderName, c.token).
		SetContext(ctx).
		Delete(p)
	if err != nil {
		return fmt.Errorf("an error occured while removing user record, err: %w", err)
	}

	return nil
}
