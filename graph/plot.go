package graph

import (
	"bytes"
	"fmt"
	"math"
)

type Legend struct {
	Name       string
	LineStyle  *Style
	Shape      Shape
	ShapeStyle *Style
}

type BoundsModifier func(xBounds, yBounds Bounds, p *Plot, canvas Canvas) (Bounds, Bounds)

func Zoom(p Point, f float64) BoundsModifier {
	return func(xBounds, yBounds Bounds, _ *Plot, canvas Canvas) (Bounds, Bounds) {
		if xBounds.valid {
			d := xBounds.Width() / (2 * f)
			xBounds.Min = p.X - d
			xBounds.Max = p.X + d
		}
		if yBounds.valid {
			d := yBounds.Width() / (2 * f)
			yBounds.Min = p.Y - d
			yBounds.Max = p.Y + d
		}
		return xBounds, yBounds
	}
}

type Plot struct {
	XAxis          Axis
	YAxis          Axis
	XBounds        Bounds
	YBounds        Bounds
	LeftBorder     float64
	RightBorder    float64
	Grid           *Style
	Title          string
	XLabel         string
	YLabel         string
	Content        []PlotContent
	Legend         []Legend
	FillBackground bool
	BoundsModifier BoundsModifier
	xTicks         []Tick
	yTicks         []Tick
	legendPosGiven bool
	legendPos      Point
	trans          Transform
	canvas         Canvas
}

func (p *Plot) DrawTo(canvas Canvas) error {
	p.canvas = canvas
	c := canvas.Context()
	rect := canvas.Rect()
	textStyle := Black.Text()
	textSize := c.TextSize
	if textSize <= rect.Height()/200 {
		textSize = rect.Height() / 200
	}

	if p.FillBackground {
		canvas.DrawPath(rect.Poly(), White.SetStrokeWidth(0).SetFill(White))
	}

	lb := p.LeftBorder
	if lb <= 0 {
		lb = 5
	}
	rb := p.RightBorder
	if rb <= 0 {
		rb = 2
	}

	innerRect := Rect{
		Min: Point{rect.Min.X + textSize*lb*0.75, rect.Min.Y + textSize*2},
		Max: Point{rect.Max.X - textSize*rb*0.75, rect.Max.Y - textSize/2},
	}

	xBounds := p.XBounds
	yBounds := p.YBounds

	if !(xBounds.valid && yBounds.valid) {
		mergeX := !xBounds.valid
		mergeY := !yBounds.valid
		for _, plotContent := range p.Content {
			x, y, err := plotContent.PreferredBounds(p.XBounds, p.YBounds)
			if err != nil {
				return err
			}
			if mergeX {
				xBounds.MergeBounds(x)
			}
			if mergeY {
				yBounds.MergeBounds(y)
			}
		}
	}

	if p.BoundsModifier != nil {
		xBounds, yBounds = p.BoundsModifier(xBounds, yBounds, p, canvas)
	}

	if !xBounds.valid {
		xBounds = NewBounds(-1, 1)
	}
	if !yBounds.valid {
		yBounds = NewBounds(-1, 1)
	}

	xAxis := p.XAxis
	if xAxis == nil {
		xAxis = LinearAxis
	}
	yAxis := p.YAxis
	if yAxis == nil {
		yAxis = LinearAxis
	}

	xTrans, xTicks, xBounds := xAxis(innerRect.Min.X, innerRect.Max.X, xBounds,
		func(width float64, digits int) bool {
			return width > textSize*(float64(digits+2))*0.75
		})
	yTrans, yTicks, yBounds := yAxis(innerRect.Min.Y, innerRect.Max.Y, yBounds,
		func(width float64, _ int) bool {
			return width > textSize*2
		})

	p.xTicks = xTicks
	p.yTicks = yTicks

	p.trans = func(p Point) Point {
		return Point{xTrans(p.X), yTrans(p.Y)}
	}

	inner := TransformCanvas{
		transform: p.trans,
		parent:    canvas,
		size: Rect{
			Min: Point{xBounds.Min, yBounds.Min},
			Max: Point{xBounds.Max, yBounds.Max},
		},
	}

	large := textSize / 2
	small := textSize / 4

	thickLine := Black.SetStrokeWidth(2)
	thinLine := Black.SetStrokeWidth(1)

	for _, tick := range xTicks {
		xp := xTrans(tick.Position)
		if tick.Label == "" {
			canvas.DrawPath(NewPointsPath(false, Point{xp, innerRect.Min.Y - small}, Point{xp, innerRect.Min.Y}), thinLine)
		} else {
			canvas.DrawText(Point{xp, innerRect.Min.Y - large - small}, tick.Label, Top|HCenter, textStyle, textSize)
			canvas.DrawPath(NewPointsPath(false, Point{xp, innerRect.Min.Y - large}, Point{xp, innerRect.Min.Y}), thickLine)
		}
		if p.Grid != nil {
			canvas.DrawPath(NewPointsPath(false, Point{xp, innerRect.Min.Y}, Point{xp, innerRect.Max.Y}), p.Grid)
		}
	}
	canvas.DrawText(Point{innerRect.Max.X - small, innerRect.Min.Y + small}, p.XLabel, Bottom|Right, textStyle, textSize)
	for _, tick := range yTicks {
		yp := yTrans(tick.Position)
		if tick.Label == "" {
			canvas.DrawPath(NewPointsPath(false, Point{innerRect.Min.X - small, yp}, Point{innerRect.Min.X, yp}), thinLine)
		} else {
			canvas.DrawText(Point{innerRect.Min.X - large, yp}, tick.Label, Right|VCenter, textStyle, textSize)
			canvas.DrawPath(NewPointsPath(false, Point{innerRect.Min.X - large, yp}, Point{innerRect.Min.X, yp}), thickLine)
		}
		if p.Grid != nil {
			canvas.DrawPath(NewPointsPath(false, Point{innerRect.Min.X, yp}, Point{innerRect.Max.X, yp}), p.Grid)
		}
	}
	canvas.DrawText(Point{innerRect.Min.X + small, innerRect.Max.Y - small}, p.YLabel, Top|Left, textStyle, textSize)
	if p.Title != "" {
		canvas.DrawText(Point{innerRect.Max.X - small, innerRect.Max.Y - small}, p.Title, Top|Right, textStyle, textSize)
	}

	for _, plotContent := range p.Content {
		err := plotContent.DrawTo(p, inner)
		if err != nil {
			return err
		}
	}

	canvas.DrawPath(innerRect.Poly(), thickLine)

	if len(p.Legend) > 0 {
		var lp Point
		if p.legendPosGiven {
			lp = Point{xTrans(p.legendPos.X), yTrans(p.legendPos.Y)}
		} else {
			lp = Point{innerRect.Min.X + textSize*3, innerRect.Min.Y + textSize*(float64(len(p.Legend))*1.5-0.5)}
		}
		for _, leg := range p.Legend {
			canvas.DrawText(lp, leg.Name, Left|VCenter, textStyle, textSize)
			if leg.Shape != nil && leg.ShapeStyle != nil {
				canvas.DrawShape(lp.Add(Point{-1*textSize - small, 0}), leg.Shape, leg.ShapeStyle)
			}
			if leg.LineStyle != nil {
				canvas.DrawPath(NewPointsPath(false, lp.Add(Point{-2*textSize - small, 0}), lp.Add(Point{-small, 0})), leg.LineStyle)
			}
			lp = lp.Add(Point{0, -textSize * 1.5})
		}

	}
	return nil
}

