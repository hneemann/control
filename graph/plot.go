package graph

import (
	"bytes"
	"fmt"
	"math"
)

type ShapeLineStyle struct {
	LineStyle  *Style
	Shape      Shape
	ShapeStyle *Style
}

var defShapeLineStyle = ShapeLineStyle{
	LineStyle:  Black.SetDash(7, 7),
	Shape:      NewCircleMarker(4),
	ShapeStyle: Black.SetFill(White),
}

func (ml ShapeLineStyle) IsShape() bool {
	return ml.Shape != nil && ml.ShapeStyle != nil
}

func (ml ShapeLineStyle) IsLine() bool {
	return ml.LineStyle != nil
}

func (ml ShapeLineStyle) EnsureSomethingIsVisible() ShapeLineStyle {
	if ml.IsShape() || ml.IsLine() {
		return ml
	}
	return defShapeLineStyle
}

type Legend struct {
	ShapeLineStyle
	Name string
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
	Frame          *Style
	Title          string
	XLabel         string
	YLabel         string
	YLabelExtend   bool
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
		rb = 1
	}

	innerRect := Rect{
		Min: Point{rect.Min.X + textSize*lb*0.75, rect.Min.Y + textSize*2},
		Max: Point{rect.Max.X - textSize*rb*0.75, rect.Max.Y - textSize/2},
	}

	xBounds := p.XBounds
	yBounds := p.YBounds

	yAutoScale := !yBounds.valid

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

	yExp := 0.02
	xTrans, xTicks, xBounds := xAxis(innerRect.Min.X, innerRect.Max.X, xBounds,
		func(width float64, digits int) bool {
			return width > textSize*(float64(digits+1))*0.5
		}, yExp)

	if p.YLabelExtend && yAutoScale && (p.XLabel != "" || p.YLabel != "") {
		yExp = 1.8 * textSize / innerRect.Height()
	}

	yTrans, yTicks, yBounds := yAxis(innerRect.Min.Y, innerRect.Max.Y, yBounds,
		func(width float64, _ int) bool {
			return width > textSize*2
		}, yExp)

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

	var thinLine *Style
	if p.Frame != nil {
		thinLine = p.Frame
	} else {
		thinLine = Black
	}
	thickLine := thinLine.SetStrokeWidth(thinLine.StrokeWidth * 2)

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

	for _, plotContent := range p.Content {
		err := plotContent.DrawTo(p, inner)
		if err != nil {
			return err
		}
	}

	canvas.DrawText(Point{innerRect.Max.X - small, innerRect.Min.Y + small}, p.XLabel, Bottom|Right, textStyle, textSize)
	canvas.DrawText(Point{innerRect.Min.X + small, innerRect.Max.Y - small}, p.YLabel, Top|Left, textStyle, textSize)
	if p.Title != "" {
		canvas.DrawText(Point{innerRect.Max.X - small, innerRect.Max.Y - small}, p.Title, Top|Right, textStyle, textSize)
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
			sls := leg.EnsureSomethingIsVisible()
			if sls.IsLine() {
				canvas.DrawPath(NewPointsPath(false, lp.Add(Point{-2*textSize - small, 0}), lp.Add(Point{-small, 0})), sls.LineStyle)
			}
			if sls.IsShape() {
				canvas.DrawShape(lp.Add(Point{-1*textSize - small, 0}), sls.Shape, sls.ShapeStyle)
			}
			lp = lp.Add(Point{0, -textSize * 1.5})
		}

	}
	return nil
}

func (p *Plot) GetTransform() Transform {
	return p.trans
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
		Name: name,
		ShapeLineStyle: ShapeLineStyle{
			LineStyle:  lineStyle,
			Shape:      shape,
			ShapeStyle: shapeStyle,
		},
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
	// PreferredBounds returns the preferred bounds for the content.
	// The first bounds is the x-axis, the second is the y-axis.
	// The given bounds are valid if they are set by the user.
	// They are necessary only if the bounds of this instance depend
	// on the bounds given by the user.
	PreferredBounds(xGiven, yGiven Bounds) (x, y Bounds, err error)
	// DrawTo draws the content to the given Canvas.
	// The *Plot is passed to allow the content to access the plot's properties.
	DrawTo(*Plot, Canvas) error
}

type HasLine interface {
	SetLine(*Style) PlotContent
}

type HasShape interface {
	SetShape(Shape, *Style) PlotContent
}

type Function struct {
	Function func(x float64) (float64, error)
	Style    *Style
	Steps    int
}

func (f Function) SetLine(style *Style) PlotContent {
	f.Style = style
	return f
}

const functionSteps = 100

func (f Function) steps() int {
	if f.Steps <= 0 {
		return functionSteps
	}
	return f.Steps
}

func (f Function) String() string {
	return "Function"
}

