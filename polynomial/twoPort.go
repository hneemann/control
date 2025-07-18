package polynomial

import (
	"errors"
	"fmt"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"github.com/hneemann/parser2/value/export"
	"github.com/hneemann/parser2/value/export/xmlWriter"
)

type TpType rune

func (t TpType) String() string {
	return string(t)
}

const (
	HParam TpType = 'H'
	ZParam TpType = 'Z'
	YParam TpType = 'Y'
	CParam TpType = 'C'
	AParam TpType = 'A'
)

type TwoPort struct {
	m11, m12, m21, m22 complex128
	typ                TpType
}

var _ export.ToHtmlInterface = &TwoPort{}

func (tp *TwoPort) Get(key string) (value.Value, bool) {
	switch key {
	case "m11":
		return Complex(tp.m11), true
	case "m12":
		return Complex(tp.m12), true
	case "m21":
		return Complex(tp.m21), true
	case "m22":
		return Complex(tp.m22), true
	case "type":
		return value.String(tp.typ), true
	}
	return nil, false
}

func (tp *TwoPort) Iter(yield func(key string, v value.Value) bool) bool {
	if !yield("m11", Complex(tp.m11)) {
		return false
	}
	if !yield("m12", Complex(tp.m12)) {
		return false
	}
	if !yield("m21", Complex(tp.m21)) {
		return false
	}
	if !yield("m22", Complex(tp.m22)) {
		return false
	}
	if !yield("type", value.String(tp.typ)) {
		return false
	}
	return true
}

func (tp *TwoPort) Size() int {
	return 5
}

func (tp *TwoPort) ToHtml(_ funcGen.Stack[value.Value], w *xmlWriter.XMLWriter) error {
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
		Open("mtd").Write(Complex(tp.m11).Format(6)).Close().
		Open("mtd").Write(Complex(tp.m12).Format(6)).Close().
		Close().
		Open("mtr").
		Open("mtd").Write(Complex(tp.m21).Format(6)).Close().
		Open("mtd").Write(Complex(tp.m22).Format(6)).Close().
		Close().
		Close()

	w.Open("mo").Write(")").Close()

	w.Close()
	w.Close()
	w.Close()
	return nil
}

func (tp *TwoPort) ToList() (*value.List, bool) {
	return nil, false
}

func (tp *TwoPort) ToMap() (value.Map, bool) {
	return value.NewMap(tp), true
}

func (tp *TwoPort) ToInt() (int, bool) {
	return 0, false
}

func (tp *TwoPort) ToFloat() (float64, bool) {
	return 0, false
}

func (tp *TwoPort) String() string {
	return tp.typ.String() + "=(" +
		Complex(tp.m11).String() + ", " +
		Complex(tp.m12).String() + "; " +
		Complex(tp.m21).String() + ", " +
		Complex(tp.m22).String() + ")"
}

