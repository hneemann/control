package graph

import (
	"bytes"
	"fmt"
	"github.com/hneemann/control/graph/grParser/mathml"
	"github.com/hneemann/parser2/value/export/xmlWriter"
	"math"
	"strings"
)

type SVG struct {
	rect    Rect
	w       *xmlWriter.XMLWriter
	context *Context
	err     error
}

func NewSVG(context *Context, w *xmlWriter.XMLWriter) *SVG {
	width := context.Width
	height := context.Height
	s := &SVG{rect: Rect{
		Point{0, 0},
		Point{width, height},
	}, w: w, context: context}

	w.Open("svg").
		Attr("class", "svg").
		Attr("xmlns:svg", "http://www.w3.org/2000/svg").
		Attr("xmlns", "http://www.w3.org/2000/svg").
		Attr("width", fmt.Sprintf("%g", width)).
		Attr("height", fmt.Sprintf("%g", height)).
		Attr("viewBox", fmt.Sprintf("0 0 %g %g", width, height))

	return s
}

func (s *SVG) Close() error {
	s.w.Close()
	return nil
}

func (s *SVG) DrawPath(path Path, style *Style) {
	var buf bytes.Buffer
	for m, p := range path.Iter {
		buf.WriteRune(m)
		buf.WriteRune(' ')
		buf.WriteString(fmt.Sprintf("%.2f,%.2f ", p.X, s.rect.Max.Y-p.Y))
	}
	if buf.Len() > 0 {
		if path.IsClosed() {
			buf.WriteRune('Z')
		}
		s.w.Open("path").
			Attr("d", buf.String()).
			Attr("style", styleString(style)).
			Close()
	}
}

func (s *SVG) DrawShape(a Point, shape Shape, style *Style) {
	shape.DrawTo(TransformCanvas{transform: Translate(a), parent: s, size: s.rect}, style)
}

func styleString(style *Style) string {
	if style == nil {
		style = Black
	}
	var buf bytes.Buffer
	if style.Stroke {
		buf.WriteString("stroke:")
		buf.WriteString(style.Color.Color())
		buf.WriteString(";stroke-opacity:")
		buf.WriteString(style.Color.Opacity())
		buf.WriteString(";stroke-width:")
		buf.WriteString(fmt.Sprintf("%0.2g", style.StrokeWidth))
		buf.WriteString(";stroke-linejoin:round")
		if len(style.Dash) > 0 {
			buf.WriteString(";stroke-dasharray:")
			for i, d := range style.Dash {
				if i > 0 {
					buf.WriteString(",")
				}
				buf.WriteString(fmt.Sprintf("%0.2f", d))
			}
		}
	} else {
		buf.WriteString("stroke:none")
	}
	if style.Fill {
		buf.WriteString(";fill:")
		buf.WriteString(style.FillColor.Color())
		buf.WriteString(";fill-opacity:")
		buf.WriteString(style.FillColor.Opacity())
		buf.WriteString(";fill-rule:evenodd")
	} else {
		buf.WriteString(";fill:none")
	}
	return buf.String()
}

func (s *SVG) DrawCircle(a Point, b Point, style *Style) {
	s.w.Open("ellipse").
		Attr("cx", fmt.Sprintf("%0.2f", (a.X+b.X)/2)).
		Attr("cy", fmt.Sprintf("%0.2f", s.rect.Max.Y-(a.Y+b.Y)/2)).
		Attr("rx", fmt.Sprintf("%0.2f", math.Abs(a.X-b.X)/2)).
		Attr("ry", fmt.Sprintf("%0.2f", math.Abs(a.Y-b.Y)/2)).
		Attr("style", styleString(style)).
		Close()
}

func (s *SVG) DrawText(a Point, text string, orientation Orientation, style *Style, textSize float64) {
	st := fmt.Sprintf(";font-size:%0.1fpx", textSize)
	switch orientation & 3 {
	case 0:
		a.X += textSize / 4
	case 1:
		st += ";text-anchor:middle"
	case 2:
		st += ";text-anchor:end"
		a.X -= textSize / 4
	}

	if strings.HasPrefix(text, "$$") {
		ast, err := mathml.ParseLaTeX(text[2:])
		if err != nil {
			s.w.Write(text)
		} else {
			switch (orientation >> 2) & 3 {
			case 0:
				a.Y += textSize * 1.2
			case 1:
				a.Y += textSize * 2 / 3
			}
			s.w.Open("foreignObject").
				Attr("style", styleString(style)+st).
				Attr("x", fmt.Sprintf("%0.2f", a.X)).
				Attr("y", fmt.Sprintf("%0.2f", s.rect.Max.Y-a.Y)).
				Attr("width", fmt.Sprintf("%0.2f", textSize*float64(len(text)))).
				Attr("height", fmt.Sprintf("%0.2f", textSize*2))

			ast.ToMathMl(s.w, nil)
			s.w.Close()
		}
	} else {
		switch (orientation >> 2) & 3 {
		case 0:
			a.Y += textSize / 4
		case 1:
			a.Y -= textSize / 3
		case 2:
			a.Y -= textSize / 10 * 9
		}
		s.w.Open("text").
			Attr("x", fmt.Sprintf("%0.2f", a.X)).
			Attr("y", fmt.Sprintf("%0.2f", s.rect.Max.Y-a.Y)).
			Attr("style", styleString(style)+st)
		s.w.Write(text)
		s.w.Close()
	}
}

func (s *SVG) Context() *Context {
	return s.context
}

func (s *SVG) Rect() Rect {
	return s.rect
}
