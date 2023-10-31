package models

import (
	"fmt"
	"time"

	"github.com/fxamacker/cbor/v2"
)

type RecordData interface {
	BinData() ([]byte, error)
}

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
