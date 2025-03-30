package polynomial

import (
	"errors"
	"fmt"
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
	return p.intString(false)
}

func (p Polynomial) StringToParse() string {
	return p.intString(true)
}

func (p Polynomial) intString(parser bool) string {
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
			if parser {
				switch n {
				case 0:
				case 1:
					result += "x"
				default:
					result += fmt.Sprintf("x^%d", n)
				}
			} else {
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
	if len(p) == 0 {
		return Roots{}, errors.New("no coefficients given")
	}
	if p[len(p)-1] == 0 {
		return Roots{}, errors.New("not canonical")
	}
	switch len(p) {
	case 1:
		return Roots{roots: nil, factor: p[0]}, nil
	case 2:
		return Roots{roots: []complex128{complex(-p[0]/p[1], 0)}, factor: p[1]}, nil
	case 3:
		a, b, c := p[2], p[1], p[0]
		d := b*b - 4*a*c
		if d < 0 {
			sqrtD := math.Sqrt(-d)
			return Roots{roots: []complex128{complex(-b/(2*a), math.Abs(sqrtD/(2*a)))}, factor: a}, nil
		}
		sqrtD := math.Sqrt(d)
		return Roots{roots: []complex128{complex((-b+sqrtD)/(2*a), 0), complex((-b-sqrtD)/(2*a), 0)}, factor: a}, nil
	default:
		if math.Abs(p[0]) < eps {
			np := p[1:]
			r, err := np.Roots()
			if err != nil {
				return Roots{}, err
			}
			r.roots = append(r.roots, 0)
			return r, nil
		} else {
			zero, err := p.findRootNewton(1e-9)
			if err != nil {
				return Roots{}, err
			}
			rp := FromRoot(zero)
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
			return r, nil
		}
	}
}

func (p Polynomial) findRootNewton(zEps float64) (complex128, error) {
	deriv := p.Derivative()
	var lastz complex128
	z := complex(1, 1)
	for range 1000 {
		f := p.EvalCplx(z)
		if cmplx.Abs(z-lastz) < eps && cmplx.Abs(f) < zEps {
			return cleanUp(z), nil
		}
		lastz = z
		z = z - f/deriv.EvalCplx(z)
	}
	z = cleanUp(z)
	return z, fmt.Errorf("no convergence in %v, s=%v, f(s)=%v", p, z, p.EvalCplx(z))
}

func cleanUp(z complex128) complex128 {
	absImag := math.Abs(imag(z))
	if absImag < eps {
		return complex(real(z), 0)
	}
	return complex(real(z), absImag)
}

func FromRoot(zero complex128) Polynomial {
	if math.Abs(imag(zero)) < eps {
		return Polynomial{-real(zero), 1}
	} else {
		return Polynomial{real(zero)*real(zero) + imag(zero)*imag(zero), -2 * real(zero), 1}
	}
}

func Equals(a, b complex128) bool {
	return math.Abs(real(a)-real(b)) < eps &&
		math.Abs(imag(a)-imag(b)) < eps
}

func (r Roots) Valid() bool {
	return r.factor != 0
}

func (r Roots) Polynomial() Polynomial {
	p := Polynomial{r.factor}
	for _, root := range r.roots {
		m := FromRoot(root)
		p = p.Mul(m)
	}
	return p
}

func (r Roots) MulFloat(k float64) Roots {
	return Roots{
		roots:  r.roots,
		factor: r.factor * k,
	}
}

func (r Roots) Real(a, b float64) (Roots, error) {
	if a == 0 {
		return Roots{}, errors.New("not a real root")
	}
	return Roots{
		roots:  append(r.roots, complex(-b/a, 0)),
		factor: r.factor * a,
	}, nil
}

func (r Roots) Complex(a, b, c float64) (Roots, error) {
	if a == 0 {
		return r.Real(b, c)
	}
	d := b*b - 4*a*c
	if d < 0 {
		sqrtD := math.Sqrt(-d)
		z := complex(-b/(2*a), math.Abs(sqrtD/(2*a)))
		return Roots{
			roots:  append(r.roots, z),
			factor: r.factor * a,
		}, nil
	} else {
		sqrtD := math.Sqrt(d)
		z1 := complex((-b-sqrtD)/(2*a), 0)
		z2 := complex((-b+sqrtD)/(2*a), 0)
		return Roots{
			roots:  append(r.roots, z1, z2),
			factor: r.factor * a,
		}, nil
	}
}

func (r Roots) String() string {
	return r.intString(false)
}

func (r Roots) StringToParse() string {
	return r.intString(true)
}

func (r Roots) intString(parse bool) string {
	var b strings.Builder
	if r.factor != 1 {
		b.WriteString(fmt.Sprintf("%.6g", r.factor))
	}
	for _, root := range r.roots {
		if b.Len() > 0 {
			b.WriteString("*")
		}
		p := FromRoot(root)
		b.WriteString("(")
		b.WriteString(p.intString(parse))
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

func (r Roots) Equals(b Roots) bool {
	if math.Abs(r.factor-b.factor) > eps {
		return false
	}
	if len(r.roots) != len(b.roots) {
		return false
	}
	for i := range r.roots {
		if !Equals(r.roots[i], b.roots[i]) {
			return false
		}
	}
	return true
}

func (r Roots) reduce(poles Roots) (Roots, Roots, bool) {
	nZeros := Roots{factor: r.factor}
	nPoles := Roots{factor: poles.factor, roots: append([]complex128{}, poles.roots...)}
	success := false
	for _, z := range r.roots {
		found := -1
		for i, p := range nPoles.roots {
			if Equals(z, p) {
				found = i
				break
			}
		}
		if found == -1 {
			nZeros.roots = append(nZeros.roots, z)
		} else {
			nPoles.roots = append(nPoles.roots[:found], nPoles.roots[found+1:]...)
			success = true
		}
	}
	return nZeros, nPoles, success
}
