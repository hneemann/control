package polynomial

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hneemann/control/graph"
	"github.com/hneemann/control/graph/grParser"
	"github.com/hneemann/iterator"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"github.com/hneemann/parser2/value/export"
	"github.com/hneemann/parser2/value/export/xmlWriter"
	"math"
	"math/cmplx"
	"sort"
)

type Linear struct {
	Numerator   Polynomial
	Denominator Polynomial
	zeros       Roots
	poles       Roots
}

var _ export.ToHtmlInterface = &Linear{}

func (l *Linear) EvalCplx(s complex128) complex128 {
	c := l.Numerator.EvalCplx(s) / l.Denominator.EvalCplx(s)
	return c
}

func (l *Linear) Eval(s float64) float64 {
	c := l.Numerator.Eval(s) / l.Denominator.Eval(s)
	return c
}

func (l *Linear) Equals(b *Linear) bool {
	return l.Numerator.Equals(b.Numerator) && l.Denominator.Equals(b.Denominator)
}

func (l *Linear) String() string {
	var n string
	if l.zerosCalculated() {
		n = l.zeros.String()
		if l.Denominator.IsOne() {
			return n
		}
	} else {
		is := l.Numerator.String()
		if l.Denominator.IsOne() {
			return is
		}
		if l.Numerator.IsSum() {
			n = "(" + is + ")"
		} else {
			n = is
		}
	}
	var d string
	if l.polesCalculated() {
		d = l.poles.String()
	} else {
		d = l.Denominator.String()
	}
	return fmt.Sprintf("%s/(%s)", n, d)
}

func (l *Linear) ToHtml(_ funcGen.Stack[value.Value], w *xmlWriter.XMLWriter) error {
	w.Open("math").
		Attr("xmlns", "http://www.w3.org/1998/Math/MathML")

	w.Open("mstyle").
		Attr("displaystyle", "true").
		Attr("scriptlevel", "0")

	w.Open("mfrac")

	l.Numerator.ToMathML(w)
	l.Denominator.ToMathML(w)

	w.Close()
	w.Close()
	w.Close()
	return nil
}

func (l *Linear) ToLaTeX(w *bytes.Buffer) {
	w.WriteString("\\frac{")
	l.Numerator.ToLaTeX(w)
	w.WriteString("}{")
	l.Denominator.ToLaTeX(w)
	w.WriteString("}")
}

