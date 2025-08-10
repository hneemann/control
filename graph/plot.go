package graph

import (
	"bytes"
	"fmt"
	"math"
	"slices"
)

type ShapeLineStyle struct {
	LineStyle  *Style
	Shape      Shape
	ShapeStyle *Style
}

var defShapeLineStyle = ShapeLineStyle{
	Shape:      NewCircleMarker(4),
	ShapeStyle: Black.SetFill(White),
	LineStyle:  Black.SetDash(7, 7),
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
	return func(xBounds, yBounds Bounds, _ *Plot, _ Canvas) (Bounds, Bounds) {
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
	NoBorder       bool
	NoXExpand      bool
	NoYExpand      bool
	Cross          bool
	Grid           *Style
	Frame          *Style
	Title          string
	XLabel         string
	YLabel         string
	ProtectLabels  bool
	Content        []PlotContent
	FillBackground bool
	HideXAxis      bool
	HideYAxis      bool
	BoundsModifier BoundsModifier
	xTicks         Ticks
	yTicks         Ticks
	XTickSep       float64
	YTickSep       float64
	HideLegend     bool
	legendPosGiven bool
	legendPos      Point
	trans          Transform
	canvas         Canvas
	inner          Canvas
	textSize       float64
	cross          bool
}

func (p *Plot) DrawTo(canvas Canvas) error {
	p.canvas = canvas
	c := canvas.Context()
	rect := canvas.Rect()
	textStyle := Black.Text()
	p.textSize = c.TextSize
	if p.textSize <= rect.Height()/200 {
		p.textSize = rect.Height() / 200
	}
	if p.Frame == nil {
		p.Frame = Black.SetStrokeWidth(2)
	}

	if p.FillBackground {
		canvas.DrawPath(rect.Poly(), White.SetStrokeWidth(0).SetFill(White))
	}

	xBounds := p.XBounds
	yBounds := p.YBounds

	yAutoScale := !yBounds.valid

	if !(xBounds.valid && yBounds.valid) {
		mergeX := !xBounds.valid
		mergeY := !yBounds.valid
		for _, plotContent := range p.Content {
			x, y, err := plotContent.Bounds()
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
		var xPrefBounds, yPrefBounds Bounds
		for _, plotContent := range p.Content {
			x, y, err := plotContent.DependantBounds(xBounds, yBounds)
			if err != nil {
				return err
			}
			xPrefBounds.MergeBounds(x)
			yPrefBounds.MergeBounds(y)
		}
		if mergeX {
			xBounds.MergeBounds(xPrefBounds)
		}
		if mergeY {
			yBounds.MergeBounds(yPrefBounds)
		}
	}

	if !xBounds.valid {
		xBounds = NewBounds(-1, 1)
	}
	if !yBounds.valid {
		yBounds = NewBounds(-1, 1)
	}

	if p.BoundsModifier != nil {
		xBounds, yBounds = p.BoundsModifier(xBounds, yBounds, p, canvas)
	}

	p.cross = p.Cross && xBounds.Min < 0 && xBounds.Max > 0 && yBounds.Min < 0 && yBounds.Max > 0

	xAxis := p.XAxis
	if xAxis == nil {
		xAxis = LinearAxis
	}
	yAxis := p.YAxis
	if yAxis == nil {
		yAxis = LinearAxis
	}

	large := p.textSize / 2
	small := p.textSize / 4

	thinLine := p.Frame.SetStrokeWidth(p.Frame.StrokeWidth / 2)

	innerRect := p.calculateInnerRect(rect)

	xTickSep := p.XTickSep
	if xTickSep <= 0 {
		xTickSep = 1
	}
	yTickSep := p.YTickSep
	if yTickSep <= 0 {
		yTickSep = 1
	}

	xExp := 0.02
	if p.NoXExpand {
		xExp = 0
	}
	xTrans, xTicks, xBounds, xUnit := xAxis(innerRect.Min.X, innerRect.Max.X, xBounds,
		func(width float64, digits int) bool {
			return width > p.textSize*(float64(digits)+1+xTickSep)*0.5
		}, xExp)

	yExp := 0.0
	if !p.NoYExpand {
		yExp = 0.02
		if p.ProtectLabels && yAutoScale && (p.XLabel != "" || p.YLabel != "") {
			yExp = 1.8 * p.textSize / innerRect.Height()
		}
	}

	yTrans, yTicks, yBounds, yUnit := yAxis(innerRect.Min.Y, innerRect.Max.Y, yBounds,
		func(width float64, _ int) bool {
			return width > p.textSize*(1+yTickSep)
		}, yExp)

	p.xTicks = xTicks
	p.yTicks = yTicks

	p.trans = func(p Point) Point {
		return Point{xTrans(p.X), yTrans(p.Y)}
	}

	p.inner = TransformCanvas{
		transform: p.trans,
		parent:    canvas,
		size: Rect{
			Min: Point{xBounds.Min, yBounds.Min},
			Max: Point{xBounds.Max, yBounds.Max},
		},
	}

	if !p.HideXAxis {
		yp := innerRect.Min.Y
		if p.cross {
			yp = yTrans(0)
		}
		for _, tick := range xTicks {
			if !p.cross || math.Abs(tick.Position) > 1e-8 {
				xp := xTrans(tick.Position)
				if tick.Label == "" {
					canvas.DrawPath(NewPointsPath(false, Point{xp, yp - small}, Point{xp, yp}), thinLine)
				} else {
					canvas.DrawText(Point{xp, yp - large - small}, tick.Label, Top|HCenter, textStyle, p.textSize)
					canvas.DrawPath(NewPointsPath(false, Point{xp, yp - large}, Point{xp, yp}), p.Frame)
				}
				if p.Grid != nil {
					canvas.DrawPath(NewPointsPath(false, Point{xp, innerRect.Min.Y}, Point{xp, innerRect.Max.Y}), p.Grid)
				}
			}
		}
	}
	if !p.HideYAxis {
		xp := innerRect.Min.X
		if p.cross {
			xp = xTrans(0)
		}
		for _, tick := range yTicks {
			if !p.cross || math.Abs(tick.Position) > 1e-8 {
				yp := yTrans(tick.Position)
				if tick.Label == "" {
					canvas.DrawPath(NewPointsPath(false, Point{xp - small, yp}, Point{xp, yp}), thinLine)
				} else {
					canvas.DrawText(Point{xp - large, yp}, tick.Label, Right|VCenter, textStyle, p.textSize)
					canvas.DrawPath(NewPointsPath(false, Point{xp - large, yp}, Point{xp, yp}), p.Frame)
				}
				if p.Grid != nil {
					canvas.DrawPath(NewPointsPath(false, Point{innerRect.Min.X, yp}, Point{innerRect.Max.X, yp}), p.Grid)
				}
			}
		}
	}

	var legends []Legend
	for _, plotContent := range slices.Backward(p.Content) {
		err := plotContent.DrawTo(p, p.inner)
		if err != nil {
			return err
		}
		l := plotContent.Legend()
		if l.Name != "" && !p.HideLegend {
			legends = append(legends, l)
		}
	}

	if p.XLabel != "" || xUnit != "" {
		yp := innerRect.Min.Y
		if p.cross {
			yp = yTrans(0)
		}
		canvas.DrawText(Point{innerRect.Max.X - small, yp + small}, p.XLabel+" "+xUnit, Bottom|Right, textStyle, p.textSize)
	}
	if p.YLabel != "" || yUnit != "" {
		xp := innerRect.Min.X
		if p.cross {
			xp = xTrans(0)
		}
		canvas.DrawText(Point{xp + small, innerRect.Max.Y - small}, p.YLabel+" "+yUnit, Top|Left, textStyle, p.textSize)
	}
	if p.Title != "" {
		canvas.DrawText(Point{innerRect.Max.X - small, innerRect.Max.Y - small}, p.Title, Top|Right, textStyle, p.textSize)
	}
	if p.cross {
		xp := xTrans(0)
		yp := yTrans(0)
		canvas.DrawPath(NewPointsPath(false, Point{xp, innerRect.Min.Y}, Point{xp, innerRect.Max.Y}), p.Frame)
		canvas.DrawPath(NewPointsPath(false, Point{innerRect.Min.X, yp}, Point{innerRect.Max.X, yp}), p.Frame)
	} else {
		canvas.DrawPath(innerRect.Poly(), p.Frame)
	}

	if len(legends) > 0 {
		var lp Point
		if p.legendPosGiven {
			lp = Point{xTrans(p.legendPos.X), yTrans(p.legendPos.Y)}
		} else {
			lp = Point{innerRect.Min.X + p.textSize*3, innerRect.Min.Y + p.textSize*(float64(len(legends))*1.5-0.5)}
		}
		for _, leg := range slices.Backward(legends) {
			canvas.DrawText(lp, leg.Name, Left|VCenter, textStyle, p.textSize)
			sls := leg.EnsureSomethingIsVisible()
			if sls.IsLine() {
				canvas.DrawPath(NewPointsPath(false, lp.Add(Point{-2*p.textSize - small, 0}), lp.Add(Point{-small, 0})), sls.LineStyle)
			}
			if sls.IsShape() {
				canvas.DrawShape(lp.Add(Point{-1*p.textSize - small, 0}), sls.Shape, sls.ShapeStyle)
			}
			lp = lp.Add(Point{0, -p.textSize * 1.5})
		}

	}
	return nil
}

func (p *Plot) calculateInnerRect(rect Rect) Rect {
	if p.cross {
		return rect
	}

	rMin := rect.Min
	rMax := rect.Max

	lb := p.LeftBorder
	if lb <= 0 {
		if p.HideYAxis {
			if p.NoXExpand {
				lb = 1
			} else {
				lb = 0
			}
		} else {
			lb = 5
		}
	}
	rb := p.RightBorder
	if rb < 0 {
		rb = 0
	}
	if rb == 0 && p.NoXExpand {
		rb = 1
	}

	stroke := p.Frame.StrokeWidth

	if p.NoBorder {
		rMin.X += stroke / 2
		rMax.X -= stroke / 2
	} else {
		if lb == 0 {
			rMin.X += stroke / 2
		} else {
			rMin.X += p.textSize * lb * 0.75
		}
		if rb == 0 {
			rMax.X -= stroke / 2
		} else {
			rMax.X -= p.textSize * rb * 0.75
		}
	}

	// calculate y-borders
	if p.NoBorder {
		rMin.Y += stroke / 2
		rMax.Y -= stroke / 2
	} else {
		if p.HideXAxis {
			if p.NoYExpand {
				rMin.Y += p.textSize / 3 * 2
				rMax.Y -= p.textSize / 3 * 2
			} else {
				rMin.Y += stroke / 2
				rMax.Y -= stroke / 2
			}
		} else {
			rMin.Y += p.textSize * 2
			if p.NoYExpand && !p.HideYAxis {
				rMax.Y -= p.textSize / 3 * 2
			} else {
				rMax.Y -= stroke / 2
			}
		}
	}

	return Rect{Min: rMin, Max: rMax}
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
	return Bounds{valid: true, Min: min, Max: max}
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

// PlotContent is the interface that all plot contents must implement.
// If the plot is created at first, all Bounds methods are called and the
// returned bounds are merged. After that, the DependantBounds method
// is called with the merged bounds.
type PlotContent interface {
	// Bounds returns the fixed bounds of the content.
	// There may be non. For example, if the content is a function
	// that has by definition no bounds because it is defined for all x
	// and the corresponding y=f(x) depends on the x bounds.
	// A set of given data points on the other hand has fixed bounds.
	Bounds() (x, y Bounds, err error)
	// DependantBounds returns the preferred bounds for the content
	// that depend on the bounds given by all other plot contents or the user.
	// A function y=f(x), for example, has certain y bounds if the x bounds
	// are given.
	DependantBounds(xGiven, yGiven Bounds) (x, y Bounds, err error)
	// DrawTo draws the content to the given Canvas.
	// The *Plot is passed to allow the content to access the plot's properties.
	DrawTo(*Plot, Canvas) error
	// Legend returns the legend of this Content
	Legend() Legend
}

type HasLine interface {
	SetLine(*Style) PlotContent
}

type HasShape interface {
	SetShape(Shape, *Style) PlotContent
}

type HasTitle interface {
	SetTitle(title string) PlotContent
}

// Function represents a mathematical function that can be plotted.
type Function struct {
	Function func(x float64) (float64, error)
	Style    *Style
	Title    string
	Steps    int
}

func (f Function) SetLine(style *Style) PlotContent {
	f.Style = style
	return f
}

func (f Function) SetTitle(title string) PlotContent {
	f.Title = title
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
	if f.Title != "" {
		return fmt.Sprintf("Function: %s", f.Title)
	}
	return "Function"
}

func (f Function) Bounds() (Bounds, Bounds, error) {
	return Bounds{}, Bounds{}, nil
}

func (f Function) DependantBounds(xGiven, _ Bounds) (Bounds, Bounds, error) {
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
	p, err := NewLinearParameterFunc(r.Min.X, r.Max.X, f.steps())
	if err != nil {
		return fmt.Errorf("error creating linear parameter function: %w", err)
	}
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

func (f Function) Legend() Legend {
	return Legend{
		Name:           f.Title,
		ShapeLineStyle: ShapeLineStyle{LineStyle: f.Style},
	}
}

// Scatter represents a scatter plot with points represented by a Shape
// and can have a line style for connecting the points.
type Scatter struct {
	ShapeLineStyle
	Points Points
	Title  string
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

func (s Scatter) SetTitle(title string) PlotContent {
	s.Title = title
	return s
}

func (s Scatter) String() string {
	if s.Title != "" {
		return fmt.Sprintf("Scatter: %s", s.Title)
	}
	return "Scatter"
}

func (s Scatter) Bounds() (Bounds, Bounds, error) {
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

func (s Scatter) DependantBounds(_, _ Bounds) (Bounds, Bounds, error) {
	return Bounds{}, Bounds{}, nil
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

func (s Scatter) Legend() Legend {
	return Legend{
		Name:           s.Title,
		ShapeLineStyle: s.EnsureSomethingIsVisible(),
	}
}

// Hint is a simple marker that can be used to indicate a point of interest
type Hint struct {
	Text  string
	Style *Style
	Pos   Point
}

func (h Hint) String() string {
	if h.Text != "" {
		return fmt.Sprintf("Hint: %s at %s", h.Text, h.Pos)
	}
	return fmt.Sprintf("Hint at %s", h.Pos)
}

func (h Hint) Bounds() (Bounds, Bounds, error) {
	return NewBounds(h.Pos.X, h.Pos.X), NewBounds(h.Pos.Y, h.Pos.Y), nil
}

func (h Hint) DependantBounds(_, _ Bounds) (Bounds, Bounds, error) {
	return Bounds{}, Bounds{}, nil
}

func (h Hint) DrawTo(plot *Plot, canvas Canvas) error {
	r := canvas.Rect()
	if r.Contains(h.Pos) {
		r := plot.canvas.Rect()
		hPos := plot.trans(h.Pos)
		tPos := hPos
		dx := r.Width() / 30
		if r.IsInLeftHalf(hPos) {
			tPos = tPos.Add(Point{dx, 0})
		} else {
			tPos = tPos.Add(Point{-dx, 0})
		}
		dy := r.Height() / 30
		if r.IsInTopHalf(hPos) {
			tPos = tPos.Add(Point{0, -dy})
		} else {
			tPos = tPos.Add(Point{0, dy})
		}
		drawArrow(plot, tPos, hPos, h.Style, 1, h.Text)
	}
	return nil
}

func (h Hint) Legend() Legend {
	return Legend{}
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

// Arrow represents a directed line segment with an optional label.
// The Mode field controls the arrow style: The first bit indicates an
// arrowhead at the end, the second bit indicates an arrowhead at the start.
type Arrow struct {
	From, To Point
	Style    *Style
	Label    string
	Mode     int
}

func (a Arrow) String() string {
	if a.Label != "" {
		return fmt.Sprintf("Arrow from %s to %s: %s", a.From, a.To, a.Label)
	}
	return fmt.Sprintf("Arrow from %s to %s", a.From, a.To)
}

func (a Arrow) Bounds() (Bounds, Bounds, error) {
	var x, y Bounds
	x.Merge(a.From.X)
	x.Merge(a.To.X)
	y.Merge(a.From.Y)
	y.Merge(a.To.Y)
	return x, y, nil
}

func (a Arrow) DependantBounds(_, _ Bounds) (Bounds, Bounds, error) {
	return Bounds{}, Bounds{}, nil
}

func (a Arrow) DrawTo(plot *Plot, _ Canvas) error {
	from := plot.trans(a.From)
	to := plot.trans(a.To)
	drawArrow(plot, from, to, a.Style, a.Mode, a.Label)
	return nil
}

func (a Arrow) Legend() Legend {
	return Legend{}
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

func NewDiamondMarker(r float64) Shape {
	r = r * math.Sqrt2
	return pointsPath{
		points: []Point{{-r, 0}, {0, r}, {r, 0}, {0, -r}},
		closed: true,
	}
}

type Cross struct {
	Style *Style
}

func (c Cross) SetLine(style *Style) PlotContent {
	return Cross{
		Style: style,
	}
}

func (c Cross) String() string {
	return "coordinate cross"
}

func (c Cross) Bounds() (Bounds, Bounds, error) {
	return Bounds{}, Bounds{}, nil
}

func (c Cross) DependantBounds(_, _ Bounds) (Bounds, Bounds, error) {
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

func (c Cross) Legend() Legend {
	return Legend{}
}

type ParameterFunc struct {
	Func     func(t float64) (Point, error)
	Points   int
	InitialT float64
	NextT    func(float64) float64
	Style    *Style
	Title    string
}

func (p *ParameterFunc) SetTitle(title string) PlotContent {
	p.Title = title
	return p
}

func (p *ParameterFunc) SetLine(style *Style) PlotContent {
	p.Style = style
	return p
}

func NewLinearParameterFunc(tMin, tMax float64, steps int) (*ParameterFunc, error) {
	if steps <= 0 {
		steps = functionSteps
	}
	if tMax <= tMin {
		return nil, fmt.Errorf("invalid parameters for lin parameter function: tMin=%f, tMax=%f", tMin, tMax)
	}
	delta := (tMax - tMin) / float64(steps)
	return &ParameterFunc{
		Points:   steps,
		InitialT: tMin,
		NextT: func(t float64) float64 {
			return t + delta
		},
	}, nil
}

func NewLogParameterFunc(tMin, tMax float64, steps int) (*ParameterFunc, error) {
	if steps <= 0 {
		steps = functionSteps
	}
	if tMin <= 0 || tMax <= tMin {
		return nil, fmt.Errorf("invalid parameters for log parameter function: tMin=%f, tMax=%f", tMin, tMax)
	}
	f := math.Pow(tMax/tMin, 1/float64(steps))
	return &ParameterFunc{
		Points:   steps,
		InitialT: tMin,
		NextT: func(t float64) float64 {
			return t * f
		},
	}, nil
}

func (p *ParameterFunc) String() string {
	return "Parameter curve"
}

func (p *ParameterFunc) Bounds() (Bounds, Bounds, error) {
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

func (p *ParameterFunc) DependantBounds(_, _ Bounds) (Bounds, Bounds, error) {
	return Bounds{}, Bounds{}, nil
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

func (p *ParameterFunc) Legend() Legend {
	return Legend{
		Name:           p.Title,
		ShapeLineStyle: ShapeLineStyle{LineStyle: p.Style},
	}
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
		if !p.refine(t0, p0, d0, t1, p1, d1, yield, 10, 0) {
			return
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

func cosAngleBetween(d0, d1 Point) float64 {
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
	return cos
}

func (p *pFuncPath) refine(w0 float64, p0, d0 Point, w1 float64, p1, d1 Point, yield func(rune, Point) bool, depth int, lastDist float64) bool {
	dw := w1 - w0
	dist := p.plot.Dist(p0, p1)
	if dist > p.maxDist || // the distance of two points is too large
		p.plot.Dist(p1, p0.Add(d0.Mul(dw))) > p.maxDist/50 || // the distance to the tangent is too large
		cosAngleBetween(d0, d1) < 0.98 { // the angle is larger than approx 10 degrees
		if depth > 0 {
			w := (w0 + w1) / 2
			point, delta, err := p.f(w, dw)
			if err != nil {
				p.e = err
				return false
			}
			if !p.refine(w0, p0, d0, w, point, delta, yield, depth-1, dist) {
				return false
			}
			if !yield('L', point) {
				return false
			}
			if !p.refine(w, point, delta, w1, p1, d1, yield, depth-1, dist) {
				return false
			}
		} else {
			// detecting poles
			if dist > lastDist*1.001 && dist > p.maxDist {
				// if a pole is detected, do not draw a line
				if !yield('M', p1) {
					return false
				}
			}
		}
	}
	return true
}

type ImageInset struct {
	Location    Rect
	Image       Image
	VisualGuide *Style
}

func (s ImageInset) Bounds() (Bounds, Bounds, error) {
	var x, y Bounds
	x.Merge(s.Location.Min.X)
	x.Merge(s.Location.Max.X)
	y.Merge(s.Location.Min.Y)
	y.Merge(s.Location.Max.Y)
	return x, y, nil
}

func (s ImageInset) DependantBounds(_, _ Bounds) (Bounds, Bounds, error) {
	return Bounds{}, Bounds{}, nil
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
	err := s.Image.DrawTo(inner)
	if err != nil {
		return fmt.Errorf("error drawing image inset: %w", err)
	}

	if s.VisualGuide != nil {
		if insetPlot, ok := s.Image.(*Plot); ok {
			ir := insetPlot.inner.Rect()

			sMin := p.trans(ir.Min)
			sMax := p.trans(ir.Max)

			p.canvas.DrawPath(SlicePath{Closed: true}.
				Add(sMin).
				Add(Point{X: sMax.X, Y: sMin.Y}).
				Add(sMax).
				Add(Point{X: sMin.X, Y: sMax.Y}), s.VisualGuide)

			lMin := insetPlot.trans(ir.Min)
			lMax := insetPlot.trans(ir.Max)
			s12 := Point{X: sMin.X, Y: sMax.Y}
			l12 := Point{X: lMin.X, Y: lMax.Y}
			s21 := Point{X: sMax.X, Y: sMin.Y}
			l21 := Point{X: lMax.X, Y: lMin.Y}

			if (lMin.X < sMin.X && lMin.Y > sMin.Y) || (lMin.X > sMin.X && lMin.Y < sMin.Y) {
				p.canvas.DrawPath(NewPath(false).Add(sMin).Add(lMin), s.VisualGuide)
			}
			if (l12.X > s12.X && l12.Y > s12.Y) || (l12.X < s12.X && l12.Y < s12.Y) {
				p.canvas.DrawPath(NewPath(false).Add(s12).Add(l12), s.VisualGuide)
			}
			if (lMax.X < sMax.X && lMax.Y > sMax.Y) || (lMax.X > sMax.X && lMax.Y < sMax.Y) {
				p.canvas.DrawPath(NewPath(false).Add(sMax).Add(lMax), s.VisualGuide)
			}
			if (l21.X > s21.X && l21.Y > s21.Y) || (l21.X < s21.X && l21.Y < s21.Y) {
				p.canvas.DrawPath(NewPath(false).Add(s21).Add(l21), s.VisualGuide)
			}
		}
	}
	return nil
}

func (s ImageInset) Legend() Legend {
	return Legend{}
}

type YConst struct {
	Y     float64
	Style *Style
}

func (yc YConst) SetLine(style *Style) PlotContent {
	return YConst{
		Y:     yc.Y,
		Style: style,
	}
}

func (yc YConst) String() string {
	return fmt.Sprintf("yConst: %0.2f", yc.Y)
}

func (yc YConst) Bounds() (x, y Bounds, err error) {
	return Bounds{}, NewBounds(yc.Y, yc.Y), nil
}

func (yc YConst) DependantBounds(_, _ Bounds) (x, y Bounds, err error) {
	return Bounds{}, Bounds{}, nil
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

func (yc YConst) Legend() Legend {
	return Legend{}
}

type XConst struct {
	X     float64
	Style *Style
}

func (xc XConst) SetLine(style *Style) PlotContent {
	return XConst{
		X:     xc.X,
		Style: style,
	}
}

func (xc XConst) String() string {
	return fmt.Sprintf("xConst: %0.2f", xc.X)
}

func (xc XConst) Bounds() (Bounds, Bounds, error) {
	return NewBounds(xc.X, xc.X), Bounds{}, nil
}

func (xc XConst) DependantBounds(_, _ Bounds) (Bounds, Bounds, error) {
	return Bounds{}, Bounds{}, nil
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

func (xc XConst) Legend() Legend {
	return Legend{}
}

type Text struct {
	Pos   Point
	Text  string
	Style *Style
}

func (t Text) Bounds() (x, y Bounds, err error) {
	return NewBounds(t.Pos.X, t.Pos.X), NewBounds(t.Pos.Y, t.Pos.Y), nil
}

func (t Text) DependantBounds(_, _ Bounds) (x, y Bounds, err error) {
	return Bounds{}, Bounds{}, nil
}

func (t Text) DrawTo(_ *Plot, canvas Canvas) error {
	canvas.DrawText(t.Pos, t.Text, Left|Bottom, t.Style.Text(), canvas.Context().TextSize)
	return nil
}

func (t Text) Legend() Legend {
	return Legend{}
}
