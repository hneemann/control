package polynomial

import (
	"fmt"
	"github.com/hneemann/control/graph"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"github.com/hneemann/parser2/value/export/xmlWriter"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLinear_Mul(t *testing.T) {
	l1 := &Linear{
		Numerator:   Polynomial{4, 5},
		Denominator: Polynomial{1, 2, 3},
	}
	l2 := &Linear{
		Numerator:   Polynomial{5},
		Denominator: Polynomial{-2, 1, 4},
	}
	// externally checked
	expected := &Linear{
		Numerator:   Polynomial{20, 25},
		Denominator: Polynomial{-2, -3, 0, 11, 12},
	}

	testOperation(t, l1, l2, (*Linear).Mul, expected)
}

func TestLinear_Div(t *testing.T) {
	l1 := &Linear{
		Numerator:   Polynomial{4, 5},
		Denominator: Polynomial{1, 2, 3},
	}
	l2 := &Linear{
		Numerator:   Polynomial{5},
		Denominator: Polynomial{-2, 1, 4},
	}

	// externally checked
	expected := &Linear{
		Numerator:   Polynomial{-8, -6, 21, 20},
		Denominator: Polynomial{5, 10, 15},
	}

	testOperation(t, l1, l2, (*Linear).Div, expected)
}

func TestLinear_Add(t *testing.T) {
	l1 := &Linear{
		Numerator:   Polynomial{4, 2},
		Denominator: Polynomial{4, 4, 2},
	}
	l2 := &Linear{
		Numerator:   Polynomial{5},
		Denominator: Polynomial{-4, 2, 2},
	}
	// externally checked
	expected := &Linear{
		Numerator:   Polynomial{4, 20, 22, 4},
		Denominator: Polynomial{-16, -8, 8, 12, 4},
	}

	testOperation(t, l1, l2, func(a, b *Linear) *Linear {
		r, err := a.Add(b)
		assert.NoError(t, err)
		return r
	}, expected)
}

func testOperation(t *testing.T, a, b *Linear, op func(a, b *Linear) *Linear, expected *Linear) {
	a1, a2, a3, a4 := addRoots(t, a)
	b1, b2, b3, b4 := addRoots(t, b)

	assert.True(t, expected.Equals(op(a1, b1)))
	assert.True(t, expected.Equals(op(a1, b2)))
	assert.True(t, expected.Equals(op(a1, b3)))
	assert.True(t, expected.Equals(op(a1, b4)))

	assert.True(t, expected.Equals(op(a2, b1)))
	assert.True(t, expected.Equals(op(a2, b2)))
	assert.True(t, expected.Equals(op(a2, b3)))
	assert.True(t, expected.Equals(op(a2, b4)))

	assert.True(t, expected.Equals(op(a3, b1)))
	assert.True(t, expected.Equals(op(a3, b2)))
	assert.True(t, expected.Equals(op(a3, b3)))
	assert.True(t, expected.Equals(op(a3, b4)))

	assert.True(t, expected.Equals(op(a4, b1)))
	assert.True(t, expected.Equals(op(a4, b2)))
	assert.True(t, expected.Equals(op(a4, b3)))
	assert.True(t, expected.Equals(op(a4, b4)))
}

func addRoots(t *testing.T, a *Linear) (*Linear, *Linear, *Linear, *Linear) {
	nr, err := a.Zeros()
	assert.Nil(t, err)
	dr, err := a.Poles()
	assert.Nil(t, err)
	return &Linear{Numerator: a.Numerator, Denominator: a.Denominator},
		&Linear{Numerator: a.Numerator, zeros: nr, Denominator: a.Denominator},
		&Linear{Numerator: a.Numerator, Denominator: a.Denominator, poles: dr},
		&Linear{Numerator: a.Numerator, zeros: nr, Denominator: a.Denominator, poles: dr}
}

func TestLinear_Deriv(t *testing.T) {
	lin := &Linear{
		Numerator:   Polynomial{1, 3},
		Denominator: Polynomial{1, 2, 3},
	}

	// externally checked
	expected := &Linear{
		Numerator:   Polynomial{1, -6, -9},
		Denominator: Polynomial{1, 4, 10, 12, 9},
	}

	derivative := lin.Derivative()
	assert.True(t, expected.Equals(derivative))
}

func TestFromRoots(t *testing.T) {
	tests := []struct {
		name        string
		numerator   Roots
		denominator Roots
		want        string
	}{
		{"simple",
			NewRoots(complex(-2, 0), complex(1, 0)),
			NewRoots(complex(-1, 0), complex(2, 0)),
			"(s+2)*(s-1)/((s+1)*(s-2))"},
		{"reduce",
			NewRoots(complex(-2, 0), complex(1, 0)),
			NewRoots(complex(-1, 0), complex(1, 0)),
			"(s+2)/((s+1))"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.EqualValues(t, tt.want, FromRoots(tt.numerator, tt.denominator).String())
		})
	}
}

