package graph

import (
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func TestBounds_Merge(t *testing.T) {
	tests := []struct {
		name string
		b    Bounds
		v    float64
		want Bounds
	}{
		{name: "empty", b: Bounds{}, v: 1, want: Bounds{isSet: true, Min: 1, Max: 1}},
		{name: "max", b: Bounds{isSet: true, Min: 1, Max: 1}, v: 2, want: Bounds{isSet: true, Min: 1, Max: 2}},
		{name: "min", b: Bounds{isSet: true, Min: 1, Max: 1}, v: 0, want: Bounds{isSet: true, Min: 0, Max: 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.b.Merge(tt.v)
			assert.Equalf(t, tt.want, tt.b, "Merge(%v)", tt.b)
		})
	}
}

func TestBounds_MergeBounds(t *testing.T) {
	tests := []struct {
		name  string
		b     Bounds
		other Bounds
		want  Bounds
	}{
		{name: "empty", b: Bounds{}, other: Bounds{}, want: Bounds{isSet: false}},
		{name: "this empty", b: Bounds{}, other: Bounds{true, 1, 2}, want: Bounds{true, 1, 2}},
		{name: "other empty", b: Bounds{true, 1, 2}, other: Bounds{}, want: Bounds{true, 1, 2}},
		{name: "both", b: Bounds{true, 1, 2}, other: Bounds{true, 0, 3}, want: Bounds{true, 0, 3}},
		{name: "both2", b: Bounds{true, 1, 2}, other: Bounds{true, 3, 4}, want: Bounds{true, 1, 4}},
		{name: "both3", b: Bounds{true, 3, 4}, other: Bounds{true, 1, 2}, want: Bounds{true, 1, 4}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.b.MergeBounds(tt.other)
			assert.Equalf(t, tt.want, tt.b, "Merge(%v)", tt.b)
		})
	}
}

func Test_angleBetween(t *testing.T) {
	alpha := 10 * math.Pi / 180
	expected := math.Cos(alpha)
	for i := range 400 {
		a0 := float64(i) * math.Pi / 180
		a1 := a0 + alpha

		d0 := Point{X: math.Cos(a0), Y: math.Sin(a0)}
		d1 := Point{X: math.Cos(a1), Y: math.Sin(a1)}

		angle := cosAngleBetween(d0, d1)
		assert.InDelta(t, expected, angle, 1e-7, "angleBetween(%v, %v)", d0, d1)

	}
}
