package polynomial

import (
	"errors"
	"fmt"
	"github.com/hneemann/control/nelderMead"
	"math"
	"math/cmplx"
	"strings"
)

const eps = 1e-10

type Polynomial []float64

func (p Polynomial) Degree() int {
	return len(p) - 1
}

func (p Polynomial) Eval(x float64) float64 {
	var result float64
	var xPower float64 = 1
	for _, c := range p {
		result += c * xPower
		xPower *= x
	}
	return result
}

func (p Polynomial) EvalCplx(x complex128) complex128 {
	var result complex128
	var xPower complex128 = 1
	for _, c := range p {
		result += complex(c, 0) * xPower
		xPower *= x
	}
	return result
}

func (p Polynomial) String() string {
	result := ""
	for i := range p {
		n := len(p) - i - 1
		c := p[n]
		if math.Abs(c) > eps {
			neg := false
			if c < 0 {
				neg = true
			}
			c = math.Abs(c)
			if neg || i > 0 {
				if neg {
					result += "-"
				} else {
					result += "+"
				}
			}
			if c != 1 || n == 0 {
				result += fmt.Sprintf("%.6g", c)
			}
			switch n {
			case 0:
			case 1:
				result += "x"
			case 2:
				result += "x²"
			case 3:
				result += "x³"
			case 4:
				result += "x⁴"
			case 5:
				result += "x⁵"
			case 6:
				result += "x⁶"
			case 7:
				result += "x⁷"
			case 8:
				result += "x⁸"
			case 9:
				result += "x⁹"
			default:
				result += fmt.Sprintf("x^%d", n)
			}
		}
	}
	return result
}

func (p Polynomial) Derivative() Polynomial {
	if len(p) == 0 {
		return Polynomial{}
	}
	result := make(Polynomial, len(p)-1)
	for i := 1; i < len(p); i++ {
		result[i-1] = float64(i) * p[i]
	}
	return result
}

func (p Polynomial) Add(q Polynomial) Polynomial {
	if len(p) >= len(q) {
		result := make(Polynomial, len(p))
		copy(result, p)
		for i := range q {
			result[i] += q[i]
		}
		return result
	} else {
		return q.Add(p)
	}
}

func (p Polynomial) Mul(q Polynomial) Polynomial {
	result := make(Polynomial, p.Degree()+q.Degree()+1)
	for i := range p {
		for j := range q {
			result[i+j] += p[i] * q[j]
		}
	}
	return result
}

func (p Polynomial) MulFloat(f float64) Polynomial {
	mp := make(Polynomial, len(p))
	copy(mp, p)
	for i := range p {
		mp[i] *= f
	}
	return mp
}

func (p Polynomial) Div(q Polynomial) (Polynomial, Polynomial, error) {
	if q.Degree() < 0 {
		return Polynomial{}, Polynomial{}, errors.New("division by zero")
	}
	if p.Degree() < q.Degree() {
		return Polynomial{}, p, nil
	}
	result := make(Polynomial, p.Degree()-q.Degree()+1)
	remainder := make(Polynomial, len(p))
	copy(remainder, p)
	for i := 0; i < len(result); i++ {
		result[i] = remainder[i] / q[0]
		for j := 0; j < len(q); j++ {
			remainder[i+j] -= result[i] * q[j]
		}
	}
	return result, remainder.Canonical(), nil
}

func (p Polynomial) Pow(n int) Polynomial {
	if n < 0 {
		panic("negative exponent")
	}
	result := Polynomial{1}
	for i := 0; i < n; i++ {
		result = result.Mul(p)
	}
	return result
}

func (p Polynomial) Canonical() Polynomial {
	for i := len(p) - 1; i >= 0; i-- {
		if math.Abs(p[i]) > eps {
			return p[:i+1]
		}
	}
	return Polynomial{}
}

func (p Polynomial) Equals(b Polynomial) bool {
	if len(p) != len(b) {
		return false
	}
	for i := range p {
		if math.Abs(p[i]-b[i]) > eps {
			return false
		}
	}
	return true
}

