package models

import (
	"strconv"
	"strings"
)

// Metadata - for storing arbitrary textual meta-information
// (data belonging to a website, an individual or a bank, lists of one-time activation codes, etc.).
type Metadata struct {
	// Key - The record key may not be unique.
	Key string `cbor:"key"`
	// Value - The text value of the record.
	Value string `cbor:"value"`
}

// NewMetadataFromStringArray - The function converts an array of strings into an object Metadata.
// Example strings:
//   - "key1:value1" <- A key value pair separated by a colon.
//   - "keyvalue" <- Just a string, in this case a sequence number will be added to the key.
func NewMetadataFromStringArray(ss []string) ([]*Metadata, error) {
	var mi []*Metadata
	if len(ss) == 0 {
		return mi, nil
	}

	const kvCount = 2

	for i, s := range ss {
		sp := strings.Split(s, ":")
		switch len(sp) {
		case kvCount:
			mi = append(mi, &Metadata{
				Key:   sp[0],
				Value: sp[1],
			})
		case 1:
			mi = append(mi, &Metadata{
				Key:   strconv.Itoa(i),
				Value: sp[0],
			})
		default:
			continue
		}
	}

	return mi, nil
}
