package graph

import (
	"fmt"
	"github.com/hneemann/control/nErr"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/listMap"
	"github.com/hneemann/parser2/value"
	"math"
	"sort"
)

var Vector3dType value.Type

type Vector3d struct {
	X, Y, Z float64
}

func (v Vector3d) ToList() (*value.List, bool) {
	return nil, false
}

func (v Vector3d) ToMap() (value.Map, bool) {
	return value.NewMap(listMap.New[value.Value](3).Append("x", value.Float(v.X)).Append("y", value.Float(v.Y)).Append("z", value.Float(v.Z))), true
}

func (v Vector3d) ToInt() (int, bool) {
	return 0, false
}

func (v Vector3d) ToFloat() (float64, bool) {
	return 0, false
}

func (v Vector3d) ToString(st funcGen.Stack[value.Value]) (string, error) {
	return fmt.Sprintf("(%g,%g,%g)", v.X, v.Y, v.Z), nil
}

func (v Vector3d) ToBool() (bool, bool) {
	return false, false
}

func (v Vector3d) ToClosure() (funcGen.Function[value.Value], bool) {
	return funcGen.Function[value.Value]{}, false
}

func (v Vector3d) GetType() value.Type {
	return Vector3dType
}

func (v Vector3d) Sub(p2 Vector3d) Vector3d {
	return Vector3d{
		X: v.X - p2.X,
		Y: v.Y - p2.Y,
		Z: v.Z - p2.Z,
	}
}

func (v Vector3d) Cross(p Vector3d) Vector3d {
	return Vector3d{
		X: v.Y*p.Z - v.Z*p.Y,
		Y: v.Z*p.X - v.X*p.Z,
		Z: v.X*p.Y - v.Y*p.X,
	}
}

func (v Vector3d) Normalize() Vector3d {
	l := math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
	if l == 0 {
		return Vector3d{0, 0, 0}
	}
	return Vector3d{
		X: v.X / l,
		Y: v.Y / l,
		Z: v.Z / l,
	}
}

func (v Vector3d) Scalar(p2 Vector3d) float64 {
	return v.X*p2.X + v.Y*p2.Y + v.Z*p2.Z
}

func (v Vector3d) Mul(f float64) Vector3d {
	return Vector3d{
		X: v.X * f,
		Y: v.Y * f,
		Z: v.Z * f,
	}
}

func (v Vector3d) Add(d Vector3d) Vector3d {
	return Vector3d{
		X: v.X + d.X,
		Y: v.Y + d.Y,
		Z: v.Z + d.Z,
	}
}

func (v Vector3d) Abs() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

func (v Vector3d) Zero() bool {
	return v.X == 0 && v.Y == 0 && v.Z == 0
}

func (v Vector3d) Neg() value.Value {
	return Vector3d{-v.X, -v.Y, -v.Z}
}

type PathElement3d struct {
	Mode rune
	Vector3d
}

type Path3d interface {
	Iter(func(PathElement3d, error) bool)
	IsClosed() bool
}

type SlicePath3d struct {
	Elements []PathElement3d
	Closed   bool
}

func NewPath3d(closed bool) SlicePath3d {
	return SlicePath3d{Closed: closed}
}

func (p SlicePath3d) Iter(yield func(PathElement3d, error) bool) {
	for _, e := range p.Elements {
		if !yield(e, nil) {
			break
		}
	}
}

func (p SlicePath3d) IsClosed() bool {
	return p.Closed
}

func (p SlicePath3d) Add(point Vector3d) SlicePath3d {
	if len(p.Elements) == 0 {
		return SlicePath3d{append(p.Elements, PathElement3d{Mode: 'M', Vector3d: point}), p.Closed}
	} else {
		return SlicePath3d{append(p.Elements, PathElement3d{Mode: 'L', Vector3d: point}), p.Closed}
	}
}

type LinePath3d struct {
	Func   func(t float64) (Vector3d, error)
	Bounds Bounds
	Steps  int
}

