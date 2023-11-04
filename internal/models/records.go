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

// MaxFileSize - The maximum size of the record data that the service can accept.
var MaxFileSize = 40 * 1024 * 1024 // 40 Mb

// ErrLargeFile - An error that is returned in case of an attempt to transfer data whose size exceeds
// the size set in the constant MaxFileSize.
var ErrLargeFile = fmt.Sprintf("the file size should not exceed %d bytes", MaxFileSize)

// ErrEmptyID - An error that is returned in case of empty ID.
var ErrEmptyID = errors.New("field ID cannot be empty")

// ErrRecordNotFound - An error that is returned in case when record not found in database.
var ErrRecordNotFound = errors.New("record not found")

// ErrUserStorageNotFound - An error that is returned when the user is logged in,
// but the client has not created a storage for him.
var ErrUserStorageNotFound = errors.New("user cache not found")

type RecordStorage interface {
	// List - used to retrieving user records.
	List(ctx context.Context, userID string, offset int, limit int) ([]*Record, error)
	// Get - used to retrieving record.
	Get(ctx context.Context, userID string, recordID string) (*Record, error)
	// Delete - mark records as deleted.
	Delete(ctx context.Context, userID string, recordID string) error
	// Add - add new record to the storage.
	Add(ctx context.Context, userID string, record *RecordDTO) (*Record, error)
	// Update - Update record to the storage.
	Update(ctx context.Context, userID string, record *Record) (*Record, error)
}

// DataType - the object contains the data directly related to the record.
// There can be four types in the current implementation:
//   - AUTH - Encoded username and password. Identifies the Auth type.
//   - TEXT -  Identifies the Text type.
//   - BINARY - file data. Identifies the Binary type.
//   - CARD - Bank card details including number, term and owner. cvv code is not stored.
type DataType string

const (
	// AuthType - encoded username and password. Identifies the Auth type.
	AuthType DataType = "AUTH"
	// TextType - contains strings. Lots of lines.
	TextType DataType = "TEXT"
	// BinaryType - the file data.
	BinaryType DataType = "BINARY"
	// CardType - is the bank card details including: number, term and owner. cvv code is not stored.
	CardType DataType = "CARD"
)

// RecordDTO - Data transfer object for Record.
type RecordDTO struct {
	// Description - contains various kinds of descriptive information depending on the file type.
	//   - If the file is stored in a record, the field contains the name of the file;
	//   - In all other cases, the user is free to specify the description himself.
	Description string `cbor:"description"`
	// Type - allows you to determine the type of stored content in the file.
	Type string `cbor:"type"`
	// Data - the field contains the data directly related to the record.
	// There can be four types in the current implementation:
	//   - Auth - Encoded username and password. Identifies the Auth type.
	//   - Text -  Identifies the Text type.
	//   - Binary - file data. Identifies the Binary type.
	//   - Card - Bank card details including number, term and owner. Cvv code is not stored.
	Data []byte `cbor:"data"`
	// Hashsum - the hash sum of the data that is stored in the repository.
	Hashsum string `cbor:"hashsum"`
	// Metadata - for storing arbitrary textual meta-information
	// (data belonging to a website, an individual or a bank, lists of one-time activation codes, etc.).
	Metadata []*Metadata `cbor:"metadata"`
}

// NewRecordDTO - Object Constructor. The constructor of the object. Automatically calculates the hashsum of data.
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

// Record - The object contains consolidated information about the user's data record in the database on the server.
type Record struct {
	// ID - uuid record in database storage.
	ID string `cbor:"uuid"`
	// Owner - id of the user owner of the file in the database.
	Owner string `cbor:"user"`
	// Description - contains various kinds of descriptive information depending on the file type.
	//   - If the file is stored in a record, the field contains the name of the file;
	//   - In all other cases, the user is free to specify the description himself.
	Description string `cbor:"description"`
	// Type - allows you to determine the type of stored content in the file.
	Type string `cbor:"type"`
	// Created - indicates the date of creation of the record in the repository.
	Created time.Time `cbor:"created"`
	// Modified - indicates the date when the record in the repository was changed.
	Modified time.Time `cbor:"modified"`
	// Data - the field contains the data directly related to the record.
	// There can be four types in the current implementation:
	//   - Auth - Encoded username and password. Identifies the Auth type.
	//   - Text -  Identifies the Text type.
	//   - Binary - file data. Identifies the Binary type.
	//   - Card - Bank card details including number, term and owner. Cvv code is not stored.
	Data []byte `cbor:"data"`
	// Hashsum - the hash sum of the data that is stored in the repository.
	Hashsum string `cbor:"hashsum"`
	// Metadata - for storing arbitrary textual meta-information
	// (data belonging to a website, an individual or a bank, lists of one-time activation codes, etc.).
	Metadata []*Metadata `cbor:"metadata"`
	// Deleted - this flag indicates that the file has been deleted.
	Deleted bool `cbor:"deleted"`
	// Version - file version.
	Version int64 `cbor:"version"`
}

// NewRecord - Object Constructor. The constructor of the object. Automatically calculates the hashsum of data.
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

// GetVersion - Returns the version number of the record.
func (r *Record) GetVersion() int64 {
	return r.Version
}

// GetHashsum - Returns the hash sum of the record data.
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
