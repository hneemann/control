package nErr

import "fmt"

type pError struct {
	e error
}

func (p pError) String() string {
	return fmt.Sprint(p.e)
}

func CatchErr(err *error) {
	if r := recover(); r != nil {
		if e, ok := r.(pError); ok {
			*err = e.e
		} else {
			panic(r)
		}
	}
}

func Try(err error) {
	if err != nil {
		panic(pError{err})
	}
}

func TryArg[T any](t T, err error) T {
	if err != nil {
		panic(pError{err})
	}
	return t
}

func TryArgs[A any, B any](a A, b B, err error) (A, B) {
	if err != nil {
		panic(pError{err})
	}
	return a, b
}
