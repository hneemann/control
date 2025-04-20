package polynomial

import (
	"github.com/hneemann/control/graph"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestSimple(t *testing.T) {

	s := NewSystem().
		AddBlock([]string{}, "w", Const(1)).
		AddBlock([]string{"ul"}, "y", Linear(&Linear{
			Numerator:   Polynomial{70},
			Denominator: NewRoots(complex(-1, 0), complex(-2, 0), complex(-2.5, 0)).Polynomial(),
		})).
		AddBlock([]string{"e"}, "u", Linear(Must(PIDReal(0.3, 1.14, 0.77, 0.05)))).
		AddBlock([]string{"w", "y"}, "e", Sub()).
		AddBlock([]string{"u"}, "ul", Limit(0, 0.8))

	err := s.Initialize()
	assert.NoError(t, err)

	data := s.Run(10)

	p := graph.Plot{YBounds: graph.NewBounds(-0.2, 2)}

	n := 0
	for name, points := range data {
		style := graph.GetColor(n)
		p.AddContent(graph.Curve{
			Path:  graph.NewPointsPath(false, points...),
			Style: style,
		})
		p.AddLegend(name, style, nil, nil)
		n++
	}

	file, err := os.Create("test.svg")
	assert.NoError(t, err)
	svg := graph.NewSVG(800, 600, 15, file)
	assert.NoError(t, p.DrawTo(svg))
	assert.NoError(t, svg.Close())
	assert.NoError(t, file.Close())
}

func Must[T any](a T, e error) T {
	if e != nil {
		panic(e)
	}
	return a
}
