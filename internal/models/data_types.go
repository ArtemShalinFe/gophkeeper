package models

import (
	"fmt"
	"time"

	"github.com/fxamacker/cbor/v2"
)

// RecordData - The interface that the types of stored data should implement.
type RecordData interface {
	// BinData - Must return binary data or error if something wen wrong.
	BinData() ([]byte, error)
}

// Auth - encoded username and password. Identifies the Auth type.
type Auth struct {
	Login    string `cbor:"login"`
	Password string `cbor:"password"`
}

func (a *Auth) BinData() ([]byte, error) {
	b, err := cbor.Marshal(a)
	if err != nil {
		return nil, fmt.Errorf("an error occured while encode auth data to bytes, err: %w", err)
	}
	return b, nil
}

// Text - contains strings. Lots of lines.
type Text struct {
	Data string `cbor:"data"`
}

func (t *Text) BinData() ([]byte, error) {
	b, err := cbor.Marshal(t)
	if err != nil {
		return nil, fmt.Errorf("an error occured while encode text data to bytes, err: %w", err)
	}
	return b, nil
}

// Binary - the file data.
type Binary struct {
	Name string `cbor:"name"`
	Ext  string `cbor:"ext"`
	Data []byte `cbor:"data"`
}

func (bin *Binary) BinData() ([]byte, error) {
	b, err := cbor.Marshal(bin)
	if err != nil {
		return nil, fmt.Errorf("an error occured while encode bin data to bytes, err: %w", err)
	}
	return b, nil
}

// Card - is the bank card details including: number, term and owner. Cvv code is not stored.
type Card struct {
	Number string    `cbor:"number"`
	Term   time.Time `cbor:"term"`
	Owner  string    `cbor:"owner"`
}

func (c *Card) BinData() ([]byte, error) {
	b, err := cbor.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("an error occured while encode card data to bytes, err: %w", err)
	}
	return b, nil
}
