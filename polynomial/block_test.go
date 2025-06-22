package polynomial

import (
	"github.com/hneemann/control/graph"
	"github.com/hneemann/parser2/value/export/xmlWriter"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSimple(t *testing.T) {

	s := NewSystem().
		AddBlock([]string{}, "w", Const(1)).
		AddBlock([]string{"ul"}, "y", BlockLinear(&Linear{
			Numerator:   Polynomial{70},
			Denominator: NewRoots(complex(-1, 0), complex(-2, 0), complex(-2.5, 0)).Polynomial(),
		})).
		AddBlock([]string{"e"}, "u", BlockLinear(Must(PID(0.3, 1.14, 0.77, 0.05)))).
		AddBlock([]string{"w", "y"}, "e", Sub()).
		AddBlock([]string{"u"}, "ul", Limit(0, 0.8))

	err := s.Initialize()
	assert.NoError(t, err)

	data, err := s.Run(10)
	assert.NoError(t, err)

	p := graph.Plot{YBounds: graph.NewBounds(-0.2, 2)}

	for n, name := range s.outputs {
		style := graph.GetColor(n)
		p.AddContent(graph.Scatter{
			Points:         data.toPoints(0, n+1),
			ShapeLineStyle: graph.ShapeLineStyle{LineStyle: style},
		})
		p.AddLegend(name, style, nil, nil)
		n++
	}

	assert.NoError(t, err)
	svg := graph.NewSVG(&graph.DefaultContext, xmlWriter.New())
	assert.NoError(t, p.DrawTo(svg))
	assert.NoError(t, svg.Close())

}
