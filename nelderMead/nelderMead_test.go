package nelderMead

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNelderMead(t *testing.T) {

	solvable := func(v Vector) float64 {
		return sqr(v[0]-1) + sqr(v[1]-2) + 1
	}
	min, val, err := NelderMead(solvable, []Vector{{0, 0}, {0, 1}, {1, 0}}, 100)
	assert.Nil(t, err)
	assert.InDelta(t, 1, min[0], 1e-6)
	assert.InDelta(t, 2, min[1], 1e-6)
	assert.InDelta(t, 1, val, 1e-6)
}

func sqr(x float64) float64 {
	return x * x
}
