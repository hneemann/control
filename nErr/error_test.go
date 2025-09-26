package nErr

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func testFunc() (err error) {
	defer CatchErr(&err)
	Try(fmt.Errorf("throw error"))
	return nil
}

func TestError(t *testing.T) {
	err := testFunc()
	assert.Error(t, err)
	assert.Equal(t, "throw error", err.Error())
}

func testFunc2() (err error) {
	defer CatchErr(&err)

	panic("test panic")

	return nil
}

func TestErrorPanic(t *testing.T) {
	defer func() {
		r := recover()
		assert.NotNil(t, r)
		assert.Equal(t, "test panic", r)
	}()
	testFunc2()
	assert.Fail(t, "should not reach here")
}

func testInnerFuncArg(e bool) (int, error) {
	if e {
		return 1, fmt.Errorf("throw error")
	} else {
		return 2, nil
	}
}

func testFuncArg() (i int, err error) {
	defer CatchErr(&err)
	j := TryArg(testInnerFuncArg(true))
	return j, nil
}

func testFuncArg2() (i int, err error) {
	defer CatchErr(&err)
	j := TryArg(testInnerFuncArg(false))
	return j, nil
}

func TestErrorArg(t *testing.T) {
	i, err := testFuncArg()
	assert.Error(t, err)
	assert.Equal(t, 0, i)
	assert.Equal(t, "throw error", err.Error())

	i, err = testFuncArg2()
	assert.NoError(t, err)
	assert.Equal(t, 2, i)
}
