package graph

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func p(x, y float64) Point {
	return Point{X: x, Y: y}
}

func TestRect_Intersect(t *testing.T) {
	tests := []struct {
		name   string
		p0, p1 Point
		w0, w1 Point
		cut    IntersectResult
	}{
		{"inside", p(-0.5, -0.5), p(0.5, 0.5), p(-0.5, -0.5), p(0.5, 0.5), BothInside},
		{"Outside", p(-1.5, -0.5), p(-1.5, 0.5), p(-1.5, -0.5), p(-1.5, 0.5), BothOutside},
		{"Outside2", p(1.5, -0.5), p(1.5, 0.5), p(1.5, -0.5), p(1.5, 0.5), BothOutside},
		{"oneInside0", p(0, 0), p(3, 0), p(0, 0), p(1, 0), P1Outside},
		{"oneInside1", p(-3, 0), p(0, 0), p(-1, 0), p(0, 0), P0Outside},
		{"oneInside2", p(0, 0), p(0, 3), p(0, 0), p(0, 1), P1Outside},
		{"oneInside3", p(0, -3), p(0, 0), p(0, -1), p(0, 0), P0Outside},
		{"both outside1", p(-1, 2), p(2, -1), p(0, 1), p(1, 0), BothOutsidePartVisible},
		{"both outside2", p(-2, 1), p(1, -2), p(-1, 0), p(0, -1), BothOutsidePartVisible},
		{"both outside3", p(-1, -2), p(2, 1), p(0, -1), p(1, 0), BothOutsidePartVisible},
		{"both outside4", p(-2, -1), p(1, 2), p(-1, 0), p(0, 1), BothOutsidePartVisible},

		{"both outside nc", p(-2.1, 0), p(0, 2.1), p(-2.1, 0), p(0, 2.1), BothOutside},
		{"both outside nc", p(-2, 0), p(0, 2), p(-2, 0), p(0, 2), BothOutside},
	}
	r := NewRect(-1, 1, -1, 1)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g0, g1, cut := r.Intersect(tt.p0, tt.p1)
			assert.Equal(t, tt.cut, cut)
			assert.Equal(t, tt.w0, g0)
			assert.Equal(t, tt.w1, g1)
		})
	}
}
