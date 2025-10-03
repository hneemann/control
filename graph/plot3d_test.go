package graph

import (
	"github.com/hneemann/parser2/value/export/xmlWriter"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func TestPlot3d_DrawTo(t *testing.T) {

	context := &DefaultContext
	writer := xmlWriter.New()
	svg := NewSVG(context, writer)

	plot := &Plot3d{
		X: AxisDescription{
			Bounds:  NewBounds(-7, 7),
			Factory: LinearAxis,
		},
		Y: AxisDescription{
			Bounds:  NewBounds(-7, 7),
			Factory: LinearAxis,
		},
		Z: AxisDescription{
			Bounds:  NewBounds(-1, 3),
			Factory: LinearAxis,
		},
		Contents: []Plot3dContent{
			&Grid3d{
				Func: func(x, y float64) (float64, error) {
					r := math.Sqrt(x*x + y*y)
					return math.Cos(r), nil
				},
				Style: Black,
				Steps: 40,
			},
		},
	}

	err := plot.DrawTo(svg)
	assert.NoError(t, err)
	assert.NoError(t, svg.Close())
}
