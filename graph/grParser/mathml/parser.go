package mathml

import (
	"fmt"
	"github.com/hneemann/parser2/value/export/xmlWriter"
	"html/template"
	"strings"
)

type parser struct {
	tok *Tokenizer
}

type Walker func(a Ast)

type Ast interface {
	ToMathMl(w *xmlWriter.XMLWriter, attr map[string]string)
	Walk(Walker)
}

type MathMl struct {
	ast Ast
}

func (m MathMl) ToMathMl(w *xmlWriter.XMLWriter, _ map[string]string) {
	w.Open("math").Attr("xmlns", "http://www.w3.org/1998/Math/MathML")
	m.ast.ToMathMl(w, nil)
	w.Close()
}

func (m MathMl) Walk(walker Walker) {
	walker(m)
	m.ast.Walk(walker)
}

type Line []Ast

func (l Line) ToMathMl(w *xmlWriter.XMLWriter, attr map[string]string) {
	for _, item := range l {
		item.ToMathMl(w, attr)
	}
}

func (l Line) Walk(walker Walker) {
	for _, item := range l {
		walker(item)
	}
}

type Plain string

func (p Plain) ToMathMl(w *xmlWriter.XMLWriter, _ map[string]string) {
	w.Write(string(p))
}

func (p Plain) Walk(walker Walker) {
	walker(p)
}

type SimpleItem struct {
	tok      Token
	fontsize string
}

func (s *SimpleItem) setFontSize(size string) *SimpleItem {
	s.fontsize = size
	return s
}

func openWithAttr(w *xmlWriter.XMLWriter, name string, attr map[string]string) {
	w.Open(name)
	for k, v := range attr {
		w.Attr(k, v)
	}
}

func (s *SimpleItem) ToMathMl(w *xmlWriter.XMLWriter, attr map[string]string) {
	switch s.tok.kind {
	case Number:
		openWithAttr(w, "mn", attr)
	case Identifier:
		openWithAttr(w, "mi", attr)
	default:
		openWithAttr(w, "mo", attr)
	}
	w.WriteHTML(template.HTML(s.tok.value))
	w.Close()
}

func (s *SimpleItem) Walk(walker Walker) {
	walker(s)
}

type Row struct {
	items []Ast
}

func (f *Row) ToMathMl(w *xmlWriter.XMLWriter, attr map[string]string) {
	openWithAttr(w, "mrow", attr)
	for _, item := range f.items {
		item.ToMathMl(w, nil)
	}
	w.Close()
}

func (f *Row) Items() []Ast {
	return f.items
}

func (f *Row) Walk(walker Walker) {
	walker(f)
	for _, item := range f.items {
		item.Walk(walker)
	}
}

type Empty struct {
}

func (e *Empty) ToMathMl(*xmlWriter.XMLWriter, map[string]string) {
}

func (e *Empty) Walk(walker Walker) {
	walker(e)
}

func NewRow(l ...Ast) Ast {
	switch len(l) {
	case 0:
		return &Empty{}
	case 1:
		return l[0]
	default:
		return &Row{l}
	}
}

type AddAttribute struct {
	inner Ast
	attr  map[string]string
}

func (a *AddAttribute) ToMathMl(w *xmlWriter.XMLWriter, attr map[string]string) {
	a.inner.ToMathMl(w, a.attr)
}

func addAttribute(key, value string, inner Ast) Ast {
	if a, ok := inner.(*AddAttribute); ok {
		a.attr[key] = value
		return a
	}
	return &AddAttribute{inner: inner, attr: map[string]string{key: value}}
}

func (a *AddAttribute) Walk(walker Walker) {
	walker(a)
	a.inner.Walk(walker)
}

type Fraction struct {
	Top    Ast
	Bottom Ast
}

func (f *Fraction) ToMathMl(w *xmlWriter.XMLWriter, attr map[string]string) {
	openWithAttr(w, "mfrac", attr)
	f.Top.ToMathMl(w, nil)
	f.Bottom.ToMathMl(w, nil)
	w.Close()
}

