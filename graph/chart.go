package graph

import (
	"bytes"
	"fmt"
	"github.com/hneemann/control/nErr"
	img "image"
	col "image/color"
	"math"
	"runtime"
	"slices"
	"sync"
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

func (ml ShapeLineStyle) GetStyle() *Style {
	if ml.IsLine() {
		return ml.LineStyle
	}
	if ml.IsShape() {
		return ml.ShapeStyle
	}
	return Black
}

type Legend struct {
	ShapeLineStyle
	Name      string
	ColorOnly bool
}

type BoundsModifier func(xBounds, yBounds Bounds, p *Chart, canvas Canvas) (Bounds, Bounds)

func Zoom(p Point, f float64) BoundsModifier {
	return func(xBounds, yBounds Bounds, _ *Chart, _ Canvas) (Bounds, Bounds) {
		if xBounds.isSet {
			d := xBounds.Width() / (2 * f)
			xBounds.Min = p.X - d
			xBounds.Max = p.X + d
		}
		if yBounds.isSet {
			d := yBounds.Width() / (2 * f)
			yBounds.Min = p.Y - d
			yBounds.Max = p.Y + d
		}
		return xBounds, yBounds
	}
}

type AxisDescription struct {
	Factory     AxisFactory
	TickSep     float64
	Bounds      Bounds
	NoExpand    bool
	Label       string
	HideAxis    bool
	Grid        *Style
	Style       *Style
	CustomTicks Ticks
}

// GetAxis returns the Axis
func (a *AxisDescription) GetAxis(minParent, maxParent float64, bounds Bounds, ctw CheckTextWidth, expand float64) Axis {
	fac := a.Factory
	if fac == nil {
		fac = LinearAxis
	}
	axis := fac(minParent, maxParent, bounds, ctw, expand)
	if len(a.CustomTicks) > 0 {
		axis.Ticks = a.CustomTicks
	}
	return axis
}

type axisStyles struct {
	text      *Style
	ticks     *Style
	thinTicks *Style
}

func (a *AxisDescription) getTextStyle(style, frame *Style) axisStyles {
	s := a.Style
	if s == nil {
		s = style
	}
	return axisStyles{
		text:      s.Text(),
		ticks:     s.SetStrokeWidth(frame.StrokeWidth),
		thinTicks: s.SetStrokeWidth(frame.StrokeWidth / 2),
	}
}

type PosMode int

const (
	NoPos PosMode = iota
	AbsPos
	RelPos
)

type RelativePos struct {
	Pos  Point
	Mode PosMode
}

func NewRelativePos(pos Point, rel bool) RelativePos {
	var rp RelativePos
	rp.Set(pos, rel)
	return rp
}

func (lp *RelativePos) Set(pos Point, rel bool) {
	if rel {
		lp.Mode = RelPos
	} else {
		lp.Mode = AbsPos
	}
	lp.Pos = pos
}

func (lp *RelativePos) Get(trans Transform, rect Rect) Point {
	if p, ok := lp.GetNoDef(trans, rect); ok {
		return p
	}
	return trans(lp.Pos)
}

func (lp *RelativePos) GetNoDef(trans Transform, rect Rect) (Point, bool) {
	switch lp.Mode {
	case AbsPos:
		return trans(lp.Pos), true
	case RelPos:
		return Point{
			rect.Min.X + lp.Pos.X*(rect.Max.X-rect.Min.X)/100,
			rect.Min.Y + lp.Pos.Y*(rect.Max.Y-rect.Min.Y)/100,
		}, true
	default:
		return Point{}, false
	}
}

type Chart struct {
	X, Y, Y2       AxisDescription
	Square         bool
	SquareYCenter  float64
	LeftBorder     float64
	RightBorder    float64
	NoBorder       bool
	Cross          bool
	NoFrame        bool
	Frame          *Style
	Title          string
	ProtectLabels  bool
	StackBothYAxes bool
	Background     *Style
	BoundsModifier BoundsModifier
	HideLegend     bool
	LegendPos      RelativePos
	LegendPosY2    RelativePos

	content chartContent
}

type chartContent []contentHolder

func (pc chartContent) hasY2Content() bool {
	for _, holder := range pc {
		if holder.toY2 {
			return true
		}
	}
	return false
}

func (pc chartContent) getByYType(toY2 bool) chartContent {
	var c chartContent
	for _, holder := range pc {
		if holder.toY2 == toY2 {
			c = append(c, contentHolder{
				content: holder.content,
				toY2:    false,
			})
		}
	}
	return c
}

type contentHolder struct {
	content ChartContent
	toY2    bool
}

func (h contentHolder) String() string {
	return fmt.Sprint(h.content)
}

type ChartContentEnvironment struct {
	Canvas       Canvas
	ParentCanvas Canvas
	Chart        *Chart
	InnerRect    Rect
	Transform    Transform
	XAxis        *Axis
	YAxis        *Axis
}

func (p *ChartContentEnvironment) Dist(p1, p2 Point) float64 {
	return p.Transform(p1).DistTo(p.Transform(p2))
}

const (
	arrowLenFactor   = 0.5
	arrowWidthFactor = arrowLenFactor / 4
)

func (p *ChartContentEnvironment) getArrowHeadSize() (len, width float64) {
	textSize := p.ParentCanvas.Context().TextSize
	return textSize * arrowLenFactor, textSize * arrowWidthFactor
}

func (p *Chart) DrawTo(canvas Canvas) error {
	err, _ := p.DrawToAsInset(canvas)
	return err
}

func (p *Chart) DrawToAsInset(canvas Canvas) (err error, env *ChartContentEnvironment) {
	defer nErr.CatchErr(&err)
	if p.StackBothYAxes && p.content.hasY2Content() {
		upper := *p
		upper.content = p.content.getByYType(false)

		lower := *p
		lower.content = p.content.getByYType(true)
		lower.Title = ""
		lower.Y = p.Y2
		lower.LegendPos = p.LegendPosY2

		err = SplitHorizontal{&upper, &lower}.DrawTo(canvas)
		return err, nil
	}
	return p.drawToInternal(canvas)
}

func (p *Chart) drawToInternal(canvas Canvas) (error, *ChartContentEnvironment) {
	c := canvas.Context()
	rect := canvas.Rect()
	textStyle := Black.Text()
	if p.Frame == nil {
		p.Frame = Black.SetStrokeWidth(2)
	}
	xStyles := p.X.getTextStyle(textStyle, p.Frame)
	yStyles := p.Y.getTextStyle(textStyle, p.Frame)
	y2Styles := p.Y2.getTextStyle(textStyle, p.Frame)

	hasY2Content := p.content.hasY2Content()

	textSize := c.TextSize
	if textSize <= rect.Height()/200 {
		textSize = rect.Height() / 200
	}

	if p.Background != nil && p.Background.Color.A > 0 {
		nErr.Try(canvas.DrawPath(rect.Path(), p.Background.SetStrokeWidth(0).SetFill(p.Background)))
	}

	xBoundsPre, yBoundsPre, y2BoundsPre, err := p.calcBounds()
	if err != nil {
		return fmt.Errorf("error calculating chart bounds: %w", err), nil
	}

	if p.BoundsModifier != nil {
		xBoundsPre, yBoundsPre = p.BoundsModifier(xBoundsPre, yBoundsPre, p, canvas)
	}

	cross := p.Cross && xBoundsPre.Min <= 0 && xBoundsPre.Max >= 0 && yBoundsPre.Min <= 0 && yBoundsPre.Max >= 0 && !hasY2Content

	large := textSize / 2
	small := textSize / 4

	nonCrossInner := p.calculateInnerRect(rect, textSize, hasY2Content)
	innerRect := rect
	if !cross {
		innerRect = nonCrossInner
	}

	xExp := 0.02
	if cross && !p.X.HideAxis {
		// space for the arrow head
		xExp = 1.1 * textSize / innerRect.Width()
	}
	if p.X.NoExpand {
		xExp = 0
	}
	xAxis := p.X.GetAxis(innerRect.Min.X, innerRect.Max.X, xBoundsPre,
		func(width float64, digits int) bool {
			return width > math.Max(1, textSize*(float64(digits)+2+p.X.TickSep)*0.5)
		}, xExp)

	yExp := 0.0
	if !p.Y.NoExpand {
		yAutoScale := !p.Y.Bounds.isSet
		if cross {
			// space for the arrow head
			yExp = 1.1 * textSize / innerRect.Height()
		} else {
			yExp = 0.02
		}
		if p.ProtectLabels && yAutoScale && !cross && (p.X.Label != "" || p.Y.Label != "" || p.Y2.Label != "" || p.Title != "") {
			yExp = 1.8 * textSize / innerRect.Height()
		}
	}
	y2Exp := 0.0
	if !p.Y2.NoExpand {
		yAutoScale := !p.Y2.Bounds.isSet
		y2Exp = 0.02
		if p.ProtectLabels && yAutoScale && !cross && (p.X.Label != "" || p.Y.Label != "" || p.Y2.Label != "" || p.Title != "") {
			y2Exp = 1.8 * textSize / innerRect.Height()
		}
	}

	if p.Square && xAxis.IsLinear {
		yBoundsWidth := innerRect.Height() * xAxis.Bounds.Width() / innerRect.Width()
		dif := yBoundsWidth / 2
		yBoundsPre.Min = p.SquareYCenter - dif
		yBoundsPre.Max = p.SquareYCenter + dif
		yExp = 0
	}

	yTickMax := math.Max(1, textSize*(2+p.Y.TickSep))
	yAxis := p.Y.GetAxis(innerRect.Min.Y, innerRect.Max.Y, yBoundsPre,
		func(width float64, _ int) bool { return width > yTickMax }, yExp)

	y2TickMax := math.Max(1, textSize*(2+p.Y2.TickSep))
	y2Axis := p.Y2.GetAxis(innerRect.Min.Y, innerRect.Max.Y, y2BoundsPre,
		func(width float64, _ int) bool { return width > y2TickMax }, y2Exp)

	if p.Square && (!xAxis.IsLinear || !yAxis.IsLinear) {
		return fmt.Errorf("square charts are only possible if both axis are linear"), nil
	}

	if !p.X.HideAxis {
		topBottom := 1.0
		orient := Top | HCenter
		yp := innerRect.Min.Y
		if cross {
			yp = yAxis.Trans(0)
			if yp < nonCrossInner.Min.Y {
				topBottom = -1
				orient = Bottom | HCenter
			}
		}
		for _, tick := range xAxis.Ticks {
			if !cross || math.Abs(tick.Position) > 1e-8 {
				xp := xAxis.Trans(tick.Position)
				if p.X.Grid != nil {
					nErr.Try(canvas.DrawPath(PointsFromSlice(Point{xp, innerRect.Min.Y}, Point{xp, innerRect.Max.Y}), p.X.Grid))
				}
				if tick.Label == "" {
					nErr.Try(canvas.DrawPath(PointsFromSlice(Point{xp, yp - small*topBottom}, Point{xp, yp}), xStyles.thinTicks))
				} else {
					canvas.DrawText(Point{xp, yp - (large+small)*topBottom}, tick.Label, orient, xStyles.text, textSize)
					nErr.Try(canvas.DrawPath(PointsFromSlice(Point{xp, yp - large*topBottom}, Point{xp, yp}), xStyles.ticks))
				}
			}
		}
	}

	if !p.Y.HideAxis {
		rightLeft := 1.0
		orient := Right | VCenter
		xp := innerRect.Min.X
		if cross {
			xp = xAxis.Trans(0)
			if xp < nonCrossInner.Min.X {
				rightLeft = -1
				orient = Left | VCenter
			}
		}
		for _, tick := range yAxis.Ticks {
			if !cross || math.Abs(tick.Position) > 1e-8 {
				yp := yAxis.Trans(tick.Position)
				if p.Y.Grid != nil {
					nErr.Try(canvas.DrawPath(PointsFromSlice(Point{innerRect.Min.X, yp}, Point{innerRect.Max.X, yp}), p.Y.Grid))
				}
				if tick.Label == "" {
					nErr.Try(canvas.DrawPath(PointsFromSlice(Point{xp - small*rightLeft, yp}, Point{xp, yp}), yStyles.thinTicks))
				} else {
					canvas.DrawText(Point{xp - large*rightLeft, yp}, tick.Label, orient, yStyles.text, textSize)
					nErr.Try(canvas.DrawPath(PointsFromSlice(Point{xp - large*rightLeft, yp}, Point{xp, yp}), yStyles.ticks))
				}
			}
		}
	}

	if !p.Y2.HideAxis && hasY2Content {
		orient := Left | VCenter
		xp := innerRect.Max.X
		for _, tick := range y2Axis.Ticks {
			yp := y2Axis.Trans(tick.Position)
			if p.Y2.Grid != nil {
				nErr.Try(canvas.DrawPath(PointsFromSlice(Point{innerRect.Min.X, yp}, Point{innerRect.Max.X, yp}), p.Y2.Grid))
			}
			if tick.Label == "" {
				nErr.Try(canvas.DrawPath(PointsFromSlice(Point{xp + small, yp}, Point{xp, yp}), y2Styles.thinTicks))
			} else {
				canvas.DrawText(Point{xp + large, yp}, tick.Label, orient, y2Styles.text, textSize)
				nErr.Try(canvas.DrawPath(PointsFromSlice(Point{xp + large, yp}, Point{xp, yp}), y2Styles.ticks))
			}
		}
	}

	transY := func(point Point) Point {
		return Point{xAxis.Trans(point.X), yAxis.Trans(point.Y)}
	}
	envY := &ChartContentEnvironment{
		Chart:        p,
		ParentCanvas: canvas,
		InnerRect:    innerRect,
		Transform:    transY,
		Canvas: TransformCanvas{
			transform: transY,
			parent:    canvas,
			size: Rect{
				Min: Point{xAxis.Bounds.Min, yAxis.Bounds.Min},
				Max: Point{xAxis.Bounds.Max, yAxis.Bounds.Max},
			},
		},
		XAxis: &xAxis,
		YAxis: &yAxis,
	}
	var envY2 *ChartContentEnvironment
	var transY2 Transform
	if hasY2Content {
		transY2 = func(point Point) Point {
			return Point{xAxis.Trans(point.X), y2Axis.Trans(point.Y)}
		}
		envY2 = &ChartContentEnvironment{
			Chart:        p,
			ParentCanvas: canvas,
			InnerRect:    innerRect,
			Transform:    transY2,
			Canvas: TransformCanvas{
				transform: transY2,
				parent:    canvas,
				size: Rect{
					Min: Point{xAxis.Bounds.Min, y2Axis.Bounds.Min},
					Max: Point{xAxis.Bounds.Max, y2Axis.Bounds.Max},
				},
			},
			XAxis: &xAxis,
			YAxis: &y2Axis,
		}
	}

	if cross {
		xp := xAxis.Trans(0)
		yp := yAxis.Trans(0)
		cs := p.Frame.StrokeWidth / 2
		headLen, headWidth := envY.getArrowHeadSize()
		nErr.Try(canvas.DrawPath(PointsFromSlice(
			Point{xp, innerRect.Min.Y},
			Point{xp, innerRect.Max.Y - cs}), p.Frame))
		nErr.Try(canvas.DrawPath(PointsFromSlice(
			Point{xp - headWidth, innerRect.Max.Y - headLen},
			Point{xp, innerRect.Max.Y - cs},
			Point{xp + headWidth, innerRect.Max.Y - headLen},
		), p.Frame))
		nErr.Try(canvas.DrawPath(PointsFromSlice(
			Point{innerRect.Min.X, yp},
			Point{innerRect.Max.X - cs, yp}), p.Frame))
		nErr.Try(canvas.DrawPath(PointsFromSlice(
			Point{innerRect.Max.X - headLen, yp + headWidth},
			Point{innerRect.Max.X - cs, yp},
			Point{innerRect.Max.X - headLen, yp - headWidth},
		), p.Frame))
	}

	var legends []Legend
	var legendsY2 []Legend
	for _, holder := range slices.Backward(p.content) {
		if holder.toY2 {
			nErr.Try(holder.content.DrawTo(envY2))
		} else {
			nErr.Try(holder.content.DrawTo(envY))
		}
		l := holder.content.Legend()
		if len(l) > 0 && !p.HideLegend {
			if holder.toY2 && p.LegendPosY2.Mode != NoPos {
				legendsY2 = append(legendsY2, l...)
			} else {
				legends = append(legends, l...)
			}
		}
	}

	if lab := xAxis.LabelFormat(p.X.Label); lab != "" {
		yp := innerRect.Min.Y
		if cross {
			yp = yAxis.Trans(0)
		}
		canvas.DrawText(Point{innerRect.Max.X - small, yp + small}, lab, Bottom|Right, xStyles.text, textSize)
	}
	if lab := yAxis.LabelFormat(p.Y.Label); lab != "" {
		xp := innerRect.Min.X
		if cross {
			xp = xAxis.Trans(0)
		}
		canvas.DrawText(Point{xp + small, innerRect.Max.Y - small}, lab, Top|Left, yStyles.text, textSize)
	}
	if !p.Y2.HideAxis && hasY2Content {
		if lab := y2Axis.LabelFormat(p.Y2.Label); lab != "" {
			xp := innerRect.Max.X
			canvas.DrawText(Point{xp - small, innerRect.Max.Y - small}, lab, Top|Right, y2Styles.text, textSize)
		}
	}
	if p.Title != "" {
		if p.Y2.HideAxis || !hasY2Content {
			canvas.DrawText(Point{innerRect.Max.X - small, innerRect.Max.Y - small}, p.Title, Top|Right, textStyle, textSize)
		} else {
			canvas.DrawText(Point{(innerRect.Max.X + innerRect.Min.X) / 2, innerRect.Max.Y - small}, p.Title, Top|HCenter, textStyle, textSize)
		}
	}

	if !cross && !p.NoFrame {
		nErr.Try(canvas.DrawPath(innerRect.Path(), p.Frame))
	}

	// user wants a cross but origin is not visible, so cross could not be plotted
	if p.Cross && !cross {
		if xAxis.Bounds.Min < 0 && xAxis.Bounds.Max > 0 {
			xp := xAxis.Trans(0)
			nErr.Try(canvas.DrawPath(PointsFromSlice(
				Point{xp, innerRect.Min.Y},
				Point{xp, innerRect.Max.Y}), p.Frame))
		}
		if yAxis.Bounds.Min < 0 && yAxis.Bounds.Max > 0 {
			yp := yAxis.Trans(0)
			nErr.Try(canvas.DrawPath(PointsFromSlice(
				Point{innerRect.Min.X, yp},
				Point{innerRect.Max.X, yp}), p.Frame))
		}
	}

	if len(legends) > 0 || len(legendsY2) > 0 {
		var lp Point
		if rp, ok := p.LegendPos.GetNoDef(transY, innerRect); ok {
			lp = rp
		} else {
			lp = Point{innerRect.Min.X + textSize*3, innerRect.Min.Y + textSize*(float64(len(legends))*1.5-0.5)}
		}
		for _, leg := range slices.Backward(legends) {
			lp = p.drawLegend(canvas, lp, leg, textStyle, textSize)
		}
		if p.LegendPosY2.Mode != NoPos {
			lp = p.LegendPosY2.Get(transY2, innerRect)
		}
		for _, leg := range slices.Backward(legendsY2) {
			lp = p.drawLegend(canvas, lp, leg, textStyle, textSize)
		}
	}
	return nil, envY
}

func (p *Chart) drawLegend(canvas Canvas, lp Point, leg Legend, textStyle *Style, textSize float64) Point {
	small := textSize / 4
	canvas.DrawText(lp, leg.Name, Left|VCenter, textStyle, textSize)
	if leg.ColorOnly {
		nErr.Try(canvas.DrawPath(NewPath(true).
			MoveTo(lp.Add(Point{-2*textSize - small, textSize / 2})).
			LineTo(lp.Add(Point{-small, textSize / 2})).
			LineTo(lp.Add(Point{-small, -textSize / 2})).
			LineTo(lp.Add(Point{-2*textSize - small, -textSize / 2})), leg.LineStyle))
	} else {
		sls := leg.EnsureSomethingIsVisible()
		if sls.IsLine() {
			nErr.Try(canvas.DrawPath(PointsFromSlice(lp.Add(Point{-2*textSize - small, 0}), lp.Add(Point{-small, 0})), sls.LineStyle))
		}
		if sls.IsShape() {
			nErr.Try(canvas.DrawShape(lp.Add(Point{-1*textSize - small, 0}), sls.Shape, sls.ShapeStyle))
		}
	}
	return lp.Add(Point{0, -textSize * 1.5})
}

func (p *Chart) calcBounds() (Bounds, Bounds, Bounds, error) {
	xBounds := p.X.Bounds
	yBounds := p.Y.Bounds
	y2Bounds := p.Y2.Bounds

	if !(xBounds.isSet && yBounds.isSet && y2Bounds.isSet) {
		mergeX := !xBounds.isSet
		mergeY := !yBounds.isSet
		mergeY2 := !y2Bounds.isSet
		for _, holder := range p.content {
			x, y, err := holder.content.Bounds()
			if err != nil {
				return Bounds{}, Bounds{}, Bounds{}, err
			}
			if mergeX {
				xBounds.MergeBounds(x)
			}
			if mergeY && !holder.toY2 {
				yBounds.MergeBounds(y)
			}
			if mergeY2 && holder.toY2 {
				y2Bounds.MergeBounds(y)
			}
		}
		var xPrefBounds, yPrefBounds, y2PrefBounds Bounds
		for _, holder := range p.content {
			if holder.toY2 {
				x, y := nErr.TryArgs(holder.content.DependantBounds(xBounds, y2Bounds))
				xPrefBounds.MergeBounds(x)
				y2PrefBounds.MergeBounds(y)
			} else {
				x, y := nErr.TryArgs(holder.content.DependantBounds(xBounds, yBounds))
				xPrefBounds.MergeBounds(x)
				yPrefBounds.MergeBounds(y)
			}
		}
		if mergeX {
			xBounds.MergeBounds(xPrefBounds)
		}
		if mergeY {
			yBounds.MergeBounds(yPrefBounds)
		}
		if mergeY2 {
			y2Bounds.MergeBounds(y2PrefBounds)
		}
	}

	return xBounds, yBounds, y2Bounds, nil
}

func (p *Chart) calculateInnerRect(rect Rect, textSize float64, hasY2Content bool) Rect {
	rMin := rect.Min
	rMax := rect.Max

	lb := p.LeftBorder
	if lb <= 0 {
		if p.Y.HideAxis {
			if p.X.NoExpand {
				lb = 1
			} else {
				lb = 0
			}
		} else {
			lb = 5
		}
	}
	rb := p.RightBorder
	if rb <= 0 {
		if p.Y2.HideAxis || !hasY2Content {
			if p.X.NoExpand {
				rb = 1
			} else {
				rb = 0
			}
		} else {
			rb = 5
		}
	}

	stroke := p.Frame.StrokeWidth

	if p.NoBorder {
		rMin.X += stroke / 2
		rMax.X -= stroke / 2
	} else {
		if lb == 0 {
			rMin.X += stroke / 2
		} else {
			rMin.X += textSize * lb * 0.75
		}
		if rb == 0 {
			rMax.X -= stroke / 2
		} else {
			rMax.X -= textSize * rb * 0.75
		}
	}

	// calculate y-borders
	if p.NoBorder {
		rMin.Y += stroke / 2
		rMax.Y -= stroke / 2
	} else {
		if p.X.HideAxis {
			if p.Y.NoExpand {
				rMin.Y += textSize / 3 * 2
				rMax.Y -= textSize / 3 * 2
			} else {
				rMin.Y += stroke / 2
				rMax.Y -= stroke / 2
			}
		} else {
			rMin.Y += textSize * 2
			if p.Y.NoExpand && !p.Y.HideAxis {
				rMax.Y -= textSize / 3 * 2
			} else {
				rMax.Y -= stroke / 2
			}
		}
	}

	return Rect{Min: rMin, Max: rMax}
}

func (p *Chart) AddContent(content ChartContent, toY2 bool) {
	p.content = append(p.content, contentHolder{content, toY2})
}

func (p *Chart) AddContentAtTop(content ChartContent, toY2 bool) {
	p.content = append(chartContent{contentHolder{content, toY2}}, p.content...)
}

func (p *Chart) String() string {
	bu := bytes.Buffer{}
	bu.WriteString("Chart: ")
	for i, content := range p.content {
		if i > 0 {
			bu.WriteString(", ")
		}
		bu.WriteString(fmt.Sprint(content))
	}
	return bu.String()
}

type Bounds struct {
	isSet    bool
	Min, Max float64
}

func NewBounds(min, max float64) Bounds {
	if min > max {
		min, max = max, min
	}
	return Bounds{isSet: true, Min: min, Max: max}
}

func (b *Bounds) Width() float64 {
	return b.Max - b.Min
}

func (b *Bounds) MergeBounds(other Bounds) {
	if other.isSet {
		// other is available
		if !b.isSet {
			b.isSet = true
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
		if !b.isSet {
			b.isSet = true
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

func (b *Bounds) Limit(x float64) float64 {
	if b.isSet {
		if x < b.Min {
			return b.Min
		} else if x > b.Max {
			return b.Max
		}
	}
	return x
}

// ChartContent is the interface that all chart contents must implement.
// If the chart is created at first, all Bounds methods are called and the
// returned bounds are merged. After that, the DependantBounds method
// is called with the merged bounds.
type ChartContent interface {
	// Bounds returns the fixed bounds of the content.
	// There may be non. For example, if the content is a function
	// that has by definition, no bounds because it is defined for all x
	// and the corresponding y=f(x) depends on the x bounds.
	// A set of given data points on the other hand has fixed bounds.
	Bounds() (x, y Bounds, err error)
	// DependantBounds returns the preferred bounds for the content
	// that depend on the bounds given by all other chart contents or the user.
	// A function y=f(x), for example, has certain y bounds if the x bounds
	// are given.
	DependantBounds(xGiven, yGiven Bounds) (x, y Bounds, err error)
	// DrawTo draws the content.
	// The *ChartContentEnvironment is passed to allow the content to access the
	// chart's properties, including the Canvas.
	DrawTo(*ChartContentEnvironment) error
	// Legend returns the legend of this Content
	Legend() []Legend
}

type HasLine interface {
	SetLine(*Style) ChartContent
}

type HasShape interface {
	SetShape(Shape, *Style) ChartContent
}

type HasTitle interface {
	SetTitle(title string) ChartContent
}

type IsCloseable interface {
	Close() ChartContent
}

// Function represents a mathematical function that can be plotted.
type Function struct {
	Function func(x float64) (float64, error)
	Style    *Style
	Title    string
	Steps    int
	closed   bool
}

func (f Function) SetLine(style *Style) ChartContent {
	f.Style = style
	return f
}

func (f Function) SetTitle(title string) ChartContent {
	f.Title = title
	return f
}

func (f Function) Close() ChartContent {
	f.closed = true
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
	if xGiven.isSet {
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

func (f Function) DrawTo(env *ChartContentEnvironment) error {
	r := env.Canvas.Rect()
	p, err := NewLinearParameterFunc(r.Min.X, r.Max.X, f.steps())
	if err != nil {
		return fmt.Errorf("error creating linear parameter function: %w", err)
	}
	p.Func = func(x float64) (Point, error) {
		y, err := f.Function(x)
		return Point{x, y}, err
	}
	p.closed = f.closed
	path := pFuncPath{
		pf:  p,
		env: env,
		r:   r,
	}
	return env.Canvas.DrawPath(r.IntersectPath(&path), f.Style)
}

func (f Function) Legend() []Legend {
	if f.Title == "" {
		return nil
	}
	return []Legend{{
		Name:           f.Title,
		ShapeLineStyle: ShapeLineStyle{LineStyle: f.Style},
	}}
}

// Scatter represents a scatter chart with points represented by a Shape
// and can have a line style for connecting the points.
type Scatter struct {
	ShapeLineStyle
	Points Points
	Closed bool
	Title  string
}

func (s Scatter) SetShape(shape Shape, style *Style) ChartContent {
	s.Shape = shape
	s.ShapeStyle = style
	return s
}

func (s Scatter) SetLine(style *Style) ChartContent {
	s.LineStyle = style
	return s
}

func (s Scatter) SetTitle(title string) ChartContent {
	s.Title = title
	return s
}

func (s Scatter) Close() ChartContent {
	s.Closed = true
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

func (s Scatter) DrawTo(env *ChartContentEnvironment) error {
	canvas := env.Canvas
	rect := canvas.Rect()

	sls := s.EnsureSomethingIsVisible()
	if sls.IsLine() {
		err := canvas.DrawPath(
			canvas.Rect().IntersectPath(
				CloseablePointsPath{
					Points: s.Points,
					Closed: s.Closed,
				}), sls.LineStyle)
		if err != nil {
			return err
		}
	}
	if sls.IsShape() {
		for p, err := range s.Points {
			if err != nil {
				return err
			}
			if rect.Contains(p) {
				err := canvas.DrawShape(p, sls.Shape, sls.ShapeStyle)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s Scatter) Legend() []Legend {
	if s.Title == "" {
		return nil
	}
	return []Legend{{
		Name:           s.Title,
		ShapeLineStyle: s.EnsureSomethingIsVisible(),
	}}
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

const arrowLen = 2.5

func (h Hint) DrawTo(env *ChartContentEnvironment) error {
	canvas := env.Canvas
	r := canvas.Rect()
	textSize := canvas.Context().TextSize
	if r.Contains(h.Pos) {
		r := env.ParentCanvas.Rect()
		hPos := env.Transform(h.Pos)
		tPos := hPos
		delta := textSize * arrowLen / math.Sqrt(2)
		if r.IsInLeftHalf(hPos) {
			tPos = tPos.Add(Point{delta, 0})
		} else {
			tPos = tPos.Add(Point{-delta, 0})
		}
		if r.IsInTopHalf(hPos) {
			tPos = tPos.Add(Point{0, -delta})
		} else {
			tPos = tPos.Add(Point{0, delta})
		}
		return drawArrow(env, tPos, hPos, h.Style, 1, h.Text)
	}
	return nil
}

func (h Hint) Legend() []Legend {
	return nil
}

type HintDir struct {
	Hint
	PosDir Point
}

func (h HintDir) DrawTo(env *ChartContentEnvironment) error {
	r := env.Canvas.Rect()
	if r.Contains(h.Pos) {
		p1 := env.Transform(h.Pos)
		p2 := env.Transform(h.PosDir)

		delta := p1.Sub(p2).Norm().Rot90().Mul(env.Canvas.Context().TextSize * arrowLen)
		tPos := p1.Add(delta)

		return drawArrow(env, tPos, p1, h.Style, 1, h.Text)
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

func (a Arrow) SetLine(s *Style) ChartContent {
	a.Style = s
	return a
}

func (a Arrow) SetTitle(s string) ChartContent {
	a.Label = s
	return a
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

func (a Arrow) DrawTo(env *ChartContentEnvironment) error {
	from := env.Transform(a.From)
	to := env.Transform(a.To)
	return drawArrow(env, from, to, a.Style, a.Mode, a.Label)
}

func (a Arrow) Legend() []Legend {
	return nil
}

func drawArrow(env *ChartContentEnvironment, from, to Point, style *Style, mode int, label string) error {
	textSize := env.ParentCanvas.Context().TextSize
	headLen, headWidth := env.getArrowHeadSize()

	d := to.Sub(from).Norm()
	dif := d.Mul(headLen)
	norm := d.Rot90().Mul(headWidth)

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
		err := env.ParentCanvas.DrawPath(p, style)
		if err != nil {
			return err
		}
	}

	if label != "" {
		if mode > 3 {
			mode >>= 2
		}
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
		env.ParentCanvas.DrawText(textPos, label, o, style.Text(), textSize)
	}
	return nil
}

func orientationByDelta(delta Point) Orientation {
	const f = 0.41
	var o Orientation
	if delta.X > math.Abs(delta.Y*f) {
		o = Left
	} else if delta.X < -math.Abs(delta.Y*f) {
		o = Right
	} else {
		o = HCenter
	}
	if delta.Y > math.Abs(delta.X*f) {
		o |= Bottom
	} else if delta.Y < -math.Abs(delta.X*f) {
		o |= Top
	} else {
		o |= VCenter
	}
	return o
}

type Arc struct {
	Pos            Point
	Radius         float64
	Alpha0, Alpha1 float64
	Style          *Style
	Label          string
	Mode           int
}

func (a Arc) String() string {
	if a.Label != "" {
		return fmt.Sprintf("Arc at %s: %s", a.Pos, a.Label)
	}
	return fmt.Sprintf("Arc at %s", a.Pos)
}

func (a Arc) SetLine(s *Style) ChartContent {
	a.Style = s
	return a
}

func (a Arc) SetTitle(s string) ChartContent {
	a.Label = s
	return a
}

func (a Arc) Bounds() (Bounds, Bounds, error) {
	var x, y Bounds
	x.Merge(a.Pos.X)
	y.Merge(a.Pos.Y)
	return x, y, nil
}

func (a Arc) DependantBounds(_, _ Bounds) (Bounds, Bounds, error) {
	return Bounds{}, Bounds{}, nil
}

func (a Arc) DrawTo(env *ChartContentEnvironment) error {
	textSize := env.ParentCanvas.Context().TextSize
	r := a.Radius * textSize

	pos := env.Transform(a.Pos)
	path := NewPath(false)
	if math.Abs(a.Alpha0-a.Alpha1) > 1e-6 {
		path = drawArcTo(path, pos, r, r, a.Alpha0, a.Alpha1, 4)
	}
	headLen, headWidth := env.getArrowHeadSize()
	if a.Mode&1 != 0 {
		da := math.Atan(headLen / r)
		path = drawArcTo(path, pos, r+headWidth, r, a.Alpha1-da, a.Alpha1, 4)
		path = drawArcTo(path, pos, r-headWidth, r, a.Alpha1-da, a.Alpha1, 4)
	}
	if a.Mode&2 != 0 {
		da := math.Atan(headLen / r)
		path = drawArcTo(path, pos, r, r+headWidth, a.Alpha0, a.Alpha0+da, 4)
		path = drawArcTo(path, pos, r, r-headWidth, a.Alpha0, a.Alpha0+da, 4)
	}
	err := env.ParentCanvas.DrawPath(path, a.Style)
	if err != nil {
		return err
	}

	if a.Label != "" {
		alpha := (a.Alpha0 + a.Alpha1) / 2
		p := Point{X: math.Cos(alpha), Y: math.Sin(alpha)}.Mul(r)
		env.ParentCanvas.DrawText(pos.Add(p), a.Label, orientationByDelta(p), a.Style.Text(), textSize)
	}
	return nil
}

func drawArcTo(path SlicePath, pos Point, r0, r1, a0, a1 float64, nMin int) SlicePath {
	n := int(math.Abs(a1-a0) / (2 * math.Pi) * 90)
	if nMin > n {
		n = nMin
	}
	if n > 0 {
		dAlpha := (a1 - a0) / float64(n)
		dR := (r1 - r0) / float64(n)
		for i := 0; i <= n; i++ {
			alpha := a0 + dAlpha*float64(i)
			r := r0 + dR*float64(i)
			p := Point{X: pos.X + r*math.Cos(alpha), Y: pos.Y + r*math.Sin(alpha)}
			if i == 0 {
				path = path.MoveTo(p)
			} else {
				path = path.LineTo(p)
			}
		}
	}
	return path
}

func (a Arc) Legend() []Legend {
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

func (c circleMarker) DrawTo(canvas Canvas, style *Style) error {
	canvas.DrawCircle(c.p1, c.p2, style)
	return nil
}

func NewCrossMarker(r float64) Shape {
	return NewPath(false).
		AddMode('M', Point{-r, -r}).
		AddMode('L', Point{r, r}).
		AddMode('M', Point{-r, r}).
		AddMode('L', Point{r, -r})
}

func NewSquareMarker(r float64) Shape {
	return CloseablePointsPath{
		Points: PointsFromSlice(Point{-r, -r}, Point{-r, r}, Point{r, r}, Point{r, -r}),
		Closed: true,
	}
}

func NewTriangleMarker(r float64) Shape {
	return CloseablePointsPath{
		Points: PointsFromSlice(Point{0, r}, Point{-r, -r}, Point{r, -r}),
		Closed: true,
	}
}

func NewDiamondMarker(r float64) Shape {
	r = r * math.Sqrt2
	return CloseablePointsPath{
		Points: PointsFromSlice(Point{-r, 0}, Point{0, r}, Point{r, 0}, Point{0, -r}),
		Closed: true,
	}
}

type Cross struct {
	Style *Style
}

func (c Cross) SetLine(style *Style) ChartContent {
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

func (c Cross) DrawTo(env *ChartContentEnvironment) error {
	var err error
	r := env.Canvas.Rect()
	if r.Contains(Point{0, 0}) {
		err = env.Canvas.DrawPath(NewPath(false).
			MoveTo(Point{r.Min.X, 0}).
			LineTo(Point{r.Max.X, 0}).
			MoveTo(Point{0, r.Min.Y}).
			LineTo(Point{0, r.Max.Y}), c.Style)
	}
	return err
}

func (c Cross) Legend() []Legend {
	return nil
}

type ParameterFunc struct {
	Func       func(t float64) (Point, error)
	Points     int
	InitialT   float64
	NextT      func(float64) float64
	Style      *Style
	Title      string
	ModifyPath func(Path) Path
	closed     bool
}

func (p *ParameterFunc) SetTitle(title string) ChartContent {
	np := *p
	np.Title = title
	return &np
}

func (p *ParameterFunc) SetLine(style *Style) ChartContent {
	np := *p
	np.Style = style
	return &np
}

func (p *ParameterFunc) Close() ChartContent {
	np := *p
	np.closed = true
	return &np
}

func NewLinearParameterFunc(tMin, tMax float64, steps int) (*ParameterFunc, error) {
	if steps <= 0 {
		steps = functionSteps
	}
	delta := (tMax - tMin) / float64(steps)
	return &ParameterFunc{
		Points:   steps,
		Style:    Black,
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
		Style:    Black,
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

func (p *ParameterFunc) DrawTo(env *ChartContentEnvironment) error {
	canvas := env.Canvas
	path := pFuncPath{
		pf:  p,
		env: env,
		r:   canvas.Rect(),
	}
	if p.ModifyPath != nil {
		return canvas.DrawPath(canvas.Rect().IntersectPath(p.ModifyPath(&path)), p.Style)
	}
	return canvas.DrawPath(canvas.Rect().IntersectPath(&path), p.Style)
}

func (p *ParameterFunc) Legend() []Legend {
	if p.Title == "" {
		return nil
	}
	return []Legend{{
		Name:           p.Title,
		ShapeLineStyle: ShapeLineStyle{LineStyle: p.Style},
	}}
}

type pFuncPath struct {
	pf      *ParameterFunc
	env     *ChartContentEnvironment
	r       Rect
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

func (p *pFuncPath) Iter(yield func(PathElement, error) bool) {
	if p.maxDist == 0 {
		p.maxDist = p.env.ParentCanvas.Rect().Width() / float64(p.pf.Points) * 2
	}
	pf := p.pf
	t0 := pf.InitialT
	p0, d0, err := p.f(t0, pf.NextT(t0)-t0)
	if !yield(PathElement{Mode: 'M', Point: p0}, err) {
		return
	}
	for i := 1; i <= pf.Points; i++ {
		t1 := pf.NextT(t0)
		p1, d1, err := p.f(t1, t1-t0)
		if !p.refine(t0, p0, d0, t1, p1, d1, yield, 10, 0) {
			return
		}
		if !yield(PathElement{Mode: 'L', Point: p1}, err) {
			return
		}
		t0 = t1
		p0 = p1
		d0 = d1
	}
}

func (p *pFuncPath) IsClosed() bool {
	return p.pf.closed
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

func (p *pFuncPath) refine(w0 float64, p0, d0 Point, w1 float64, p1, d1 Point, yield func(PathElement, error) bool, depth int, lastDist float64) bool {
	dw := w1 - w0
	dist := p.env.Dist(p0, p1)
	if dist > p.maxDist || // the distance of two points is too large
		p.env.Dist(p1, p0.Add(d0.Mul(dw))) > p.maxDist/50 || // the distance to the tangent is too large
		cosAngleBetween(d0, d1) < 0.98 { // the angle is larger than approx 10 degrees
		if depth > 0 {
			w := (w0 + w1) / 2
			point, delta, err := p.f(w, dw)
			if !p.refine(w0, p0, d0, w, point, delta, yield, depth-1, dist) {
				return false
			}
			if !yield(PathElement{Mode: 'L', Point: point}, err) {
				return false
			}
			if !p.refine(w, point, delta, w1, p1, d1, yield, depth-1, dist) {
				return false
			}
		} else {
			// detecting poles
			if dist > lastDist*1.001 && dist > p.maxDist {
				// if a pole is detected, do not draw a line
				if !yield(PathElement{Mode: 'M', Point: p1}, nil) {
					return false
				}
			}
		}
	}
	return true
}

type ImageInset struct {
	Min         RelativePos
	Max         RelativePos
	Chart       *Chart
	VisualGuide *Style
}

func (s ImageInset) SetLine(style *Style) ChartContent {
	s.Chart.Background = style
	return s
}

func (s ImageInset) String() string {
	return "ImageInset(" + s.Chart.String() + ")"
}

func (s ImageInset) Bounds() (Bounds, Bounds, error) {
	var x, y Bounds
	if s.Min.Mode == AbsPos {
		x.Merge(s.Min.Pos.X)
		y.Merge(s.Min.Pos.Y)
	}
	if s.Max.Mode == AbsPos {
		x.Merge(s.Max.Pos.X)
		y.Merge(s.Max.Pos.Y)
	}
	return x, y, nil
}

func (s ImageInset) DependantBounds(_, _ Bounds) (Bounds, Bounds, error) {
	return Bounds{}, Bounds{}, nil
}

func (s ImageInset) DrawTo(env *ChartContentEnvironment) (cErr error) {
	defer nErr.CatchErr(&cErr)
	inner := ResizeCanvas{
		parent: env.ParentCanvas,
		size: Rect{
			Min: s.Min.Get(env.Transform, env.InnerRect),
			Max: s.Max.Get(env.Transform, env.InnerRect),
		},
	}
	err, ie := s.Chart.DrawToAsInset(inner)
	if err != nil {
		return fmt.Errorf("error drawing image inset: %w", err)
	}

	if s.VisualGuide != nil {
		ir := ie.Canvas.Rect()
		sMin := env.Transform(ir.Min)
		sMax := env.Transform(ir.Max)

		nErr.Try(
			env.ParentCanvas.DrawPath(SlicePath{Closed: true}.
				Add(sMin).
				Add(Point{X: sMax.X, Y: sMin.Y}).
				Add(sMax).
				Add(Point{X: sMin.X, Y: sMax.Y}), s.VisualGuide),
		)

		lMin := ie.Transform(ir.Min)
		lMax := ie.Transform(ir.Max)
		s12 := Point{X: sMin.X, Y: sMax.Y}
		l12 := Point{X: lMin.X, Y: lMax.Y}
		s21 := Point{X: sMax.X, Y: sMin.Y}
		l21 := Point{X: lMax.X, Y: lMin.Y}

		if (lMin.X < sMin.X && lMin.Y > sMin.Y) || (lMin.X > sMin.X && lMin.Y < sMin.Y) {
			nErr.Try(env.ParentCanvas.DrawPath(NewPath(false).Add(sMin).Add(lMin), s.VisualGuide))
		}
		if (l12.X > s12.X && l12.Y > s12.Y) || (l12.X < s12.X && l12.Y < s12.Y) {
			nErr.Try(env.ParentCanvas.DrawPath(NewPath(false).Add(s12).Add(l12), s.VisualGuide))
		}
		if (lMax.X < sMax.X && lMax.Y > sMax.Y) || (lMax.X > sMax.X && lMax.Y < sMax.Y) {
			nErr.Try(env.ParentCanvas.DrawPath(NewPath(false).Add(sMax).Add(lMax), s.VisualGuide))
		}
		if (l21.X > s21.X && l21.Y > s21.Y) || (l21.X < s21.X && l21.Y < s21.Y) {
			nErr.Try(env.ParentCanvas.DrawPath(NewPath(false).Add(s21).Add(l21), s.VisualGuide))
		}

	}
	return nil
}

func (s ImageInset) Legend() []Legend {
	return nil
}

type YConst struct {
	Y     float64
	Style *Style
	Title string
}

func (yc YConst) SetLine(style *Style) ChartContent {
	return YConst{
		Y:     yc.Y,
		Title: yc.Title,
		Style: style,
	}
}

func (yc YConst) SetTitle(t string) ChartContent {
	return YConst{
		Y:     yc.Y,
		Title: t,
		Style: yc.Style,
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

func (yc YConst) DrawTo(env *ChartContentEnvironment) error {
	var err error
	r := env.Canvas.Rect()
	if r.Max.Y >= yc.Y && r.Min.Y <= yc.Y {
		err = env.Canvas.DrawPath(NewPath(false).
			MoveTo(Point{r.Min.X, yc.Y}).
			LineTo(Point{r.Max.X, yc.Y}), yc.Style)
	}
	return err
}

func (yc YConst) Legend() []Legend {
	if yc.Title == "" {
		return nil
	}
	return []Legend{{ShapeLineStyle: ShapeLineStyle{LineStyle: yc.Style}, Name: yc.Title}}
}

type XConst struct {
	X     float64
	Style *Style
	Title string
}

func (xc XConst) SetLine(style *Style) ChartContent {
	return XConst{
		X:     xc.X,
		Style: style,
		Title: xc.Title,
	}
}

func (xc XConst) SetTitle(t string) ChartContent {
	return XConst{
		X:     xc.X,
		Title: t,
		Style: xc.Style,
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

func (xc XConst) DrawTo(env *ChartContentEnvironment) error {
	var err error
	r := env.Canvas.Rect()
	if r.Max.X >= xc.X && r.Min.X <= xc.X {
		err = env.Canvas.DrawPath(NewPath(false).
			MoveTo(Point{xc.X, r.Min.Y}).
			LineTo(Point{xc.X, r.Max.Y}), xc.Style)
	}
	return err
}

func (xc XConst) Legend() []Legend {
	if xc.Title == "" {
		return nil
	}
	return []Legend{{ShapeLineStyle: ShapeLineStyle{LineStyle: xc.Style}, Name: xc.Title}}
}

type Text struct {
	Pos         Point
	Orientation Orientation
	Text        string
	Style       *Style
}

func (t Text) Bounds() (x, y Bounds, err error) {
	return NewBounds(t.Pos.X, t.Pos.X), NewBounds(t.Pos.Y, t.Pos.Y), nil
}

func (t Text) DependantBounds(_, _ Bounds) (x, y Bounds, err error) {
	return Bounds{}, Bounds{}, nil
}

func (t Text) DrawTo(env *ChartContentEnvironment) error {
	env.Canvas.DrawText(t.Pos, t.Text, t.Orientation, t.Style.Text(), env.Canvas.Context().TextSize)
	return nil
}

func (t Text) Legend() []Legend {
	return nil
}

type Heat struct {
	AxisFactory AxisFactory
	FuncFac     func() func(x, y float64) (col.RGBA, error)
	Steps       int
}

func (h Heat) Bounds() (x, y Bounds, err error) {
	return Bounds{}, Bounds{}, nil
}

func (h Heat) DependantBounds(_, _ Bounds) (x, y Bounds, err error) {
	return Bounds{}, Bounds{}, nil
}

type pixLine struct {
	y   int
	pix []col.RGBA
	err error
}

func (h Heat) DrawTo(env *ChartContentEnvironment) error {
	if h.AxisFactory == nil {
		h.AxisFactory = LinearAxis
	}

	xa := env.XAxis
	if xa.Reverse == nil {
		return fmt.Errorf("heat chart requires a reverable x axis")
	}
	ya := env.YAxis
	if ya.Reverse == nil {
		return fmt.Errorf("heat chart requires a reverable y axis")
	}
	if h.FuncFac == nil {
		return fmt.Errorf("heat chart requires a function")
	}

	steps := h.Steps
	if steps <= 0 {
		steps = 400
	}

	im := img.NewRGBA(img.Rectangle{Min: img.Point{}, Max: img.Point{X: steps, Y: steps}})

	r := env.InnerRect

	lines := make(chan int)
	stop := make(chan struct{})
	go func() {
		for y := 0; y < steps; y++ {
			select {
			case <-stop:
				return
			case lines <- y:
			}
		}
		close(lines)
	}()
	wg := sync.WaitGroup{}
	results := make(chan pixLine)
	for range runtime.NumCPU() {
		wg.Add(1)
		f := h.FuncFac()
		go func() {
			defer wg.Done()

			pix := [][]col.RGBA{make([]col.RGBA, steps), make([]col.RGBA, steps)}
			l := 0
			for y := range lines {
				yp := ya.Reverse(r.Min.Y + (r.Max.Y-r.Min.Y)*float64(y)/float64(steps-1))
				var co col.RGBA
				var err error
				for x := 0; x < steps; x++ {
					xp := xa.Reverse(r.Min.X + (r.Max.X-r.Min.X)*float64(x)/float64(steps-1))
					co, err = f(xp, yp)
					if err != nil {
						break
					}
					pix[l][x] = co
				}
				select {
				case <-stop:
					return
				case results <- pixLine{y: y, pix: pix[l], err: err}:
				}
				l = 1 - l
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for pixL := range results {
		if pixL.err != nil {
			close(stop)
			return fmt.Errorf("error evaluating heat function: %w", pixL.err)
		}
		for x := 0; x < steps; x++ {
			im.Set(x, steps-pixL.y-1, pixL.pix[x])
		}
	}

	p1 := Point{X: xa.Reverse(r.Min.X), Y: ya.Reverse(r.Min.Y)}
	p2 := Point{X: xa.Reverse(r.Max.X), Y: ya.Reverse(r.Max.Y)}
	return env.Canvas.DrawImage(p1, p2, im)
}

func (h Heat) Legend() []Legend {
	return nil
}

type BarSet struct {
	Points Points
	Style  *Style
	Title  string
}

type Bars struct {
	BarSets    []BarSet
	Horizontal bool
}

func (b Bars) String() string {
	return "BarChart"
}

func (b Bars) SetLine(style *Style) ChartContent {
	l := len(b.BarSets)
	if l > 0 {
		b.BarSets[l-1].Style = style
	}
	return b
}

func (b Bars) SetTitle(title string) ChartContent {
	l := len(b.BarSets)
	if l > 0 {
		b.BarSets[l-1].Title = title
	}
	return b
}

func (b Bars) Add(cc ChartContent) (ChartContent, error) {
	if other, ok := cc.(Bars); ok {
		return Bars{
			BarSets:    append(b.BarSets, other.BarSets...),
			Horizontal: b.Horizontal || other.Horizontal,
		}, nil
	}
	return nil, fmt.Errorf("cannot add %T to Bars", cc)
}

func (b Bars) barWidth() float64 {
	var w Bounds
	nMax := 0
	for _, sets := range b.BarSets {
		n := 0
		for p, err := range sets.Points {
			if err != nil {
				return 0
			}
			w.Merge(p.X)
			n++
		}
		if n > nMax {
			nMax = n
		}
	}
	bars := (len(b.BarSets)+1)*(nMax-1) + 1
	width := w.Width() / float64(bars)
	if len(b.BarSets) == 1 {
		width *= 1.75
	}
	return width
}

func (b Bars) Bounds() (Bounds, Bounds, error) {
	var x, y Bounds
	y.Merge(0)
	for _, sets := range b.BarSets {
		for p, err := range sets.Points {
			if err != nil {
				return Bounds{}, Bounds{}, err
			}
			x.Merge(p.X)
			y.Merge(p.Y)
		}
	}

	bw := b.barWidth() * float64(len(b.BarSets)) / 2
	x.Merge(x.Min - bw)
	x.Merge(x.Max + bw)

	if b.Horizontal {
		return y, x, nil
	}
	return x, y, nil
}

func (b Bars) DependantBounds(_, _ Bounds) (Bounds, Bounds, error) {
	return Bounds{}, Bounds{}, nil
}

func (b Bars) DrawTo(environment *ChartContentEnvironment) error {
	canvas := environment.Canvas
	barWidth := b.barWidth()
	gap := barWidth * 0.1
	ofs := float64(len(b.BarSets)) * barWidth / 2
	for i, set := range b.BarSets {
		style := set.Style
		if !style.Fill {
			style = style.SetFill(style)
		}
		for p, err := range set.Points {
			if err != nil {
				return fmt.Errorf("error plotting bars: %w", err)
			}
			var x, y0, y1 float64
			if b.Horizontal {
				x = environment.YAxis.Bounds.Limit(p.X)
				y0 = environment.XAxis.Bounds.Limit(0)
				y1 = environment.XAxis.Bounds.Limit(p.Y)
			} else {
				x = environment.XAxis.Bounds.Limit(p.X)
				y0 = environment.YAxis.Bounds.Limit(0)
				y1 = environment.YAxis.Bounds.Limit(p.Y)
			}
			if y0 != y1 {
				o := barWidth*float64(i) - ofs
				var path Path
				if barWidth == 0 {
					path = NewPath(false).
						MoveTo(Point{X: x + o, Y: y0}).
						LineTo(Point{X: x + o, Y: y1})
				} else {
					path = NewPath(true).
						MoveTo(Point{X: x + o + gap, Y: y0}).
						LineTo(Point{X: x + o + barWidth - gap, Y: y0}).
						LineTo(Point{X: x + o + barWidth - gap, Y: y1}).
						LineTo(Point{X: x + o + gap, Y: y1})

				}
				if b.Horizontal {
					path = swapXYPath{path: path}
				}
				err = canvas.DrawPath(path, style)
				if err != nil {
					return fmt.Errorf("error plotting bars: %w", err)
				}
			}
		}
	}
	return nil
}

type swapXYPath struct {
	path Path
}

func (s swapXYPath) Iter(yield func(PathElement, error) bool) {
	for pe, err := range s.path.Iter {
		if !yield(PathElement{Mode: pe.Mode, Point: Point{X: pe.Point.Y, Y: pe.Point.X}}, err) {
			return
		}
	}
}

func (s swapXYPath) IsClosed() bool {
	return s.path.IsClosed()
}

func (b Bars) Legend() []Legend {
	var l []Legend
	for _, set := range slices.Backward(b.BarSets) {
		if set.Title != "" {
			style := set.Style
			if !style.Fill {
				style = style.SetFill(style)
			}
			l = append(l, Legend{
				Name:           set.Title,
				ShapeLineStyle: ShapeLineStyle{LineStyle: style},
				ColorOnly:      true,
			})
		}
	}
	return l
}