func (p *Plot) GetXTicks() []Tick {
	return p.xTicks
}

func (p *Plot) GetYTicks() []Tick {
	return p.yTicks
}

func (p *Plot) AddContent(content PlotContent) {
	p.Content = append(p.Content, content)
}

func (p *Plot) AddLegend(name string, lineStyle *Style, shape Shape, shapeStyle *Style) {
	p.Legend = append(p.Legend, Legend{
		Name:       name,
		LineStyle:  lineStyle,
		Shape:      shape,
		ShapeStyle: shapeStyle,
	})
}

func (p *Plot) SetLegendPosition(pos Point) {
	p.legendPosGiven = true
	p.legendPos = pos
}

func (p *Plot) String() string {
	bu := bytes.Buffer{}
	bu.WriteString("Plot: ")
	for i, content := range p.Content {
		if i > 0 {
			bu.WriteString(", ")
		}
		bu.WriteString(fmt.Sprint(content))
	}
	return bu.String()
}

func (p *Plot) Dist(p1, p2 Point) float64 {
	return p.trans(p1).DistTo(p.trans(p2))
}

type Bounds struct {
	valid    bool
	Min, Max float64
}

func NewBounds(min, max float64) Bounds {
	if min > max {
		min, max = max, min
	}
	return Bounds{true, min, max}
}

func (b *Bounds) Valid() bool {
	return b.valid
}