func TestLinear_Loop(t *testing.T) {
	g0 := &Linear{
		Numerator:   Polynomial{4, 2},
		Denominator: Polynomial{4, 4, 2},
	}
	fmt.Println(g0.StringToParse())

	expected := &Linear{
		Numerator:   Polynomial{2, 1},
		Denominator: Polynomial{4, 3, 1},
	}

	testFunc(t, g0, func(a *Linear) *Linear {
		return Must(a.Loop())
	}, expected)
}

func testFunc(t *testing.T, a *Linear, op func(a *Linear) *Linear, expected *Linear) {
	a1, a2, a3, a4 := addRoots(t, a)

	assert.True(t, expected.Equals(op(a1)))
	assert.True(t, expected.Equals(op(a2)))
	assert.True(t, expected.Equals(op(a3)))
	assert.True(t, expected.Equals(op(a4)))
}

func Must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

func Test_Integration(t *testing.T) {
	n := Must(NewRoots().Real(1.5, 1))
	d := Must(Must(Must(NewRoots().Real(2, 1)).Real(1, 1)).Complex(1, 3, 3.1))
	g := FromRoots(n, d)

	k, err := PID(12, 1.5, 1, 0)
	assert.NoError(t, err)

	g0 := g.Mul(k)
	gw, err := g0.Loop()
	assert.NoError(t, err)

	p, err := gw.Poles()
	assert.NoError(t, err)

	// externally checked
	expected := NewRoots(complex(-1.4838920018993484, 3.04283839228145), -0.6814635644285129+0i, complex(-0.42537621588638014, 0.5755095234855198))
	assert.True(t, expected.Equals(p))
}

const testFolder = "/home/hneemann/temp/control"

func Test_Evans1(t *testing.T) {
	n := NewRoots(complex(-3, 0), complex(-4, 0))
	d := NewRoots(complex(-2, 0), complex(-1, 0))
	g0 := FromRoots(n, d)

	pl := &graph.Plot{Content: Must(g0.CreateEvans(15))}
	pl.XBounds = graph.NewBounds(-4, 0.1)
	if pl != nil {
		err := exportPlot(pl, "wok1.svg")
		assert.NoError(t, err)
	}
}

func Test_Evans2(t *testing.T) {
	n := NewRoots(complex(0.9, 0), complex(-0.5, 0))
	d := NewRoots(complex(1, 0), complex(2, 0))
	g0 := FromRoots(n, d)

	pl := &graph.Plot{Content: Must(g0.CreateEvans(25))}
	fmt.Println(pl)
	if pl != nil {
		pl.XBounds = graph.NewBounds(-1, 3)
		pl.YBounds = graph.NewBounds(-1.5, 1.5)

		err := exportPlot(pl, "wok2.svg")
		assert.NoError(t, err)
	}
}

func Test_Evans3(t *testing.T) {
	n := NewRoots()
	d := NewRoots(complex(1, 0), complex(2, 0))
	g := FromRoots(n, d)

	pid, err := PID(1, 0.7, 0.45, 0)
	assert.NoError(t, err)

	g0 := g.Mul(pid)

	pl := &graph.Plot{Content: Must(g0.CreateEvans(100))}
	if pl != nil {
		pl.XBounds = graph.NewBounds(-6, 3)
		pl.YBounds = graph.NewBounds(-4, 4)

		err = exportPlot(pl, "wok3.svg")
		assert.NoError(t, err)
	}
}

func Test_Evans4(t *testing.T) {
	n := Must(NewRoots().Real(1.5, 1))
	d := Must(Must(Must(NewRoots().Real(2, 1)).Real(1, 1)).Complex(1, 3, 3.1))
	g := FromRoots(n, d)

	pid, err := PID(1, 1.5, 1, 0)
	assert.NoError(t, err)

	g0 := g.Mul(pid)

	pl := &graph.Plot{Content: Must(g0.CreateEvans(10))}
	if pl != nil {
		pl.XBounds = graph.NewBounds(-2, 0.5)
		pl.YBounds = graph.NewBounds(-3, 3)

		err = exportPlot(pl, "wok4.svg")
		assert.NoError(t, err)
	}
}

