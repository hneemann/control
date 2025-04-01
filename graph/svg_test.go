package graph

import (
	"math"
	"os"
	"testing"
)

func Test_Simple(t *testing.T) {
	f, _ := os.Create("/home/hneemann/temp/z.svg")
	defer f.Close()
	s := NewSVG(800, 600, f)

	p := Plot{
		xAxis:   NewLinear(-4.5, 4),
		yAxis:   NewLinear(-3, 3),
		Content: []Node{Function(math.Sin), Function(math.Cos), Function(math.Tan)},
	}
	p.Draw(s)
	s.Close()

}
