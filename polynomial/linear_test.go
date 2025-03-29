package polynomial

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func TestLinear_Mul(t *testing.T) {
	l1 := Linear{
		Numerator:   Polynomial{4, 5},
		Denominator: Polynomial{1, 2, 3},
	}
	l2 := Linear{
		Numerator:   Polynomial{5},
		Denominator: Polynomial{-2, 1, 4},
	}
	mul, err := l1.Mul(l2)

	expected := Linear{
		Numerator:   Polynomial{20, 25},
		Denominator: Polynomial{-2, -3, 0, 11, 12},
	}

	assert.Nil(t, err)
	assert.True(t, expected.Equals(mul), mul.String())
}

func TestLinear_Div(t *testing.T) {
	l1 := Linear{
		Numerator:   Polynomial{4, 5},
		Denominator: Polynomial{1, 2, 3},
	}
	l2 := Linear{
		Numerator:   Polynomial{5},
		Denominator: Polynomial{-2, 1, 4},
	}
	div, err := l1.Div(l2)

	expected := Linear{
		Numerator:   Polynomial{-8, -6, 21, 20},
		Denominator: Polynomial{5, 10, 15},
	}

	assert.Nil(t, err)
	assert.True(t, expected.Equals(div), div.String())
}

func TestLinear_Add(t *testing.T) {
	l1 := Linear{
		Numerator:   Polynomial{4, 5},
		Denominator: Polynomial{1, 2, 3},
	}
	l2 := Linear{
		Numerator:   Polynomial{5},
		Denominator: Polynomial{-2, 1, 4},
	}
	add, err := l1.Add(l2)

	expected := Linear{
		Numerator:   Polynomial{-3, 4, 36, 20},
		Denominator: Polynomial{-2, -3, 0, 11, 12},
	}

	assert.Nil(t, err)
	assert.True(t, expected.Equals(add), add.String())
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
	g0 := Linear{
		Numerator:   Polynomial{4, 5},
		Denominator: Polynomial{1, 2, 3, 4},
	}

	fmt.Println(g0)
	fmt.Println(Must(g0.Zeros()))
	fmt.Println(Must(g0.Poles()))

	kp := 0.001
	for i := 0; i < 7; i++ {
		gw, _ := g0.MulFloat(kp).Loop()
		fmt.Println(kp, Must(gw.Poles()))
		kp = kp * 10
	}
}

func Must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

func Test_Integration(t *testing.T) {
	n := NewRoots().Real(1.5, 1)
	d := NewRoots().Real(2, 1).Real(1, 1).Complex(1, 3, 3.1)
	g := FromRoots(n, d)

	fac := math.Sqrt(10)
	kp := 0.001
	for range 13 {

		//  $k_p=12$,& $T_I=1.5\sek$, & $T_D=1\sek$
		k := PID(kp, 1.5, 1)

		//fmt.Println(g)
		//fmt.Println(k)

		g0, err1 := k.Mul(g)
		assert.Nil(t, err1)
		//fmt.Println(g0)

		gw, err2 := g0.Loop()
		assert.Nil(t, err2)
		//fmt.Println(gw)

		p, err3 := gw.Poles()
		assert.Nil(t, err3)

		fmt.Println(kp, p)

		kp *= fac
	}
}
