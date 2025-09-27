package nelderMead

import (
	"errors"
	"sort"
)

type Vector []float64

func (v Vector) Add(w Vector) Vector {
	if len(v) != len(w) {
		panic("Vector.Add: different sizes")
	}
	result := make(Vector, len(v))
	for i := range v {
		result[i] = v[i] + w[i]
	}
	return result
}

func (v Vector) AddInPlace(w Vector) {
	if len(v) != len(w) {
		panic("Vector.Add: different sizes")
	}
	for i := range v {
		v[i] += w[i]
	}
}

func (v Vector) Sub(w Vector) Vector {
	if len(v) != len(w) {
		panic("Vector.Sub: different sizes")
	}
	result := make(Vector, len(v))
	for i := range v {
		result[i] = v[i] - w[i]
	}
	return result
}

func (v Vector) Mul(a float64) Vector {
	result := make(Vector, len(v))
	for i := range v {
		result[i] = v[i] * a
	}
	return result
}

type SimplexPoint struct {
	x   Vector
	val float64
}

type Simplex []SimplexPoint

func (s Simplex) Len() int {
	return len(s)
}

func (s Simplex) Less(i, j int) bool {
	return s[i].val < s[j].val
}

func (s Simplex) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Simplex) size() float64 {
	var delta float64
	for xi := 0; xi < len(s)-1; xi++ {
		min := s[0].x[xi]
		max := min
		for i := 1; i < len(s); i++ {
			v := s[i].x[xi]
			if v < min {
				min = v
			}
			if v > max {
				max = v
			}
		}
		d := (max - min) / max
		if d > delta {
			delta = d
		}
	}
	return delta
}

type Solvable func(Vector) (float64, error)

func (s Solvable) operate(ofs, a, b Vector, mul float64) (SimplexPoint, error) {
	x := ofs.Add(a.Sub(b).Mul(mul))
	val, err := s(x)
	if err != nil {
		return SimplexPoint{}, err
	}
	return SimplexPoint{
		x:   x,
		val: val,
	}, nil
}

func NelderMead(f Solvable, initial []Vector, maxIter int) (Vector, float64, error) {
	return NelderMeadBase(f, initial, 1, 2, 0.5, 0.5, maxIter)
}

func NelderMeadBase(f Solvable, initial []Vector, alpha, gamma, beta, sigma float64, maxIter int) (Vector, float64, error) {
	s := make(Simplex, len(initial))
	for i := range s {
		val, err := f(initial[i])
		if err != nil {
			return nil, 0, err
		}
		s[i] = SimplexPoint{
			x:   initial[i],
			val: val,
		}
	}
	n := len(s) - 1
	mid := make(Vector, n)

	for {
		//fmt.Println(s)

		maxIter--
		if maxIter < 0 {
			return nil, 0, errors.New("max iterations reached")
		}

		if s.size() < 1e-13 {
			return s[0].x, s[0].val, nil
		}

		sort.Sort(s)

		for i := 0; i < n; i++ {
			mid[i] = 0
		}
		for i := 0; i < n; i++ {
			mid.AddInPlace(s[i].x)
		}
		mid = mid.Mul(1 / float64(n))

		// reflection
		sr, err := f.operate(mid, mid, s[n].x, alpha)
		if err != nil {
			return nil, 0, err
		}
		if sr.val < s[0].val {
			// expansion
			xe, err := f.operate(sr.x, sr.x, mid, gamma)
			if err != nil {
				return nil, 0, err
			}
			if xe.val < sr.val {
				s[n] = xe
			} else {
				s[n] = sr
			}
		} else {
			if sr.val < s[n-1].val {
				s[n] = sr
			} else {
				// contraction
				h := s[n]
				if sr.val < s[n].val {
					h = sr
				}

				sc, err := f.operate(h.x, mid, h.x, beta)
				if err != nil {
					return nil, 0, err
				}
				if sc.val < s[n].val {
					s[n] = sc
				} else {
					// shrink
					for i := 1; i <= n; i++ {
						s[i], err = f.operate(s[i].x, s[0].x, s[i].x, sigma)
						if err != nil {
							return nil, 0, err
						}
					}
				}
			}
		}
	}
}
