package vectors

type Vector interface {
	GetVesrion() int64
	GetHashsum() string
}

const (
	VectorAIsEqualsVectorB   = "equals"
	VectorAIsHigherVectorB   = "higher"
	VectorAIsLowerVectorB    = "lower"
	VectorAIsConflictVectorB = "conflict"
)

type Comparison struct {
	VectorA Vector
	VectorB Vector
}

func NewComparison(vectorA Vector, vectorB Vector) *Comparison {
	return &Comparison{
		VectorA: vectorA,
		VectorB: vectorB,
	}
}

func (c *Comparison) Compare() string {
	if c.VectorA.GetVesrion() > c.VectorB.GetVesrion() {
		return VectorAIsHigherVectorB
	}

	if c.VectorA.GetVesrion() < c.VectorB.GetVesrion() {
		return VectorAIsLowerVectorB
	}

	if c.VectorA.GetVesrion() == c.VectorB.GetVesrion() && c.VectorA.GetHashsum() == c.VectorB.GetHashsum() {
		return VectorAIsEqualsVectorB
	}

	return VectorAIsConflictVectorB
}
