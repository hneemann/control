package graph

import (
	"math"
	"os"
	"testing"
)

func Test_Simple(t *testing.T) {
	f, _ := os.Create("/home/hneemann/temp/control/z.svg")
	defer f.Close()
	s := NewSVG(800, 600, 15, f)

	p := Plot{
		XBounds: NewBounds(-4.5, 4),
		YBounds: NewBounds(-3, 3),
		Content: []PlotContent{Function{Function: math.Sin}, Function{Function: math.Cos}, Function{Function: math.Tan}},
	}
	p.DrawTo(s)
	s.Close()

}
