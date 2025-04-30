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
	expected := tp.VoltageGain(10)

	checkVG(t, expected, tp.GetY())
	checkVG(t, expected, tp.GetZ())
	checkVG(t, expected, tp.GetA())
	checkVG(t, expected, tp.GetH())
	checkVG(t, expected, tp.GetC())
}

func checkVG(t *testing.T, expected complex128, tp *TwoPort) {
	checkCmplx(t, expected, tp.GetY().VoltageGain(10))
	checkCmplx(t, expected, tp.GetZ().VoltageGain(10))
	checkCmplx(t, expected, tp.GetA().VoltageGain(10))
	checkCmplx(t, expected, tp.GetH().VoltageGain(10))
	checkCmplx(t, expected, tp.GetC().VoltageGain(10))
}

func TestTwoPort_CurrentGain(t *testing.T) {
	tp := NewTwoPort(1, 2+2i, 3, 4-1i, YParam)
	expected := tp.CurrentGain(10)

	checkCG(t, expected, tp.GetY())
	checkCG(t, expected, tp.GetZ())
	checkCG(t, expected, tp.GetA())
	checkCG(t, expected, tp.GetH())
	checkCG(t, expected, tp.GetC())
}

func checkCG(t *testing.T, expected complex128, tp *TwoPort) {
	checkCmplx(t, expected, tp.GetY().CurrentGain(10))
	checkCmplx(t, expected, tp.GetZ().CurrentGain(10))
	checkCmplx(t, expected, tp.GetA().CurrentGain(10))
	checkCmplx(t, expected, tp.GetH().CurrentGain(10))
	checkCmplx(t, expected, tp.GetC().CurrentGain(10))
}

func checkCmplx(t *testing.T, expected complex128, got complex128) {
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