func (l LinePath3d) Iter(f func(PathElement3d, error) bool) {
	for i := 0; i <= l.Steps; i++ {
		c := l.Bounds.Min + float64(i)*l.Bounds.Width()/float64(l.Steps)
		p, err := l.Func(c)
		if i == 0 {
			if !f(PathElement3d{'M', p}, err) {
				return
			}
		} else {
			if !f(PathElement3d{'L', p}, err) {
				return
			}
		}
		if err != nil {
			return
		}
	}
}

func (l LinePath3d) IsClosed() bool {
	return false
}

type Cube interface {
	DrawPath(Path3d, *Style) error
	DrawTriangle(Vector3d, Vector3d, Vector3d, *Style, *Style) error
	DrawLine(Vector3d, Vector3d, *Style)
	DrawText(Vector3d, Vector3d, string, *Style)
	Bounds() (x, y, z Bounds)
}

type Plot3dContent interface {
	DrawTo(*Plot3d, Cube) error
	Legend() Legend
	SetStyle(s *Style) Plot3dContent
}
type UBoundsSetter interface {
	SetUBounds(bounds Bounds) Plot3dContent
}

type VBoundsSetter interface {
	SetVBounds(bounds Bounds) Plot3dContent
}

type TitleSetter interface {
	SetTitle(title string) Plot3dContent
}

type Plot3d struct {
	X, Y, Z     AxisDescription
	Contents    []Plot3dContent
	alpha       float64
	beta        float64
	gamma       float64
	HideCube    bool
	Size        float64
	Perspective float64
}

func NewPlot3d() *Plot3d {
	return &Plot3d{
		alpha:       0.3,
		beta:        0.2,
		gamma:       0,
		Size:        0.98,
		Perspective: 1,
	}
}

type unityCube struct {
	parent     Cube
	X, Y, Z    Bounds
	ax, ay, az Axis
}

func newUnityCube(parent Cube, x, y, z AxisDescription) *unityCube {
	ctw := func(width float64, digits int) bool {
		return width > float64(digits)*10
	}
	ax := x.Factory(-100, 100, x.Bounds, ctw, 0.02)
	ay := y.Factory(-100, 100, y.Bounds, ctw, 0.02)
	az := z.Factory(-100, 100, z.Bounds, ctw, 0.02)
	return &unityCube{
		parent: parent,
		X:      x.Bounds, Y: y.Bounds, Z: z.Bounds,
		ax: ax, ay: ay, az: az,
	}
}

func (t *unityCube) transform(p Vector3d) Vector3d {
	return Vector3d{
		X: t.ax.Trans(p.X),
		Y: t.ay.Trans(p.Y),
		Z: t.az.Trans(p.Z),
	}
}

type unityPath3d struct {
	p Path3d
	u *unityCube
}

func (t unityPath3d) Iter(yield func(PathElement3d, error) bool) {
	for pe, err := range t.p.Iter {
		if !yield(PathElement3d{Mode: pe.Mode, Vector3d: t.u.transform(pe.Vector3d)}, err) {
			return
		}
		if err != nil {
			return
		}
	}
}

func (t unityPath3d) IsClosed() bool {
	return t.p.IsClosed()
}

func (uc *unityCube) DrawPath(d Path3d, style *Style) error {
	return uc.parent.DrawPath(&unityPath3d{d, uc}, style)
}

func (uc *unityCube) DrawTriangle(p1, p2, p3 Vector3d, s1, s2 *Style) error {
	return uc.parent.DrawTriangle(uc.transform(p1), uc.transform(p2), uc.transform(p3), s1, s2)
}

func (uc *unityCube) DrawLine(p1, p2 Vector3d, lineStyle *Style) {
	uc.parent.DrawLine(uc.transform(p1), uc.transform(p2), lineStyle)
}

func (uc *unityCube) DrawText(p, d Vector3d, s string, style *Style) {
	uc.parent.DrawText(uc.transform(p), uc.transform(d), s, style)
}