func (b *Bounds) MergeBounds(other Bounds) {
	if other.valid {
		// other is available
		if !b.valid {
			b.valid = true
			b.Min = other.Min
			b.Max = other.Max
		} else {
			// both are available
			if b.Min > other.Min {
				b.Min = other.Min
			}
			if b.Max < other.Max {
				b.Max = other.Max
			}
		}
	}
}

func (b *Bounds) Merge(p float64) {
	if !math.IsNaN(p) {
		if !b.valid {
			b.valid = true
			b.Min = p
			b.Max = p
		} else {
			if p < b.Min {
				b.Min = p
			}
			if p > b.Max {
				b.Max = p
			}
		}
	}
}

func (b *Bounds) Width() float64 {
	return b.Max - b.Min
}

type PlotContent interface {
	// DrawTo draws the content to the given canvas
	// The *Plot is passed to allow the content to access the plot's properties
	DrawTo(*Plot, Canvas) error
	// PreferredBounds returns the preferred bounds for the content
	// The first bounds is the x-axis, the second is the y-axis
	// The given bounds are valid if they are set by the user
	PreferredBounds(xGiven, yGiven Bounds) (x, y Bounds, err error)
}

type Function struct {
	Function func(x float64) (float64, error)
	Style    *Style
}

const functionSteps = 200

func (f Function) String() string {
	return "Function"
}

func (f Function) PreferredBounds(xGiven, _ Bounds) (Bounds, Bounds, error) {
	if xGiven.valid {
		yBounds := Bounds{}
		width := xGiven.Width()
		for i := 0; i <= functionSteps; i++ {
			x := xGiven.Min + width*float64(i)/functionSteps
			y, err := f.Function(x)
			if err != nil {
				return Bounds{}, Bounds{}, err
			}
			yBounds.Merge(y)
		}
		return Bounds{}, yBounds, nil
	}
	return Bounds{}, Bounds{}, nil
}

type funcPath struct {
	f func(x float64) (float64, error)
	r Rect
	e error
}

func (f *funcPath) Iter(yield func(rune, Point) bool) {
	width := f.r.Width()
	xo := f.r.Min.X
	for i := 0; i <= functionSteps; i++ {
		x := xo + width*float64(i)/functionSteps
		y, err := f.f(x)
		if err != nil {
			f.e = err
			return
		}
		mode := 'L'
		if i == 0 {
			mode = 'M'
		}
		if !yield(mode, Point{x, y}) {
			return
		}
	}
}

func (f *funcPath) IsClosed() bool {
	return false
}

func (f Function) DrawTo(_ *Plot, canvas Canvas) error {
	path := funcPath{
		f: f.Function,
		r: canvas.Rect(),
	}
	canvas.DrawPath(canvas.Rect().IntersectPath(&path), f.Style)
	return path.e
}

type Scatter struct {
	Points     Points
	Shape      Shape
	ShapeStyle *Style
	LineStyle  *Style
}

func (s Scatter) String() string {
	return "Scatter"
}

func (s Scatter) PreferredBounds(_, _ Bounds) (Bounds, Bounds, error) {
	var x, y Bounds
	for p, err := range s.Points {
		if err != nil {
			return Bounds{}, Bounds{}, err
		}
		x.Merge(p.X)
		y.Merge(p.Y)
	}
	return x, y, nil
}

func (s Scatter) DrawTo(_ *Plot, canvas Canvas) error {
	rect := canvas.Rect()
	if s.ShapeStyle != nil && s.Shape != nil {
		for p := range s.Points {
			if rect.Inside(p) {
				canvas.DrawShape(p, s.Shape, s.ShapeStyle)
			}
		}
	}
	if s.LineStyle != nil {
		canvas.DrawPath(canvas.Rect().IntersectPath(s.Points), s.LineStyle)
	}
	return nil
}

type Hint struct {
	Text        string
	Pos         Point
	Marker      Shape
	MarkerStyle *Style
}

func (h Hint) PreferredBounds(_, _ Bounds) (Bounds, Bounds, error) {
	var x Bounds
	x.Merge(h.Pos.X)
	var y Bounds
	y.Merge(h.Pos.Y)
	return x, y, nil
}

