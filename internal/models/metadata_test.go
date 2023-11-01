package models

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

func TestNewMetadataFromStringArray(t *testing.T) {
	type args struct {
		ss []string
	}
	s1 := uuid.NewString()
	s2 := uuid.NewString()
	s3 := uuid.NewString()

	var mi []*Metadata
	tests := []struct {
		name    string
		args    args
		want    []*Metadata
		wantErr bool
	}{
		{
			name: "positive case empty array",
			args: args{
				ss: []string{},
			},
			want:    mi,
			wantErr: false,
		},
		{
			name: "positive case",
			args: args{
				ss: []string{fmt.Sprintf("%s:%s", s1, s1),
					fmt.Sprintf("%s:%s", s2, s2),
					fmt.Sprintf("%s:%s", s3, s3)},
			},
			want: []*Metadata{{Key: s1, Value: s1},
				{Key: s2, Value: s2},
				{Key: s3, Value: s3},
			},
			wantErr: false,
		},
		{
			name: "positive case three elements",
			args: args{
				ss: []string{fmt.Sprintf("%s:%s", s1, s1),
					fmt.Sprintf("%s:%s", s2, s2),
					fmt.Sprintf("%s:%s:%s", s3, s3, s3)},
			},
			want: []*Metadata{{Key: s1, Value: s1},
				{Key: s2, Value: s2},
			},
			wantErr: false,
		},

		{
			name: "positive case with 1 string",
			args: args{
				ss: []string{fmt.Sprintf("%s%s", "", s1),
					fmt.Sprintf("%s%s", "", s2),
					fmt.Sprintf("%s%s", "", s3)},
			},
			want: []*Metadata{{Key: "0", Value: s1},
				{Key: "1", Value: s2},
				{Key: "2", Value: s3},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMetadataFromStringArray(tt.args.ss)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewMetadataFromStringArray() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMetadataFromStringArray() = %v, want %v", got, tt.want)
			}
		})
	}
}
