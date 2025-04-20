package graph

import (
	"bufio"
	"fmt"
	"html"
	"io"
	"math"
)

type SVG struct {
	rect    Rect
	w       *bufio.Writer
	context *Context
	err     error
}

func NewSVG(width, height, textSize float64, writer io.Writer) *SVG {
	var w *bufio.Writer
	if bw, ok := writer.(*bufio.Writer); ok {
		w = bw
	} else {
		w = bufio.NewWriter(writer)
	}
	s := &SVG{rect: Rect{
		Point{0, 0},
		Point{width, height},
	}, w: w, context: &Context{TextSize: textSize}}

	//s.write("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	s.write(fmt.Sprintf("<svg class=\"svg\" xmlns:svg=\"http://www.w3.org/2000/svg\"\n   xmlns=\"http://www.w3.org/2000/svg\"\n   width=\"%g\"\n   height=\"%g\"\n   viewBox=\"0 0 %g %g\">\n",
		width, height, width, height))

	return s
}

func (s *SVG) Close() error {
	s.write("</svg>")
	e := s.w.Flush()
	if s.err == nil {
		s.err = e
	}
	return s.err
}

func (s *SVG) DrawPath(path Path, style *Style) {
	isOpen := false
	for m, p := range path.Iter {
		if !isOpen {
			s.write("  <path d=\"")
			isOpen = true
		}
		s.writeRune(m)
		s.writeRune(' ')
		s.write(fmt.Sprintf("%.2f,%.2f ", p.X, s.rect.Max.Y-p.Y))
	}
	if isOpen {
		if path.IsClosed() {
			s.write("Z")
		}
		s.write("\"")
		s.writeStyle(style, "")
		s.write("/>\n")
	}
}

func (s *SVG) DrawShape(a Point, shape Shape, style *Style) {
	shape.DrawTo(TransformCanvas{transform: Translate(a), parent: s, size: s.rect}, style)
}

func (s *SVG) writeStyle(style *Style, extra string) {
	if style == nil {
		style = Black
	}
	s.write(" style=\"")
	if style.Stroke {
		s.write("stroke:")
		s.write(style.Color.Color())
		s.write(";stroke-opacity:")
		s.write(style.Color.Opacity())
		s.write(";stroke-width:")
		s.write(fmt.Sprintf("%0.2g", style.StrokeWidth))
		s.write(";stroke-linejoin:round")
		if len(style.Dash) > 0 {
			s.write(";stroke-dasharray:")
			for i, d := range style.Dash {
				if i > 0 {
					s.write(",")
				}
				s.write(fmt.Sprintf("%0.2f", d))
			}
		}
	} else {
		s.write("stroke:none")
	}
	if style.Fill {
		s.write(";fill:")
		s.write(style.FillColor.Color())
		s.write(";fill-opacity:")
		s.write(style.FillColor.Opacity())
		s.write(";fill-rule:evenodd")
	} else {
		s.write(";fill:none")
	}
	if extra != "" {
		s.write(";" + extra)
	}
	s.write("\"")
}

func (s *SVG) DrawCircle(a Point, b Point, style *Style) {
	s.write("  <ellipse ")
	s.write(fmt.Sprintf("cx=\"%0.2f\" cy=\"%0.2f\" ", (a.X+b.X)/2, s.rect.Max.Y-(a.Y+b.Y)/2))
	s.write(fmt.Sprintf("rx=\"%0.2f\" ry=\"%0.2f\"", math.Abs(a.X-b.X)/2, math.Abs(a.Y-b.Y)/2))
	s.writeStyle(style, "")
	s.write("/>\n")
}

func (s *SVG) DrawText(a Point, text string, orientation Orientation, style *Style, textSize float64) {
	st := fmt.Sprintf("font-size:%0.2gpx", textSize)
	switch orientation & 3 {
	case 0:
		a.X += textSize / 4
	case 1:
		st += ";text-anchor:middle"
	case 2:
		st += ";text-anchor:end"
		a.X -= textSize / 4
	}
	switch (orientation >> 2) & 3 {
	case 0:
		a.Y += textSize / 4
	case 1:
		a.Y -= textSize / 3
	case 2:
		a.Y -= textSize / 10 * 9
	}

	s.write("  <text ")
	s.write(fmt.Sprintf("x=\"%0.2f\" y=\"%0.2f\" ", a.X, s.rect.Max.Y-a.Y))
	s.writeStyle(style, st)
	s.write(">")
	s.write(html.EscapeString(text))
	s.write("</text>\n")
}

func (s *SVG) Context() *Context {
	return s.context
}

func (s *SVG) Rect() Rect {
	return s.rect
}

func (s *SVG) write(str string) {
	_, e := s.w.WriteString(str)
	if s.err == nil {
		s.err = e
	}
}

func (s *SVG) writeRune(mode rune) {
	_, e := s.w.WriteRune(mode)
	if s.err == nil {
		s.err = e
	}
}
