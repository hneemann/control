package polynomial

import (
	"fmt"
	"github.com/hneemann/control/graph"
	"github.com/stretchr/testify/assert"
	"math"
	"os"
	"path/filepath"
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
			"(x+2)*(x-1)/((x+1)*(x-2))"},
		{"reduce",
			NewRoots(complex(-2, 0), complex(1, 0)),
			NewRoots(complex(-1, 0), complex(1, 0)),
			"(x+2)/((x+1))"},
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

	k := PID(12, 1.5, 1)

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

	pl, err := g0.CreateEvans(15)
	pl.XBounds = graph.NewBounds(-4, 0.1)
	assert.NoError(t, err)
	if pl != nil {
		err = exportPlot(pl, "wok1.svg")
		assert.NoError(t, err)
	}
}

func Test_Evans2(t *testing.T) {
	n := NewRoots(complex(0.9, 0), complex(-0.5, 0))
	d := NewRoots(complex(1, 0), complex(2, 0))
	g0 := FromRoots(n, d)

	pl, err := g0.CreateEvans(25)
	fmt.Println(pl)
	assert.NoError(t, err)
	if pl != nil {
		pl.XBounds = graph.NewBounds(-1, 3)
		pl.YBounds = graph.NewBounds(-1.5, 1.5)

		err = exportPlot(pl, "wok2.svg")
		assert.NoError(t, err)
	}
}

func Test_Evans3(t *testing.T) {
	n := NewRoots()
	d := NewRoots(complex(1, 0), complex(2, 0))
	g := FromRoots(n, d)

	pid := PID(1, 0.7, 0.45)

	g0 := g.Mul(pid)

	pl, err := g0.CreateEvans(100)
	assert.NoError(t, err)
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

	pid := PID(1, 1.5, 1)

	g0 := g.Mul(pid)

	pl, err := g0.CreateEvans(10)
	assert.NoError(t, err)
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

	pl, err := g0.CreateEvans(10)
	assert.NoError(t, err)
	if pl != nil {
		pl.XBounds = graph.NewBounds(-2, 0.5)
		pl.YBounds = graph.NewBounds(-2, 2)

		err = exportPlot(pl, "wok5.svg")
		assert.NoError(t, err)
	}
}

func Test_Evans6(t *testing.T) {
	n := NewRoots(complex(-3, 0), complex(-4, 0.3))
	d := NewRoots(complex(-2, 0), complex(-1, 0), complex(-4, 0))
	g0 := FromRoots(n, d)

	pl, err := g0.CreateEvans(20)
	assert.NoError(t, err)
	if pl != nil {
		err = exportPlot(pl, "wok6.svg")
		assert.NoError(t, err)
	}
}

func exportPlot(pl graph.Image, name string) error {
	f, _ := os.Create(filepath.Join(testFolder, name))
	defer f.Close()
	c := graph.NewSVG(800, 600, 15, f)
	pl.DrawTo(c)
	return c.Close()
}

func TestLinear_EvansSplitPoints(t *testing.T) {
	tests := []struct {
		name string
		lin  Linear
		want []float64
	}{
		{"simple", Linear{Numerator: Polynomial{3, 1}, Denominator: Polynomial{2, 3, 1}}, []float64{math.Sqrt(2) - 3, -math.Sqrt(2) - 3}},
		{"cplx", Linear{Numerator: Polynomial{3, 1}, Denominator: Polynomial{2, 2, 1}}, []float64{-math.Sqrt(5) - 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.lin.EvansSplitPoints()
			assert.NoError(t, err)
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

	k := PID(10, 2, 1)

	pl := NewBode(0.01, 100)
	g.AddToBode(pl, graph.Green)
	k.AddToBode(pl, graph.Blue)
	k.Mul(g).AddToBode(pl, graph.Black)

	err := exportPlot(pl, "bode1.svg")
	assert.NoError(t, err)
}

func Test_Nyquist1(t *testing.T) {
	n := Must(NewRoots().Real(6, 6))
	d := Must(Must(NewRoots().Real(1, -1)).Real(1, -3))
	g := FromRoots(n, d)

	err := exportPlot(g.Nyquist(), "nyquist1.svg")
	assert.NoError(t, err)
}

func Test_Nyquist2(t *testing.T) {
	n := Must(NewRoots().Real(1, 0.2)).MulFloat(70)
	d := Must(Must(Must(NewRoots().Complex(1, 2, 10)).Real(1, 4)).Complex(1, 0.2, 0.1))
	g := FromRoots(n, d)

	err := exportPlot(g.Nyquist(), "nyquist2.svg")
	assert.NoError(t, err)
}
