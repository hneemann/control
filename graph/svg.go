package graph

import (
	"bufio"
	"fmt"
	"io"
	"math"
)

type SVG struct {
	size    Rect
	w       *bufio.Writer
	context *Context
}

func NewSVG(width, height, textsize float64, writer io.Writer) *SVG {
	var w *bufio.Writer
	if bw, ok := writer.(*bufio.Writer); ok {
		w = bw
	} else {
		w = bufio.NewWriter(writer)
	}
	w.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	w.WriteString(fmt.Sprintf("<svg xmlns:svg=\"http://www.w3.org/2000/svg\"\n   xmlns=\"http://www.w3.org/2000/svg\"\n   width=\"%g\"\n   height=\"%g\"\n   viewBox=\"0 0 %g %g\">\n",
		width, height, width, height))
	return &SVG{size: Rect{
		Point{0, 0},
		Point{width, height},
	}, w: w, context: &Context{TextSize: textsize}}
}

func (s *SVG) Close() {
	s.w.WriteString("</svg>")
	s.w.Flush()
}

func (s *SVG) Path(polygon Path, style *Style) {
	s.w.WriteString("  <path d=\"")
	for _, p := range polygon.Elements {
		s.w.WriteRune(p.Mode)
		s.w.WriteRune(' ')
		s.w.WriteString(fmt.Sprintf("%.2f,%.2f ", p.X, s.size.Max.Y-p.Y))
	}
	if polygon.Closed {
		s.w.WriteString("Z")
	}
	s.w.WriteString("\"")
	writeStyle(s.w, style, "")
	s.w.WriteString("/>\n")
}

func (s *SVG) Shape(a Point, shape Shape, style *Style) {
	shape.DrawTo(TransformCanvas{transform: Translate(a), parent: s, size: s.size}, style)
}

func writeStyle(w *bufio.Writer, style *Style, extra string) {
	w.WriteString(" style=\"")
	if style.Stroke {
		w.WriteString("stroke:")
		w.WriteString(style.Color.Color())
		w.WriteString(";stroke-opacity:")
		w.WriteString(style.Color.Opacity())
		w.WriteString(";stroke-width:")
		w.WriteString(fmt.Sprintf("%0.2g", style.StrokeWidth))
		w.WriteString(";stroke-linejoin:round")
	} else {
		w.WriteString("stroke:none")
	}
	if style.Fill {
		w.WriteString(";fill:")
		w.WriteString(style.FillColor.Color())
		w.WriteString(";fill-opacity:")
		w.WriteString(style.FillColor.Opacity())
		w.WriteString(";fill-rule:evenodd")
	} else {
		w.WriteString(";fill:none")
	}
	if extra != "" {
		w.WriteString(";" + extra)
	}
	w.WriteString("\"")
}

func (s *SVG) Circle(a Point, b Point, style *Style) {
	s.w.WriteString("  <ellipse ")
	s.w.WriteString(fmt.Sprintf("cx=\"%0.2f\" cy=\"%0.2f\" ", (a.X+b.X)/2, s.size.Max.Y-(a.Y+b.Y)/2))
	s.w.WriteString(fmt.Sprintf("rx=\"%0.2f\" ry=\"%0.2f\"", math.Abs(a.X-b.X)/2, math.Abs(a.Y-b.Y)/2))
	writeStyle(s.w, style, "")
	s.w.WriteString("/>\n")
}

func (s *SVG) Text(a Point, text string, orientation Orientation, style *Style, textSize float64) {
	st := fmt.Sprintf("font-size:%0.2gpx", textSize)
	switch orientation & 3 {
	case 1:
		st += ";text-anchor:middle"
	case 2:
		st += ";text-anchor:end"
	}
	switch (orientation >> 2) & 3 {
	case 1:
		st += ";dominant-baseline:middle"
	case 2:
		st += ";dominant-baseline:hanging"
	}

	s.w.WriteString("  <text ")
	s.w.WriteString(fmt.Sprintf("x=\"%0.2f\" y=\"%0.2f\" ", a.X, s.size.Max.Y-a.Y))
	writeStyle(s.w, style, st)
	s.w.WriteString(">")
	s.w.WriteString(text)
	s.w.WriteString("</text>\n")
}

func (s *SVG) Context() *Context {
	return s.context
}

func (s *SVG) Size() Rect {
	return s.size
}
