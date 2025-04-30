package twoPort

import (
	"bytes"
	"fmt"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"github.com/hneemann/parser2/value/export"
	"github.com/hneemann/parser2/value/export/xmlWriter"
)

type TpType int

func (t TpType) String() string {
	switch t {
	case HParam:
		return "H"
	case ZParam:
		return "Z"
	case YParam:
		return "Y"
	case CParam:
		return "C"
	case AParam:
		return "A"
	default:
		return "Unknown"
	}
}

const TwoPortValueType value.Type = 39

const (
	HParam TpType = iota
	ZParam
	YParam
	CParam
	AParam
)

type TwoPort struct {
	m11, m12, m21, m22 complex128
	typ                TpType
}

func (tp *TwoPort) ToHtml(st funcGen.Stack[value.Value], w *xmlWriter.XMLWriter) error {
	w.Open("math").
		Attr("xmlns", "http://www.w3.org/1998/Math/MathML")

	w.Open("mstyle").
		Attr("displaystyle", "true").
		Attr("scriptlevel", "0")

	w.Open("mrow").
		Open("mi").Write(tp.typ.String()).Close().
		Open("mo").Write("=").Close().
		Open("mo").Write("(").Close()

	w.Open("mtable").
		Open("mtr").
		Open("mtd").Write(cmplxStr(tp.m11)).Close().
		Open("mtd").Write(cmplxStr(tp.m12)).Close().
		Close().
		Open("mtr").
		Open("mtd").Write(cmplxStr(tp.m21)).Close().
		Open("mtd").Write(cmplxStr(tp.m22)).Close().
		Close().
		Close()

	w.Open("mo").Write(")").Close()

	w.Close()
	w.Close()
	w.Close()
	return nil
}

func cmplxStr(c complex128) string {
	if imag(c) == 0 {
		return fmt.Sprintf("%g", real(c))
	}
	return fmt.Sprintf("%g", c)
}

var _ export.ToHtmlInterface = &TwoPort{}

func (tp *TwoPort) ToList() (*value.List, bool) {
	return nil, false
}

func (tp *TwoPort) ToMap() (value.Map, bool) {
	return value.Map{}, false
}

func (tp *TwoPort) ToInt() (int, bool) {
	return 0, false
}

func (tp *TwoPort) ToFloat() (float64, bool) {
	return 0, false
}

func (tp *TwoPort) ToString(st funcGen.Stack[value.Value]) (string, error) {
	var buf bytes.Buffer
	switch tp.typ {
	case YParam:
		buf.WriteString("Y")
	case AParam:
		buf.WriteString("A")
	case HParam:
		buf.WriteString("H")
	case ZParam:
		buf.WriteString("Z")
	case CParam:
		buf.WriteString("C")
	}
	buf.WriteString("=(")
	buf.WriteString(fmt.Sprintf("%f", tp.m11))
	buf.WriteString(",")
	buf.WriteString(fmt.Sprintf("%f", tp.m12))
	buf.WriteString(";")
	buf.WriteString(fmt.Sprintf("%f", tp.m21))
	buf.WriteString(",")
	buf.WriteString(fmt.Sprintf("%f", tp.m22))
	buf.WriteString(")")
	return buf.String(), nil
}

func (tp *TwoPort) ToBool() (bool, bool) {
	return false, false
}

func (tp *TwoPort) ToClosure() (funcGen.Function[value.Value], bool) {
	return funcGen.Function[value.Value]{}, false
}

func (tp *TwoPort) GetType() value.Type {
	return TwoPortValueType
}

func NewTwoPort(m11, m12, m21, m22 complex128, typ TpType) *TwoPort {
	return &TwoPort{
		m11: m11,
		m12: m12,
		m21: m21,
		m22: m22,
		typ: typ,
	}
}

func NewSeries(z complex128) *TwoPort {
	return &TwoPort{
		m11: 1,
		m12: z,
		m21: 0,
		m22: 1,
		typ: AParam,
	}
}

func NewShunt(z complex128) *TwoPort {
	return &TwoPort{
		m11: 1,
		m12: 0,
		m21: 1 / z,
		m22: 1,
		typ: AParam,
	}
}

func Cascade(a ...*TwoPort) *TwoPort {
	if len(a) == 0 {
		return nil
	}
	tp := a[0]
	for i := 1; i < len(a); i++ {
		tp = tp.cascade(a[i])
	}
	return tp
}

func (tp *TwoPort) det() complex128 {
	return tp.m11*tp.m22 - tp.m12*tp.m21
}