type Roots struct {
	roots  []complex128
	factor float64
}

func NewRoots(roots ...complex128) Roots {
	return Roots{roots: roots, factor: 1}
}

func (p Polynomial) Roots() (Roots, error) {
	if p[len(p)-1] == 0 {
		return Roots{}, errors.New("not canonical")
	}
	switch len(p) {
	case 0:
		return Roots{}, nil
	case 1:
		return Roots{roots: nil, factor: p[0]}, nil
	case 2:
		return Roots{roots: []complex128{complex(-p[0]/p[1], 0)}, factor: p[1]}, nil
	case 3:
		a, b, c := p[2], p[1], p[0]
		d := b*b - 4*a*c
		if d < 0 {
			sqrtD := math.Sqrt(-d)
			return Roots{roots: []complex128{complex(-b/(2*a), sqrtD/(2*a)), complex(-b/(2*a), -sqrtD/(2*a))}, factor: a}, nil
		}
		sqrtD := math.Sqrt(d)
		return Roots{roots: []complex128{complex((-b+sqrtD)/(2*a), 0), complex((-b-sqrtD)/(2*a), 0)}, factor: a}, nil
	default:
		zero, err := p.findRoot()
		if err != nil {
			return Roots{}, err
		}
		rp, cplx := FromRoot(zero)
		var np Polynomial
		np, _, err = p.Div(rp)
		if err != nil {
			return Roots{}, err
		}
		var r Roots
		r, err = np.Roots()
		if err != nil {
			return Roots{}, err
		}
		r.roots = append(r.roots, zero)
		if cplx {
			r.roots = append(r.roots, complex(real(zero), -imag(zero)))
		}
		return r, nil
	}
}

func (p Polynomial) findRoot() (complex128, error) {
	r, z, err := nelderMead.NelderMead(func(x nelderMead.Vector) float64 {
		return cmplx.Abs(p.EvalCplx(complex(x[0], x[1])))
	}, []nelderMead.Vector{{0, 0}, {0, 1}, {1, 0}}, 1000)
	if err != nil {
		return 0, err
	}
	if math.Abs(z) > eps {
		return 0, errors.New("Nebenminimum")
	}
	return complex(r[0], r[1]), nil
}

func FromRoot(zero complex128) (Polynomial, bool) {
	if math.Abs(imag(zero)) < eps {
		return Polynomial{-real(zero), 1}, false
	} else {
		return Polynomial{real(zero)*real(zero) + imag(zero)*imag(zero), -2 * real(zero), 1}, true
	}
}

func (r Roots) RemoveConjugate() Roots {
	var nr []complex128
	for _, z := range r.roots {
		if math.Abs(imag(z)) < eps {
			nr = append(nr, z)
		} else {
			if imag(z) > 0 {
				nr = append(nr, z)
			}
		}
	}
	return Roots{
		roots:  nr,
		factor: r.factor,
	}
}

func Equals(a, b complex128) bool {
	return math.Abs(real(a)-real(b)) < eps &&
		math.Abs(imag(a)-imag(b)) < eps
}

func (r Roots) Polynomial() Polynomial {
	p := Polynomial{r.factor}
	for _, root := range r.RemoveConjugate().roots {
		m, _ := FromRoot(root)
		p = p.Mul(m)
	}
	return p
}

func (r Roots) String() string {
	var b strings.Builder
	if r.factor != 1 {
		b.WriteString(fmt.Sprintf("%g", r.factor))
	}
	for _, root := range r.RemoveConjugate().roots {
		if b.Len() > 0 {
			b.WriteString("*")
		}
		p, _ := FromRoot(root)
		b.WriteString("(")
		b.WriteString(p.String())
		b.WriteString(")")
	}
	return b.String()
}

func (r Roots) Mul(b Roots) Roots {
	return Roots{
		roots:  append(r.roots, b.roots...),
		factor: r.factor * b.factor,
	}
}
