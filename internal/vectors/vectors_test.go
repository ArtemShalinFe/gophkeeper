package vectors

import (
	"testing"
)

type vector struct {
	vers    int64
	hashsum string
}

func newVector(vers int64, hashsum string) *vector {
	return &vector{
		vers:    vers,
		hashsum: hashsum,
	}
}

func (v *vector) GetVersion() int64 {
	return v.vers
}

func (v *vector) GetHashsum() string {
	return v.hashsum
}

func TestComparison_Compare(t *testing.T) {
	type fields struct {
		VectorA Vector
		VectorB Vector
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "IsEquals",
			fields: fields{
				VectorA: newVector(1, "1"),
				VectorB: newVector(1, "1"),
			},
			want: VectorAIsEqualsVectorB,
		},
		{
			name: "IsLower",
			fields: fields{
				VectorA: newVector(1, "1"),
				VectorB: newVector(2, "1"),
			},
			want: VectorAIsLowerVectorB,
		},
		{
			name: "IsHigher",
			fields: fields{
				VectorA: newVector(2, "1"),
				VectorB: newVector(1, "1"),
			},
			want: VectorAIsHigherVectorB,
		},
		{
			name: "IsConflict",
			fields: fields{
				VectorA: newVector(1, "1"),
				VectorB: newVector(1, "2"),
			},
			want: VectorAIsConflictVectorB,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewComparison(
				tt.fields.VectorA,
				tt.fields.VectorB,
			)
			if got := c.Compare(); got != tt.want {
				t.Errorf("Comparison.Compare() = %v, want %v", got, tt.want)
			}
		})
	}
}
