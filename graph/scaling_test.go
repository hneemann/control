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
			ad := LinearAxis(tt.min, tt.max, tt.bounds, tt.ctw, 0.02)
			assert.EqualValues(t, tt.want, ad.Ticks)

			for _, ti := range tt.want {
				pix := ad.Trans(ti.Position)
				rev := ad.Reverse(pix)
				assert.InDelta(t, ti.Position, rev, 0.0001, "position %v pix %v rev %v", ti.Position, pix, rev)
			}

		})
	}
}

func Test_LinAxisInf(t *testing.T) {
	bounds := NewBounds(18, math.Inf(-1))
	ad := LinearAxis(0, 60, bounds,
		func(width float64, _ int) bool {
			return width > 25
		}, 0.02)

	assert.False(t, math.IsNaN(ad.Bounds.Max) || math.IsInf(ad.Bounds.Max, -1) || math.IsInf(ad.Bounds.Max, 1))
	assert.False(t, math.IsNaN(ad.Bounds.Min) || math.IsInf(ad.Bounds.Min, -1) || math.IsInf(ad.Bounds.Min, 1))
}

func Test_LogAxisRev(t *testing.T) {
	bounds := NewBounds(0.01, 1000)
	ad := LogAxis(0, 60, bounds,
		func(width float64, _ int) bool {
			return width > 25
		}, 0.02)

	v := ad.Trans(2.3)
	rev := ad.Reverse(v)
	assert.InDelta(t, 2.3, rev, 0.0001)
}

func Test_LogAxisInf(t *testing.T) {
	bounds := NewBounds(1, math.Inf(1))
	ad := LogAxis(0, 60, bounds,
		func(width float64, _ int) bool {
			return width > 25
		}, 0.02)

	assert.False(t, math.IsNaN(ad.Bounds.Max) || math.IsInf(ad.Bounds.Max, -1) || math.IsInf(ad.Bounds.Max, 1))
	assert.False(t, math.IsNaN(ad.Bounds.Min) || math.IsInf(ad.Bounds.Min, -1) || math.IsInf(ad.Bounds.Min, 1))
}

func Test_dBAxisRev(t *testing.T) {
	bounds := NewBounds(0.01, 1000)
	ad := DBAxis(0, 60, bounds,
		func(width float64, _ int) bool {
			return width > 25
		}, 0.02)

	v := ad.Trans(2.3)
	rev := ad.Reverse(v)
	assert.InDelta(t, 2.3, rev, 0.0001)
}

func Test_dBAxisInf(t *testing.T) {
	bounds := NewBounds(1, math.Inf(1))
	ad := DBAxis(0, 60, bounds,
		func(width float64, _ int) bool {
			return width > 25
		}, 0.02)

	assert.False(t, math.IsNaN(ad.Bounds.Max) || math.IsInf(ad.Bounds.Max, -1) || math.IsInf(ad.Bounds.Max, 1))
	assert.False(t, math.IsNaN(ad.Bounds.Min) || math.IsInf(ad.Bounds.Min, -1) || math.IsInf(ad.Bounds.Min, 1))
}

func Test_FixedStepIsInf(t *testing.T) {
	bounds := NewBounds(1, math.Inf(1))
	ax := CreateFixedStepAxis(100)
	ad := ax(0, 60, bounds,
		func(width float64, _ int) bool {
			return width > 25
		}, 0.02)

	assert.False(t, math.IsNaN(ad.Bounds.Max) || math.IsInf(ad.Bounds.Max, -1) || math.IsInf(ad.Bounds.Max, 1))
	assert.False(t, math.IsNaN(ad.Bounds.Min) || math.IsInf(ad.Bounds.Min, -1) || math.IsInf(ad.Bounds.Min, 1))
}
