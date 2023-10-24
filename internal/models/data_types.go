package models

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"
)

var errEncodeToByteTemplate = "an error occured while encode to bytes, err: %w"
var errDecodeToByteTemplate = "an error occured while decode to bytes, err: %w"

type Byter interface {
	ToByte() ([]byte, error)
	FromByte([]byte) error
}

type Auth struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (a *Auth) ToByte() ([]byte, error) {
	b, err := encode(a)
	if err != nil {
		return nil, fmt.Errorf(errEncodeToByteTemplate, err)
	}
	return b, nil
}

func (a *Auth) FromByte(b []byte) error {
	if err := decode(a, b); err != nil {
		return fmt.Errorf(errDecodeToByteTemplate, err)
	}
	return nil
}

type Text struct {
	Data string `json:"data"`
}

func (t *Text) ToByte() ([]byte, error) {
	b, err := encode(t.Data)
	if err != nil {
		return nil, fmt.Errorf(errEncodeToByteTemplate, err)
	}
	return b, nil
}

func (t *Text) FromByte(b []byte) error {
	if err := decode(t.Data, b); err != nil {
		return fmt.Errorf(errDecodeToByteTemplate, err)
	}
	return nil
}

type Binary struct {
	Data []byte `json:"data"`
}

func (b *Binary) ToByte() ([]byte, error) {
	bs, err := encode(b.Data)
	if err != nil {
		return nil, fmt.Errorf(errEncodeToByteTemplate, err)
	}
	return bs, nil
}

func (b *Binary) FromByte(bs []byte) error {
	if err := decode(b.Data, bs); err != nil {
		return fmt.Errorf(errDecodeToByteTemplate, err)
	}
	return nil
}

type Card struct {
	Number string    `json:"number"`
	Term   time.Time `json:"term"`
	Owner  string    `json:"owner"`
}

func (c *Card) ToByte() ([]byte, error) {
	bs, err := encode(c)
	if err != nil {
		return nil, fmt.Errorf(errEncodeToByteTemplate, err)
	}
	return bs, nil
}

func (c *Card) FromByte(b []byte) error {
	if err := decode(c, b); err != nil {
		return fmt.Errorf(errDecodeToByteTemplate, err)
	}
	return nil
}

func encode(val any) ([]byte, error) {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(val)
	if err != nil {
		return nil, fmt.Errorf("an error occurred when encode the value to bytes, err: %w", err)
	}
	return buff.Bytes(), nil
}

func decode(val any, b []byte) error {
	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&val); err != nil {
		return fmt.Errorf("an error occured while decode from binary, err: %w", err)
	}

	return nil
}
