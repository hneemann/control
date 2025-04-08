package polynomial

import (
	"fmt"
	"github.com/hneemann/control/graph"
	"math"
	"math/cmplx"
	"sort"
)

type Linear struct {
	Numerator   Polynomial
	Denominator Polynomial
	Latency     float64
	zeros       Roots
	poles       Roots
}

func (l *Linear) Eval(s complex128) complex128 {
	c := l.Numerator.EvalCplx(s) / l.Denominator.EvalCplx(s)
	if l.Latency != 0 {
		c *= cmplx.Exp(complex(-l.Latency, 0) * s)
	}
	return c
}

func (l *Linear) Equals(b *Linear) bool {
	return l.Numerator.Equals(b.Numerator) && l.Denominator.Equals(b.Denominator)
}

func (l *Linear) StringPoly(parse bool) string {
	s := "(" + l.Numerator.intString(parse) + ")/(" + l.Denominator.intString(parse) + ")"
	if l.Latency > 0 {
		s += fmt.Sprintf("*exp(-%gs)", l.Latency)
	}
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
	if l.Latency != 0 {
		sp += fmt.Sprintf("*exp(-%gs)", l.Latency)
	}
	return sp
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
		Latency:     l.Latency + b.Latency,
	}).reduce()
}

func (l *Linear) Inv() *Linear {
	return &Linear{
		Numerator:   l.Denominator,
		Denominator: l.Numerator,
		zeros:       l.poles,
		poles:       l.zeros,
		Latency:     -l.Latency,
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
		Latency:     l.Latency - b.Latency,
	}).reduce()
}

func (l *Linear) Add(b *Linear) (*Linear, error) {
	if l.Latency != 0 || b.Latency != 0 {
		return nil, fmt.Errorf("cannot add linear systems with latency")
	}
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
	if l.Latency != 0 {
		return nil, fmt.Errorf("cannot close the loop if linear system has a latency")
	}

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
			Latency:     l.Latency,
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
			Latency:     l.Latency,
		}
	}
	return l
}

