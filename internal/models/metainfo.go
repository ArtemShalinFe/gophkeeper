package models

import (
	"fmt"
	"strings"
)

type Metainfo struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func NewMetainfoFromStringArray(ss []string) ([]*Metainfo, error) {
	mi := make([]*Metainfo, len(ss))
	const kvCount = 2

	for i := 0; i < len(ss); i++ {
		s := ss[i]
		sp := strings.Split(s, ":")
		if len(sp) != kvCount {
			return nil, fmt.Errorf("an error occured while parse metainfo %s", s)
		}
		mi[i] = &Metainfo{
			Key:   sp[0],
			Value: sp[1],
		}
	}

	return mi, nil
}