func (f *Fraction) Walk(walker Walker) {
	walker(f)
	f.Top.Walk(walker)
	f.Bottom.Walk(walker)
}

type Index struct {
	Base Ast
	Up   Ast
	Down Ast
}

func (i *Index) ToMathMl(w *xmlWriter.XMLWriter, attr map[string]string) {
	if i.Up == nil {
		openWithAttr(w, "msub", attr)
		i.Base.ToMathMl(w, nil)
		i.Down.ToMathMl(w, nil)
		w.Close()
		return
	}
	if i.Down == nil {
		openWithAttr(w, "msup", attr)
		i.Base.ToMathMl(w, nil)
		i.Up.ToMathMl(w, nil)
		w.Close()
		return
	}
	openWithAttr(w, "msubsup", attr)
	i.Base.ToMathMl(w, nil)
	i.Down.ToMathMl(w, nil)
	i.Up.ToMathMl(w, nil)
	w.Close()
}

func (i *Index) Walk(walker Walker) {
	walker(i)
	i.Base.Walk(walker)
	if i.Up != nil {
		i.Up.Walk(walker)
	}
	if i.Down != nil {
		i.Down.Walk(walker)
	}
}

func NewIndex(base Ast, up Ast, down Ast) Ast {
	if i, ok := base.(*SimpleItem); ok && i.tok.kind == Operator {
		return &UnderOver{base: base, over: up, under: down}
	}
	return &Index{Base: base, Up: up, Down: down}
}

type UnderOver struct {
	base  Ast
	over  Ast
	under Ast
}

func (o *UnderOver) ToMathMl(w *xmlWriter.XMLWriter, attr map[string]string) {
	if o.over == nil {
		openWithAttr(w, "munder", attr)
		o.base.ToMathMl(w, nil)
		o.under.ToMathMl(w, nil)
		w.Close()
		return
	}
	if o.under == nil {
		openWithAttr(w, "mover", attr)
		o.base.ToMathMl(w, nil)
		o.over.ToMathMl(w, nil)
		w.Close()
		return
	}
	openWithAttr(w, "munderover", attr)
	o.base.ToMathMl(w, nil)
	o.under.ToMathMl(w, nil)
	o.over.ToMathMl(w, nil)
	w.Close()
}

func (o *UnderOver) Walk(walker Walker) {
	walker(o)
	o.base.Walk(walker)
	if o.over != nil {
		o.over.Walk(walker)
	}
	if o.under != nil {
		o.under.Walk(walker)
	}
}

type Sqrt struct {
	Inner Ast
}

func (s Sqrt) ToMathMl(w *xmlWriter.XMLWriter, attr map[string]string) {
	openWithAttr(w, "msqrt", attr)
	s.Inner.ToMathMl(w, nil)
	w.Close()
}

func (s Sqrt) Walk(walker Walker) {
	walker(s)
	s.Inner.Walk(walker)
}

type align int

const (
	left align = iota
	center
	right
)

type cellStyle struct {
	leftBorder  bool
	rightBorder bool
	align       align
}

func (s cellStyle) style() string {
	style := ""
	if s.leftBorder {
		style += "border-left:1px solid black;"
	}
	if s.rightBorder {
		style += "border-right:1px solid black;"
	}
	switch s.align {
	case left:
		style += "text-align:left;"
	case right:
		style += "text-align:right;"
	}
	return style
}

type Table struct {
	table [][]Ast
	style []cellStyle
}

func (t *Table) ToMathMl(w *xmlWriter.XMLWriter, attr map[string]string) {
	openWithAttr(w, "mtable", attr)
	topLine := false
	for _, row := range t.table {
		w.Open("mtr")
		if len(row) == 0 {
			topLine = true
		} else {
			for i, item := range row {
				style := ""
				if i < len(t.style) {
					style = t.style[i].style()
				}
				if topLine {
					style += "border-top:1px solid black;"
				}
				w.Open("mtd")
				if style != "" {
					w.Attr("style", style)
				}
				item.ToMathMl(w, nil)
				w.Close()
			}
			topLine = false
		}
		w.Close() // mtr
	}

	w.Close() //table
}