func (l *Linear) MulFloat(f float64) *Linear {
	return &Linear{
		Numerator:   l.Numerator.MulFloat(f),
		Denominator: l.Denominator,
		Latency:     l.Latency,
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

func (p Polar) PreferredBounds(_, _ graph.Bounds) (x, y graph.Bounds) {
	return graph.Bounds{}, graph.Bounds{}
}

func (p Polar) DrawTo(plot *graph.Plot, canvas graph.Canvas) {
	r := canvas.Rect()
	text := graph.Gray.Text()

	textSize := canvas.Context().TextSize * 0.8
	var zero graph.Point
	if r.Inside(zero) {
		radius := r.Diagonal()
		path := graph.NewPath(false)
		for angle := 90; angle <= 270; angle += 15 {
			if angle != 180 {
				x := radius * math.Cos(float64(angle)*math.Pi/180)
				y := radius * math.Sin(float64(angle)*math.Pi/180)
				ep := r.Cut(zero, graph.Point{X: x, Y: y})
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
				path = path.MoveTo(zero).LineTo(ep)
				canvas.DrawText(ep, fmt.Sprintf("%d°", 180-angle), o, text, textSize)
			}
		}
		canvas.DrawPath(path, graph.Gray.SetDash(4, 4))

		path = graph.NewPath(false)
		for _, t := range plot.GetXTicks() {
			radius = -t.Position
			if radius > 1e-5 {
				for angle := 90; angle <= 270; angle += 5 {
					x := radius * math.Cos(float64(angle)*math.Pi/180)
					y := radius * math.Sin(float64(angle)*math.Pi/180)
					point := graph.Point{X: x, Y: y}
					if angle == 90 {
						if r.Inside(point) {
							canvas.DrawText(point, t.Label, graph.VCenter|graph.Left, text, textSize)
						}
						path = path.MoveTo(point)
					} else {
						path = path.LineTo(point)
					}
				}
			}
		}
		canvas.DrawPath(path.Intersect(r), graph.Gray.SetDash(4, 4))
	}
}

type Asymptotes struct {
	Point graph.Point
	Order int
}

var asymptotesStyle = graph.Gray.SetStrokeWidth(2)

func (a Asymptotes) String() string {
	return "Asymptotes"
}

func (a Asymptotes) PreferredBounds(_, _ graph.Bounds) (graph.Bounds, graph.Bounds) {
	return graph.Bounds{}, graph.Bounds{}
}

func (a Asymptotes) DrawTo(_ *graph.Plot, canvas graph.Canvas) {
	r := canvas.Rect()
	if r.Inside(a.Point) {
		d := r.Diagonal()

		dAlpha := 2 * math.Pi / float64(a.Order)
		alpha := dAlpha / 2
		for i := 0; i < a.Order; i++ {
			x := a.Point.X + d*math.Cos(alpha)
			y := a.Point.Y + d*math.Sin(alpha)

			l := graph.NewPath(false).Add(a.Point).Add(graph.Point{X: x, Y: y})
			canvas.DrawPath(l.Intersect(r), asymptotesStyle)

			alpha += dAlpha
		}
	}
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
	if l.Latency != 0 {
		return nil, fmt.Errorf("cannot create Evans plot for linear system with latency")
	}

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

	pathList := make([]graph.Path, poleCount)
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

func (b *BodePlot) DrawTo(canvas graph.Canvas) {
	b.bode.DrawTo(canvas)
}

func (l *Linear) AddToBode(b *BodePlot, style *graph.Style) {
	cZero := l.Eval(complex(0, 0))
	lastAngle := 0.0
	if real(cZero) < 0 {
		lastAngle = -180
	}

	wMult := math.Pow(b.wMax/b.wMin, 0.01)
	amplitude := graph.NewPath(false)
	phase := graph.NewPath(false)
	angleOffset := 0.0
	w := b.wMin
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
		amplitude = amplitude.Add(graph.Point{X: w, Y: 20 * math.Log10(amp)})
		phase = phase.Add(graph.Point{X: w, Y: angle + angleOffset})
		w *= wMult
	}

	b.amplitude.AddContent(graph.Curve{Path: amplitude, Style: style})
	b.phase.AddContent(graph.Curve{Path: phase, Style: style})
}

func NewBode(wMin, wMax float64) *BodePlot {
	amplitude := &graph.Plot{
		XBounds: graph.NewBounds(wMin, wMax),
		XAxis:   graph.LogAxis,
		YAxis:   graph.CreateFixedStepAxis(20),
		Grid:    graph.Gray,
		XLabel:  "ω [rad/s]",
		YLabel:  "Amplitude [dB]",
	}
	phase := &graph.Plot{
		XBounds: graph.NewBounds(wMin, wMax),
		XAxis:   graph.LogAxis,
		YAxis:   graph.CreateFixedStepAxis(45),
		Grid:    graph.Gray,
		XLabel:  "ω [rad/s]",
		YLabel:  "Phase [°]",
	}
	b := BodePlot{wMin, wMax,
		amplitude, phase,
		graph.SplitImage{Top: amplitude, Bottom: phase}}
	return &b
}

func (l *Linear) Nyquist() *graph.Plot {
	cZero := l.Eval(complex(0, 0))
	pZero := graph.Point{X: real(cZero), Y: imag(cZero)}

	path := graph.NewPath(false).Add(pZero)

	wMult := 1.25
	lastW := 0.0
	lastP := pZero
	w := 1.0 / 128
	for range 100 {
		c := l.Eval(complex(0, w))
		p := graph.Point{X: real(c), Y: imag(c)}
		l.refineNy(lastW, lastP, w, p, &path)
		path = path.Add(p)

		lastP = p
		lastW = w
		w *= wMult
	}

	pathNeg := graph.NewPath(false)
	for _, pe := range path.Elements {
		pathNeg = pathNeg.Add(graph.Point{X: pe.X, Y: -pe.Y})
	}

	posStyle := graph.Black.SetStrokeWidth(2)
	negStyle := graph.Black.SetDash(4, 4).SetStrokeWidth(2)

	zeroMarker := graph.NewCircleMarker(4)
	return &graph.Plot{
		XLabel: "Re",
		YLabel: "Im",
		Content: []graph.PlotContent{
			graph.Cross{Style: graph.Gray},
			graph.Curve{Path: path, Style: posStyle},
			graph.Curve{Path: pathNeg, Style: negStyle},
			graph.Scatter{Points: []graph.Point{{X: -1, Y: 0}}, Shape: graph.NewCrossMarker(4), Style: graph.Red},
			graph.Scatter{Points: []graph.Point{pZero}, Shape: zeroMarker, Style: graph.Black},
		},
		Legend: []graph.Legend{
			{Name: "k>0", LineStyle: posStyle},
			{Name: "k=0", Shape: zeroMarker, ShapeStyle: graph.Black},
			{Name: "k<0", LineStyle: negStyle},
		},
	}
}

func (l *Linear) refineNy(w0 float64, p0 graph.Point, w1 float64, p1 graph.Point, path *graph.Path) {
	if p0.DistTo(p1) > 0.05 {
		w := (w0 + w1) / 2
		c := l.Eval(complex(0, w))
		p := graph.Point{X: real(c), Y: imag(c)}
		l.refineNy(w0, p0, w, p, path)
		*path = path.Add(p)
		l.refineNy(w, p, w1, p1, path)
	}
}
