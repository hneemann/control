package graph

import (
	"bytes"
	"fmt"
	"github.com/hneemann/control/nErr"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"math"
)

var Vector3dType value.Type

type Vector3d struct {
	X, Y, Z float64
}

type Vectors func(func(Vector3d, error) bool)

func (v Vector3d) ToList() (*value.List, bool) {
	return nil, false
}

var vecMapFac = value.NewFuncMapFactory(func(v Vector3d, key string) (value.Value, bool) {
	switch key {
	case "x":
		return value.Float(v.X), true
	case "y":
		return value.Float(v.Y), true
	case "z":
		return value.Float(v.Z), true
	default:
		return nil, false
	}
}, "x", "y", "z")

func (v Vector3d) ToMap() (value.Map, bool) {
	return vecMapFac.Create(v), true
}

func (v Vector3d) ToInt() (int, bool) {
	return 0, false
}

func (v Vector3d) ToFloat() (float64, bool) {
	return 0, false
}

func (v Vector3d) String() string {
	return fmt.Sprintf("(%g,%g,%g)", v.X, v.Y, v.Z)
}

func (v Vector3d) ToString(_ funcGen.Stack[value.Value]) (string, error) {
	return v.String(), nil
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

func (v Vector3d) Div(f float64) Vector3d {
	return Vector3d{
		X: v.X / f,
		Y: v.Y / f,
		Z: v.Z / f,
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

func (v Vector3d) Neg() Vector3d {
	return Vector3d{-v.X, -v.Y, -v.Z}
}

func (v Vector3d) RotX(alpha float64) Vector3d {
	cos := math.Cos(alpha)
	sin := math.Sin(alpha)
	return Vector3d{X: v.X, Y: v.Y*cos - v.Z*sin, Z: v.Y*sin + v.Z*cos}
}

func (v Vector3d) RotY(beta float64) Vector3d {
	cos := math.Cos(beta)
	sin := math.Sin(beta)
	return Vector3d{X: v.X*cos + v.Z*sin, Y: v.Y, Z: -v.X*sin + v.Z*cos}
}

func (v Vector3d) RotZ(gamma float64) Vector3d {
	cos := math.Cos(gamma)
	sin := math.Sin(gamma)
	return Vector3d{X: v.X*cos - v.Y*sin, Y: v.X*sin + v.Y*cos, Z: v.Z}
}

func (v Vector3d) ToPoint() Point {
	return Point{
		X: v.X,
		Y: v.Y,
	}
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

type Plot3dContent interface {
	DrawTo(*Plot3d, Cube) error
	Legend() Legend
	SetStyle(s *Style) Plot3dContent
	RequiresHiddenLineRemoval() bool
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

type IsCloseable3d interface {
	Close() Plot3dContent
}

type HasShape3d interface {
	SetShape(Shape, *Style) Plot3dContent
}

type hlrMode int

const (
	hlrAuto hlrMode = iota
	hlrOn
	hlrOff
)

type Plot3d struct {
	X, Y, Z     AxisDescription
	Contents    []Plot3dContent
	alpha       float64
	beta        float64
	gamma       float64
	HideCube    bool
	Size        float64
	Perspective float64
	hlrMode     hlrMode
}

const (
	DefAlpha = 0.177
	DefBeta  = 0.4
	pFac     = 800
)

func NewPlot3d() *Plot3d {
	return &Plot3d{
		alpha:       DefAlpha,
		beta:        DefBeta,
		gamma:       0,
		Size:        1,
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
	ax := x.GetFactory()(-100, 100, x.Bounds, ctw, 0.02)
	ay := y.GetFactory()(-100, 100, y.Bounds, ctw, 0.02)
	az := z.GetFactory()(-100, 100, z.Bounds, ctw, 0.02)
	return &unityCube{
		parent: parent,
		X:      x.Bounds, Y: y.Bounds, Z: z.Bounds,
		ax: ax, ay: ay, az: az,
	}
}

func (uc *unityCube) transform(p Vector3d) Vector3d {
	return Vector3d{
		X: uc.ax.Trans(p.X),
		Y: uc.ay.Trans(p.Y),
		Z: uc.az.Trans(p.Z),
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

func (uc *unityCube) DrawLine(p1, p2 Vector3d, lineStyle *Style) error {
	return uc.parent.DrawLine(uc.transform(p1), uc.transform(p2), lineStyle)
}

func (uc *unityCube) DrawShape(p Vector3d, shape Shape, style *Style) error {
	return uc.parent.DrawShape(uc.transform(p), shape, style)
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

func (r RotCube) DrawLine(p1, p2 Vector3d, lineStyle *Style) error {
	return r.parent.DrawLine(r.matrix.MulPoint(p1), r.matrix.MulPoint(p2), lineStyle)
}
func (r RotCube) DrawShape(p Vector3d, shape Shape, style *Style) error {
	return r.parent.DrawShape(r.matrix.MulPoint(p), shape, style)
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

func (p *Plot3d) String() string {
	bu := bytes.Buffer{}
	bu.WriteString("Plot3d: ")
	for i, content := range p.Contents {
		if i > 0 {
			bu.WriteString(", ")
		}
		bu.WriteString(fmt.Sprint(content))
	}
	return bu.String()
}

func (p *Plot3d) DrawTo(canvas Canvas) (err error) {
	defer nErr.CatchErr(&err)

	var requiresHiddenLineRemoval bool
	switch p.hlrMode {
	case hlrOn:
		requiresHiddenLineRemoval = true
	case hlrOff:
		requiresHiddenLineRemoval = false
	default:
		requiresHiddenLineRemoval = false
		for _, content := range p.Contents {
			if content.RequiresHiddenLineRemoval() {
				requiresHiddenLineRemoval = true
				break
			}
		}
	}

	canvasCube := newCanvasCube(canvas, p.Size, p.Perspective, requiresHiddenLineRemoval)
	rot := NewRotCube(canvasCube, p.alpha, p.beta, p.gamma)

	cubeColor := Gray
	textColor := cubeColor.Text()
	if !p.HideCube {
		nErr.Try(rot.DrawLine(Vector3d{100, 100, 100}, Vector3d{100, -100, 100}, cubeColor))
		nErr.Try(rot.DrawLine(Vector3d{100, 100, -100}, Vector3d{100, -100, -100}, cubeColor))
		nErr.Try(rot.DrawLine(Vector3d{-100, 100, 100}, Vector3d{-100, -100, 100}, cubeColor))

		nErr.Try(rot.DrawLine(Vector3d{100, 100, 100}, Vector3d{-100, 100, 100}, cubeColor))
		nErr.Try(rot.DrawLine(Vector3d{100, 100, -100}, Vector3d{-100, 100, -100}, cubeColor))
		nErr.Try(rot.DrawLine(Vector3d{100, -100, 100}, Vector3d{-100, -100, 100}, cubeColor))

		nErr.Try(rot.DrawLine(Vector3d{100, 100, -100}, Vector3d{100, 100, 100}, cubeColor))
		nErr.Try(rot.DrawLine(Vector3d{-100, 100, -100}, Vector3d{-100, 100, 100}, cubeColor))
		nErr.Try(rot.DrawLine(Vector3d{100, -100, -100}, Vector3d{100, -100, 100}, cubeColor))
	}
	cube := newUnityCube(rot, p.X, p.Y, p.Z)
	const facShortLabel = 1.02
	const facLongLabel = 1.04
	const facText = 1.1
	if !p.X.HideAxis {
		for _, tick := range cube.ax.Ticks {
			xp := cube.ax.Trans(tick.Position)
			yp := -100.0
			zp := -100.0
			nErr.Try(rot.DrawLine(Vector3d{xp, yp, zp}, Vector3d{xp, yp * facLongLabel, zp}, cubeColor))
			rot.DrawText(Vector3d{xp, yp, zp}, Vector3d{xp, yp * facLongLabel, zp}, tick.Label, textColor)
		}
		t := Vector3d{100 * facText, -100, -100}
		rot.DrawText(Vector3d{t.X, t.Y, t.Z}, Vector3d{t.X, t.Y * facLongLabel, t.Z}, checkEmpty(p.X.Label, "x"), textColor)
		nErr.Try(rot.DrawLine(Vector3d{-100, -100, -100}, Vector3d{100 * facText, -100, -100}, cubeColor))
		nErr.Try(rot.DrawLine(Vector3d{102, -100, -102}, Vector3d{100 * facText, -100, -100}, cubeColor))
		nErr.Try(rot.DrawLine(Vector3d{102, -100, -98}, Vector3d{100 * facText, -100, -100}, cubeColor))
	}
	if !p.Y.HideAxis {
		for _, tick := range cube.ay.Ticks {
			xp := -100.0
			yp := cube.ay.Trans(tick.Position)
			zp := -100.0
			nErr.Try(rot.DrawLine(Vector3d{xp, yp, zp}, Vector3d{xp * facShortLabel, yp, zp}, cubeColor))
			rot.DrawText(Vector3d{xp, yp, zp}, Vector3d{xp * facShortLabel, yp, zp}, tick.Label, textColor)
		}
		t := Vector3d{-100, 100 * facText, -100}
		rot.DrawText(Vector3d{t.X, t.Y, t.Z}, Vector3d{t.X * facLongLabel, t.Y, t.Z}, checkEmpty(p.Y.Label, "y"), textColor)
		nErr.Try(rot.DrawLine(Vector3d{-100, -100, -100}, Vector3d{-100, 100 * facText, -100}, cubeColor))
		nErr.Try(rot.DrawLine(Vector3d{-102, 102, -100}, Vector3d{-100, 100 * facText, -100}, cubeColor))
		nErr.Try(rot.DrawLine(Vector3d{-98, 102, -100}, Vector3d{-100, 100 * facText, -100}, cubeColor))
	}
	if !p.Z.HideAxis {
		for _, tick := range cube.az.Ticks {
			xp := -100.0
			yp := -100.0
			zp := cube.az.Trans(tick.Position)
			nErr.Try(rot.DrawLine(Vector3d{xp, yp, zp}, Vector3d{xp * facShortLabel, yp * facShortLabel, zp}, cubeColor))
			rot.DrawText(Vector3d{xp, yp, zp}, Vector3d{xp * facShortLabel, yp * facShortLabel, zp}, tick.Label, textColor)
		}
		t := Vector3d{-100, -100, 100 * facText}
		rot.DrawText(Vector3d{t.X, t.Y, t.Z}, Vector3d{t.X * facLongLabel, t.Y * facLongLabel, t.Z}, checkEmpty(p.Z.Label, "z"), textColor)
		nErr.Try(rot.DrawLine(Vector3d{-100, -100, -100}, Vector3d{-100, -100, 100 * facText}, cubeColor))
		nErr.Try(rot.DrawLine(Vector3d{-102, -100, 102}, Vector3d{-100, -100, 100 * facText}, cubeColor))
		nErr.Try(rot.DrawLine(Vector3d{-98, -100, 102}, Vector3d{-100, -100, 100 * facText}, cubeColor))
	}

	textSize := canvas.Context().TextSize
	yPos := canvas.Rect().Max.Y - textSize
	for _, c := range p.Contents {
		err := c.DrawTo(p, cube)
		if err != nil {
			return err
		}

		leg := c.Legend()
		if leg.Name != "" {
			canvas.DrawText(Point{X: textSize * 2, Y: yPos}, leg.Name, Left|VCenter, textColor, textSize)
			nErr.Try(canvas.DrawPath(PointsFromSlice(Point{X: 0, Y: yPos}, Point{X: textSize * 2, Y: yPos}), leg.ShapeLineStyle.LineStyle))
			yPos -= textSize * 1.5
		}

	}

	return canvasCube.DrawObjects()
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

func (p *Plot3d) EnableHLR(b bool) *Plot3d {
	np := *p
	if b {
		np.hlrMode = hlrOn
	} else {
		np.hlrMode = hlrOff
	}
	return &np
}

type SecondaryStyle interface {
	SetSecondaryStyle(s *Style) Plot3dContent
}

type Graph3d struct {
	Func   func(x, y float64) (Vector3d, error)
	U      Bounds
	V      Bounds
	Style  *Style
	USteps int
	VSteps int
	Title  string
}

func (g *Graph3d) RequiresHiddenLineRemoval() bool {
	return false
}

func (g *Graph3d) String() string {
	return "Graph3d"
}

func (g *Graph3d) SetUBounds(bounds Bounds) Plot3dContent {
	ng := *g
	ng.U = bounds
	return &ng
}

func (g *Graph3d) SetVBounds(bounds Bounds) Plot3dContent {
	ng := *g
	ng.V = bounds
	return &ng
}

func (g *Graph3d) SetStyle(s *Style) Plot3dContent {
	ng := *g
	ng.Style = s
	return &ng
}

func (g *Graph3d) SetTitle(s string) Plot3dContent {
	ng := *g
	ng.Title = s
	return &ng
}

func (g *Graph3d) DrawTo(_ *Plot3d, cube Cube) error {
	uSteps := g.USteps
	if uSteps <= 0 {
		uSteps = 31
	}
	vSteps := g.VSteps
	if vSteps <= 0 {
		vSteps = uSteps
	}
	style := g.Style
	if style == nil {
		style = Black.SetStrokeWidth(0.5)
	}

	x, y, _ := cube.Bounds()

	uB := g.U
	if !uB.isSet {
		uB = x
	}
	vB := g.V
	if !vB.isSet {
		vB = y
	}

	for xn := 0; xn <= uSteps; xn++ {
		uv := uB.Min + float64(xn)*uB.Width()/float64(uSteps)
		err := cube.DrawPath(LinePath3d{
			Func: func(vv float64) (Vector3d, error) {
				return g.Func(uv, vv)
			},
			Bounds: vB,
			Steps:  vSteps * 3,
		}, style)
		if err != nil {
			return err
		}
	}
	for yn := 0; yn <= vSteps; yn++ {
		vv := vB.Min + float64(yn)*vB.Width()/float64(vSteps)
		err := cube.DrawPath(LinePath3d{
			Func: func(uv float64) (Vector3d, error) {
				return g.Func(uv, vv)
			},
			Bounds: uB,
			Steps:  uSteps * 3,
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
	Func      func(x, y float64) (Vector3d, error)
	U         Bounds
	V         Bounds
	Style1    *Style
	Style2    *Style
	USteps    int
	VSteps    int
	Title     string
	Hexagonal bool
}

var _ SecondaryStyle = (*Solid3d)(nil)

func (g *Solid3d) RequiresHiddenLineRemoval() bool {
	return true
}

func (g *Solid3d) String() string {
	return "Solid3d"
}

func (g *Solid3d) SetStyle(s *Style) Plot3dContent {
	ng := *g
	ng.Style1 = s
	return &ng
}

func (g *Solid3d) SetTitle(s string) Plot3dContent {
	ng := *g
	ng.Title = s
	return &ng
}

func (g *Solid3d) SetUBounds(bounds Bounds) Plot3dContent {
	ng := *g
	ng.U = bounds
	return &ng
}

func (g *Solid3d) SetVBounds(bounds Bounds) Plot3dContent {
	ng := *g
	ng.V = bounds
	return &ng
}

func (g *Solid3d) SetSecondaryStyle(s *Style) Plot3dContent {
	ng := *g
	ng.Style2 = s
	return &ng
}

func (g *Solid3d) DrawTo(_ *Plot3d, cube Cube) (err error) {
	defer nErr.CatchErr(&err)

	uSteps := g.USteps
	if uSteps <= 0 {
		uSteps = 31
	}
	vSteps := g.VSteps
	if vSteps <= 0 {
		if g.Hexagonal {
			vSteps = int(float64(uSteps) / 2 * math.Sqrt(3))
		} else {
			vSteps = uSteps
		}
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

	x, y, _ := cube.Bounds()

	uB := g.U
	if !uB.isSet {
		uB = x
	}
	vB := g.V
	if !vB.isSet {
		vB = y
	}

	matrix := make([][]Vector3d, uSteps+1)
	for u := 0; u <= uSteps; u++ {
		var vOfs float64
		if g.Hexagonal && u&1 != 0 {
			vOfs = 0.5
		}
		matrix[u] = make([]Vector3d, vSteps+1)
		uVal := uB.Min + float64(u)*uB.Width()/float64(uSteps)
		for v := 0; v <= vSteps; v++ {
			vVal := vB.Min + (float64(v)+vOfs)*vB.Width()/float64(vSteps)
			matrix[u][v] = nErr.TryArg(g.Func(uVal, vVal))
		}
	}

	if g.Hexagonal {
		for u := 0; u < uSteps; u++ {
			if u&1 == 0 {
				for v := 0; v < vSteps; v++ {
					nErr.Try(cube.DrawTriangle(matrix[u][v], matrix[u+1][v], matrix[u][v+1], style1, style2))
					nErr.Try(cube.DrawTriangle(matrix[u+1][v], matrix[u+1][v+1], matrix[u][v+1], style1, style2))
				}
			} else {
				for v := 0; v < vSteps; v++ {
					nErr.Try(cube.DrawTriangle(matrix[u][v], matrix[u+1][v], matrix[u+1][v+1], style1, style2))
					nErr.Try(cube.DrawTriangle(matrix[u][v], matrix[u+1][v+1], matrix[u][v+1], style1, style2))
				}
			}
		}
	} else {
		for u := 0; u < uSteps; u++ {
			for v := 0; v < vSteps; v++ {
				nErr.Try(cube.DrawTriangle(matrix[u][v], matrix[u+1][v], matrix[u][v+1], style1, style2))
				nErr.Try(cube.DrawTriangle(matrix[u+1][v], matrix[u+1][v+1], matrix[u][v+1], style1, style2))
			}
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

func (g *Line3d) RequiresHiddenLineRemoval() bool {
	return false
}

func (g *Line3d) String() string {
	return "Line3d"
}

func (g *Line3d) SetUBounds(bounds Bounds) Plot3dContent {
	ng := *g
	ng.U = bounds
	return &ng
}

func (g *Line3d) SetStyle(s *Style) Plot3dContent {
	ng := *g
	ng.Style = s
	return &ng
}
func (g *Line3d) SetTitle(s string) Plot3dContent {
	ng := *g
	ng.Title = s
	return &ng
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

	uB := g.U
	if !uB.isSet {
		uB = NewBounds(0, 1)
	}

	err := cube.DrawPath(LinePath3d{
		Func: func(vv float64) (Vector3d, error) {
			v, err := g.Func(vv)
			return Vector3d{X: v.X, Y: v.Y, Z: v.Z}, err
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
	From, To    Vector3d
	PlaneDefVec Vector3d
	Style       *Style
	Label       string
	Mode        int
}

func (g Arrow3d) RequiresHiddenLineRemoval() bool {
	return false
}

func (a Arrow3d) String() string {
	return fmt.Sprintf("Arrow3d(From=%v, To=%v)", a.From, a.To)
}

func (a Arrow3d) DrawTo(_ *Plot3d, cube Cube) (e error) {
	defer nErr.CatchErr(&e)
	nErr.Try(cube.DrawLine(a.From, a.To, a.Style))

	const headLen = 0.2

	dist := a.To.Sub(a.From)

	dNorm := dist.Normalize()
	inPlaneVec := dNorm.Cross(a.PlaneDefVec).Neg()
	if inPlaneVec.Zero() {
		// if no plane is given, make the two reverse tips of the arrow head
		// having the same z-value
		if dNorm.X == 0 {
			inPlaneVec = Vector3d{1, 0, 0}
		} else {
			inPlaneVec = Vector3d{-dNorm.Y / dNorm.X, 1, 0}.Normalize()
		}
	} else {
		// If a plane is given, the given plane is the normal vector of the plane
		// created by the tips of the arrow head and the two reverse tips.
		inPlaneVec = inPlaneVec.Normalize()
	}

	if dist.Abs() > headLen {
		d := dNorm.Mul(headLen)
		plane := inPlaneVec.Mul(headLen / 4)
		if a.Mode&1 != 0 {
			nErr.Try(cube.DrawLine(a.To, a.To.Sub(d).Add(plane), a.Style))
			nErr.Try(cube.DrawLine(a.To, a.To.Sub(d).Sub(plane), a.Style))
		}
		if a.Mode&2 != 0 {
			nErr.Try(cube.DrawLine(a.From, a.From.Add(d).Add(plane), a.Style))
			nErr.Try(cube.DrawLine(a.From, a.From.Add(d).Sub(plane), a.Style))
		}
	}
	if a.Label != "" {
		t1 := dist.Cross(inPlaneVec).Normalize().Mul(headLen / 3)
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

type ListBasedLine3d struct {
	Vectors        Vectors
	ShapeLineStyle ShapeLineStyle
	Title          string
	closed         bool
}

func (s ListBasedLine3d) RequiresHiddenLineRemoval() bool {
	return false
}

func (s ListBasedLine3d) String() string {
	return "ListBasedLine3d"
}

func (s ListBasedLine3d) Close() Plot3dContent {
	s.closed = true
	return s
}

type vecPath struct {
	Vectors Vectors
	Closed  bool
}

func (v vecPath) Iter(yield func(PathElement3d, error) bool) {
	mode := 'M'
	for vec, err := range v.Vectors {
		if !yield(PathElement3d{Mode: mode, Vector3d: vec}, err) {
			return
		}
		mode = 'L'
	}
}

func (v vecPath) IsClosed() bool {
	return v.Closed
}

func (s ListBasedLine3d) DrawTo(_ *Plot3d, cube Cube) error {
	sls := s.ShapeLineStyle.EnsureSomethingIsVisible()
	if sls.IsLine() {
		err := cube.DrawPath(vecPath{s.Vectors, s.closed}, sls.LineStyle)
		if err != nil {
			return err
		}
	}
	if sls.IsShape() {
		for v, err := range s.Vectors {
			if err != nil {
				return err
			}
			err = cube.DrawShape(v, sls.Shape, sls.ShapeStyle)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s ListBasedLine3d) Legend() Legend {
	return Legend{
		ShapeLineStyle: s.ShapeLineStyle.EnsureSomethingIsVisible(),
		Name:           s.Title,
	}
}

func (s ListBasedLine3d) SetStyle(style *Style) Plot3dContent {
	s.ShapeLineStyle.LineStyle = style
	return s
}

func (s ListBasedLine3d) SetShape(shape Shape, style *Style) Plot3dContent {
	s.ShapeLineStyle.Shape = shape
	s.ShapeLineStyle.ShapeStyle = style
	return s
}
