package models

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/fxamacker/cbor/v2"
)

var MaxFileSize = 40 * 1024 * 1024 // 40 Mb
var ErrLargeFile = fmt.Sprintf("the file size should not exceed %d bytes", MaxFileSize)

var ErrEmptyID = errors.New("field ID cannot be empty")
var ErrRecordNotFound = errors.New("record not found")
var ErrUserStorageNotFound = errors.New("user cache not found")

type RecordStorage interface {
	List(ctx context.Context, userID string, offset int, limit int) ([]*Record, error)
	Get(ctx context.Context, userID string, recordID string) (*Record, error)
	Delete(ctx context.Context, userID string, recordID string) error
	Add(ctx context.Context, userID string, record *RecordDTO) (*Record, error)
	Update(ctx context.Context, userID string, record *Record) (*Record, error)
}

type DataType string

const (
	AuthType   DataType = "AUTH"
	TextType   DataType = "TEXT"
	BinaryType DataType = "BINARY"
	CardType   DataType = "CARD"
)

type RecordDTO struct {
	Description string      `cbor:"description"`
	Type        string      `cbor:"type"`
	Data        []byte      `cbor:"data"`
	Hashsum     string      `cbor:"hashsum"`
	Metadata    []*Metadata `cbor:"metadata"`
}

func NewRecordDTO(description string, dataType DataType, data RecordData, metadata []*Metadata) (*RecordDTO, error) {
	b, err := cbor.Marshal(data)
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
		Metadata:    metadata,
	}, nil
}

type Record struct {
	ID          string      `cbor:"uuid"`
	Owner       string      `cbor:"user"`
	Description string      `cbor:"description"`
	Type        string      `cbor:"type"`
	Created     time.Time   `cbor:"created"`
	Modified    time.Time   `cbor:"modified"`
	Data        []byte      `cbor:"data"`
	Hashsum     string      `cbor:"hashsum"`
	Metadata    []*Metadata `cbor:"metadata"`
	Deleted     bool        `cbor:"deleted"`
	Version     int64       `cbor:"version"`
}

func NewRecord(id string,
	description string,
	dataType DataType,
	created time.Time,
	modified time.Time,
	data RecordData,
	metadata []*Metadata,
	deleted bool,
	version int64) (*Record, error) {
	b, err := cbor.Marshal(data)
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
		Metadata:    metadata,
		Deleted:     deleted,
		Version:     version,
	}, nil
}

func (r *Record) GetVersion() int64 {
	return r.Version
}

func (r *Record) GetHashsum() string {
	return r.Hashsum
}

func hashsum(b []byte) (string, error) {
	h := sha256.New()
	f := bytes.NewReader(b)
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("an error occured while calculate hashsum, err: %w", err)
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