func (h Hint) DrawTo(_ *Plot, canvas Canvas) error {
	r := canvas.Rect()
	if r.Inside(h.Pos) {
		if h.Marker != nil && h.MarkerStyle != nil {
			canvas.DrawShape(h.Pos, h.Marker, h.MarkerStyle)
		}
		tPos := h.Pos
		var o Orientation
		dx := r.Width() / 30
		if r.IsInLeftHalf(h.Pos) {
			o = Left
			tPos = tPos.Add(Point{dx, 0})
		} else {
			o = Right
			tPos = tPos.Add(Point{-dx, 0})
		}
		dy := r.Height() / 30
		if r.IsInTopHalf(h.Pos) {
			o |= Top
			tPos = tPos.Add(Point{0, -dy})
		} else {
			o |= Bottom
			tPos = tPos.Add(Point{0, dy})
		}

		canvas.DrawPath(NewPointsPath(false, h.Pos, tPos), Black)
		canvas.DrawText(tPos, h.Text, o, Black.Text(), canvas.Context().TextSize)
	}
	return nil
}

type HintDir struct {
	Hint
	PosDir Point
}

func (h HintDir) DrawTo(_ *Plot, canvas Canvas) error {
	r := canvas.Rect()
	if r.Inside(h.Pos) {
		if h.Marker != nil && h.MarkerStyle != nil {
			canvas.DrawShape(h.Pos, h.Marker, h.MarkerStyle)
		}

		if tc, ok := canvas.(TransformCanvas); ok {
			parentCanvas := tc.parent
			p1 := tc.transform(h.Pos)
			p2 := tc.transform(h.PosDir)

			delta := p1.Sub(p2).Norm().Rot90().Mul(parentCanvas.Rect().Width() / 30)
			tPos := p1.Add(delta)

			var o Orientation
			if delta.X > 0 {
				o = Left
			} else if delta.X < 0 {
				o = Right
			} else {
				o = HCenter
			}
			if delta.Y > 0 {
				o |= Bottom
			} else if delta.Y < 0 {
				o |= Top
			} else {
				o |= VCenter
			}
			parentCanvas.DrawPath(NewPointsPath(false, p1, tPos), Black)
			parentCanvas.DrawText(tPos, h.Text, o, Black.Text(), canvas.Context().TextSize)

		}
	}
	return nil
}

type circleMarker struct {
	p1, p2 Point
}

func NewCircleMarker(r float64) Shape {
	p1 := Point{X: -r, Y: -r}
	p2 := Point{X: r, Y: r}
	return circleMarker{p1: p1, p2: p2}
}

func (c circleMarker) DrawTo(canvas Canvas, style *Style) {
	canvas.DrawCircle(c.p1, c.p2, style)
}

func NewCrossMarker(r float64) Shape {
	return NewPath(false).
		AddMode('M', Point{-r, -r}).
		AddMode('L', Point{r, r}).
		AddMode('M', Point{-r, r}).
		AddMode('L', Point{r, -r})
}

func NewSquareMarker(r float64) Shape {
	return pointsPath{
		points: []Point{{-r, -r}, {-r, r}, {r, r}, {r, -r}},
		closed: true,
	}
}

func NewTriangleMarker(r float64) Shape {
	return pointsPath{
		points: []Point{{0, r}, {-r, -r}, {r, -r}},
		closed: true,
	}
}

type Cross struct {
	Style *Style
}

func (c Cross) String() string {
	return "coordinate cross"
}

func (c Cross) PreferredBounds(_, _ Bounds) (Bounds, Bounds, error) {
	return Bounds{}, Bounds{}, nil
}

func (c Cross) DrawTo(_ *Plot, canvas Canvas) error {
	r := canvas.Rect()
	if r.Inside(Point{0, 0}) {
		canvas.DrawPath(NewPath(false).
			MoveTo(Point{r.Min.X, 0}).
			LineTo(Point{r.Max.X, 0}).
			MoveTo(Point{0, r.Min.Y}).
			LineTo(Point{0, r.Max.Y}), c.Style)
	}
	return nil
}

type ParameterFunc struct {
	Func     func(t float64) (Point, error)
	Points   int
	InitialT float64
	NextT    func(float64) float64
	Style    *Style
}

func NewLinearParameterFunc(tMin, tMax float64) *ParameterFunc {
	delta := (tMax - tMin) / float64(functionSteps)
	return &ParameterFunc{
		Points:   functionSteps,
		InitialT: tMin,
		NextT: func(t float64) float64 {
			return t + delta
		},
	}
}

func NewLogParameterFunc(tMin, tMax float64) *ParameterFunc {
	f := math.Pow(tMax/tMin, 1/float64(functionSteps))
	return &ParameterFunc{
		Points:   functionSteps,
		InitialT: tMin,
		NextT: func(t float64) float64 {
			return t * f
		},
	}
}

