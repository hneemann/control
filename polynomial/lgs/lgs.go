package lgs

import (
	"errors"
	"fmt"
	"math"
	"math/cmplx"
)

type ComplexVector []complex128

func (v ComplexVector) String() string {
	s := "["
	for i, val := range v {
		if i > 0 {
			s += " "
		}
		if math.Abs(imag(val)) < eps {
			s += fmt.Sprintf("%v", real(val))
		} else {
			s += fmt.Sprintf("%v", val)
		}
	}
	s += "]"
	return s
}

type ComplexMatrix []ComplexVector

func (m ComplexMatrix) String() string {
	s := "["
	for i, row := range m {
		if i > 0 {
			s += "\n "
		}
		s += fmt.Sprintf("%v", row)
	}
	s += "]"
	return s
}

func SolveLinearSystem(A ComplexMatrix, b ComplexVector) (ComplexVector, error) {
	rowIndex := makeIndex(A)
	for col := 0; col < len(A[0])-1; col++ {
		err := doStep(A, b, col, rowIndex)
		if err != nil {
			return nil, err
		}
	}

	result := make(ComplexVector, len(b))
	for i := len(b) - 1; i >= 0; i-- {
		sum := complex(0, 0)
		for j := i + 1; j < len(b); j++ {
			sum += A[rowIndex[i]][j] * result[j]
		}
		div := A[rowIndex[i]][i]
		if cmplx.Abs(div) < eps {
			return nil, errors.New("matrix is singular")
		}
		result[i] = (b[rowIndex[i]] - sum) / div
	}
	return result, nil
}

func makeIndex(A ComplexMatrix) []int {
	rowIndex := make([]int, len(A))
	for i := range rowIndex {
		rowIndex[i] = i
	}
	return rowIndex
}

const eps = 1e-10

func doStep(A ComplexMatrix, b ComplexVector, col int, rowIndex []int) error {
	pivotRow := col
	for row := col + 1; row < len(A); row++ {
		if cmplx.Abs(A[rowIndex[row]][col]) > cmplx.Abs(A[rowIndex[pivotRow]][col]) {
			pivotRow = row
		}
	}
	rowIndex[col], rowIndex[pivotRow] = rowIndex[pivotRow], rowIndex[col]

	if cmplx.Abs(A[rowIndex[col]][col]) < eps {
		return errors.New("matrix is singular")
	}

	for row := col + 1; row < len(A); row++ {
		factor := A[rowIndex[row]][col] / A[rowIndex[col]][col]
		for c := col; c < len(A[0]); c++ {
			A[rowIndex[row]][c] = A[rowIndex[row]][c] - factor*A[rowIndex[col]][c]
		}
		A[rowIndex[row]][col] = 0
		b[rowIndex[row]] = b[rowIndex[row]] - factor*b[rowIndex[col]]
	}
	return nil
}