func Test_Evans5(t *testing.T) {
	n := Must(NewRoots().Real(1.5, 1))
	d := Must(Must(Must(NewRoots().Real(2, 1)).Real(1, 1)).Complex(1, 3, 3.1))
	g0 := FromRoots(n, d)

	pl := &graph.Plot{Content: Must(g0.CreateEvans(10))}
	if pl != nil {
		pl.XBounds = graph.NewBounds(-2, 0.5)
		pl.YBounds = graph.NewBounds(-2, 2)

		err := exportPlot(pl, "wok5.svg")
		assert.NoError(t, err)
	}
}

func Test_Evans6(t *testing.T) {
	n := NewRoots(complex(-1, 0), complex(-2, 0.3))
	d := NewRoots(complex(0, 0), complex(1, 0), complex(-2, 0))
	g0 := FromRoots(n, d)

	pl := &graph.Plot{Content: Must(g0.CreateEvans(50))}
	if pl != nil {
		err := exportPlot(pl, "wok6.svg")
		assert.NoError(t, err)
	}
}

func Test_Evans7(t *testing.T) {
	n := NewRoots()
	d := NewRoots(complex(-1, 1))
	g0 := FromRoots(n, d)

	pl := &graph.Plot{Content: Must(g0.CreateEvans(5))}
	pl.XBounds = graph.NewBounds(-2, 0.2)
	if pl != nil {
		err := exportPlot(pl, "wok7.svg")
		assert.NoError(t, err)
	}
}

func exportPlot(pl graph.Image, name string) error {
	w := xmlWriter.New()
	c := graph.NewSVG(&graph.DefaultContext, w)
	err := pl.DrawTo(c)
	if err != nil {
		return err
	}
	err = c.Close()

	//f, _ := os.Create(filepath.Join(testFolder, name))
	//defer f.Close()
	//_, err := f.Write(w.Bytes())
	return err
}

func TestLinear_EvansSplitPoints(t *testing.T) {
	tests := []struct {
		name string
		lin  Linear
		want []float64
	}{
		{"simple", Linear{Numerator: Polynomial{3, 1}, Denominator: Polynomial{2, 3, 1}}, []float64{0.17157287525380963, 5.82842712474619}},
		{"cplx", Linear{Numerator: Polynomial{3, 1}, Denominator: Polynomial{2, 2, 1}}, []float64{8.47213595499958}},
		// (7s²+7s+3.5)/(s⁴+2s³-s²-2s)
		{"cplx2", Linear{Numerator: Polynomial{3.5, 7, 7}, Denominator: Polynomial{0, -2, -1, 2, 1}}, []float64{0.10913314607145859, 0.7480097110713985}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.lin.EvansSplitGains()
			assert.NoError(t, err)
			fmt.Printf("got: %v\n", got)
			assert.Equal(t, len(tt.want), len(got))
			for i, w := range tt.want {
				assert.InDelta(t, w, got[i], 1e-6)
			}
		})
	}
}

func Test_Bode1(t *testing.T) {
	n := Must(NewRoots().Real(1.5, 1))
	d := Must(Must(Must(NewRoots().Real(2, 1)).Real(1, 1)).Complex(1, 3, 3.1))
	g := FromRoots(n, d)

	k, err := PID(10, 2, 1, 0)
	assert.NoError(t, err)

	pl := NewBode(0.01, 100)
	pl.Add(g.CreateBode(graph.Green, "G"))
	pl.Add(k.CreateBode(graph.Blue, "K"))
	pl.Add(k.Mul(g).CreateBode(graph.Black, "G#0"))

	err = exportPlot(pl, "bode1.svg")
	assert.NoError(t, err)
}

func Test_Nyquist1(t *testing.T) {
	n := Must(NewRoots().Real(6, 6))
	d := Must(Must(NewRoots().Real(1, -1)).Real(1, -3))
	g := FromRoots(n, d)

	pl := &graph.Plot{Content: Must(g.Nyquist(1000, true))}
	err := exportPlot(pl, "nyquist1.svg")
	assert.NoError(t, err)
}