func (t *Table) Walk(walker Walker) {
	walker(t)
	for _, row := range t.table {
		for _, item := range row {
			item.Walk(walker)
		}
	}
}

func ParseLaTeX(latex string) (ast Ast, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("parse error: %v", r)
			}
		}
	}()
	p := &parser{tok: NewTokenizer(latex)}
	return p.ParseFull(), nil
}

func (p *parser) ParseFull() Ast {
	var li Line
	var sb strings.Builder
	for {
		r := p.tok.Read()
		if r == 0 {
			break
		}
		if r == '$' {
			if sb.Len() > 0 {
				li = append(li, Plain(sb.String()))
				sb.Reset()
			}
			ast := p.Parse(Dollar)
			li = append(li, MathMl{ast})
		} else {
			sb.WriteRune(r)
		}
	}
	if sb.Len() > 0 {
		li = append(li, Plain(sb.String()))
	}
	return li
}

func (p *parser) Parse(end Kind) Ast {
	a, _ := p.ParseFunc(func(t Token) bool { return t.kind == end })
	return a
}

func (p *parser) ParseFunc(isEnd func(token Token) bool) (Ast, Token) {
	var list []Ast
	for {
		tok := p.tok.NextToken()
		if isEnd(tok) {
			return NewRow(list...), tok
		}
		switch tok.kind {
		case Number:
			list = append(list, &SimpleItem{tok: tok})
		case Identifier:
			list = append(list, &SimpleItem{tok: tok})
		case Operator:
			list = append(list, &SimpleItem{tok: tok})
		case OpenParen:
			inner := p.Parse(CloseParen)
			list = append(list, NewRow(SimpleOperator("("), inner, SimpleOperator(")")))
		case Command:
			list = append(list, p.ParseCommand(tok.value))
		case Up:
			up := p.ParseInBrace()
			var down Ast
			if p.tok.PeekToken().kind == Down {
				p.tok.NextToken()
				down = p.ParseInBrace()
			}
			list[len(list)-1] = NewIndex(list[len(list)-1], up, down)
		case Down:
			down := p.ParseInBrace()
			var up Ast
			if p.tok.PeekToken().kind == Up {
				p.tok.NextToken()
				up = p.ParseInBrace()
			}
			list[len(list)-1] = NewIndex(list[len(list)-1], up, down)
		default:
			panic(fmt.Sprintf("unexpected token: %v", tok))
		}
	}
}

func (p *parser) ParseCommand(value string) Ast {
	switch value {
	case "frac":
		top := p.ParseInBrace()
		bottom := p.ParseInBrace()
		return &Fraction{top, bottom}
	case "pm":
		return SimpleOperator("&PlusMinus;")
	case "left":
		open := p.getBrace(OpenParen)
		inner, _ := p.ParseFunc(func(t Token) bool { return t.kind == Command && t.value == "right" })
		close := p.getBrace(CloseParen)
		return NewRow(open, inner, close)
	case "sqrt":
		return Sqrt{p.ParseInBrace()}
	case "vec":
		return &UnderOver{base: p.ParseInBrace(), over: addAttribute("mathsize", "75%", SimpleOperator("&rarr;"))}
	case "u":
		inner := p.ParseInBrace()
		inner.Walk(func(a Ast) {
			if si, ok := a.(*SimpleItem); ok {
				if si.tok.kind == Identifier {
					si.tok.kind = Number
				}
			}
		})
		return inner
	case "table":
		return p.parseTable()
	case "overset":
		over := p.ParseInBrace()
		return &UnderOver{base: p.ParseInBrace(), over: over}
	case "ds":
		inner := p.ParseInBrace()
		return addAttribute("displaystyle", "true", inner)
	case "underset":
		under := p.ParseInBrace()
		return &UnderOver{base: p.ParseInBrace(), under: under}
	case "sum":
		return SimpleOperator("&sum;")
	case "int":
		return SimpleOperator("&int;")
	case "oint":
		return SimpleOperator("&oint;")
	case "cdot":
		return SimpleOperator("&middot;")
	case "dif":
		return SimpleNumber("d")
	case "infty":
		return SimpleNumber("&infin;")
	case "rightarrow":
		return SimpleOperator("&rightarrow;")
	case "Rightarrow":
		return SimpleOperator("&Rightarrow;")
	case "leftarrow":
		return SimpleOperator("&leftarrow;")
	case "Leftarrow":
		return SimpleOperator("&Leftarrow;")
	case "sin":
		return SimpleIdent("sin")
	case "cos":
		return SimpleIdent("cos")
	case "tan":
		return SimpleIdent("tan")
	case "ln":
		return SimpleIdent("ln")
	case "lim":
		return SimpleIdent("lim")
	default:
		// assuming it's a symbol
		return SimpleIdent("&" + value + ";")
	}
}

