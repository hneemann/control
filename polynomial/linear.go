package polynomial

import (
	"fmt"
	"github.com/hneemann/control/graph"
	"math"
)

type Linear struct {
	Numerator   Polynomial
	Denominator Polynomial
	zeros       Roots
	poles       Roots
}

func (l *Linear) Eval(s complex128) complex128 {
	return l.Numerator.EvalCplx(s) / l.Denominator.EvalCplx(s)
}

func (l *Linear) Equals(b *Linear) bool {
	return l.Numerator.Equals(b.Numerator) && l.Denominator.Equals(b.Denominator)
}

func (l *Linear) StringPoly(parse bool) string {
	return "(" + l.Numerator.intString(parse) + ")/(" + l.Denominator.intString(parse) + ")"
}

func (l *Linear) String() string {
	return l.intString(false)
}

func (l *Linear) StringToParse() string {
	return l.intString(true)
}

func (l *Linear) intString(parse bool) string {
	var n string
	if l.zerosCalculated() {
		n = l.zeros.intString(parse)
	} else {
		n = "(" + l.Numerator.intString(parse) + ")"
	}
	var d string
	if l.polesCalculated() {
		d = l.poles.intString(parse)
	} else {
		d = l.Denominator.intString(parse)
	}
	return fmt.Sprintf("%s/(%s)", n, d)
}

func (l *Linear) zerosCalculated() bool {
	return l.zeros.Valid()
}

func (l *Linear) Zeros() (Roots, error) {
	if !l.zeros.Valid() {
		roots, err := l.Numerator.Roots()
		if err != nil {
			return Roots{}, fmt.Errorf("error in calculating zeros of %v: %w", l, err)
		}
		l.zeros = roots
	}
	return l.zeros, nil
}

func (l *Linear) polesCalculated() bool {
	return l.poles.Valid()
}

func (l *Linear) Poles() (Roots, error) {
	if !l.poles.Valid() {
		roots, err := l.Denominator.Roots()
		if err != nil {
			return Roots{}, fmt.Errorf("error in calculating poles of %v: %w", l, err)
		}
		l.poles = roots
	}
	return l.poles, nil
}

func FromRoots(zeros, poles Roots) *Linear {
	nZeros, nPoles, _ := zeros.reduce(poles)
	return &Linear{
		Numerator:   nZeros.Polynomial(),
		Denominator: nPoles.Polynomial(),
		zeros:       nZeros,
		poles:       nPoles,
	}
}

func (l *Linear) reduce() *Linear {
	if l.zerosCalculated() && l.polesCalculated() {
		if nZeros, nPoles, ok := l.zeros.reduce(l.poles); ok {
			return &Linear{
				Numerator:   nZeros.Polynomial(),
				Denominator: nPoles.Polynomial(),
				zeros:       nZeros,
				poles:       nPoles,
			}
		}
	}
	return l
}

func (l *Linear) Mul(b *Linear) *Linear {
	var n Polynomial
	var z Roots
	if l.zerosCalculated() && b.zerosCalculated() {
		z = l.zeros.Mul(b.zeros)
		n = z.Polynomial()
	} else {
		n = l.Numerator.Mul(b.Numerator)
	}

	var d Polynomial
	var p Roots
	if l.polesCalculated() && b.polesCalculated() {
		p = l.poles.Mul(b.poles)
		d = p.Polynomial()
	} else {
		d = l.Denominator.Mul(b.Denominator)
	}

	return (&Linear{
		Numerator:   n,
		Denominator: d,
		zeros:       z,
		poles:       p,
	}).reduce()
}

func (l *Linear) Inv() *Linear {
	return &Linear{
		Numerator:   l.Denominator,
		Denominator: l.Numerator,
		zeros:       l.poles,
		poles:       l.zeros,
	}
}

func (l *Linear) Div(b *Linear) *Linear {
	var n Polynomial
	var z Roots
	if l.zerosCalculated() && b.polesCalculated() {
		z = l.zeros.Mul(b.poles)
		n = z.Polynomial()
	} else {
		n = l.Numerator.Mul(b.Denominator)
	}

	var d Polynomial
	var p Roots
	if l.polesCalculated() && b.zerosCalculated() {
		p = l.poles.Mul(b.zeros)
		d = p.Polynomial()
	} else {
		d = l.Denominator.Mul(b.Numerator)
	}

	return (&Linear{
		Numerator:   n,
		Denominator: d,
		zeros:       z,
		poles:       p,
	}).reduce()
}

func (l *Linear) Add(b *Linear) *Linear {
	n := l.Numerator.Mul(b.Denominator).Add(b.Numerator.Mul(l.Denominator))
	if l.polesCalculated() && b.polesCalculated() {
		adr, _ := l.Poles()
		bdr, _ := b.Poles()
		d := adr.Mul(bdr)
		return &Linear{
			Numerator:   n,
			Denominator: d.Polynomial(),
			poles:       d,
		}
	} else {
		d := l.Denominator.Mul(b.Denominator)
		return &Linear{
			Numerator:   n,
			Denominator: d,
		}
	}
}

func (l *Linear) Pow(n int) *Linear {
	return (&Linear{
		Numerator:   l.Numerator.Pow(n),
		Denominator: l.Denominator.Pow(n),
	})
}

func NewConst(c float64) *Linear {
	return &Linear{
		Numerator:   Polynomial{c},
		Denominator: Polynomial{1},
	}
}

func (l *Linear) Loop() (*Linear, error) {
	l, err := l.Reduce()
	if err != nil {
		return nil, err
	}
	return &Linear{
		Numerator:   l.Numerator,
		zeros:       l.zeros,
		Denominator: l.Numerator.Add(l.Denominator),
	}, nil
}

