package polynomial

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math"
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

	testOperation(t, l1, l2, (*Linear).Add, expected)
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
	assert.Nil(t, err)

	p, err := gw.Poles()
	assert.Nil(t, err)

	// externally checked
	expected := NewRoots(complex(-1.4838920018993484, 3.04283839228145), -0.6814635644285129+0i, complex(-0.42537621588638014, 0.5755095234855198))
	assert.True(t, expected.Equals(p))

	fac := math.Sqrt(10)
	kp := 0.001
	for range 13 {

		k := PID(kp, 1.5, 1)

		g0 := k.Mul(g)

		gw, err := g0.Loop()
		assert.Nil(t, err)

		_, err3 := gw.Poles()
		assert.Nil(t, err3)
		kp *= fac
	}
}
