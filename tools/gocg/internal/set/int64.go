package set

type Int64 map[int64]struct{}

func NewInt64Slice(d []int64) Int64 {
	m := make(Int64)
	for _, dp := range d {
		m[dp] = struct{}{}
	}
	return m
}

// NewIntFromTo creates a set with values ranging from (inclusive) to (exclusive)
func NewInt64FromTo(from, to int64) Int64 {
	m := make(Int64)
	for i := from; i < to; i++ {
		m[i] = struct{}{}
	}
	return m
}

func DisjointInt64(tg, cg Int64) bool {
	for k := range tg {
		if _, exists := cg[k]; exists {
			return false
		}
	}
	return true
}

func EqualInt64(tg, cg Int64) bool {
	if len(tg) != len(cg) {
		return false
	}

	for k := range tg {
		if _, exists := cg[k]; !exists {
			return false
		}
	}
	return true
}

func SubSetInt64(super, sub Int64) bool {
	for k := range sub {
		if _, contained := super[k]; !contained {
			return false
		}
	}
	return true
}

func UnionInt64(a, b Int64) Int64 {
	s := make(Int64)
	for k := range a {
		s[k] = struct{}{}
	}
	for k := range b {
		s[k] = struct{}{}
	}
	return s
}

func IntersectionInt64(a, b Int64) Int64 {
	s := make(Int64)
	for k := range a {
		if _, ok := b[k]; ok {
			s[k] = struct{}{}
		}
	}
	return s
}

func ComplementInt64(a, b Int64) Int64 {
	s := make(Int64)
	for k := range a {
		if _, ok := b[k]; !ok {
			s[k] = struct{}{}
		}
	}
	return s
}
