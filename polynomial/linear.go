package polynomial

import (
	"errors"
	"fmt"
	"github.com/hneemann/control/graph"
	"github.com/hneemann/control/graph/grParser"
	"github.com/hneemann/iterator"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
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

func (l *Linear) Eval(s complex128) complex128 {
	c := l.Numerator.EvalCplx(s) / l.Denominator.EvalCplx(s)
	return c
}

func (l *Linear) Equals(b *Linear) bool {
	return l.Numerator.Equals(b.Numerator) && l.Denominator.Equals(b.Denominator)
}

func (l *Linear) StringPoly(parse bool) string {
	s := "(" + l.Numerator.intString(parse) + ")/(" + l.Denominator.intString(parse) + ")"
	return s
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
	sp := fmt.Sprintf("%s/(%s)", n, d)
	return sp
}

func (l *Linear) ToMathML() string {
	result := "<mfrac>"
	result += l.Numerator.ToMathML()
	result += l.Denominator.ToMathML()
	return result + "</mfrac>"
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

type evansPoint struct {
	points []graph.Point
	gain   float64
}

func (e evansPoint) dist(other evansPoint) float64 {
	var maxDist float64
	op := other.points
	for i, ep := range e.points {
		var best int
		bestDist := math.Inf(1)
		for j := i; j < len(op); j++ {
			d := ep.DistTo(op[j])
			if d < bestDist {
				best = j
				bestDist = d
			}
		}
		if best != i {
			op[i], op[best] = op[best], op[i]
		}
		d := ep.DistTo(op[i])
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

type Polar struct{}

func (p Polar) String() string {
	return "Polar Grid"
}

func (p Polar) PreferredBounds(_, _ graph.Bounds) (x, y graph.Bounds, e error) {
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
	r := canvas.Rect()
	text := graph.Gray.Text()
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
				canvas.DrawPath(graph.NewPointsPath(false, ap, ep), grParser.GridStyle)
				canvas.DrawText(ep, fmt.Sprintf("%d°", 180-angle), o, text, textSize)
			}
		}
	}

	if r.Inside(zero) {
		for _, t := range plot.GetXTicks() {
			radius = -t.Position
			if radius > 1e-5 {
				canvas.DrawPath(r.IntersectPath(polarPath{radius: radius, r: r}), grParser.GridStyle)
				point := graph.Point{X: 0, Y: radius}
				if r.Inside(point) {
					canvas.DrawText(point, t.Label, graph.VCenter|graph.Left, text, textSize)
				}
			}
		}
	}
	return nil
}

type Asymptotes struct {
	Point graph.Point
	Order int
}

var asymptotesStyle = graph.Gray.SetStrokeWidth(2)

func (a Asymptotes) String() string {
	return "Asymptotes"
}

func (a Asymptotes) PreferredBounds(_, _ graph.Bounds) (graph.Bounds, graph.Bounds, error) {
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

var styleList = []*graph.Style{
	graph.Red.SetStrokeWidth(2).Darker(),
	graph.Green.SetStrokeWidth(2).Darker(),
	graph.Blue.SetStrokeWidth(2).Darker(),
	graph.Cyan.SetStrokeWidth(2).Darker(),
	graph.Magenta.SetStrokeWidth(2).Darker(),
	graph.Yellow.SetStrokeWidth(2).Darker(),
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

	var evPoints evansPoints
	evPoints = append(evPoints, evansPoint{points: p.ToPoints(), gain: 0})

	poleCount := p.Count()

	points, err := l.getPoles(kMax, poleCount)
	if err != nil {
		return nil, err
	}
	evPoints = append(evPoints, evansPoint{points: points, gain: kMax})

	splitGains, err := l.EvansSplitGains()
	if err != nil {
		return nil, err
	}

	for _, k := range splitGains {
		if k < kMax {
			points, err = l.getPoles(k, poleCount)
			if err != nil {
				return nil, err
			}
			evPoints = append(evPoints, evansPoint{points: points, gain: k})
		}
	}

	sort.Sort(evPoints)

	le := len(evPoints)
	for i := 1; i < le; i++ {
		err = l.refine(evPoints[i-1], evPoints[i], &evPoints, poleCount)
		if err != nil {
			return nil, err
		}
	}

	sort.Sort(evPoints)
	for i := 1; i < len(evPoints); i++ {
		evPoints[i-1].dist(evPoints[i])
	}

	pathList := make([]graph.SlicePath, poleCount)
	for _, pl := range evPoints {
		for i := range poleCount {
			pathList[i] = pathList[i].Add(pl.points[i])
		}
	}

	curveList := make([]graph.PlotContent, 0, len(pathList)+2)
	curveList = append(curveList, Polar{})
	var legend []graph.Legend

	as, order, err := l.EvansAsymptotesIntersect()
	if err != nil {
		return nil, err
	}
	if order > 0 {
		curveList = append(curveList, Asymptotes{Point: graph.Point{X: as, Y: 0}, Order: order})
	}

	for i, pa := range pathList {
		curveList = append(curveList, graph.Curve{Path: pa, Style: styleList[i%len(styleList)]})
	}

	markerStyle := graph.Black.SetStrokeWidth(2)
	if poleCount > 0 {
		polesMarker := graph.NewCrossMarker(4)
		curveList = append(curveList,
			graph.Scatter{
				Points: p.ToPoints(),
				Shape:  polesMarker,
				Style:  markerStyle,
			},
		)
		legend = append(legend,
			graph.Legend{
				Name:       "Poles",
				Shape:      polesMarker,
				ShapeStyle: markerStyle,
			},
		)
	}
	if z.Count() > 0 {
		zeroMarker := graph.NewCircleMarker(4)
		curveList = append(curveList,
			graph.Scatter{
				Points: z.ToPoints(),
				Shape:  zeroMarker,
				Style:  markerStyle,
			},
		)
		legend = append(legend,
			graph.Legend{
				Name:       "Zeros",
				Shape:      zeroMarker,
				ShapeStyle: markerStyle,
			},
		)
	}

	if order > 0 {
		legend = append(legend, graph.Legend{
			Name:      "Asymptotes",
			LineStyle: asymptotesStyle,
		})
	}

	return &graph.Plot{
		XLabel:  "Re",
		YLabel:  "Im",
		Content: curveList,
		Legend:  legend,
	}, nil

}
func (l *Linear) refine(p1 evansPoint, p2 evansPoint, e *evansPoints, poleCount int) error {
	dist := p1.dist(p2)
	if dist > 0.1 {
		nk := (p1.gain + p2.gain) / 2
		points, err := l.getPoles(nk, poleCount)
		if err != nil {
			return err
		}
		evPoint := evansPoint{points: points, gain: nk}
		*e = append(*e, evPoint)

		err = l.refine(p1, evPoint, e, poleCount)
		if err != nil {
			return err
		}
		err = l.refine(evPoint, p2, e, poleCount)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *Linear) getPoles(k float64, poleCount int) ([]graph.Point, error) {
	gw, err := l.MulFloat(k).Loop()
	if err != nil {
		return nil, err
	}
	poles, err := gw.Poles()
	if err != nil {
		return nil, err
	}

	points := poles.ToPoints()

	if len(points) != poleCount {
		return nil, fmt.Errorf("unexpected pole count: %d", len(points))
	}

	return points, nil
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

func (l *Linear) EvansSplitPoints() ([]float64, error) {
	a := l.Numerator.Mul(l.Denominator.Derivative())
	b := l.Denominator.Mul(l.Numerator.Derivative())
	g := a.Add(b.MulFloat(-1)).Canonical()

	r, err := g.Roots()
	if err != nil {
		return nil, err
	}

	p, err := l.Poles()
	if err != nil {
		return nil, err
	}

	z, err := l.Zeros()
	if err != nil {
		return nil, err
	}

	sp := append(p.OnlyReal(), z.OnlyReal()...)
	sort.Float64s(sp)

	var f []float64
	for _, can := range r.OnlyReal() {
		n := 0
		for _, s := range sp {
			if can < s {
				n++
			}
		}
		if n&1 == 1 {
			f = append(f, can)
		}
	}

	return f, nil
}

func (l *Linear) EvansSplitGains() ([]float64, error) {
	f, err := l.EvansSplitPoints()
	if err != nil {
		return nil, err
	}
	for i, sp := range f {
		f[i] = -l.Denominator.Eval(sp) / l.Numerator.Eval(sp)
	}
	return f, nil
}

type BodePlot struct {
	wMin, wMax float64
	amplitude  *graph.Plot
	phase      *graph.Plot
	bode       graph.SplitImage
}

func (b *BodePlot) DrawTo(canvas graph.Canvas) error {
	return b.bode.DrawTo(canvas)
}

func (b *BodePlot) AddLegend(s string, style *graph.Style) {
	b.amplitude.AddLegend(s, style, nil, nil)
}

func (b *BodePlot) SetAmplitudeBounds(min, max float64) {
	b.amplitude.YBounds = graph.NewBounds(min, max)
}

func (b *BodePlot) SetPhaseBounds(min, max float64) {
	b.phase.YBounds = graph.NewBounds(min, max)
}

func (l *Linear) AddToBode(b *BodePlot, style *graph.Style, latency float64) {
	cZero := l.Eval(complex(0, 0))
	lastAngle := 0.0
	if real(cZero) < 0 {
		lastAngle = -180
	}

	wMult := math.Pow(b.wMax/b.wMin, 0.01)
	var amplitude []graph.Point
	var phase []graph.Point
	angleOffset := 0.0
	w := b.wMin
	latFactor := latency / math.Pi * 180
	for i := 0; i <= 100; i++ {
		c := l.Eval(complex(0, w))
		amp := cmplx.Abs(c)
		angle := cmplx.Phase(c) / math.Pi * 180
		if lastAngle-angle > 180 {
			angleOffset += 360
		}
		if lastAngle-angle < -180 {
			angleOffset -= 360
		}

		lastAngle = angle
		amplitude = append(amplitude, graph.Point{X: w, Y: 20 * math.Log10(amp)})
		phase = append(phase, graph.Point{X: w, Y: angle + angleOffset - latFactor*w})
		w *= wMult
	}

	b.amplitude.AddContent(graph.Curve{Path: graph.NewPointsPath(false, amplitude...), Style: style})
	b.phase.AddContent(graph.Curve{Path: graph.NewPointsPath(false, phase...), Style: style})
}

func NewBode(wMin, wMax float64) *BodePlot {
	amplitude := &graph.Plot{
		XBounds: graph.NewBounds(wMin, wMax),
		XAxis:   graph.LogAxis,
		YAxis:   graph.CreateFixedStepAxis(20),
		Grid:    grParser.GridStyle,
		XLabel:  "ω [rad/s]",
		YLabel:  "Amplitude [dB]",
	}
	phase := &graph.Plot{
		XBounds: graph.NewBounds(wMin, wMax),
		XAxis:   graph.LogAxis,
		YAxis:   graph.CreateFixedStepAxis(45),
		Grid:    grParser.GridStyle,
		XLabel:  "ω [rad/s]",
		YLabel:  "Phase [°]",
	}
	b := BodePlot{wMin, wMax,
		amplitude, phase,
		graph.SplitImage{Top: amplitude, Bottom: phase}}
	return &b
}

func (l *Linear) NyquistPos(style *graph.Style) graph.PlotContent {
	pfPos := graph.NewLogParameterFunc(0.001, 1000)
	pfPos.Func = func(w float64) (graph.Point, error) {
		c := l.Eval(complex(0, w))
		return graph.Point{X: real(c), Y: imag(c)}, nil
	}
	pfPos.Style = style
	return pfPos
}

func (l *Linear) NyquistNeg(style *graph.Style) graph.PlotContent {
	pfNeg := graph.NewLogParameterFunc(0.001, 1000)
	pfNeg.Func = func(w float64) (graph.Point, error) {
		c := l.Eval(complex(0, -w))
		return graph.Point{X: real(c), Y: imag(c)}, nil
	}
	pfNeg.Style = style
	return pfNeg
}

var (
	posStyle = graph.Black.SetStrokeWidth(2)
	negStyle = graph.Black.SetDash(4, 4).SetStrokeWidth(2)
)

func (l *Linear) Nyquist(alsoNeg bool) (*graph.Plot, error) {
	cZero := l.Eval(complex(0, 0))
	isZero := !(math.IsNaN(real(cZero)) || math.IsNaN(imag(cZero)))

	var cp []graph.PlotContent
	cp = append(cp, graph.Cross{Style: graph.Gray})
	cp = append(cp, l.NyquistPos(posStyle))
	if alsoNeg {
		cp = append(cp, l.NyquistNeg(negStyle))
		cp = append(cp, graph.Scatter{Points: []graph.Point{{X: -1, Y: 0}}, Shape: graph.NewCrossMarker(4), Style: graph.Red})
	}
	zeroMarker := graph.NewCircleMarker(4)
	if isZero {
		cp = append(cp, graph.Scatter{Points: []graph.Point{{X: real(cZero), Y: imag(cZero)}}, Shape: zeroMarker, Style: graph.Black})
	}

	var legend []graph.Legend
	legend = append(legend, graph.Legend{Name: "ω>0", LineStyle: posStyle})
	if isZero {
		legend = append(legend, graph.Legend{Name: "ω=0", Shape: zeroMarker, ShapeStyle: graph.Black})
	}
	if alsoNeg {
		legend = append(legend, graph.Legend{Name: "ω<0", LineStyle: negStyle})
	}

	return &graph.Plot{
		XLabel:  "Re",
		YLabel:  "Im",
		Content: cp,
		Legend:  legend,
	}, nil
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

func (v dataSet) toList() *value.List {
	return value.NewListFromIterable(func(st funcGen.Stack[value.Value], yield iterator.Consumer[value.Value]) error {
		o := 0
		for range v.rows {
			row := value.NewListConvert[float64](func(v float64) value.Value {
				return value.Float(v)
			}, v.elements[o:o+v.cols])

			if err := yield(row); err != nil {
				return err
			}
			o += v.cols
		}
		return nil
	})
}

func (l *Linear) Simulate(tMax float64, u func(float64) (float64, error)) (*value.List, error) {
	if tMax <= 0 {
		return nil, fmt.Errorf("tMax must be greater than 0")
	}

	lin := l.reduce()

	a, c, d, err := lin.GetStateSpaceRepresentation()
	if err != nil {
		return nil, err
	}

	const pointsExported = 1000
	const pointsInternal = 100000
	const skip = pointsInternal / pointsExported
	dt := tMax / pointsInternal
	t := 0.0
	n := len(lin.Denominator) - 1
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
				return data.toList(), nil
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
	n := len(l.Denominator) - 1

	if len(l.Numerator) > n+1 {
		return nil, nil, 0, fmt.Errorf("not a propper transfer function, numerator has higher order than denominator")
	}

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

func bisection(x0, x1 float64, f func(float64) float64) (float64, error) {
	x0Pos := f(x0) > 0
	for i := 0; i < 100; i++ {
		x := (x0 + x1) / 2
		if (f(x) > 0) == x0Pos {
			x0 = x
		} else {
			x1 = x
		}
		if math.Abs(x1-x0) < eps {
			return x, nil
		}
	}
	return 0, fmt.Errorf("bisection failed")
}

func (l *Linear) PMargin() (float64, float64, error) {
	w0 := 0.0
	g := cmplx.Abs(l.Eval(complex(0, w0)))

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
		g = cmplx.Abs(l.Eval(complex(0, w1)))
		if g < 1 {
			var err error
			w0, err = bisection(w0, w1, func(w float64) float64 {
				return cmplx.Abs(l.Eval(complex(0, w))) - 1
			})
			if err != nil {
				return 0, 0, err
			}
			break
		}
		w0 = w1
	}

	ph := cmplx.Phase(l.Eval(complex(0, w0))) / math.Pi * 180
	if ph > 0 {
		ph = ph - 180
	} else {
		ph = ph + 180
	}

	return w0, ph, nil
}

func (l *Linear) GMargin() (float64, float64, error) {
	w0 := 0.0
	im0 := imag(l.Eval(complex(0, w0)))
	for {
		var w1 float64
		if w0 == 0 {
			w1 = 0.01
		} else {
			w1 = w0 * 1.1
		}
		im1 := imag(l.Eval(complex(0, w1)))

		if (im0 > 0) != (im1 > 0) {
			var err error
			w0, err = bisection(w0, w1, func(w float64) float64 {
				return imag(l.Eval(complex(0, w)))
			})
			if err != nil {
				return 0, 0, err
			}
			break
		}
		w0 = w1
		im0 = im1
	}
	gm := cmplx.Abs(l.Eval(complex(0, w0)))
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
