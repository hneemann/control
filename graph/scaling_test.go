package graph

import (
	"github.com/stretchr/testify/assert"
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
			func(width float64, vks, nks int) bool {
				return width > 10
			},
			[]Tick{{-4, "-4"}, {-3, "-3"}, {-2, "-2"}, {-1, "-1"},
				{0, "0"}, {1, "1"}, {2, "2"}, {3, "3"}, {4, "4"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ti, _ := LinearAxis(tt.min, tt.max, tt.ctw, tt.bounds)
			assert.EqualValues(t, tt.want, ti)
		})
	}
}
