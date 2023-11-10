package vectors

type Vector interface {
	GetVersion() int64
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
	if c.VectorA.GetVersion() > c.VectorB.GetVersion() {
		return VectorAIsHigherVectorB
	}

	if c.VectorA.GetVersion() < c.VectorB.GetVersion() {
		return VectorAIsLowerVectorB
	}

	if c.VectorA.GetVersion() == c.VectorB.GetVersion() && c.VectorA.GetHashsum() == c.VectorB.GetHashsum() {
		return VectorAIsEqualsVectorB
	}

	return VectorAIsConflictVectorB
}
