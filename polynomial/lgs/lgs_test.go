package lgs

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_doLgsStep(t *testing.T) {
	tests := []struct {
		name  string
		A     ComplexMatrix
		b     ComplexVector
		AList []ComplexMatrix
		BList []ComplexVector
	}{
		{name: "simple", A: ComplexMatrix{{1, 1, -1}, {10, 20, 0}, {0, 20, 40}}, b: ComplexVector{0, 4, 8},
			AList: []ComplexMatrix{
				{{0, -1, -1}, {10, 20, 0}, {0, 20, 40}},
				{{0, 0, 1}, {10, 20, 0}, {0, 20, 40}},
			}, BList: []ComplexVector{
				{-0.4, 4, 8},
				{0, 4, 8},
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			rowIndex := makeIndex(tt.A)
			for col := 0; col < len(tt.A[0])-1; col++ {
				err := doStep(tt.A, tt.b, col, rowIndex)
				fmt.Println("A:", tt.A)
				fmt.Println("b:", tt.b)
				assert.NoError(t, err)
				assert.EqualValues(t, tt.AList[col], tt.A)
				assert.EqualValues(t, tt.BList[col], tt.b)
			}

		})
	}
}

func TestSolveLinearSystem(t *testing.T) {
	tests := []struct {
		name string
		A    ComplexMatrix
		b    ComplexVector
		err  bool
		want ComplexVector
	}{
		{name: "simple", A: ComplexMatrix{{1, 1, -1}, {10, -20, 0}, {0, 20, 40}}, b: ComplexVector{0, 4, 8}, want: ComplexVector{0.2857142857142856, -0.057142857142857204, 0.2285714285714286}},
		{name: "4", A: ComplexMatrix{{2, 2, 3, 4}, {2, 4, 4, 5}, {3, 4, 6, 6}, {4, 5, 6, 8}}, b: ComplexVector{5, 6, 7, 8}, want: ComplexVector{-5, -2, 1, 4}},
		{name: "vl", A: ComplexMatrix{{70, -45, 0}, {-45, 120, -35}, {0, -35, 65}}, b: ComplexVector{6, -6, 12}, want: ComplexVector{1632.0 / 13145.0, 786.0 / 13145.0, 570.0 / 2629.0}},
		{name: "vl2", A: ComplexMatrix{{30, -10, -10}, {-10, 30, -10}, {-10, -10, 30}}, b: ComplexVector{30, 50, 0}, want: ComplexVector{2.75, 3.25, 2}},
		{name: "err1", A: ComplexMatrix{{1, 1, 1}, {2, 2, 2}, {3, 3, 3}}, b: ComplexVector{1, 2, 3}, err: true},
		{name: "err2", A: ComplexMatrix{{1, 1, 1}, {2, 2, 2}, {3, 4, 5}}, b: ComplexVector{1, 2, 3}, err: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SolveLinearSystem(tt.A, tt.b)
			if !tt.err {
				assert.NoError(t, err)
				checkVectors(t, tt.want, got)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func checkVectors(t *testing.T, want ComplexVector, got ComplexVector) {
	for i := range want {
		assert.InDelta(t, real(want[i]), real(got[i]), 1e-6, "real part of index %d", i)
		assert.InDelta(t, imag(want[i]), imag(got[i]), 1e-6, "imag part of index %d", i)
	}
}
