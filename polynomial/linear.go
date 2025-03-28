package polynomial

import "fmt"

type Linear struct {
	Numerator   Polynomial
	Denominator Polynomial
	nRoots      Roots
	dRoots      Roots
}

func (l Linear) Eval(s complex128) complex128 {
	return l.Numerator.EvalCplx(s) / l.Denominator.EvalCplx(s)
}

func (l Linear) String() string {
	return fmt.Sprintf("(%s)/(%s)", l.Numerator.String(), l.Denominator.String())
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
	denominator = denominator.RemoveConjugate()
	for _, nr := range numerator.RemoveConjugate().roots {
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
