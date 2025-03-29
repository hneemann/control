package polynomial

import (
	"fmt"
	"math"
)

type Linear struct {
	Numerator   Polynomial
	Denominator Polynomial
	nRoots      Roots
	dRoots      Roots
}

func (l Linear) Eval(s complex128) complex128 {
	return l.Numerator.EvalCplx(s) / l.Denominator.EvalCplx(s)
}

func (l Linear) Equals(b Linear) bool {
	return l.Numerator.Equals(b.Numerator) && l.Denominator.Equals(b.Denominator)
}

func (l Linear) String() string {
	var n string
	nr, err := l.getNumRoots()
	if err == nil {
		n = nr.String()
	} else {
		n = l.Numerator.String()
	}
	var d string
	dr, err := l.getDenRoots()
	if err == nil {
		d = dr.String()
	} else {
		d = l.Denominator.String()
	}
	return fmt.Sprintf("%s/(%s)", n, d)
}

func (l *Linear) getNumRoots() (Roots, error) {
	if l.nRoots.factor == 0 {
		roots, err := l.Numerator.Roots()
		if err != nil {
			return Roots{}, err
		}
		l.nRoots = roots
	}
	return l.nRoots, nil
}

func (l *Linear) getDenRoots() (Roots, error) {
	if l.dRoots.factor == 0 {
		roots, err := l.Denominator.Roots()
		if err != nil {
			return Roots{}, err
		}
		l.dRoots = roots
	}
	return l.dRoots, nil
}

func (l Linear) Zeros() ([]complex128, error) {
	r, err := l.getNumRoots()
	if err != nil {
		return []complex128{}, err
	}
	return r.roots, nil
}

func (l Linear) Poles() ([]complex128, error) {
	r, err := l.getDenRoots()
	if err != nil {
		return []complex128{}, err
	}
	return r.roots, nil
}

func FromRoots(numerator, denominator Roots) Linear {
	nn := Roots{factor: numerator.factor}
	for _, nr := range numerator.roots {
		found := -1
		for i, dr := range denominator.roots {
			if Equals(nr, dr) {
				found = i
				break
			}
		}
		if found >= 0 {
			denominator.roots = append(denominator.roots[:found], denominator.roots[found+1:]...)
		} else {
			nn.roots = append(nn.roots, nr)
		}
	}

	return Linear{
		Numerator:   nn.Polynomial(),
		Denominator: denominator.Polynomial(),
		nRoots:      nn,
		dRoots:      denominator,
	}
}

func (l Linear) Mul(b Linear) (Linear, error) {
	anr, err := l.getNumRoots()
	if err != nil {
		return Linear{}, err
	}
	adr, err := l.getDenRoots()
	if err != nil {
		return Linear{}, err
	}
	bnr, err := b.getNumRoots()
	if err != nil {
		return Linear{}, err
	}
	bdr, err := b.getDenRoots()
	if err != nil {
		return Linear{}, err
	}
	return FromRoots(anr.Mul(bnr), adr.Mul(bdr)), nil
}

func (l Linear) Div(b Linear) (Linear, error) {
	anr, err := l.getNumRoots()
	if err != nil {
		return Linear{}, err
	}
	adr, err := l.getDenRoots()
	if err != nil {
		return Linear{}, err
	}
	bnr, err := b.getNumRoots()
	if err != nil {
		return Linear{}, err
	}
	bdr, err := b.getDenRoots()
	if err != nil {
		return Linear{}, err
	}
	return FromRoots(anr.Mul(bdr), adr.Mul(bnr)), nil
}

func (l Linear) Add(b Linear) (Linear, error) {
	r, err := l.Numerator.Mul(b.Denominator).Add(b.Numerator.Mul(l.Denominator)).Roots()
	if err != nil {
		return Linear{}, err
	}
	adr, err := l.getDenRoots()
	if err != nil {
		return Linear{}, err
	}
	bdr, err := b.getDenRoots()
	if err != nil {
		return Linear{}, err
	}
	return FromRoots(r, adr.Mul(bdr)), nil
}

func (l Linear) Loop() (Linear, error) {
	one := Linear{
		Numerator:   Polynomial{1},
		Denominator: Polynomial{1},
	}
	d, err := l.Add(one)
	if err != nil {
		return Linear{}, err
	}
	return l.Div(d)
}

func (l Linear) MulFloat(f float64) Linear {
	return Linear{
		Numerator:   l.Numerator.MulFloat(f),
		Denominator: l.Denominator,
	}
}

func PID(kp, ti, td float64) Linear {
	denom := Roots{factor: 1 / ti, roots: []complex128{0}}
	sq := 1/(4*td*td) - 1/(ti*td)
	if sq < 0 {
		return FromRoots(NewRoots(complex(-1/(2*td), math.Sqrt(-sq))), denom).MulFloat(kp)
	} else {
		return FromRoots(NewRoots(complex(-1/(2*td)-math.Sqrt(sq), 0), complex(-1/(2*td)+math.Sqrt(sq), 0)), denom).MulFloat(kp)
	}
}
