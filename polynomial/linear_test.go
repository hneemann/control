package polynomial

import (
	"fmt"
	"github.com/stretchr/testify/assert"
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
	assert.Nil(t, err)
	assert.Equal(t, "(25x+20)/(12x⁴+11x³-3x-2)", mul.String())
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
	assert.Nil(t, err)
	assert.Equal(t, "(20x³+21x²-6x-8)/(15x²+10x+5)", div.String())
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
	assert.Nil(t, err)
	assert.Equal(t, "(20x³+36x²+4x-3)/(12x⁴+11x³-3x-2)", add.String())
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
			"(x²+x-2)/(x²-x-2)"},
		{"reduce",
			NewRoots(complex(-2, 0), complex(1, 0)),
			NewRoots(complex(-1, 0), complex(1, 0)),
			"(x+2)/(x+1)"},
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