func (tp *TwoPort) GetY() *TwoPort {
	switch tp.typ {
	case YParam:
		return tp
	case AParam:
		return TwoPort{tp.m22, -tp.det(), -1, tp.m11, YParam}.div(tp.m12)
	case HParam:
		return TwoPort{1, -tp.m12, tp.m21, tp.det(), YParam}.div(tp.m11)
	case ZParam:
		return TwoPort{tp.m22, -tp.m12, -tp.m21, tp.m11, YParam}.div(tp.det())
	case CParam:
		return TwoPort{tp.det(), tp.m12, -tp.m21, 1, YParam}.div(tp.m22)
	}
	panic("Invalid type")
}

func (tp *TwoPort) GetZ() *TwoPort {
	switch tp.typ {
	case YParam:
		return TwoPort{tp.m22, -tp.m12, -tp.m21, tp.m11, ZParam}.div(tp.det())
	case AParam:
		return TwoPort{tp.m11, tp.det(), 1, tp.m22, ZParam}.div(tp.m21)
	case HParam:
		return TwoPort{tp.det(), tp.m12, -tp.m21, 1, ZParam}.div(tp.m22)
	case ZParam:
		return tp
	case CParam:
		return TwoPort{1, -tp.m12, tp.m21, tp.det(), ZParam}.div(tp.m11)
	}
	panic("Invalid type")
}

func (tp *TwoPort) GetA() *TwoPort {
	switch tp.typ {
	case YParam:
		return TwoPort{-tp.m22, -1, -tp.det(), -tp.m11, AParam}.div(tp.m21)
	case AParam:
		return tp
	case HParam:
		return TwoPort{-tp.det(), -tp.m11, -tp.m22, -1, AParam}.div(tp.m21)
	case ZParam:
		return TwoPort{tp.m11, tp.det(), 1, tp.m22, AParam}.div(tp.m21)
	case CParam:
		return TwoPort{1, tp.m22, tp.m11, tp.det(), AParam}.div(tp.m21)
	}
	panic("Invalid type")
}

func (tp *TwoPort) GetH() *TwoPort {
	switch tp.typ {
	case YParam:
		return TwoPort{1, -tp.m12, tp.m21, tp.det(), HParam}.div(tp.m11)
	case AParam:
		return TwoPort{tp.m12, tp.det(), -1, tp.m21, HParam}.div(tp.m22)
	case HParam:
		return tp
	case ZParam:
		return TwoPort{tp.det(), tp.m12, -tp.m21, 1, HParam}.div(tp.m22)
	case CParam:
		return TwoPort{tp.m22, -tp.m12, -tp.m21, tp.m11, HParam}.div(tp.det())
	}
	panic("Invalid type")
}

func (tp *TwoPort) GetC() *TwoPort {
	switch tp.typ {
	case YParam:
		return TwoPort{tp.det(), tp.m12, -tp.m21, 1, CParam}.div(tp.m22)
	case AParam:
		return TwoPort{tp.m21, -tp.det(), 1, tp.m12, CParam}.div(tp.m11)
	case HParam:
		return TwoPort{tp.m22, -tp.m12, -tp.m21, tp.m11, CParam}.div(tp.det())
	case ZParam:
		return TwoPort{1, -tp.m12, tp.m21, tp.det(), CParam}.div(tp.m11)
	case CParam:
		return tp
	}
	panic("Invalid type")
}

func (tp *TwoPort) VoltageGain(load complex128) complex128 {
	switch tp.typ {
	case YParam:
		return -tp.m21 / (tp.m22 + 1/load)
	case AParam:
		return 1 / (tp.m11 + tp.m12/load)
	case HParam:
		return -tp.m21 / (tp.det() + tp.m11/load)
	case ZParam:
		return tp.m21 / (tp.m11 + tp.det()/load)
	case CParam:
		return tp.m21 / (1 + tp.m22/load)
	}
	panic("Invalid type")
}

func (tp *TwoPort) CurrentGain(load complex128) complex128 {
	switch tp.typ {
	case YParam:
		return tp.m21 / load / (tp.det() + tp.m11/load)
	case AParam:
		return -1 / load / (tp.m21 + tp.m22/load)
	case HParam:
		return tp.m21 / load / (tp.m22 + 1/load)
	case ZParam:
		return -tp.m21 / load / (1 + tp.m22/load)
	case CParam:
		return -tp.m21 / load / (tp.m11 + tp.det()/load)
	}
	panic("Invalid type")
}

func (tp TwoPort) div(d complex128) *TwoPort {
	tp.m11 = tp.m11 / d
	tp.m12 = tp.m12 / d
	tp.m21 = tp.m21 / d
	tp.m22 = tp.m22 / d
	return &tp
}

func (tp *TwoPort) cascade(port *TwoPort) *TwoPort {
	a := tp.GetA()
	b := port.GetA()

	return &TwoPort{
		m11: a.m11*b.m11 + a.m12*b.m21,
		m12: a.m11*b.m12 + a.m12*b.m22,
		m21: a.m21*b.m11 + a.m22*b.m21,
		m22: a.m21*b.m12 + a.m22*b.m22,
		typ: AParam,
	}
}