func (uc *unityCube) Bounds() (x, y, z Bounds) {
	return uc.X, uc.Y, uc.Z
}

type RotCube struct {
	parent Cube
	matrix Matrix3d
}

func NewRotCube(cube Cube, alpha float64, beta float64, gamma float64) Cube {
	return RotCube{
		parent: cube,
		matrix: NewRotX(-alpha).MulMatrix(NewRotY(gamma)).MulMatrix(NewRotZ(beta)),
	}
}

func (r RotCube) DrawPath(p Path3d, style *Style) error {
	return r.parent.DrawPath(&rotPath3d{p, r}, style)
}

func (r RotCube) DrawText(p, d Vector3d, s string, style *Style) {
	r.parent.DrawText(r.matrix.MulPoint(p), r.matrix.MulPoint(d), s, style)
}

func (r RotCube) DrawTriangle(p1, p2, p3 Vector3d, s1, s2 *Style) error {
	return r.parent.DrawTriangle(r.matrix.MulPoint(p1), r.matrix.MulPoint(p2), r.matrix.MulPoint(p3), s1, s2)
}

func (r RotCube) DrawLine(p1, p2 Vector3d, lineStyle *Style) {
	r.parent.DrawLine(r.matrix.MulPoint(p1), r.matrix.MulPoint(p2), lineStyle)
}

func (r RotCube) Bounds() (x, y, z Bounds) {
	return NewBounds(-100, 100), NewBounds(-100, 100), NewBounds(-100, 100)
}

type rotPath3d struct {
	p Path3d
	r RotCube
}

func (t *rotPath3d) Iter(yield func(PathElement3d, error) bool) {
	for pe, err := range t.p.Iter {
		if !yield(PathElement3d{Mode: pe.Mode, Vector3d: t.r.matrix.MulPoint(pe.Vector3d)}, err) {
			return
		}
		if err != nil {
			return
		}
	}
}

type Matrix3d [3][3]float64

func (m Matrix3d) MulMatrix(n Matrix3d) Matrix3d {
	var r Matrix3d
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			for k := 0; k < 3; k++ {
				r[i][j] += m[i][k] * n[k][j]
			}
		}
	}
	return r
}

func (m Matrix3d) MulPoint(p Vector3d) Vector3d {
	return Vector3d{
		X: m[0][0]*p.X + m[0][1]*p.Y + m[0][2]*p.Z,
		Y: m[1][0]*p.X + m[1][1]*p.Y + m[1][2]*p.Z,
		Z: m[2][0]*p.X + m[2][1]*p.Y + m[2][2]*p.Z,
	}
}

func NewRotX(alpha float64) Matrix3d {
	return Matrix3d{
		{1, 0, 0},
		{0, math.Cos(alpha), -math.Sin(alpha)},
		{0, math.Sin(alpha), math.Cos(alpha)},
	}
}

func NewRotY(beta float64) Matrix3d {
	return Matrix3d{
		{math.Cos(beta), 0, math.Sin(beta)},
		{0, 1, 0},
		{-math.Sin(beta), 0, math.Cos(beta)},
	}
}

func NewRotZ(gamma float64) Matrix3d {
	return Matrix3d{
		{math.Cos(gamma), -math.Sin(gamma), 0},
		{math.Sin(gamma), math.Cos(gamma), 0},
		{0, 0, 1},
	}
}

func (t *rotPath3d) IsClosed() bool {
	return t.p.IsClosed()
}

type Object interface {
	DrawTo(cube *CanvasCube) error
	dist() float64
}

type triangle3d struct {
	p1, p2, p3     Vector3d
	style1, style2 *Style
}

func (d triangle3d) DrawTo(cube *CanvasCube) error {
	s := d.style1
	if d.style2 != nil {
		a := math.Abs(d.lightAngle())
		c1 := shade(d.style1, d.style2, a)
		s = &c1
	}
	return cube.canvas.DrawPath(cPath{p: NewPath3d(true).Add(d.p1).Add(d.p2).Add(d.p3), cc: cube}, s)
}