func Test_Nyquist2(t *testing.T) {
	n := Must(NewRoots().Real(1, 0.2)).MulFloat(70)
	d := Must(Must(Must(NewRoots().Complex(1, 2, 10)).Real(1, 4)).Complex(1, 0.2, 0.1))
	g := FromRoots(n, d)

	pl := &graph.Plot{Content: Must(g.Nyquist(1000, true))}
	err := exportPlot(pl, "nyquist2.svg")
	assert.NoError(t, err)
}

func Test_Nyquist3(t *testing.T) {
	n := NewRoots().MulFloat(8)
	d := Must(Must(Must(Must(Must(NewRoots().Real(1, 1)).Real(1, 1)).Real(1, 1)).Real(1, 1)).Real(1, 1))
	g := FromRoots(n, d)

	pl := &graph.Plot{Content: Must(g.Nyquist(1000, true))}
	//pl.BoundsModifier = graph.Zoom(graph.Point{}, 100)

	err := exportPlot(pl, "nyquist3.svg")
	assert.NoError(t, err)
}

func TestLinear_GetStateSpace_PT1(t *testing.T) {
	type fields struct {
	}
	tests := []struct {
		name  string
		kp, T float64
	}{
		{"11", 1, 1},
		{"21", 2, 1},
		{"12", 1, 2},
		{"22", 2, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lin := &Linear{
				Numerator:   Polynomial{tt.kp},
				Denominator: Polynomial{1, tt.T},
			}
			a, c, d, err := lin.GetStateSpaceRepresentation()
			assert.NoError(t, err)
			assert.EqualValues(t, Matrix{Vector{-1 / tt.T}}, a)
			assert.EqualValues(t, Vector{tt.kp / tt.T}, c)
			assert.EqualValues(t, 0, d)
		})
	}
}

func TestLinear_GetStateSpace_PT2(t *testing.T) {
	type fields struct {
	}
	tests := []struct {
		name     string
		kp, T, d float64
	}{
		{"111", 1, 1, 1},
		{"211", 2, 1, 1},
		{"121", 1, 2, 1},
		{"221", 2, 2, 1},
		{"112", 1, 1, 2},
		{"212", 2, 1, 2},
		{"122", 1, 2, 2},
		{"222", 2, 2, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lin := &Linear{
				Numerator:   Polynomial{tt.kp},
				Denominator: Polynomial{1, 2 * tt.d * tt.T, tt.T * tt.T},
			}
			a, c, d, err := lin.GetStateSpaceRepresentation()
			assert.NoError(t, err)
			assert.EqualValues(t, Matrix{Vector{0, 1}, Vector{-1 / (tt.T * tt.T), -2 * tt.d / tt.T}}, a)
			assert.EqualValues(t, Vector{tt.kp / (tt.T * tt.T), 0}, c)
			assert.EqualValues(t, 0, d)
		})
	}
}

func TestLinear_GetStateSpace_PHase(t *testing.T) {
	type fields struct {
	}
	tests := []struct {
		name   string
		T1, T2 float64
	}{
		{"11", 1, 1},
		{"21", 2, 1},
		{"12", 1, 2},
		{"22", 2, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lin := &Linear{
				Numerator:   Polynomial{1, tt.T1},
				Denominator: Polynomial{1, tt.T2},
			}
			a, c, d, err := lin.GetStateSpaceRepresentation()
			assert.NoError(t, err)
			assert.EqualValues(t, Matrix{Vector{-1 / tt.T2}}, a)
			assert.EqualValues(t, Vector{1/tt.T2 - tt.T1/(tt.T2*tt.T2)}, c)
			assert.EqualValues(t, tt.T1/tt.T2, d)
		})
	}
}

func Benchmark_DataSet(b *testing.B) {
	d := newDataSet(1000, 2)
	for i := 0; i < 1000; i++ {
		d.set(i, 0, float64(i))
		d.set(i, 1, float64(i))
	}
	st := funcGen.NewEmptyStack[value.Value]()

	ss := 0.0
	b.ResetTimer()
	for range b.N {
		sum := 0.0
		d.toPointList(0, 1).Iterate(st, func(v value.Value) error {
			if l, ok := v.ToList(); ok {
				p, _ := l.ToSlice(st)
				if x, ok := p[0].ToFloat(); ok {
					if y, ok := p[1].ToFloat(); ok {
						sum += x + y
					}
				}
			}
			return nil
		})
		ss += sum
	}
	fmt.Println(ss)
}