func (f Function) PreferredBounds(xGiven, _ Bounds) (Bounds, Bounds, error) {
	if xGiven.valid {
		yBounds := Bounds{}
		width := xGiven.Width()
		steps := f.steps()
		for i := 0; i <= steps; i++ {
			x := xGiven.Min + width*float64(i)/float64(steps)
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

func (f Function) DrawTo(plot *Plot, canvas Canvas) error {
	r := canvas.Rect()
	p := NewLinearParameterFunc(r.Min.X, r.Max.X, f.steps())
	p.Func = func(x float64) (Point, error) {
		y, err := f.Function(x)
		return Point{x, y}, err
	}
	path := pFuncPath{
		pf:   p,
		plot: plot,
		r:    r,
	}
	canvas.DrawPath(r.IntersectPath(&path), f.Style)
	return path.e
}

type Scatter struct {
	ShapeLineStyle
	Points Points
}

func (s Scatter) SetShape(shape Shape, style *Style) PlotContent {
	s.Shape = shape
	s.ShapeStyle = style
	return s
}

func (s Scatter) SetLine(style *Style) PlotContent {
	s.LineStyle = style
	return s
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

	sls := s.EnsureSomethingIsVisible()
	if sls.IsLine() {
		canvas.DrawPath(canvas.Rect().IntersectPath(s.Points), sls.LineStyle)
	}
	if sls.IsShape() {
		for p := range s.Points {
			if rect.Contains(p) {
				canvas.DrawShape(p, sls.Shape, sls.ShapeStyle)
			}
		}
	}
	return nil
}

type Hint struct {
	Text  string
	Style *Style
	Pos   Point
}

func (h Hint) PreferredBounds(_, _ Bounds) (Bounds, Bounds, error) {
	var x Bounds
	x.Merge(h.Pos.X)
	var y Bounds
	y.Merge(h.Pos.Y)
	return x, y, nil
}

func (h Hint) DrawTo(plot *Plot, canvas Canvas) error {
	r := canvas.Rect()
	if r.Contains(h.Pos) {
		tPos := h.Pos
		dx := r.Width() / 30
		if r.IsInLeftHalf(h.Pos) {
			tPos = tPos.Add(Point{dx, 0})
		} else {
			tPos = tPos.Add(Point{-dx, 0})
		}
		dy := r.Height() / 30
		if r.IsInTopHalf(h.Pos) {
			tPos = tPos.Add(Point{0, -dy})
		} else {
			tPos = tPos.Add(Point{0, dy})
		}
		drawArrow(plot, plot.trans(tPos), plot.trans(h.Pos), h.Style, 1, h.Text)
	}
	return nil
}

type HintDir struct {
	Hint
	PosDir Point
}

func (h HintDir) DrawTo(plot *Plot, canvas Canvas) error {
	r := canvas.Rect()
	if r.Contains(h.Pos) {
		p1 := plot.trans(h.Pos)
		p2 := plot.trans(h.PosDir)

		delta := p1.Sub(p2).Norm().Rot90().Mul(plot.canvas.Rect().Width() / 30)
		tPos := p1.Add(delta)

		drawArrow(plot, tPos, p1, h.Style, 1, h.Text)
	}
	return nil
}

type Arrow struct {
	From, To Point
	Style    *Style
	Label    string
	Mode     int
}

func (a Arrow) PreferredBounds(_, _ Bounds) (Bounds, Bounds, error) {
	var x, y Bounds
	x.Merge(a.From.X)
	x.Merge(a.To.X)
	y.Merge(a.From.Y)
	y.Merge(a.To.Y)
	return x, y, nil
}

func (a Arrow) DrawTo(plot *Plot, _ Canvas) error {
	from := plot.trans(a.From)
	to := plot.trans(a.To)
	drawArrow(plot, from, to, a.Style, a.Mode, a.Label)
	return nil
}

func drawArrow(plot *Plot, from, to Point, style *Style, mode int, label string) {
	textSize := plot.canvas.Context().TextSize
	w := textSize * 0.75

	dif := to.Sub(from).Norm().Mul(w)
	norm := dif.Rot90().Mul(0.25)

	var textPos Point
	var o Orientation

	if from != to {
		p := NewPath(false)
		p = p.MoveTo(from)
		p = p.LineTo(to)
		if mode&1 != 0 {
			p = p.MoveTo(to.Sub(dif).Add(norm))
			p = p.LineTo(to)
			p = p.LineTo(to.Sub(dif).Sub(norm))
		}
		if mode&2 != 0 {
			p = p.MoveTo(from.Add(dif).Add(norm))
			p = p.LineTo(from)
			p = p.LineTo(from.Add(dif).Sub(norm))
		}
		plot.canvas.DrawPath(p, style)
	}

	if label != "" {
		switch mode & 3 {
		case 1:
			textPos = from
			o = orientationByDelta(dif.Mul(-1))
		case 2:
			textPos = to
			o = orientationByDelta(dif)
		default:
			textPos = from.Add(to).Mul(0.5)
			o = orientationByDelta(norm)
		}
		plot.canvas.DrawText(textPos, label, o, style.Text(), textSize)
	}
}

func orientationByDelta(delta Point) Orientation {
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
	return o
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
	if r.Contains(Point{0, 0}) {
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

func (p *ParameterFunc) SetLine(style *Style) PlotContent {
	p.Style = style
	return p
}

func NewLinearParameterFunc(tMin, tMax float64, steps int) *ParameterFunc {
	if steps <= 0 {
		steps = functionSteps
	}
	delta := (tMax - tMin) / float64(steps)
	return &ParameterFunc{
		Points:   steps,
		InitialT: tMin,
		NextT: func(t float64) float64 {
			return t + delta
		},
	}
}

func NewLogParameterFunc(tMin, tMax float64, steps int) *ParameterFunc {
	if steps <= 0 {
		steps = functionSteps
	}
	f := math.Pow(tMax/tMin, 1/float64(steps))
	return &ParameterFunc{
		Points:   steps,
		InitialT: tMin,
		NextT: func(t float64) float64 {
			return t * f
		},
	}
}

func (p *ParameterFunc) String() string {
	return "Parameter curve"
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
	pf      *ParameterFunc
	plot    *Plot
	r       Rect
	e       error
	maxDist float64
}

// f returns the point at time t and the derivative at that point.
// The derivative is calculated by evaluating the function at t and t+dt/100.
func (p *pFuncPath) f(t, dt float64) (Point, Point, error) {
	p0, err := p.pf.Func(t)
	if err != nil {
		return Point{}, Point{}, err
	}

	dt = dt / 100

	p1, err := p.pf.Func(t + dt)
	if err != nil {
		return Point{}, Point{}, err
	}
	return p0, p1.Sub(p0).Div(dt), nil
}

func (p *pFuncPath) Iter(yield func(rune, Point) bool) {
	if p.maxDist == 0 {
		p.maxDist = p.plot.canvas.Rect().Width() / float64(p.pf.Points) * 2
	}
	pf := p.pf
	t0 := pf.InitialT
	p0, d0, err := p.f(t0, pf.NextT(t0)-t0)
	if err != nil {
		p.e = err
		return
	}
	if !yield('M', p0) {
		return
	}
	for i := 1; i <= pf.Points; i++ {
		t1 := pf.NextT(t0)
		p1, d1, err := p.f(t1, t1-t0)
		if err != nil {
			p.e = err
			return
		}
		if p.r.Contains(p1) || p.r.Contains(p0) {
			if !p.refine(t0, p0, d0, t1, p1, d1, yield, 10) {
				return
			}
		}
		if !yield('L', p1) {
			return
		}
		t0 = t1
		p0 = p1
		d0 = d1
	}
}

func (p *pFuncPath) IsClosed() bool {
	return false
}

func angleBetween(d0, d1 Point) float64 {
	d0n := d0.Norm()
	d1n := d1.Norm()
	if d0n.X == 0 && d0n.Y == 0 || d1n.X == 0 && d1n.Y == 0 {
		return 0
	}
	cos := d0n.X*d1n.X + d0n.Y*d1n.Y
	if cos < -1 {
		cos = -1
	} else if cos > 1 {
		cos = 1
	}
	return math.Acos(cos)
}

func (p *pFuncPath) refine(w0 float64, p0, d0 Point, w1 float64, p1, d1 Point, yield func(rune, Point) bool, depth int) bool {
	dw := w1 - w0
	if (p.plot.Dist(p0, p1) > p.maxDist ||
		p.plot.Dist(p1, p0.Add(d0.Mul(dw))) > p.maxDist/50 ||
		angleBetween(d0, d1) > 10*math.Pi/180) && depth > 0 {
		w := (w0 + w1) / 2
		point, delta, err := p.f(w, dw)
		if err != nil {
			p.e = err
			return false
		}
		if !p.refine(w0, p0, d0, w, point, delta, yield, depth-1) {
			return false
		}
		if !yield('L', point) {
			return false
		}
		if !p.refine(w, point, delta, w1, p1, d1, yield, depth-1) {
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

type Text struct {
	Pos   Point
	Text  string
	Style *Style
}

func (t Text) PreferredBounds(_, _ Bounds) (x, y Bounds, err error) {
	return NewBounds(t.Pos.X, t.Pos.X), NewBounds(t.Pos.Y, t.Pos.Y), nil
}

func (t Text) DrawTo(_ *Plot, canvas Canvas) error {
	canvas.DrawText(t.Pos, t.Text, Left|Bottom, t.Style.Text(), canvas.Context().TextSize)
	return nil
}