func shade(s1, s2 *Style, a float64) Style {
	col := Color{
		R: uint8(float64(s1.FillColor.R)*(1-a) + float64(s2.Color.R)*a),
		G: uint8(float64(s1.FillColor.G)*(1-a) + float64(s2.Color.G)*a),
		B: uint8(float64(s1.FillColor.B)*(1-a) + float64(s2.Color.B)*a),
		A: 255,
	}
	s := *s1
	s.FillColor = col
	return s
}

func (d triangle3d) dist() float64 {
	return (d.p1.Y + d.p2.Y + d.p3.Y) / 3
}

func (d triangle3d) lightAngle() float64 {
	n := d.p2.Sub(d.p1).Cross(d.p3.Sub(d.p1)).Normalize()
	lightDir := Vector3d{X: 1, Y: 1, Z: 1}.Normalize()
	return n.Scalar(lightDir)
}

type CanvasCube struct {
	canvas      Canvas
	textSize    float64
	dx, dy      float64
	fac         float64
	perspective float64

	objects []Object
}

func newCanvasCube(canvas Canvas, size, perspective float64) *CanvasCube {
	rect := canvas.Rect()

	fac := size * math.Min(rect.Width(), rect.Height()) / math.Sqrt(2) / 200

	return &CanvasCube{
		canvas:      canvas,
		textSize:    canvas.Context().TextSize,
		dx:          rect.Min.X + rect.Width()/2,
		dy:          rect.Min.Y + rect.Height()/2,
		fac:         fac,
		perspective: perspective,
	}
}

type cPath struct {
	p  Path3d
	cc *CanvasCube
}

func (c cPath) Iter(yield func(PathElement, error) bool) {
	for pe, err := range c.p.Iter {
		if !yield(PathElement{Mode: pe.Mode, Point: c.cc.To2d(pe.Vector3d)}, err) {
			return
		}
		if err != nil {
			return
		}
	}
}

func (c cPath) IsClosed() bool {
	return c.p.IsClosed()
}

func (c *CanvasCube) To2d(p Vector3d) Point {
	zFac := 1 + p.Y/800*c.perspective
	return Point{
		X: p.X*c.fac*zFac + c.dx,
		Y: c.dy + p.Z*c.fac*zFac,
	}
}

func (c *CanvasCube) DrawPath(p Path3d, style *Style) error {
	return c.canvas.DrawPath(cPath{p: p, cc: c}, style)
}

func (c *CanvasCube) DrawTriangle(p1, p2, p3 Vector3d, s1, s2 *Style) error {
	c.objects = append(c.objects, triangle3d{p1, p2, p3, s1, s2})
	return nil
}

type line3d struct {
	p1, p2 Vector3d
	style  *Style
}

func (l line3d) DrawTo(cube *CanvasCube) error {
	p1 := cube.To2d(l.p1)
	p2 := cube.To2d(l.p2)
	return cube.canvas.DrawPath(NewPath(false).Add(p1).Add(p2), l.style)
}

func (l line3d) dist() float64 {
	return (l.p1.Y + l.p2.Y) / 2
}

func (c *CanvasCube) DrawLine(p1, p2 Vector3d, lineStyle *Style) {
	c.objects = append(c.objects, line3d{p1, p2, lineStyle})
}

type text3d struct {
	p, d  Vector3d
	s     string
	style *Style
}

func (t text3d) DrawTo(cube *CanvasCube) error {
	p1 := cube.To2d(t.p)
	p2 := cube.To2d(t.d)
	o := orientationByDelta(p2.Sub(p1))
	cube.canvas.DrawText(p2, t.s, o, t.style, cube.canvas.Context().TextSize)
	return nil
}

func (t text3d) dist() float64 {
	return t.p.Y
}

func (c *CanvasCube) DrawText(p, d Vector3d, s string, style *Style) {
	c.objects = append(c.objects, text3d{p, d, s, style})
}

