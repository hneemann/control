package graph

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/hneemann/control/graph/grParser/mathml"
	"github.com/hneemann/parser2/value/export/xmlWriter"
	"html/template"
	img "image"
	"image/png"
	"math"
	"strings"
)

type SVG struct {
	rect             Rect
	w                *xmlWriter.XMLWriter
	context          *Context
	err              error
	strokeCorrection float64
	latex            bool
}

func NewSVG(context *Context, w *xmlWriter.XMLWriter) *SVG {
	width := context.Width
	height := context.Height
	s := &SVG{
		rect: Rect{
			Point{0, 0},
			Point{width, height},
		},
		w:                w,
		context:          context,
		strokeCorrection: context.StrokeCorrection,
		latex:            context.LaTeX,
	}

	w.Open("svg").
		Attr("class", "svg").
		Attr("style", "stroke-linejoin:round;stroke-linecap:round").
		Attr("xmlns:svg", "http://www.w3.org/2000/svg").
		Attr("xmlns", "http://www.w3.org/2000/svg").
		Attr("width", fmt.Sprintf("%g", width)).
		Attr("height", fmt.Sprintf("%g", height)).
		Attr("viewBox", fmt.Sprintf("0 0 %g %g", width, height))
	/*	w.Open("style");
		w.Write(".math {font-style: italic;font-family: 'Times New Roman', Times, serif;}\n");
		w.Write(".sub {font-size:70%;baseline-shift:sub;}\n");
		w.Write(".super {font-size:70%;baseline-shift:super}\n");
		w.Close()*/
	return s
}

func (s *SVG) Close() error {
	s.w.Close()
	return nil
}

func (s *SVG) DrawPath(path Path, style *Style) error {
	var buf bytes.Buffer
	for pe, err := range path.Iter {
		if err != nil {
			return fmt.Errorf("error writing path to svg: %w", err)
		}
		buf.WriteRune(pe.Mode)
		buf.WriteRune(' ')
		buf.WriteString(fmt.Sprintf("%0.2f,%0.2f ", pe.Point.X, s.rect.Max.Y-pe.Point.Y))
	}
	if buf.Len() > 0 {
		if path.IsClosed() {
			buf.WriteRune('Z')
		}
		s.w.Open("path").
			Attr("d", buf.String()).
			Attr("style", s.styleString(style)).
			Close()
	}
	return nil
}

func (s *SVG) DrawTriangle(p1, p2, p3 Point, style *Style) error {
	s.w.Open("polygon").
		Attr("points", fmt.Sprintf("%0.2f,%0.2f %0.2f,%0.2f %0.2f,%0.2f",
			p1.X, s.rect.Max.Y-p1.Y,
			p2.X, s.rect.Max.Y-p2.Y,
			p3.X, s.rect.Max.Y-p3.Y)).
		Attr("style", s.styleString(style)).
		Close()
	return nil
}

func (s *SVG) DrawShape(a Point, shape Shape, style *Style) error {
	return shape.DrawTo(TransformCanvas{transform: Translate(a), parent: s, size: s.rect}, style)
}

func (s *SVG) styleString(style *Style) string {
	if style == nil {
		style = Black
	}
	var buf bytes.Buffer
	if style.Stroke {
		buf.WriteString("stroke:")
		buf.WriteString(style.Color.Color())
		if style.Color.A < 255 {
			buf.WriteString(";stroke-opacity:")
			buf.WriteString(style.Color.Opacity())
		}
		buf.WriteString(";stroke-width:")
		buf.WriteString(fmt.Sprintf("%0.2g", style.StrokeWidth*s.strokeCorrection))
		if len(style.Dash) > 0 {
			buf.WriteString(";stroke-dasharray:")
			for i, d := range style.Dash {
				if i > 0 {
					buf.WriteString(",")
				}
				buf.WriteString(fmt.Sprintf("%0.2f", d*s.strokeCorrection))
			}
		}
	} else {
		buf.WriteString("stroke:none")
	}
	if style.Fill {
		buf.WriteString(";fill:")
		buf.WriteString(style.FillColor.Color())
		if style.FillColor.A < 255 {
			buf.WriteString(";fill-opacity:")
			buf.WriteString(style.FillColor.Opacity())
		}
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
		Attr("style", s.styleString(style)).
		Close()
}

func (s *SVG) DrawImage(a Point, b Point, img img.Image) error {
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		return fmt.Errorf("error encoding image to png: %w", err)
	}

	xp := math.Min(a.X, b.X)
	yp := s.rect.Max.Y - math.Max(a.Y, b.Y)
	width := math.Abs(a.X - b.X)
	height := math.Abs(a.Y - b.Y)

	s.w.Open("image").
		Attr("preserveAspectRatio", "none").
		Attr("x", fmt.Sprintf("%0.2f", xp)).
		Attr("y", fmt.Sprintf("%0.2f", yp)).
		Attr("width", fmt.Sprintf("%0.2f", width)).
		Attr("height", fmt.Sprintf("%0.2f", height)).
		Attr("href", "data:image/png;base64,"+base64.StdEncoding.EncodeToString(buf.Bytes())).
		Close()

	return nil
}