func (p *parser) getBrace(brace Kind) Ast {
	b := p.tok.NextToken()
	if !(b.kind == brace || b.kind == Operator) {
		panic(fmt.Sprintf("unexpected token behind \\left or \\right: %v", b))
	}
	if b.value == "." {
		return &Empty{}
	}
	return SimpleOperator(b.value)
}

func SimpleIdent(s string) Ast {
	return &SimpleItem{tok: Token{kind: Identifier, value: s}}
}
func SimpleOperator(s string) Ast {
	return &SimpleItem{tok: Token{kind: Operator, value: s}}
}
func SimpleNumber(s string) Ast {
	return &SimpleItem{tok: Token{kind: Number, value: s}}
}

func (p *parser) ParseInBrace() Ast {
	n := p.tok.NextToken()
	if n.kind == Number || n.kind == Identifier {
		return &SimpleItem{tok: n}
	}
	if n.kind != OpenBrace {
		panic(fmt.Sprintf("unexpected token, expected {, got %v", n))
	}
	return p.Parse(CloseBrace)
}

func (p *parser) parseTable() Ast {
	n := p.tok.NextToken()
	var styles []cellStyle
	if n.kind == Operator && n.value == "[" {
		styles = p.parseTableDef()
		n = p.tok.NextToken()
	}

	if n.kind != OpenBrace {
		panic(fmt.Sprintf("unexpected token behind \\table, expected {, got %v", n))
	}

	var table [][]Ast
	var row []Ast
	for {
		a, tok := p.ParseFunc(func(t Token) bool { return t.kind == CloseBrace || t.kind == Ampersand || t.kind == Linefeed })
		row = append(row, a)
		switch tok.kind {
		case CloseBrace:
			if len(row) > 0 {
				table = append(table, row)
			}
			return &Table{table: table, style: styles}
		case Linefeed:
			if len(row) == 1 {
				if si, ok := row[0].(*SimpleItem); ok && si.tok.value == "-" {
					row = nil
				}
			}
			table = append(table, row)
			row = nil
		}
	}
}

func (p *parser) parseTableDef() []cellStyle {
	var styles []cellStyle
	complete := false
	cs := cellStyle{}
	for {
		switch p.tok.Read() {
		case ']':
			if complete {
				styles = append(styles, cs)
			}
			return styles
		case '|':
			if complete {
				cs.rightBorder = true
				styles = append(styles, cs)
				cs = cellStyle{}
				complete = false
			} else {
				cs.leftBorder = true
			}
		case 'l':
			if complete {
				styles = append(styles, cs)
				cs = cellStyle{}
			}
			complete = true
			cs.align = left
		case 'r':
			if complete {
				styles = append(styles, cs)
				cs = cellStyle{}
			}
			complete = true
			cs.align = right
		case 'c':
			if complete {
				styles = append(styles, cs)
				cs = cellStyle{}
			}
			complete = true
			cs.align = center
		}
	}
}
