package graph

import (
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

func Test_LinAxisInf(t *testing.T) {
	bounds := NewBounds(18, math.Inf(-1))
	_, _, bounds, _ = LinearAxis(0, 60, bounds,
		func(width float64, _ int) bool {
			return width > 25
		}, 0.02)

	assert.False(t, math.IsNaN(bounds.Max) || math.IsInf(bounds.Max, -1) || math.IsInf(bounds.Max, 1))
	assert.False(t, math.IsNaN(bounds.Min) || math.IsInf(bounds.Min, -1) || math.IsInf(bounds.Min, 1))
}

func Test_LogAxisInf(t *testing.T) {
	bounds := NewBounds(1, math.Inf(1))
	_, _, bounds, _ = LogAxis(0, 60, bounds,
		func(width float64, _ int) bool {
			return width > 25
		}, 0.02)

	assert.False(t, math.IsNaN(bounds.Max) || math.IsInf(bounds.Max, -1) || math.IsInf(bounds.Max, 1))
	assert.False(t, math.IsNaN(bounds.Min) || math.IsInf(bounds.Min, -1) || math.IsInf(bounds.Min, 1))
}

func Test_dBAxisInf(t *testing.T) {
	bounds := NewBounds(1, math.Inf(1))
	_, _, bounds, _ = DBAxis(0, 60, bounds,
		func(width float64, _ int) bool {
			return width > 25
		}, 0.02)

	assert.False(t, math.IsNaN(bounds.Max) || math.IsInf(bounds.Max, -1) || math.IsInf(bounds.Max, 1))
	assert.False(t, math.IsNaN(bounds.Min) || math.IsInf(bounds.Min, -1) || math.IsInf(bounds.Min, 1))
}

func Test_FixedStepIsInf(t *testing.T) {
	bounds := NewBounds(1, math.Inf(1))
	ax := CreateFixedStepAxis(100)
	_, _, bounds, _ = ax(0, 60, bounds,
		func(width float64, _ int) bool {
			return width > 25
		}, 0.02)

	assert.False(t, math.IsNaN(bounds.Max) || math.IsInf(bounds.Max, -1) || math.IsInf(bounds.Max, 1))
	assert.False(t, math.IsNaN(bounds.Min) || math.IsInf(bounds.Min, -1) || math.IsInf(bounds.Min, 1))
}
