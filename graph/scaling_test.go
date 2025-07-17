package graph

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func Test_Axis(t *testing.T) {
	tests := []struct {
		name   string
		bounds Bounds
		min    float64
		max    float64
		ctw    CheckTextWidth
		want   []Tick
	}{
		{"Linear",
			NewBounds(-4, 4),
			0, 100,
			func(width float64, _ int) bool {
				return width > 10
			},
			[]Tick{{-4, "-4"}, {-3, "-3"}, {-2, "-2"}, {-1, "-1"},
				{0, "0"}, {1, "1"}, {2, "2"}, {3, "3"}, {4, "4"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ti, _, _ := LinearAxis(tt.min, tt.max, tt.bounds, tt.ctw, 0.02)
			assert.EqualValues(t, tt.want, ti)
		})
	}
}

func TestFormatFloat(t *testing.T) {
	tests := []struct {
		v    float64
		want string
	}{
		{0, "0"},
		{1e0, "1"},
		{1e1, "10"},
		{1e2, "100"},
		{1e3, "1000"},
		{1e4, "10000"},
		{1e5, "100000"},
		{1e6, "10⁶"},
		{1e7, "10⁷"},
		{1e8, "10⁸"},
		{1e9, "10⁹"},
		{1e10, "10¹⁰"},
		{1e11, "10¹¹"},
		{3e0, "3"},
		{3e1, "30"},
		{3e2, "300"},
		{3e3, "3000"},
		{3e4, "30000"},
		{3e5, "300000"},
		{3e6, "3⋅10⁶"},
		{3e7, "3⋅10⁷"},
		{3e8, "3⋅10⁸"},
		{3e9, "3⋅10⁹"},
		{3e10, "3⋅10¹⁰"},
		{3e11, "3⋅10¹¹"},
		{math.Pi * 1e7, "3.142⋅10⁷"},
		{-3e3, "-3000"},
		{-3e7, "-3⋅10⁷"},
		{3e-3, "0.003"},
		{5e-7, "5⋅10⁻⁷"},
		{1e-1, "0.1"},
		{1e-2, "0.01"},
		{1e-3, "0.001"},
		{1e-4, "0.0001"},
		{1e-5, "10⁻⁵"},
		{3.2e-5, "3.2⋅10⁻⁵"},
		{1e-6, "10⁻⁶"},
		{1e-7, "10⁻⁷"},
		{1e-8, "10⁻⁸"},
		{1e-11, "10⁻¹¹"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.v), func(t *testing.T) {
			assert.Equalf(t, tt.want, FormatFloat(tt.v), "FormatFloat(%v)", tt.v)
		})
	}
}