func (s *SVG) DrawText(a Point, text string, orientation Orientation, style *Style, textSize float64) {
	st := fmt.Sprintf(";font-size:%0.2fpx", textSize)
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
			switch orientation & 3 {
			case 1:
				a.X -= float64(len(text)-3) * textSize / 6
			case 2:
				a.X -= float64(len(text)-3) * textSize / 3
			}
			s.w.Open("foreignObject").
				Attr("style", s.styleString(style)+st).
				Attr("x", fmt.Sprintf("%0.2f", a.X)).
				Attr("y", fmt.Sprintf("%0.2f", s.rect.Max.Y-a.Y)).
				Attr("width", fmt.Sprintf("%0.2f", textSize*float64(len(text)))).
				Attr("height", fmt.Sprintf("%0.2f", textSize*3))

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
			Attr("style", s.styleString(style)+st)
		if s.latex {
			s.w.WriteHTML(template.HTML(unicodeToLaTeX(text)))
		} else {
			s.w.WriteHTML(template.HTML(parseSupSub(text, "tspan")))
		}
		s.w.Close()
	}
}

var unicodeMap = map[rune]string{
	'α': `\alpha`,
	'β': `\beta`,
	'γ': `\gamma`,
	'δ': `\delta`,
	'ε': `\epsilon`,
	'ζ': `\zeta`,
	'η': `\eta`,
	'θ': `\theta`,
	'ι': `\iota`,
	'κ': `\kappa`,
	'λ': `\lambda`,
	'μ': `\mu`,
	'ν': `\nu`,
	'ξ': `\xi`,
	'ο': `o`,
	'π': `\pi`,
	'ρ': `\rho`,
	'σ': `\sigma`,
	'τ': `\tau`,
	'υ': `\upsilon`,
	'φ': `\varphi`,
	'χ': `\chi`,
	'ψ': `\psi`,
	'ω': `\omega`,
	'Γ': `\Gamma`,
	'Δ': `\Delta`,
	'Θ': `\Theta`,
	'Λ': `\Lambda`,
	'Ξ': `\Xi`,
	'Π': `\Pi`,
	'Σ': `\Sigma`,
	'Υ': `\Upsilon`,
	'Φ': `\Phi`,
	'Ψ': `\Psi`,
	'Ω': `\Omega`,
	'≈': `\approx`,
	'≠': `\neq`,
	'≤': `\leq`,
	'≥': `\geq`,
	'∞': `\infty`,
	'∑': `\sum`,
	'∏': `\prod`,
	'∫': `\int`,
	'∂': `\partial`,
	'∇': `\nabla`,
	'∝': `\propto`,
}

func unicodeToLaTeX(text string) string {
	var w strings.Builder
	inMath := false
	for _, r := range text {
		if c, ok := unicodeMap[r]; ok {
			if !inMath {
				w.WriteRune('$')
			}
			w.WriteString(c)
			if !inMath {
				w.WriteRune('$')
			}
		} else {
			if r == '$' {
				inMath = !inMath
			}
			w.WriteRune(r)
		}
	}
	return w.String()
}

// ParseSupSub parses subscript and superscript in the text.
// It recognizes the following patterns:
// - G_{0} -> G<tspan style="font-size:70%;baseline-shift:sub">0</tspan>
// - G^{0} -> G<tspan style="font-size:70%;baseline-shift:super">0</tspan>
// If the text contains the LaTeX math mode character '$', a Times New Roman
// font with italic font style is used.
func ParseSupSub(text string) string {
	return parseSupSub(text, "span")
}

func parseSupSub(text string, tag string) string {
	const (
		normal = iota
		superscriptFirst
		superscript
		subscriptFirst
		subscript
	)
	var w strings.Builder
	mode := normal
	inMath := false
	arg := ""
	for i, r := range text {
		if r == '$' {
			if inMath {
				inMath = false
				w.WriteString("</")
				w.WriteString(tag)
				w.WriteString(">")
			} else {
				w.WriteString("<")
				w.WriteString(tag)
				w.WriteString(" style=\"font-style: italic;font-family: 'Times New Roman', Times, serif;\">")
				inMath = true
			}
		} else {
			switch mode {
			case normal:
				if r == '_' && i+1 < len(text) && text[i+1] == '{' {
					mode = subscriptFirst
				} else if r == '^' && i+1 < len(text) && text[i+1] == '{' {
					mode = superscriptFirst
				} else {
					w.WriteRune(r)
				}
			case subscriptFirst:
				mode = subscript
			case subscript:
				if r == '}' {
					mode = normal
					w.WriteString("<")
					w.WriteString(tag)
					w.WriteString(" style=\"font-size:70%;baseline-shift:sub\">")
					w.WriteString(arg)
					w.WriteString("</")
					w.WriteString(tag)
					w.WriteString(">")
					arg = ""
				} else {
					arg += string(r)
				}
			case superscriptFirst:
				mode = superscript
			case superscript:
				if r == '}' {
					mode = normal
					w.WriteString("<")
					w.WriteString(tag)
					w.WriteString(" style=\"font-size:70%;baseline-shift:super\">")
					w.WriteString(arg)
					w.WriteString("</")
					w.WriteString(tag)
					w.WriteString(">")
					arg = ""
				} else {
					arg += string(r)
				}
			}
		}
	}
	if (mode == subscript || mode == superscript) && len(arg) > 0 {
		w.WriteString(arg)
	}
	if inMath {
		w.WriteString("</")
		w.WriteString(tag)
		w.WriteString(">")
	}

	return w.String()
}

func (s *SVG) Context() *Context {
	return s.context
}

func (s *SVG) Rect() Rect {
	return s.rect
}
