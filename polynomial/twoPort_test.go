package polynomial

import (
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

func TestTwoPort_VoltageGainOpen(t *testing.T) {
	tp := NewTwoPort(1, 2+2i, 3, 4-1i, YParam)

	g1 := tp.VoltageGain(1e6)
	g2 := tp.VoltageGainOpen()
	assert.InDelta(t, real(g1), real(g2), 1e-6)
	assert.InDelta(t, imag(g1), imag(g2), 1e-6)

	check(t, tp, func(tp *TwoPort) complex128 {
		return tp.VoltageGainOpen()
	})
}

func TestTwoPort_CurrentGain(t *testing.T) {
	tp := NewTwoPort(1, 2+2i, 3, 4-1i, ZParam)
	check(t, tp, func(tp *TwoPort) complex128 {
		return tp.CurrentGain(10)
	})
}

func TestTwoPort_CurrentGainShort(t *testing.T) {
	tp := NewTwoPort(1, 2+2i, 3, 4-1i, ZParam)

	g1 := tp.CurrentGain(1e-6)
	g2 := tp.CurrentGainShort()
	assert.InDelta(t, real(g1), real(g2), 1e-6)
	assert.InDelta(t, imag(g1), imag(g2), 1e-6)

	check(t, tp, func(tp *TwoPort) complex128 {
		return tp.CurrentGainShort()
	})
}

func TestTwoPort_InputImpedance(t *testing.T) {
	tp := NewTwoPort(1, 2+2i, 3, 4-1i, HParam)
	check(t, tp, func(tp *TwoPort) complex128 {
		return tp.InputImpedance(10)
	})
}

func TestTwoPort_InputImpedanceOpen(t *testing.T) {
	tp := NewTwoPort(1, 2+2i, 3, 4-1i, HParam)

	g1 := tp.InputImpedance(1e+6)
	g2 := tp.InputImpedanceOpen()
	assert.InDelta(t, real(g1), real(g2), 1e-6)
	assert.InDelta(t, imag(g1), imag(g2), 1e-6)

	check(t, tp, func(tp *TwoPort) complex128 {
		return tp.InputImpedanceOpen()
	})
}

func TestTwoPort_InputImpedanceShort(t *testing.T) {
	tp := NewTwoPort(1, 2+2i, 3, 4-1i, HParam)

	g1 := tp.InputImpedance(1e-7)
	g2 := tp.InputImpedanceShort()
	assert.InDelta(t, real(g1), real(g2), 1e-6)
	assert.InDelta(t, imag(g1), imag(g2), 1e-6)

	check(t, tp, func(tp *TwoPort) complex128 {
		return tp.InputImpedanceShort()
	})
}

func TestTwoPort_OutputImpedance(t *testing.T) {
	tp := NewTwoPort(1, 2+2i, 3, 4-1i, AParam)
	check(t, tp, func(tp *TwoPort) complex128 {
		return tp.OutputImpedance(10)
	})
}

func TestTwoPort_OutputImpedanceOpen(t *testing.T) {
	tp := NewTwoPort(1, 2+2i, 3, 4-1i, AParam)

	g1 := tp.OutputImpedance(1e+6)
	g2 := tp.OutputImpedanceOpen()
	assert.InDelta(t, real(g1), real(g2), 1e-6)
	assert.InDelta(t, imag(g1), imag(g2), 1e-6)

	check(t, tp, func(tp *TwoPort) complex128 {
		return tp.OutputImpedanceOpen()
	})
}

func TestTwoPort_OutputImpedanceShort(t *testing.T) {
	tp := NewTwoPort(1, 2+2i, 3, 4-1i, AParam)

	g1 := tp.OutputImpedance(1e-7)
	g2 := tp.OutputImpedanceShort()
	assert.InDelta(t, real(g1), real(g2), 1e-6)
	assert.InDelta(t, imag(g1), imag(g2), 1e-6)

	check(t, tp, func(tp *TwoPort) complex128 {
		return tp.OutputImpedanceShort()
	})
}

func check(t *testing.T, tp *TwoPort, f func(tp *TwoPort) complex128) {
	expected := f(tp)
	checkExp(t, Must(tp.GetZ()), expected, f)
	checkExp(t, Must(tp.GetY()), expected, f)
	checkExp(t, Must(tp.GetH()), expected, f)
	checkExp(t, Must(tp.GetC()), expected, f)
	checkExp(t, Must(tp.GetA()), expected, f)
}

func checkExp(t *testing.T, tp *TwoPort, expected complex128, f func(tp *TwoPort) complex128) {
	got := f(tp)
	assert.InDelta(t, real(expected), real(got), 1e-13)
	assert.InDelta(t, imag(expected), imag(got), 1e-13)
}

func TestTwoPort_Travo(t *testing.T) {
	sum, err := Cascade(
		NewSeries(2000),
		NewShunt(complex(0, 2*math.Pi*50*10)),
		NewTwoPort(0, 10, -10, 0, HParam),
		NewSeries(5),
		NewShunt(complex(0, -1/(2*math.Pi*50*400e-6))),
	)
	assert.NoError(t, err)

	outputVoltage := cmplx.Abs(sum.VoltageGain(500)) * 240

	assert.InDelta(t, 8.3, outputVoltage, 1e-2)
}