func (c *CanvasCube) Bounds() (x, y, z Bounds) {
	return NewBounds(-100, 100), NewBounds(-100, 100), NewBounds(-100, 100)
}

func (c *CanvasCube) DrawObjects() error {
	sort.Slice(c.objects, func(i, j int) bool {
		return c.objects[i].dist() < c.objects[j].dist()
	})
	for _, o := range c.objects {
		err := o.DrawTo(c)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Plot3d) DrawTo(canvas Canvas) (err error) {
	defer nErr.CatchErr(&err)

	canvasCube := newCanvasCube(canvas, p.Size, p.Perspective)
	rot := NewRotCube(canvasCube, p.alpha, p.beta, p.gamma)

	cubeColor := Gray
	textColor := cubeColor.Text()
	if !p.HideCube {
		DrawLongLine(rot, Vector3d{100, 100, 100}, Vector3d{100, -100, 100}, cubeColor)
		DrawLongLine(rot, Vector3d{100, 100, -100}, Vector3d{100, -100, -100}, cubeColor)
		DrawLongLine(rot, Vector3d{-100, 100, 100}, Vector3d{-100, -100, 100}, cubeColor)
		DrawLongLine(rot, Vector3d{-100, 100, -100}, Vector3d{-100, -100, -100}, cubeColor)

		DrawLongLine(rot, Vector3d{100, 100, 100}, Vector3d{-100, 100, 100}, cubeColor)
		DrawLongLine(rot, Vector3d{100, 100, -100}, Vector3d{-100, 100, -100}, cubeColor)
		DrawLongLine(rot, Vector3d{100, -100, 100}, Vector3d{-100, -100, 100}, cubeColor)
		DrawLongLine(rot, Vector3d{100, -100, -100}, Vector3d{-100, -100, -100}, cubeColor)

		DrawLongLine(rot, Vector3d{100, 100, -100}, Vector3d{100, 100, 100}, cubeColor)
		DrawLongLine(rot, Vector3d{-100, 100, -100}, Vector3d{-100, 100, 100}, cubeColor)
		DrawLongLine(rot, Vector3d{100, -100, -100}, Vector3d{100, -100, 100}, cubeColor)
		DrawLongLine(rot, Vector3d{-100, -100, -100}, Vector3d{-100, -100, 100}, cubeColor)
	}
	cube := newUnityCube(rot, p.X.MakeValid(), p.Y.MakeValid(), p.Z.MakeValid())
	const facShortLabel = 1.02
	const facLongLabel = 1.04
	const facText = 1.1
	if !p.X.HideAxis {
		for _, tick := range cube.ax.Ticks {
			xp := cube.ax.Trans(tick.Position)
			yp := -100.0
			zp := -100.0
			rot.DrawLine(Vector3d{xp, yp, zp}, Vector3d{xp, yp * facLongLabel, zp}, cubeColor)
			rot.DrawText(Vector3d{xp, yp, zp}, Vector3d{xp, yp * facLongLabel, zp}, tick.Label, textColor)
		}
		t := Vector3d{100 * facText, -100, -100}
		rot.DrawText(Vector3d{t.X, t.Y, t.Z}, Vector3d{t.X, t.Y * facLongLabel, t.Z}, checkEmpty(p.X.Label, "x"), textColor)
	}
	if !p.Y.HideAxis {
		for _, tick := range cube.ay.Ticks {
			xp := -100.0
			yp := cube.ay.Trans(tick.Position)
			zp := -100.0
			rot.DrawLine(Vector3d{xp, yp, zp}, Vector3d{xp * facShortLabel, yp, zp}, cubeColor)
			rot.DrawText(Vector3d{xp, yp, zp}, Vector3d{xp * facShortLabel, yp, zp}, tick.Label, textColor)
		}
		t := Vector3d{-100, 100 * facText, -100}
		rot.DrawText(Vector3d{t.X, t.Y, t.Z}, Vector3d{t.X * facLongLabel, t.Y, t.Z}, checkEmpty(p.Y.Label, "y"), textColor)
	}
	if !p.Z.HideAxis {
		for _, tick := range cube.az.Ticks {
			xp := -100.0
			yp := -100.0
			zp := cube.az.Trans(tick.Position)
			rot.DrawLine(Vector3d{xp, yp, zp}, Vector3d{xp * facShortLabel, yp * facShortLabel, zp}, cubeColor)
			rot.DrawText(Vector3d{xp, yp, zp}, Vector3d{xp * facShortLabel, yp * facShortLabel, zp}, tick.Label, textColor)
		}
		t := Vector3d{-100, -100, 100 * facText}
		rot.DrawText(Vector3d{t.X, t.Y, t.Z}, Vector3d{t.X * facLongLabel, t.Y * facLongLabel, t.Z}, checkEmpty(p.Z.Label, "z"), textColor)
	}
	textSize := canvas.Context().TextSize
	ypos := canvas.Rect().Max.Y - textSize
	for _, c := range p.Contents {
		err := c.DrawTo(p, cube)
		if err != nil {
			return err
		}

		leg := c.Legend()
		if leg.Name != "" {
			canvas.DrawText(Point{X: textSize * 2, Y: ypos}, leg.Name, Left|VCenter, textColor, textSize)
			canvas.DrawPath(PointsFromSlice(Point{X: 0, Y: ypos}, Point{X: textSize * 2, Y: ypos}), leg.ShapeLineStyle.LineStyle)
			ypos -= textSize * 1.5
		}

	}

	return canvasCube.DrawObjects()
}

// DrawLongLine draws a long line from p1 to p2 by splitting it into n segments
// to avoid issues with very long lines in 3D rendering.
func DrawLongLine(rot Cube, p1 Vector3d, p2 Vector3d, style *Style) {
	const n = 10
	d := p2.Sub(p1).Mul(1 / float64(n))
	for i := 0; i < n; i++ {
		p2 := p1.Add(d)
		rot.DrawLine(p1, p2, style)
		p1 = p2
	}
}

func checkEmpty(str string, def string) string {
	if str == "" {
		return def
	}
	return str
}

func (p *Plot3d) AddContent(value Plot3dContent) {
	p.Contents = append(p.Contents, value)
}

func (p *Plot3d) SetAngle(alpha float64, beta float64, gamma float64) {
	p.alpha = alpha
	p.beta = beta
	p.gamma = gamma
}

type SecondaryStyle interface {
	SetSecondaryStyle(s *Style) Plot3dContent
}

type Graph3d struct {
	Func      func(x, y float64) (Vector3d, error)
	U         Bounds
	V         Bounds
	Style     *Style
	Steps     int
	StepsHigh int
	Title     string
}

func (g *Graph3d) SetUBounds(bounds Bounds) Plot3dContent {
	g.U = bounds
	return g
}

func (g *Graph3d) SetVBounds(bounds Bounds) Plot3dContent {
	g.V = bounds
	return g
}

func (g *Graph3d) SetStyle(s *Style) Plot3dContent {
	g.Style = s
	return g
}

func (g *Graph3d) SetTitle(s string) Plot3dContent {
	g.Title = s
	return g
}

func (g *Graph3d) DrawTo(_ *Plot3d, cube Cube) error {
	steps := g.Steps
	if steps <= 0 {
		steps = 31
	}
	stepsHigh := g.StepsHigh
	if stepsHigh <= 0 {
		stepsHigh = steps * 3
	}
	style := g.Style
	if style == nil {
		style = Black.SetStrokeWidth(0.5)
	}

	x, y, z := cube.Bounds()

	uB := g.U
	if !uB.isSet {
		uB = x
	}
	vB := g.V
	if !vB.isSet {
		vB = y
	}

	for xn := 0; xn <= steps; xn++ {
		uv := uB.Min + float64(xn)*uB.Width()/float64(steps)
		err := cube.DrawPath(LinePath3d{
			Func: func(vv float64) (Vector3d, error) {
				v, err := g.Func(uv, vv)
				return Vector3d{X: x.Bind(v.X), Y: y.Bind(v.Y), Z: z.Bind(v.Z)}, err
			},
			Bounds: vB,
			Steps:  stepsHigh,
		}, style)
		if err != nil {
			return err
		}
	}
	for yn := 0; yn <= steps; yn++ {
		vv := vB.Min + float64(yn)*vB.Width()/float64(steps)
		err := cube.DrawPath(LinePath3d{
			Func: func(uv float64) (Vector3d, error) {
				v, err := g.Func(uv, vv)
				return Vector3d{X: x.Bind(v.X), Y: y.Bind(v.Y), Z: z.Bind(v.Z)}, err
			},
			Bounds: uB,
			Steps:  stepsHigh,
		}, style)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *Graph3d) Legend() Legend {
	return Legend{Name: g.Title, ShapeLineStyle: ShapeLineStyle{LineStyle: g.Style}}
}

type Solid3d struct {
	Func   func(x, y float64) (Vector3d, error)
	U      Bounds
	V      Bounds
	Style1 *Style
	Style2 *Style
	USteps int
	VSteps int
	Title  string
}

var _ SecondaryStyle = (*Solid3d)(nil)

func (g *Solid3d) SetStyle(s *Style) Plot3dContent {
	g.Style1 = s
	return g
}

func (g *Solid3d) SetTitle(s string) Plot3dContent {
	g.Title = s
	return g
}

func (g *Solid3d) SetUBounds(bounds Bounds) Plot3dContent {
	g.U = bounds
	return g
}

func (g *Solid3d) SetVBounds(bounds Bounds) Plot3dContent {
	g.V = bounds
	return g
}

func (g *Solid3d) SetSecondaryStyle(s *Style) Plot3dContent {
	g.Style2 = s
	return g
}

func (g *Solid3d) DrawTo(_ *Plot3d, cube Cube) (err error) {
	defer nErr.CatchErr(&err)

	uSteps := g.USteps
	if uSteps <= 0 {
		uSteps = 31
	}
	vSteps := g.VSteps
	if vSteps <= 0 {
		vSteps = uSteps
	}

	style1 := g.Style1
	style2 := g.Style2
	if style1 == nil {
		style1 = Black.SetStrokeWidth(0.5).SetFill(White)
		style2 = nil
	} else {
		if !style1.Fill {
			style1 = style1.SetFill(White)
		}
	}

	x, y, z := cube.Bounds()

	uB := g.U
	if !uB.isSet {
		uB = x
	}
	vB := g.V
	if !vB.isSet {
		vB = y
	}

	for xn := 0; xn < uSteps; xn++ {
		xv0 := uB.Min + float64(xn)*uB.Width()/float64(uSteps)
		xv1 := uB.Min + float64(xn+1)*uB.Width()/float64(uSteps)
		for yn := 0; yn < vSteps; yn++ {
			yv0 := vB.Min + float64(yn)*vB.Width()/float64(vSteps)
			yv1 := vB.Min + float64(yn+1)*vB.Width()/float64(vSteps)

			v00 := nErr.TryArg(g.Func(xv0, yv0))
			v01 := nErr.TryArg(g.Func(xv0, yv1))
			v10 := nErr.TryArg(g.Func(xv1, yv0))
			v11 := nErr.TryArg(g.Func(xv1, yv1))

			nErr.Try(cube.DrawTriangle(
				Vector3d{X: x.Bind(v00.X), Y: y.Bind(v00.Y), Z: z.Bind(v00.Z)},
				Vector3d{X: x.Bind(v10.X), Y: y.Bind(v10.Y), Z: z.Bind(v10.Z)},
				Vector3d{X: x.Bind(v11.X), Y: y.Bind(v11.Y), Z: z.Bind(v11.Z)},
				style1, style2))
			nErr.Try(cube.DrawTriangle(
				Vector3d{X: x.Bind(v00.X), Y: y.Bind(v00.Y), Z: z.Bind(v00.Z)},
				Vector3d{X: x.Bind(v11.X), Y: y.Bind(v11.Y), Z: z.Bind(v11.Z)},
				Vector3d{X: x.Bind(v01.X), Y: y.Bind(v01.Y), Z: z.Bind(v01.Z)},
				style1, style2))
		}
	}
	return nil
}

func (g *Solid3d) Legend() Legend {
	return Legend{Name: g.Title, ShapeLineStyle: ShapeLineStyle{LineStyle: g.Style1}}
}

type Line3d struct {
	Func  func(u float64) (Vector3d, error)
	U     Bounds
	Style *Style
	Steps int
	Title string
}

func (g *Line3d) SetUBounds(bounds Bounds) Plot3dContent {
	g.U = bounds
	return g
}

func (g *Line3d) SetStyle(s *Style) Plot3dContent {
	g.Style = s
	return g
}
func (g *Line3d) SetTitle(s string) Plot3dContent {
	g.Title = s
	return g
}

func (g *Line3d) DrawTo(_ *Plot3d, cube Cube) error {
	steps := g.Steps
	if steps <= 0 {
		steps = 200
	}
	style := g.Style
	if style == nil {
		style = Black.SetStrokeWidth(0.5)
	}

	x, y, z := cube.Bounds()

	uB := g.U
	if !uB.isSet {
		uB = NewBounds(0, 1)
	}

	err := cube.DrawPath(LinePath3d{
		Func: func(vv float64) (Vector3d, error) {
			v, err := g.Func(vv)
			return Vector3d{X: x.Bind(v.X), Y: y.Bind(v.Y), Z: z.Bind(v.Z)}, err
		},
		Bounds: uB,
		Steps:  steps,
	}, style)

	if err != nil {
		return err
	}
	return nil
}

func (g *Line3d) Legend() Legend {
	return Legend{Name: g.Title, ShapeLineStyle: ShapeLineStyle{LineStyle: g.Style}}
}

type Arrow3d struct {
	From, To Vector3d
	Plane    Vector3d
	Style    *Style
	Label    string
	Mode     int
}

func (a Arrow3d) DrawTo(_ *Plot3d, cube Cube) error {
	cube.DrawLine(a.From, a.To, a.Style)

	const len = 0.2
	dist := a.To.Sub(a.From)

	d := dist.Normalize()
	plane := a.Plane
	if plane.Zero() {
		// if no plane is given, make the two reverse tips of the arrow head
		// having the same z-value
		if d.X == 0 {
			plane = Vector3d{0, 1, 0}
		} else {
			plane = Vector3d{-d.Y / d.X, 1, 0}.Normalize()
		}
	} else {
		// If a plane is given, the given plane is the normal vector of the plane
		// created by the tips of the arrow head and the two reverse tips.
		plane = d.Cross(plane).Normalize()
	}

	if dist.Abs() > len {
		d := d.Mul(len)
		plane := plane.Mul(len / 3)
		if a.Mode&1 != 0 {
			cube.DrawLine(a.To, a.To.Sub(d).Add(plane), a.Style)
			cube.DrawLine(a.To, a.To.Sub(d).Sub(plane), a.Style)
		}
		if a.Mode&2 != 0 {
			cube.DrawLine(a.From, a.From.Add(d).Add(plane), a.Style)
			cube.DrawLine(a.From, a.From.Add(d).Sub(plane), a.Style)
		}
	}
	if a.Label != "" {
		t1 := dist.Cross(plane).Normalize().Mul(len / 3)
		p1 := a.To.Add(a.From).Mul(0.5)
		cube.DrawText(p1, p1.Add(t1), a.Label, a.Style.Text())
	}

	return nil
}

func (a Arrow3d) Legend() Legend {
	return Legend{Name: "", ShapeLineStyle: ShapeLineStyle{LineStyle: a.Style}}
}

func (a Arrow3d) SetStyle(s *Style) Plot3dContent {
	a.Style = s
	return a
}
