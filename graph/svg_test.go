package graph

import (
	"math"
	"os"
	"testing"
)

func Test_Simple(t *testing.T) {
	f, _ := os.Create("/home/hneemann/temp/z.svg")
	defer f.Close()
	s := NewSVG(800, 600, 10, f)

	p := Plot{
		XAxis:   NewLinear(-4.5, 4),
		YAxis:   NewLinear(-3, 3),
		Content: []PlotContent{Function(math.Sin), Function(math.Cos), Function(math.Tan)},
	}
	p.DrawTo(s, nil)
	s.Close()

}
