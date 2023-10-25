package models

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"time"
)

var MaxFileSize = 20 * 1024 * 1024

var ErrRecordNotFound = errors.New("record not found")
var ErrUserStorageNotFound = errors.New("user cache not found")

type DataType string

const (
	AuthType   DataType = "auth"
	TextType   DataType = "text"
	BinaryType DataType = "binary"
	CardType   DataType = "card"
)

type RecordDTO struct {
	Description string      `json:"description"`
	Type        string      `json:"type"`
	Data        []byte      `json:"data"`
	Hashsum     string      `json:"hashsum"`
	Metainfo    []*Metainfo `json:"metainfo"`
}

func NewRecordDTO(description string, dataType DataType, data Byter, metainfo []*Metainfo) (*RecordDTO, error) {
	b, err := data.ToByte()
	if err != nil {
		return nil, fmt.Errorf("an error occured while converted data record dto to bytes, err: %w", err)
	}

	hs, err := hashsum(b)
	if err != nil {
		return nil, fmt.Errorf("an error occured while create record DTO, err: %w", err)
	}
	return &RecordDTO{
		Description: description,
		Type:        string(dataType),
		Data:        b,
		Hashsum:     hs,
		Metainfo:    metainfo,
	}, nil
}

type Record struct {
	ID          string      `json:"uuid"`
	Owner       string      `json:"user"`
	Description string      `json:"description"`
	Type        string      `json:"type"`
	Created     time.Time   `json:"created"`
	Modified    time.Time   `json:"modified"`
	Data        []byte      `json:"data"`
	Hashsum     string      `json:"hashsum"`
	Metainfo    []*Metainfo `json:"metainfo"`
	Deleted     bool        `json:"deleted"`
	Version     int         `json:"version"`
}

func NewRecord(id string,
	description string,
	dataType DataType,
	created time.Time,
	modified time.Time,
	data Byter,
	metainfo []*Metainfo,
	deleted bool,
	version int) (*Record, error) {
	b, err := data.ToByte()
	if err != nil {
		return nil, fmt.Errorf("an error occured while converted data record to bytes, err: %w", err)
	}

	hs, err := hashsum(b)
	if err != nil {
		return nil, fmt.Errorf("an error occured while create record, err: %w", err)
	}
	return &Record{
		ID:          id,
		Description: description,
		Type:        string(dataType),
		Created:     created,
		Modified:    modified,
		Data:        b,
		Hashsum:     hs,
		Metainfo:    metainfo,
		Deleted:     deleted,
		Version:     version,
	}, nil
}

func RecordStringHeader() string {
	return "ID\tDESCRIPTION\tTYPE\tCREATED\tMODIFIED\tHASHSUM\tDELETED\tVERSION"
}

func (r *Record) String() string {
	return fmt.Sprintf("%s\t%s\t%s\t%v\t%v\t%s\t%v\t%d",
		r.ID, r.Description, r.Type, r.Created, r.Modified, r.Hashsum, r.Deleted, r.Version)
}

type Vector interface {
	Increment(*Record)
	Clone() []*Record
	IsSame(*Record) bool
	IsLower(*Record) bool
	IsHigher(*Record) bool
}

func (r *Record) Increment(stg RecordStorage) {
	r.Version++
}

type RecordStorage interface {
	List(ctx context.Context, userID string) ([]*Record, error)
	Get(ctx context.Context, userID string, recordID string) (*Record, error)
	Delete(ctx context.Context, userID string, recordID string) error
	Add(ctx context.Context, userID string, record *RecordDTO) (*Record, error)
	Update(ctx context.Context, userID string, record *Record) (*Record, error)
}

func hashsum(b []byte) (string, error) {
	h := sha256.New()
	f := bytes.NewReader(b)
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("an error occured while calculate hashsum, err: %w", err)
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
