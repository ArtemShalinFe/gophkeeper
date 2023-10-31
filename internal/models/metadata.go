package models

import (
	"strconv"
	"strings"
)

type Metadata struct {
	Key   string `cbor:"key"`
	Value string `cbor:"value"`
}

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