func (p *ParameterFunc) String() string {
	return fmt.Sprintf("Parameter curve with %d points", p.Points)
}

func (p *ParameterFunc) PreferredBounds(_, _ Bounds) (Bounds, Bounds, error) {
	xb, yb := Bounds{}, Bounds{}
	t := p.InitialT
	for i := 0; i <= p.Points; i++ {
		point, err := p.Func(t)
		if err != nil {
			return Bounds{}, Bounds{}, err
		}
		xb.Merge(point.X)
		yb.Merge(point.Y)
		t = p.NextT(t)
	}
	return xb, yb, nil
}

func (p *ParameterFunc) DrawTo(plot *Plot, canvas Canvas) error {
	path := pFuncPath{
		pf:   p,
		plot: plot,
		r:    canvas.Rect(),
	}
	canvas.DrawPath(canvas.Rect().IntersectPath(&path), p.Style)
	return path.e
}

type pFuncPath struct {
	pf   *ParameterFunc
	plot *Plot
	r    Rect
	e    error
}

func (p *pFuncPath) Iter(yield func(rune, Point) bool) {
	pf := p.pf
	t0 := pf.InitialT
	p0, err := pf.Func(t0)
	if err != nil {
		p.e = err
		return
	}
	if !yield('M', p0) {
		return
	}
	t1 := t0
	for i := 1; i <= pf.Points; i++ {
		t1 = pf.NextT(t1)
		p1, err := pf.Func(t1)
		if err != nil {
			p.e = err
			return
		}
		if p.r.Inside(p1) || p.r.Inside(p0) {
			if !p.refine(t0, p0, t1, p1, yield, 10) {
				return
			}
		}
		if !yield('L', p1) {
			return
		}
		t0 = t1
		p0 = p1
	}
}

func (p *pFuncPath) IsClosed() bool {
	return false
}

func (p *pFuncPath) refine(w0 float64, p0 Point, w1 float64, p1 Point, yield func(rune, Point) bool, depth int) bool {
	if p.plot.Dist(p0, p1) > 5 && depth > 0 {
		w := (w0 + w1) / 2
		point, err := p.pf.Func(w)
		if err != nil {
			p.e = err
			return false
		}
		if !p.refine(w0, p0, w, point, yield, depth-1) {
			return false
		}
		if !yield('L', point) {
			return false
		}
		if !p.refine(w, point, w1, p1, yield, depth-1) {
			return false
		}
	}
	return true
}

type ImageInset struct {
	Location Rect
	Image    Image
}

func (s ImageInset) PreferredBounds(_, _ Bounds) (Bounds, Bounds, error) {
	var x, y Bounds
	x.Merge(s.Location.Min.X)
	x.Merge(s.Location.Max.X)
	y.Merge(s.Location.Min.Y)
	y.Merge(s.Location.Max.Y)
	return x, y, nil
}

func (s ImageInset) DrawTo(p *Plot, _ Canvas) error {
	minPos := p.trans(s.Location.Min).Add(Point{1, 1})
	maxPos := p.trans(s.Location.Max).Sub(Point{1, 1})
	inner := ResizeCanvas{
		parent: p.canvas,
		size: Rect{
			Min: minPos,
			Max: maxPos,
		},
	}
	return s.Image.DrawTo(inner)
}

type YConst struct {
	Y     float64
	Style *Style
}

func (yc YConst) PreferredBounds(_, _ Bounds) (x, y Bounds, err error) {
	return Bounds{}, NewBounds(yc.Y, yc.Y), nil
}

func (yc YConst) DrawTo(_ *Plot, canvas Canvas) error {
	r := canvas.Rect()
	if r.Max.Y >= yc.Y && r.Min.Y <= yc.Y {
		canvas.DrawPath(NewPath(false).
			MoveTo(Point{r.Min.X, yc.Y}).
			LineTo(Point{r.Max.X, yc.Y}), yc.Style)
	}
	return nil
}

type XConst struct {
	X     float64
	Style *Style
}

func (xc XConst) PreferredBounds(_, _ Bounds) (Bounds, Bounds, error) {
	return NewBounds(xc.X, xc.X), Bounds{}, nil
}

func (xc XConst) DrawTo(_ *Plot, canvas Canvas) error {
	r := canvas.Rect()
	if r.Max.X >= xc.X && r.Min.X <= xc.X {
		canvas.DrawPath(NewPath(false).
			MoveTo(Point{xc.X, r.Min.Y}).
			LineTo(Point{xc.X, r.Max.Y}), xc.Style)
	}
	return nil
}
