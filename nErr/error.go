package nErr

import "fmt"

func CatchErr(err *error) {
	if r := recover(); r != nil {
		if e, ok := r.(error); ok {
			*err = e
		} else {
			*err = fmt.Errorf("unexpected panic: %v", r)
		}
	}
}

func Throw(err error) {
	if err != nil {
		panic(err)
	}
}

func ThrowArg[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}