func (l *Linear) ToUnicode() string {
	var w bytes.Buffer
	if l.Numerator.Degree() > 0 {
		w.WriteString("(")
		w.WriteString(l.Numerator.ToUnicode())
		w.WriteString(")")
	} else {
		w.WriteString(l.Numerator.ToUnicode())
	}
	w.WriteString("/(")
	w.WriteString(l.Denominator.ToUnicode())
	w.WriteString(")")
	return w.String()
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

func (l *Linear) IsCausal() bool {
	return l.Numerator.Degree() <= l.Denominator.Degree()
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

func (l *Linear) Add(b *Linear) (*Linear, error) {
	n := l.Numerator.Mul(b.Denominator).Add(b.Numerator.Mul(l.Denominator))
	if l.polesCalculated() && b.polesCalculated() {
		adr, _ := l.Poles()
		bdr, _ := b.Poles()
		d := adr.Mul(bdr)
		return &Linear{
			Numerator:   n,
			Denominator: d.Polynomial(),
			poles:       d,
		}, nil
	} else {
		d := l.Denominator.Mul(b.Denominator)
		return &Linear{
			Numerator:   n,
			Denominator: d,
		}, nil
	}
}

func (l *Linear) Derivative() *Linear {
	n := l.Numerator.Derivative().Mul(l.Denominator).Add(l.Numerator.Mul(l.Denominator.Derivative()).MulFloat(-1))
	d := l.Denominator.Mul(l.Denominator)
	return &Linear{
		Numerator:   n,
		Denominator: d,
	}
}

func (l *Linear) Pow(n int) *Linear {
	return &Linear{
		Numerator:   l.Numerator.Pow(n),
		Denominator: l.Denominator.Pow(n),
	}
}

func NewConst(c float64) *Linear {
	return &Linear{
		Numerator:   Polynomial{c},
		Denominator: Polynomial{1},
	}
}

func (l *Linear) Loop() *Linear {
	return &Linear{
		Numerator:   l.Numerator,
		zeros:       l.zeros,
		Denominator: l.Numerator.Add(l.Denominator),
	}
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

func (l *Linear) MulPoly(p Polynomial) *Linear {
	return &Linear{
		Numerator:   l.Numerator.Mul(p),
		Denominator: l.Denominator,
	}
}

func (l *Linear) DivPoly(p Polynomial) *Linear {
	return &Linear{
		Numerator:   l.Numerator,
		Denominator: l.Denominator.Mul(p),
	}
}

func PID(kp, ti, td, tp float64) (*Linear, error) {
	if ti == 0 {
		return nil, errors.New("ti must not be zero")
	}
	return &Linear{
		Numerator:   Polynomial{kp, kp * (ti + tp), kp * ti * (td + tp)}.Canonical(),
		Denominator: Polynomial{0, ti, ti * tp}.Canonical(),
	}, nil
}

type evansPoint struct {
	points     []graph.Point
	gain       float64
	numComplex int
}

type (
	metric func(graph.Point, graph.Point) float64
)

func (e evansPoint) dist(other evansPoint, m metric) float64 {
	var maxDist float64
	op := other.points
	for i, ep := range e.points {
		var best int
		bestDist := math.Inf(1)
		for j := i; j < len(op); j++ {
			d := m(ep, op[j])
			if d < bestDist {
				best = j
				bestDist = d
			}
		}
		if best != i {
			op[i], op[best] = op[best], op[i]
		}
		d := m(ep, op[i])
		if d > maxDist {
			maxDist = d
		}
	}
	return maxDist
}

type evansPoints []evansPoint

func (e evansPoints) Len() int {
	return len(e)
}

func (e evansPoints) Less(i, j int) bool {
	return e[i].gain < e[j].gain
}

func (e evansPoints) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func (e evansPoints) getPoints(i int) graph.PointsPath {
	return graph.PointsPath{
		Points: func(yield func(graph.Point, error) bool) {
			for _, ep := range e {
				if !yield(ep.points[i], nil) {
					return
				}
			}
		},
	}
}

type Polar struct{}

func (p Polar) String() string {
	return "Polar Grid"
}

func (p Polar) Bounds() (x, y graph.Bounds, e error) {
	return graph.Bounds{}, graph.Bounds{}, nil
}

func (p Polar) DependantBounds(_, _ graph.Bounds) (x, y graph.Bounds, e error) {
	return graph.Bounds{}, graph.Bounds{}, nil
}

type polarPath struct {
	radius float64
	r      graph.Rect
}

func (p polarPath) Iter(yield func(rune, graph.Point) bool) {
	radius := p.radius
	for angle := 90; angle <= 270; angle += 2 {
		x := radius * math.Cos(float64(angle)*math.Pi/180)
		y := radius * math.Sin(float64(angle)*math.Pi/180)
		point := graph.Point{X: x, Y: y}
		if angle == 90 {
			if !yield('M', point) {
				return
			}
		} else {
			if !yield('L', point) {
				return
			}
		}
	}
}

func (p polarPath) IsClosed() bool {
	return false
}

func (p Polar) DrawTo(plot *graph.Plot, canvas graph.Canvas) error {
	style := plot.Grid
	if style == nil {
		style = grParser.GridStyle
	}

	r := canvas.Rect()
	text := style.Text()
	textSize := canvas.Context().TextSize * 0.8
	var zero graph.Point

	// Draw the angle lines
	radius := r.MaxDistance(zero)
	for angle := 90; angle <= 270; angle += 15 {
		if angle != 180 {
			x := radius * math.Cos(float64(angle)*math.Pi/180)
			y := radius * math.Sin(float64(angle)*math.Pi/180)
			if ap, ep, state := r.Intersect(zero, graph.Point{X: x, Y: y}); state != graph.CompleteOutside {
				var o graph.Orientation
				if r.IsNearTop(ep) {
					o |= graph.Top
				} else if r.IsNearBottom(ep) {
					o |= graph.Bottom
				} else {
					if ep.Y > 0 {
						o |= graph.Bottom
					} else {
						o |= graph.Top
					}
				}
				if r.IsNearLeft(ep) {
					o |= graph.Left
				} else {
					o |= graph.Right
				}
				canvas.DrawPath(graph.NewPointsPath(false, ap, ep), style)
				canvas.DrawText(ep, fmt.Sprintf("%d°", 180-angle), o, text, textSize)
			}
		}
	}

	if r.Contains(zero) {
		for _, t := range plot.GetXTicks() {
			radius = -t.Position
			if radius > 1e-5 {
				canvas.DrawPath(r.IntersectPath(polarPath{radius: radius, r: r}), style)
				point := graph.Point{X: 0, Y: radius}
				if r.Contains(point) {
					canvas.DrawText(point, t.Label, graph.VCenter|graph.Left, text, textSize)
				}
			}
		}
	}
	return nil
}

func (p Polar) Legend() graph.Legend {
	return graph.Legend{}
}

type Asymptotes struct {
	Point graph.Point
	Order int
}

var asymptotesStyle = graph.Gray.SetStrokeWidth(2)

func (a Asymptotes) String() string {
	return "Asymptotes"
}

func (a Asymptotes) Bounds() (graph.Bounds, graph.Bounds, error) {
	return graph.Bounds{}, graph.Bounds{}, nil
}

func (a Asymptotes) DependantBounds(_, _ graph.Bounds) (graph.Bounds, graph.Bounds, error) {
	return graph.Bounds{}, graph.Bounds{}, nil
}

func (a Asymptotes) DrawTo(_ *graph.Plot, canvas graph.Canvas) error {
	r := canvas.Rect()

	d := r.MaxDistance(a.Point)

	dAlpha := 2 * math.Pi / float64(a.Order)
	alpha := dAlpha / 2
	for i := 0; i < a.Order; i++ {
		x := a.Point.X + d*math.Cos(alpha)
		y := a.Point.Y + d*math.Sin(alpha)

		if p1, p2, state := r.Intersect(a.Point, graph.Point{X: x, Y: y}); state != graph.CompleteOutside {
			canvas.DrawPath(graph.NewPointsPath(false, p1, p2), asymptotesStyle)
		}

		alpha += dAlpha
	}
	return nil
}

func (a Asymptotes) Legend() graph.Legend {
	return graph.Legend{Name: "Asymptotes", ShapeLineStyle: graph.ShapeLineStyle{LineStyle: asymptotesStyle}}
}

// PlotPreferences allows modifying a graph.Plot after it has been created.
// It can be used to set labels, styles, or other properties of the plot.
// It can't modify the bounds of the plot, as the axes are already drawn when
// the Modify function is called.
type PlotPreferences struct {
	// Modify is a function that modifies the plot.
	Modify func(*graph.Plot)
}

func (p PlotPreferences) Bounds() (x, y graph.Bounds, err error) {
	return graph.Bounds{}, graph.Bounds{}, nil
}

func (p PlotPreferences) DependantBounds(_, _ graph.Bounds) (x, y graph.Bounds, err error) {
	return graph.Bounds{}, graph.Bounds{}, nil
}

func (p PlotPreferences) DrawTo(plot *graph.Plot, _ graph.Canvas) error {
	p.Modify(plot)
	return nil
}

func (p PlotPreferences) Legend() graph.Legend {
	return graph.Legend{}
}

func (p PlotPreferences) String() string {
	return "Plot Preferences"
}

func NewImReLabels() PlotPreferences {
	return PlotPreferences{Modify: func(plot *graph.Plot) {
		if plot.YLabel == "" {
			plot.YLabel = "Im"
		}
		if plot.XLabel == "" {
			plot.XLabel = "Re"
		}
	}}
}

func (l *Linear) CreateEvans(kMin, kMax float64) ([]graph.PlotContent, error) {

	lin, err := l.Reduce()
	if err != nil {
		return nil, err
	}

	p, err := lin.Poles()
	if err != nil {
		return nil, err
	}
	z, err := lin.Zeros()
	if err != nil {
		return nil, err
	}

	ecs := evansCurves{
		polyProvider: func(k float64) (Polynomial, error) {
			return lin.Numerator.MulFloat(k).Add(lin.Denominator), nil
		},
	}

	if kMax <= 0 {
		return nil, fmt.Errorf("kMax (%g) must be greater than 0", kMax)
	} else if kMin >= kMax {
		return nil, fmt.Errorf("kMin (%g) must be less than kMax (%g)", kMin, kMax)
	}

	const scalePoints = 40
	for i := 0; i <= scalePoints; i++ {
		k := kMin + (kMax-kMin)*float64(i)/float64(scalePoints)
		err := ecs.addPolesFor(k)
		if err != nil {
			return nil, err
		}
	}

	splitGains, err := lin.EvansSplitGains()
	if err != nil {
		return nil, err
	}

	for _, k := range splitGains {
		if k > kMin && k < kMax {
			err := ecs.addPolesFor(k)
			if err != nil {
				return nil, err
			}
		}
	}

	curveList := make([]graph.PlotContent, 0, 5)

	markerStyle := graph.Black.SetStrokeWidth(2)
	if p.Count() > 0 {
		polesMarker := graph.NewCrossMarker(4)
		curveList = append(curveList,
			graph.Scatter{
				Points: graph.PointsFromSlice(p.ToPoints()),
				ShapeLineStyle: graph.ShapeLineStyle{
					Shape:      polesMarker,
					ShapeStyle: markerStyle,
				},
				Title: "Poles",
			},
		)
	}
	if z.Count() > 0 {
		zeroMarker := graph.NewCircleMarker(4)
		curveList = append(curveList,
			graph.Scatter{
				Points: graph.PointsFromSlice(z.ToPoints()),
				ShapeLineStyle: graph.ShapeLineStyle{
					Shape:      zeroMarker,
					ShapeStyle: markerStyle,
				},
				Title: "Zeros",
			},
		)
	}

	curveList = append(curveList, &ecs)

	curveList = append(curveList, Polar{})
	as, order, err := lin.EvansAsymptotesIntersect()
	if err != nil {
		return nil, err
	}
	if order > 0 {
		curveList = append(curveList, Asymptotes{Point: graph.Point{X: as, Y: 0}, Order: order})
	}
	curveList = append(curveList, NewImReLabels())

	return curveList, nil
}

func (l *Linear) EvansAsymptotesIntersect() (float64, int, error) {
	p, err := l.Poles()
	if err != nil {
		return 0, 0, err
	}
	z, err := l.Zeros()
	if err != nil {
		return 0, 0, err
	}
	order := p.Count() - z.Count()
	if order == 0 {
		return 0, 0, nil
	}
	var s float64
	for _, z := range p.roots {
		if math.Abs(imag(z)) < eps {
			s += real(z)
		} else {
			s += real(z) * 2
		}
	}
	for _, z := range z.roots {
		if math.Abs(imag(z)) < eps {
			s -= real(z)
		} else {
			s -= real(z) * 2
		}
	}
	return s / float64(order), order, nil
}

func (l *Linear) EvansSplitPoints() (Roots, error) {
	a := l.Numerator.Mul(l.Denominator.Derivative())
	b := l.Denominator.Mul(l.Numerator.Derivative())
	g := a.Add(b.MulFloat(-1)).Canonical()
	return g.Roots()
}

func (l *Linear) EvansSplitGains() ([]float64, error) {
	f, err := l.EvansSplitPoints()
	if err != nil {
		return nil, err
	}

	var kList []float64
	for _, sp := range f.roots {
		k := -l.Denominator.EvalCplx(sp) / l.Numerator.EvalCplx(sp)
		if /*math.Abs(imag(k)) < 1e-6 && */ real(k) > 0 {
			found := false
			for _, ki := range kList {
				if math.Abs(real(k)-ki) < 1e-6 {
					found = true
					break
				}
			}
			if !found {
				kList = append(kList, real(k))
			}
		}
	}
	sort.Float64s(kList)
	return kList, nil
}

type PolynomialProvider func(k float64) (Polynomial, error)

type evansCurves struct {
	polyProvider        PolynomialProvider
	evPoints            evansPoints
	poleCount           int
	isGenerated         bool
	useComplexNumRefine bool
}

func (ec *evansCurves) String() string {
	return "Evans Curves"
}

func (ec *evansCurves) addPolesFor(k float64) error {
	points, numComp, err := ec.getPoles(k)
	if err != nil {
		return err
	}
	evPoint := evansPoint{points: points, gain: k, numComplex: numComp}
	ec.evPoints = append(ec.evPoints, evPoint)
	return nil
}

func (ec *evansCurves) getPoles(k float64) ([]graph.Point, int, error) {
	poly, err := ec.polyProvider(k)
	if err != nil {
		return nil, 0, err
	}
	poles, err := poly.Roots()
	if err != nil {
		return nil, 0, err
	}

	points := poles.ToPoints()

	if ec.poleCount == 0 {
		ec.poleCount = len(points)
	} else {
		if len(points) != ec.poleCount {
			return nil, 0, fmt.Errorf("unexpected pole count %d instead of %d for k=%g (%v); maybe the Linear System is not causal", len(points), ec.poleCount, k, poly)
		}
	}

	nc := 0
	if ec.useComplexNumRefine {
		nc = poles.NumComplex()
	}

	return points, nc, nil
}

func (ec *evansCurves) refine(p1 evansPoint, p2 evansPoint, m metric, maxDist float64, depth int) error {
	if depth > 0 && (p1.dist(p2, m) > maxDist || p1.numComplex != p2.numComplex) {
		nk := (p1.gain + p2.gain) / 2
		points, numComplex, err := ec.getPoles(nk)
		if err != nil {
			return fmt.Errorf("error in evans refine: %w", err)
		}
		evPoint := evansPoint{points: points, gain: nk, numComplex: numComplex}
		ec.evPoints = append(ec.evPoints, evPoint)

		err = ec.refine(p1, evPoint, m, maxDist, depth-1)
		if err != nil {
			return err
		}
		err = ec.refine(evPoint, p2, m, maxDist, depth-1)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ec *evansCurves) generate(tr graph.Transform) error {
	if ec.isGenerated {
		return nil
	}
	ec.isGenerated = true

	sort.Sort(ec.evPoints)

	const maxDist = 4
	var m metric = func(p1, p2 graph.Point) float64 {
		return tr(p1).DistTo(tr(p2))
	}

	le := len(ec.evPoints)
	for i := 1; i < le; i++ {
		err := ec.refine(ec.evPoints[i-1], ec.evPoints[i], m, maxDist, 15)
		if err != nil {
			return err
		}
	}

	sort.Sort(ec.evPoints)
	for i := 1; i < len(ec.evPoints); i++ {
		ec.evPoints[i-1].dist(ec.evPoints[i], m)
	}

	return nil
}

func (ec *evansCurves) Bounds() (x, y graph.Bounds, err error) {
	for _, ep := range ec.evPoints {
		for _, p := range ep.points {
			x.Merge(p.X)
			y.Merge(p.Y)
		}
	}
	return x, y, nil
}

func (ec *evansCurves) DependantBounds(_, _ graph.Bounds) (x, y graph.Bounds, err error) {
	return graph.Bounds{}, graph.Bounds{}, nil
}

func (ec *evansCurves) DrawTo(plot *graph.Plot, canvas graph.Canvas) error {
	err := ec.generate(plot.GetTransform())
	if err != nil {
		return err
	}

	r := canvas.Rect()
	for i := 0; i < ec.poleCount; i++ {
		canvas.DrawPath(r.IntersectPath(ec.evPoints.getPoints(i)), graph.GetColor(i).SetStrokeWidth(2))
	}

	return nil
}

func (ec *evansCurves) Legend() graph.Legend {
	return graph.Legend{}
}

func RootLocus(cpp PolynomialProvider, kMin, kMax float64, parName string) ([]graph.PlotContent, error) {
	ecs := evansCurves{
		polyProvider:        cpp,
		useComplexNumRefine: true,
	}

	const scalePoints = 40
	for i := 0; i <= scalePoints; i++ {
		k := kMin + (kMax-kMin)*float64(i)/float64(scalePoints)
		err := ecs.addPolesFor(k)
		if err != nil {
			return nil, err
		}
	}

	minMarker := graph.Scatter{
		Points: graph.PointsFromSlice(ecs.evPoints[0].points),
		ShapeLineStyle: graph.ShapeLineStyle{
			Shape:      graph.NewSquareMarker(4),
			ShapeStyle: graph.Black.SetStrokeWidth(2),
		},
		Title: fmt.Sprintf("%s = %g", parName, kMin),
	}
	maxMarker := graph.Scatter{
		Points: graph.PointsFromSlice(ecs.evPoints[len(ecs.evPoints)-1].points),
		ShapeLineStyle: graph.ShapeLineStyle{
			Shape:      graph.NewCircleMarker(4),
			ShapeStyle: graph.Black.SetStrokeWidth(2),
		},
		Title: fmt.Sprintf("%s = %g", parName, kMax),
	}
	return []graph.PlotContent{minMarker, maxMarker, &ecs, Polar{}}, nil
}

type BodePlotContent struct {
	Linear  *Linear
	Latency float64
	Style   *graph.Style
	Title   string
	Steps   int

	wMin, wMax float64
	amplitude  []graph.Point
	phase      []graph.Point
}

func (bpc BodePlotContent) String() string {
	return fmt.Sprintf("BodePlotContent(%s)", bpc.Linear.String())
}

type BodePlot struct {
	amplitude *graph.Plot
	phase     *graph.Plot
}

func (b *BodePlot) DrawTo(canvas graph.Canvas) error {
	bode := graph.SplitHorizontal{Top: b.amplitude, Bottom: b.phase}
	return bode.DrawTo(canvas)
}

func (b *BodePlot) SetFrequencyBounds(min, max float64) {
	b.amplitude.XBounds = graph.NewBounds(min, max)
	b.phase.XBounds = graph.NewBounds(min, max)
}

func (b *BodePlot) SetAmplitudeBounds(min, max float64) {
	b.amplitude.YBounds = graph.NewBounds(min, max)
}

func (b *BodePlot) SetPhaseBounds(min, max float64) {
	b.phase.YBounds = graph.NewBounds(min, max)
}

func (l *Linear) CreateBode(style *graph.Style, title string, steps int) BodePlotContent {
	if steps == 0 {
		steps = 200
	} else if steps < 100 {
		steps = 100
	} else if steps > 2000 {
		steps = 2000
	}
	return BodePlotContent{
		Linear: l,
		Style:  style,
		Title:  title,
		Steps:  steps,
	}
}

func NewBode(wMin, wMax float64) *BodePlot {
	amplitude := &graph.Plot{
		XBounds:       graph.NewBounds(wMin, wMax),
		XAxis:         graph.LogAxis,
		YAxis:         graph.DBAxis,
		Grid:          grParser.GridStyle,
		XLabel:        "ω [rad/s]",
		YLabel:        "Amplitude",
		ProtectLabels: true,
	}
	phase := &graph.Plot{
		XBounds:       graph.NewBounds(wMin, wMax),
		XAxis:         graph.LogAxis,
		YAxis:         graph.CreateFixedStepAxis(45),
		Grid:          grParser.GridStyle,
		XLabel:        "ω [rad/s]",
		YLabel:        "Phase [°]",
		ProtectLabels: true,
	}

	b := BodePlot{
		amplitude: amplitude,
		phase:     phase,
	}
	return &b
}

func (b *BodePlot) Add(bpc BodePlotContent) {
	b.amplitude.AddContent(bodeAmplitude{&bpc})
	b.phase.AddContent(bodePhase{&bpc})
}

func (bpc *BodePlotContent) generate(wMin, wMax float64) {
	if bpc.wMin != wMin || bpc.wMax != wMax {
		bpc.wMin = wMin
		bpc.wMax = wMax

		// compensate expansion of x-axis to make the graphs fill the complete x-range
		logMin := math.Log10(wMin)
		logMax := math.Log10(wMax)
		delta := (logMax - logMin) * 0.02
		logMin -= delta
		logMax += delta
		wMin = math.Pow(10, logMin)
		wMax = math.Pow(10, logMax)

		l := bpc.Linear
		cZero := l.EvalCplx(complex(0, 0))
		lastAngle := 0.0
		if real(cZero) < 0 {
			lastAngle = -180
		}

		wMult := math.Pow(wMax/wMin, 1/float64(bpc.Steps))
		var amplitude []graph.Point
		var phase []graph.Point
		angleOffset := 0.0
		w := wMin
		latFactor := bpc.Latency / math.Pi * 180
		for i := 0; i <= bpc.Steps; i++ {
			c := l.EvalCplx(complex(0, w))
			amp := cmplx.Abs(c)
			angle := cmplx.Phase(c) / math.Pi * 180
			if lastAngle-angle > 180 {
				angleOffset += 360
			}
			if lastAngle-angle < -180 {
				angleOffset -= 360
			}

			lastAngle = angle
			amplitude = append(amplitude, graph.Point{X: w, Y: amp})
			phase = append(phase, graph.Point{X: w, Y: angle + angleOffset - latFactor*w})
			w *= wMult
		}
		bpc.amplitude = amplitude
		bpc.phase = phase
	}
}

type bodePhase struct {
	bodeContent *BodePlotContent
}

func (b bodePhase) Bounds() (x, y graph.Bounds, err error) {
	return graph.Bounds{}, graph.Bounds{}, nil
}

func (b bodePhase) DependantBounds(xGiven, _ graph.Bounds) (x, y graph.Bounds, err error) {
	b.bodeContent.generate(xGiven.Min, xGiven.Max)
	var bounds graph.Bounds
	for _, p := range b.bodeContent.phase {
		bounds.Merge(p.Y)
	}
	return graph.Bounds{}, bounds, nil
}

func (b bodePhase) DrawTo(plot *graph.Plot, canvas graph.Canvas) error {
	b.bodeContent.generate(plot.XBounds.Min, plot.XBounds.Max)
	r := canvas.Rect()
	path := graph.NewPointsPath(false, b.bodeContent.phase...)
	canvas.DrawPath(r.IntersectPath(path), b.bodeContent.Style)
	return nil
}

func (b bodePhase) Legend() graph.Legend {
	return graph.Legend{}
}

type bodeAmplitude struct {
	bodeContent *BodePlotContent
}

func (b bodeAmplitude) Bounds() (x, y graph.Bounds, err error) {
	return graph.Bounds{}, graph.Bounds{}, nil
}

func (b bodeAmplitude) DependantBounds(xGiven, _ graph.Bounds) (x, y graph.Bounds, err error) {
	b.bodeContent.generate(xGiven.Min, xGiven.Max)
	var bounds graph.Bounds
	for _, p := range b.bodeContent.amplitude {
		bounds.Merge(p.Y)
	}
	return graph.Bounds{}, bounds, nil
}

func (b bodeAmplitude) DrawTo(plot *graph.Plot, canvas graph.Canvas) error {
	b.bodeContent.generate(plot.XBounds.Min, plot.XBounds.Max)
	r := canvas.Rect()
	path := graph.NewPointsPath(false, b.bodeContent.amplitude...)
	canvas.DrawPath(r.IntersectPath(path), b.bodeContent.Style)
	return nil
}

func (b bodeAmplitude) Legend() graph.Legend {
	return graph.Legend{ShapeLineStyle: graph.ShapeLineStyle{LineStyle: b.bodeContent.Style}, Name: b.bodeContent.Title}
}

func (l *Linear) NyquistPos(sMax float64) (*graph.ParameterFunc, error) {
	pfPos, err := graph.NewLogParameterFunc(0.001, sMax, 0)
	if err != nil {
		return nil, fmt.Errorf("error creating Nyquist positive frequency parameter function: %w", err)
	}
	pfPos.Func = func(w float64) (graph.Point, error) {
		c := l.EvalCplx(complex(0, w))
		return graph.Point{X: real(c), Y: imag(c)}, nil
	}
	pfPos.Style = posStyle
	pfPos.Title = "ω>0"
	return pfPos, nil
}

func (l *Linear) NyquistNeg(sMax float64) (*graph.ParameterFunc, error) {
	pfNeg, err := graph.NewLogParameterFunc(0.001, sMax, 0)
	if err != nil {
		return nil, fmt.Errorf("error creating Nyquist negative frequency parameter function: %w", err)
	}
	pfNeg.Func = func(w float64) (graph.Point, error) {
		c := l.EvalCplx(complex(0, -w))
		return graph.Point{X: real(c), Y: imag(c)}, nil
	}
	pfNeg.Style = negStyle
	pfNeg.Title = "ω<0"
	return pfNeg, nil
}

var (
	posStyle = graph.Black.SetStrokeWidth(2)
	negStyle = graph.Black.SetDash(4, 4).SetStrokeWidth(2)
)

func (l *Linear) Nyquist(sMax float64, alsoNeg bool) ([]graph.PlotContent, error) {
	cZero := l.EvalCplx(complex(0, 0))
	isZero := !(math.IsNaN(real(cZero)) || math.IsNaN(imag(cZero)))

	var cp []graph.PlotContent
	cp = append(cp, NewImReLabels())
	cp = append(cp, graph.Cross{Style: graph.Gray})
	if alsoNeg {
		neg, err := l.NyquistNeg(sMax)
		if err != nil {
			return nil, err
		}
		cp = append(cp, neg)
		cp = append(cp, graph.Scatter{Points: graph.PointsFromPoint(graph.Point{X: -1, Y: 0}), ShapeLineStyle: graph.ShapeLineStyle{Shape: graph.NewCrossMarker(4), ShapeStyle: graph.Red}})
	}
	zeroMarker := graph.NewCircleMarker(4)
	if isZero {
		cp = append(cp, graph.Scatter{Points: graph.PointsFromPoint(graph.Point{X: real(cZero), Y: imag(cZero)}), ShapeLineStyle: graph.ShapeLineStyle{Shape: zeroMarker, ShapeStyle: graph.Black}, Title: "ω=0"})
	}
	pos, err := l.NyquistPos(sMax)
	if err != nil {
		return nil, err
	}
	cp = append(cp, pos)

	return cp, nil
}

type dataSet struct {
	elements []float64
	cols     int
	rows     int
}

func newDataSet(rows, cols int) *dataSet {
	return &dataSet{
		elements: make([]float64, rows*cols),
		cols:     cols,
		rows:     rows,
	}
}

func (v dataSet) get(row, col int) float64 {
	return v.elements[row*v.cols+col]
}

func (v dataSet) set(row, col int, val float64) {
	v.elements[row*v.cols+col] = val
}

func (v dataSet) toPoints(i0, i1 int) graph.PointsPath {
	return graph.PointsPath{
		Points: func(yield func(graph.Point, error) bool) {
			o := 0
			for range v.rows {
				x := v.elements[o+i0]
				y := v.elements[o+i1]
				if !yield(graph.Point{X: x, Y: y}, nil) {
					return
				}
				o += v.cols
			}
		},
	}
}

func (v dataSet) toPointList(i0, i1 int) *value.List {
	return value.NewListFromSizedIterable(func(_ funcGen.Stack[value.Value], yield iterator.Consumer[value.Value]) error {
		o := 0
		for range v.rows {
			x := value.Float(v.elements[o+i0])
			y := value.Float(v.elements[o+i1])
			if err := yield(value.NewList(x, y)); err != nil {
				return err
			}
			o += v.cols
		}
		return nil
	}, v.rows)
}

func (l *Linear) Simulate(tMax, dt float64, u func(float64) (float64, error)) (*value.List, error) {
	if tMax <= 0 {
		return nil, fmt.Errorf("tMax must be greater than 0")
	}

	lin := l.reduce()

	a, c, d, err := lin.GetStateSpaceRepresentation()
	if err != nil {
		return nil, err
	}

	if dt <= 0 {
		dt = 1e-5
	}

	const pointsExported = 1000
	skip := int(tMax/dt) / pointsExported
	if skip < 1 {
		return nil, fmt.Errorf("step width (dt=%v) is too small for a meaningful simulation", dt)
	}

	t := 0.0
	n := lin.Denominator.Degree()
	x := make(Vector, n)
	xDot := make(Vector, n)

	data := newDataSet(pointsExported+1, 2)
	row := 0
	counter := 0
	for {
		ut, err := u(t)
		if err != nil {
			return nil, err
		}
		y := c.Mul(x) + d*ut
		if counter == 0 {
			data.set(row, 0, t)
			data.set(row, 1, y)
			row++
			if row > pointsExported {
				return data.toPointList(0, 1), nil
			}
			counter = skip
		}
		counter--

		a.Mul(xDot, x)
		xDot[n-1] += ut
		x.Add(dt, xDot)
		t += dt
	}
}

func (l *Linear) GetStateSpaceRepresentation() (Matrix, Vector, float64, error) {
	if !l.IsCausal() {
		return nil, nil, 0, fmt.Errorf("not a propper transfer function, numerator has higher order than denominator")
	}

	n := l.Denominator.Degree()
	norm := l.Denominator[n]
	a := NewMatrix(n, n)
	for i := 1; i < n; i++ {
		a[i-1][i] = 1
	}
	for i := 0; i < n; i++ {
		a[n-1][i] = -l.Denominator[i] / norm
	}
	c := make(Vector, n)
	d := 0.0
	if len(l.Numerator) < n+1 {
		for i := 0; i < len(l.Numerator); i++ {
			c[i] = l.Numerator[i] / norm
		}
	} else {
		d = l.Numerator[n] / norm
		for i := 0; i < n; i++ {
			c[i] = l.Numerator[i]/norm - d*l.Denominator[i]/norm
		}
	}

	return a, c, d, nil
}

func (l *Linear) PMargin() (float64, float64, error) {
	w0 := 0.0
	g := cmplx.Abs(l.EvalCplx(complex(0, w0)))

	if g < 1 {
		return 0, 0, errors.New("no crossover frequency")
	}

	for {
		var w1 float64
		if w0 == 0 {
			w1 = 0.01
		} else {
			w1 = w0 * 1.1
		}
		g = cmplx.Abs(l.EvalCplx(complex(0, w1)))
		if g < 1 {
			var err error
			w0, err = value.Bisection(
				func(w float64) (float64, error) {
					return cmplx.Abs(l.EvalCplx(complex(0, w))) - 1, nil
				},
				w0, w1, 1e-8)
			if err != nil {
				return 0, 0, err
			}
			break
		}
		w0 = w1
	}

	ph := cmplx.Phase(l.EvalCplx(complex(0, w0))) / math.Pi * 180
	if ph > 0 {
		ph = ph - 180
	} else {
		ph = ph + 180
	}

	return w0, ph, nil
}

func (l *Linear) GMargin() (float64, float64, error) {
	w0 := 0.0
	g0 := l.EvalCplx(complex(0, w0))
	for {
		var w1 float64
		if w0 == 0 {
			w1 = 0.01
		} else {
			w1 = w0 * 1.1
			if w1 > 1e8 {
				return 0, 0, errors.New("no gain crossover frequency found")
			}
		}
		g1 := l.EvalCplx(complex(0, w1))

		if real(g0) < 0 && real(g1) < 0 && (imag(g0) > 0) != (imag(g1) > 0) {
			var err error
			w0, err = value.Bisection(func(w float64) (float64, error) {
				return imag(l.EvalCplx(complex(0, w))), nil
			}, w0, w1, 1e-8)
			if err != nil {
				return 0, 0, err
			}
			break
		}
		w0 = w1
		g0 = g1
	}
	gm := cmplx.Abs(l.EvalCplx(complex(0, w0)))
	return w0, 20 * math.Log10(1/gm), nil
}

type Vector []float64

func (v Vector) Mul(x Vector) float64 {
	result := 0.0
	for i := range v {
		result += v[i] * x[i]
	}
	return result
}

func (v Vector) Add(dt float64, dot Vector) {
	for i := range v {
		v[i] += dt * dot[i]
	}
}

type Matrix []Vector

func NewMatrix(rows, cols int) Matrix {
	m := make(Matrix, rows)
	for i := range m {
		m[i] = make(Vector, cols)
	}
	return m
}

func (m Matrix) String() string {
	var str string
	for _, row := range m {
		str += fmt.Sprintf("%v\n", row)
	}
	return str
}

func (m Matrix) Mul(target Vector, x Vector) {
	for i := 0; i < len(m); i++ {
		target[i] = m[i].Mul(x)
	}
}
