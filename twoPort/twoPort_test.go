package twoPort

import (
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"github.com/stretchr/testify/assert"
	"math"
	"math/cmplx"
	"testing"
)

func TestTwoPort_VoltageGain(t *testing.T) {
	tp := NewTwoPort(1, 2+2i, 3, 4-1i, YParam)
	check(t, tp, func(tp *TwoPort) complex128 {
		return tp.VoltageGain(10)
	})
}

func TestTwoPort_CurrentGain(t *testing.T) {
	tp := NewTwoPort(1, 2+2i, 3, 4-1i, ZParam)
	check(t, tp, func(tp *TwoPort) complex128 {
		return tp.CurrentGain(10)
	})
}

func TestTwoPort_InputImpedance(t *testing.T) {
	tp := NewTwoPort(1, 2+2i, 3, 4-1i, HParam)
	check(t, tp, func(tp *TwoPort) complex128 {
		return tp.InputImpedance(10)
	})
}

func TestTwoPort_OutputImpedance(t *testing.T) {
	tp := NewTwoPort(1, 2+2i, 3, 4-1i, AParam)
	check(t, tp, func(tp *TwoPort) complex128 {
		return tp.OutputImpedance(10)
	})
}

func check(t *testing.T, tp *TwoPort, f func(tp *TwoPort) complex128) {
	expected := f(tp)
	checkExp(t, tp.GetZ(), expected, f)
	checkExp(t, tp.GetY(), expected, f)
	checkExp(t, tp.GetH(), expected, f)
	checkExp(t, tp.GetC(), expected, f)
	checkExp(t, tp.GetA(), expected, f)
}

func checkExp(t *testing.T, tp *TwoPort, expected complex128, f func(tp *TwoPort) complex128) {
	got := f(tp)
	assert.InDelta(t, real(expected), real(got), 1e-13)
	assert.InDelta(t, imag(expected), imag(got), 1e-13)
}

func TestTwoPort_Travo(t *testing.T) {
	sum := Cascade(
		NewSeries(2000),
		NewShunt(complex(0, 2*math.Pi*50*10)),
		NewTwoPort(0, 10, -10, 0, HParam),
		NewSeries(5),
		NewShunt(complex(0, -1/(2*math.Pi*50*400e-6))),
	)

	str, err := sum.ToString(funcGen.NewEmptyStack[value.Value]())
	assert.NoError(t, err)
	assert.EqualValues(t, "A=((14.000000+25.049729i),(250.000000-31.830989i);(0.002000+0.009383i),(0.100000-0.015915i))", str)

	outputVoltage := cmplx.Abs(sum.VoltageGain(500)) * 240

	assert.InDelta(t, 8.3, outputVoltage, 1e-2)
}