func (l *Linear) Reduce() (*Linear, error) {
	z, err := l.Zeros()
	if err != nil {
		return nil, err
	}
	p, err := l.Poles()
	if err != nil {
		return nil, err
	}
	if nz, np, ok := z.reduce(p); ok {
		return (&Linear{
			Numerator:   nz.Polynomial(),
			Denominator: np.Polynomial(),
			zeros:       nz,
			poles:       np,
		}).reduceFactor(), nil
	} else {
		return l.reduceFactor(), nil
	}
}

func (l *Linear) reduceFactor() *Linear {
	f := l.zeros.factor / l.poles.factor
	roundF := math.Round(f)
	if math.Abs(f-roundF) < eps {
		nz := Roots{roots: l.zeros.roots, factor: f}
		np := Roots{roots: l.poles.roots, factor: 1}
		return &Linear{
			Numerator:   nz.Polynomial(),
			Denominator: np.Polynomial(),
			zeros:       nz,
			poles:       np,
		}
	}
	return l
}

func (l *Linear) MulFloat(f float64) *Linear {
	return &Linear{
		Numerator:   l.Numerator.MulFloat(f),
		Denominator: l.Denominator,
	}
}

func PID(kp, ti, td float64) *Linear {
	n := Polynomial{kp, kp * ti, kp * ti * td}.Canonical()
	var d Polynomial
	if ti == 0 {
		d = Polynomial{1}
	} else {
		d = Polynomial{0, ti}
	}
	zeros, _ := n.Roots()
	poles, _ := d.Roots()
	return &Linear{
		Numerator:   n,
		Denominator: d,
		zeros:       zeros,
		poles:       poles,
	}
}

func (l *Linear) CreateEvans(kMax float64) (*graph.Plot, error) {
	p, err := l.Poles()
	if err != nil {
		return nil, err
	}
	z, err := l.Zeros()
	if err != nil {
		return nil, err
	}

	poleCount := p.Count()

	var pointsSet [][]graph.Point

	k := 0.001
	const kMult = 1.2
	for k < kMax {
		err = l.addPoles(&pointsSet, k, kMult, poleCount)
		if err != nil {
			return nil, err
		}
		k = k * kMult
	}
	err = l.addPoles(&pointsSet, kMax, kMult, poleCount)
	if err != nil {
		return nil, err
	}

	pathList := make([]graph.Path, poleCount)
	for _, pl := range pointsSet {
		for i := range poleCount {
			pathList[i] = pathList[i].Add(pl[i])
		}
	}

	curveList := make([]graph.PlotContent, 0, len(pathList)+2)
	for _, pa := range pathList {
		curveList = append(curveList, graph.Curve{Path: pa, Style: graph.Black})
	}
	curveList = append(curveList,
		graph.Scatter{
			Points: p.ToPoints(),
			Shape:  graph.NewCrossMarker(4),
			Style:  graph.Black,
		},
		graph.Scatter{
			Points: z.ToPoints(),
			Shape:  graph.NewCircleMarker(4),
			Style:  graph.Black,
		},
	)

	return &graph.Plot{
		XAxis:   graph.NewLinear(-5, 0.2),
		YAxis:   graph.NewLinear(-2, 2),
		Content: curveList,
	}, nil
}

func (l *Linear) addPoles(pointsSet *[][]graph.Point, k, kMult float64, poleCount int) error {
	points, split, err := l.getPoles(pointsSet, k, poleCount)
	if err != nil {
		return fmt.Errorf("error in getting poles for k=%g: %w", k, err)
	}

	if split {
		err = l.find(pointsSet, k/kMult, k, poleCount, 8)
		if err != nil {
			return err
		}
	}
	*pointsSet = append(*pointsSet, points)
	return nil
}

func (l *Linear) find(set *[][]graph.Point, k0, k1 float64, poleCount, depth int) error {
	if depth == 0 {
		return nil
	}
	km := (k0 + k1) / 2
	points, split, err := l.getPoles(set, km, poleCount)
	if err != nil {
		return err
	}
	if split {
		err = l.find(set, k0, km, poleCount, depth-1)
		if err != nil {
			return err
		}
		*set = append(*set, points)
	} else {
		*set = append(*set, points)
		err = l.find(set, km, k1, poleCount, depth-1)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *Linear) getPoles(pointsSet *[][]graph.Point, k float64, poleCount int) ([]graph.Point, bool, error) {
	gw, err := l.MulFloat(k).Loop()
	if err != nil {
		return nil, false, err
	}
	poles, err := gw.Poles()
	if err != nil {
		return nil, false, err
	}

	points := poles.ToPoints()

	if len(points) != poleCount {
		return nil, false, fmt.Errorf("unexpected pole count: %d", len(points))
	}

	split := false
	if len(*pointsSet) > 0 {
		split = order((*pointsSet)[len(*pointsSet)-1], points)
	}

	return points, split, nil
}

func order(base []graph.Point, points []graph.Point) bool {
	split := false
	for i := range base {
		var best int
		bestDist := math.Inf(1)
		for j := i; j < len(points); j++ {
			d := base[i].DistTo(points[j])
			if d < bestDist {
				best = j
				bestDist = d
			}
		}
		if best != i {
			points[i], points[best] = points[best], points[i]
		}
		if (math.Abs(base[i].Y) < 1e-6) != (math.Abs(points[i].Y) < 1e-6) {
			split = true
		}
	}
	return split
}