func (tp *TwoPort) ToString(_ funcGen.Stack[value.Value]) (string, error) {
	return tp.String(), nil
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

func Cascade(a ...*TwoPort) (*TwoPort, error) {
	if len(a) == 0 {
		return nil, errors.New("no ports to cascade")
	}
	tp := a[0]
	for i := 1; i < len(a); i++ {
		var err error
		tp, err = tp.Cascade(a[i])
		if err != nil {
			return nil, err
		}
	}
	return tp, nil
}

func (tp *TwoPort) det() complex128 {
	return tp.m11*tp.m22 - tp.m12*tp.m21
}

func (tp *TwoPort) GetY() (*TwoPort, error) {
	switch tp.typ {
	case YParam:
		return tp, nil
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

func (tp *TwoPort) GetZ() (*TwoPort, error) {
	switch tp.typ {
	case YParam:
		return TwoPort{tp.m22, -tp.m12, -tp.m21, tp.m11, ZParam}.div(tp.det())
	case AParam:
		return TwoPort{tp.m11, tp.det(), 1, tp.m22, ZParam}.div(tp.m21)
	case HParam:
		return TwoPort{tp.det(), tp.m12, -tp.m21, 1, ZParam}.div(tp.m22)
	case ZParam:
		return tp, nil
	case CParam:
		return TwoPort{1, -tp.m12, tp.m21, tp.det(), ZParam}.div(tp.m11)
	}
	panic("Invalid type")
}

func (tp *TwoPort) GetA() (*TwoPort, error) {
	switch tp.typ {
	case YParam:
		return TwoPort{-tp.m22, -1, -tp.det(), -tp.m11, AParam}.div(tp.m21)
	case AParam:
		return tp, nil
	case HParam:
		return TwoPort{-tp.det(), -tp.m11, -tp.m22, -1, AParam}.div(tp.m21)
	case ZParam:
		return TwoPort{tp.m11, tp.det(), 1, tp.m22, AParam}.div(tp.m21)
	case CParam:
		return TwoPort{1, tp.m22, tp.m11, tp.det(), AParam}.div(tp.m21)
	}
	panic("Invalid type")
}

func (tp *TwoPort) GetH() (*TwoPort, error) {
	switch tp.typ {
	case YParam:
		return TwoPort{1, -tp.m12, tp.m21, tp.det(), HParam}.div(tp.m11)
	case AParam:
		return TwoPort{tp.m12, tp.det(), -1, tp.m21, HParam}.div(tp.m22)
	case HParam:
		return tp, nil
	case ZParam:
		return TwoPort{tp.det(), tp.m12, -tp.m21, 1, HParam}.div(tp.m22)
	case CParam:
		return TwoPort{tp.m22, -tp.m12, -tp.m21, tp.m11, HParam}.div(tp.det())
	}
	panic("Invalid type")
}

func (tp *TwoPort) GetC() (*TwoPort, error) {
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
		return tp, nil
	}
	panic("Invalid type")
}

func (tp *TwoPort) VoltageGain(load complex128) complex128 {
	if load == 0 {
		return 0
	}
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

func (tp *TwoPort) VoltageGainOpen() complex128 {
	switch tp.typ {
	case YParam:
		return -tp.m21 / tp.m22
	case AParam:
		return 1 / tp.m11
	case HParam:
		return -tp.m21 / tp.det()
	case ZParam:
		return tp.m21 / tp.m11
	case CParam:
		return tp.m21
	}
	panic("Invalid type")
}

func (tp *TwoPort) CurrentGain(load complex128) complex128 {
	if load == 0 {
		return tp.CurrentGainShort()
	}
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

func (tp *TwoPort) CurrentGainShort() complex128 {
	switch tp.typ {
	case YParam:
		return tp.m21 / tp.m11
	case AParam:
		return -1 / tp.m22
	case HParam:
		return tp.m21
	case ZParam:
		return -tp.m21 / tp.m22
	case CParam:
		return -tp.m21 / tp.det()
	}
	panic("Invalid type")
}

func (tp *TwoPort) InputImpedance(load complex128) complex128 {
	if load == 0 {
		return tp.InputImpedanceShort()
	}
	switch tp.typ {
	case YParam:
		return (tp.m22 + 1/load) / (tp.det() + tp.m11/load)
	case ZParam:
		return (tp.m11 + tp.det()/load) / (1 + tp.m22/load)
	case HParam:
		return (tp.det() + tp.m11/load) / (tp.m22 + 1/load)
	case CParam:
		return (1 + tp.m22/load) / (tp.m11 + tp.det()/load)
	case AParam:
		return (tp.m11 + tp.m12/load) / (tp.m21 + tp.m22/load)
	}
	panic("Invalid type")
}

func (tp *TwoPort) InputImpedanceOpen() complex128 {
	switch tp.typ {
	case YParam:
		return (tp.m22) / tp.det()
	case ZParam:
		return tp.m11
	case HParam:
		return tp.det() / tp.m22
	case CParam:
		return 1 / tp.m11
	case AParam:
		return tp.m11 / tp.m21
	}
	panic("Invalid type")
}

func (tp *TwoPort) InputImpedanceShort() complex128 {
	switch tp.typ {
	case YParam:
		return 1 / tp.m11
	case ZParam:
		return tp.det() / tp.m22
	case HParam:
		return tp.m11
	case CParam:
		return tp.m22 / tp.det()
	case AParam:
		return tp.m12 / tp.m22
	}
	panic("Invalid type")
}

func (tp *TwoPort) OutputImpedance(load complex128) complex128 {
	if load == 0 {
		return tp.OutputImpedanceShort()
	}
	switch tp.typ {
	case YParam:
		return (tp.m11 + 1/load) / (tp.det() + tp.m22/load)
	case ZParam:
		return (tp.m22 + tp.det()/load) / (1 + tp.m11/load)
	case HParam:
		return (1 + tp.m11/load) / (tp.m22 + tp.det()/load)
	case CParam:
		return (tp.det() + tp.m22/load) / (tp.m11 + 1/load)
	case AParam:
		return (tp.m22 + tp.m12/load) / (tp.m21 + tp.m11/load)
	}
	panic("Invalid type")
}

func (tp *TwoPort) OutputImpedanceOpen() complex128 {
	switch tp.typ {
	case YParam:
		return tp.m11 / tp.det()
	case ZParam:
		return tp.m22
	case HParam:
		return 1 / tp.m22
	case CParam:
		return tp.det() / tp.m11
	case AParam:
		return tp.m22 / tp.m21
	}
	panic("Invalid type")
}

func (tp *TwoPort) OutputImpedanceShort() complex128 {
	switch tp.typ {
	case YParam:
		return 1 / tp.m22
	case ZParam:
		return tp.det() / tp.m11
	case HParam:
		return tp.m11 / tp.det()
	case CParam:
		return tp.m22
	case AParam:
		return tp.m12 / tp.m11
	}
	panic("Invalid type")
}

func (tp TwoPort) div(d complex128) (*TwoPort, error) {
	if d == 0 {
		return nil, fmt.Errorf("cannot create parameters: division by zero")
	}
	tp.m11 = tp.m11 / d
	tp.m12 = tp.m12 / d
	tp.m21 = tp.m21 / d
	tp.m22 = tp.m22 / d
	return &tp, nil
}

func (tp *TwoPort) add(o *TwoPort) (*TwoPort, error) {
	if tp.typ != o.typ {
		return nil, fmt.Errorf("cannot add two ports of different types: %s and %s", tp.typ, o.typ)
	}
	return &TwoPort{
		m11: tp.m11 + o.m11,
		m12: tp.m12 + o.m12,
		m21: tp.m21 + o.m21,
		m22: tp.m22 + o.m22,
		typ: tp.typ,
	}, nil
}

func (tp *TwoPort) Cascade(port *TwoPort) (*TwoPort, error) {
	a, err := tp.GetA()
	if err != nil {
		return nil, err
	}
	b, err := port.GetA()
	if err != nil {
		return nil, err
	}

	return &TwoPort{
		m11: a.m11*b.m11 + a.m12*b.m21,
		m12: a.m11*b.m12 + a.m12*b.m22,
		m21: a.m21*b.m11 + a.m22*b.m21,
		m22: a.m21*b.m12 + a.m22*b.m22,
		typ: AParam,
	}, nil
}

func (tp *TwoPort) Series(o *TwoPort) (*TwoPort, error) {
	z1, err := tp.GetZ()
	if err != nil {
		return nil, err
	}
	z2, err := o.GetZ()
	if err != nil {
		return nil, err
	}
	return z1.add(z2)
}

func (tp *TwoPort) Parallel(o *TwoPort) (*TwoPort, error) {
	y1, err := tp.GetY()
	if err != nil {
		return nil, err
	}
	y2, err := o.GetY()
	if err != nil {
		return nil, err
	}
	return y1.add(y2)
}

func (tp *TwoPort) SeriesParallel(o *TwoPort) (*TwoPort, error) {
	h1, err := tp.GetH()
	if err != nil {
		return nil, err
	}
	h2, err := o.GetH()
	if err != nil {
		return nil, err
	}
	return h1.add(h2)
}

func (tp *TwoPort) ParallelSeries(o *TwoPort) (*TwoPort, error) {
	c1, err := tp.GetC()
	if err != nil {
		return nil, err
	}
	c2, err := o.GetC()
	if err != nil {
		return nil, err
	}
	return c1.add(c2)
}
