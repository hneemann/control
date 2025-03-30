package polynomial

import (
	"fmt"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
)

const LinearValueType value.Type = 19

func (l *Linear) ToList() (*value.List, bool) {
	return nil, false
}

func (l *Linear) ToMap() (value.Map, bool) {
	return value.Map{}, false
}

func (l *Linear) ToInt() (int, bool) {
	return 0, false
}

func (l *Linear) ToFloat() (float64, bool) {
	return 0, false
}

func (l *Linear) ToString(st funcGen.Stack[value.Value]) (string, error) {
	return l.String(), nil
}

func (l *Linear) ToBool() (bool, bool) {
	return false, false
}

func (l *Linear) ToClosure() (funcGen.Function[value.Value], bool) {
	return funcGen.Function[value.Value]{}, false
}

func (l *Linear) GetType() value.Type {
	return LinearValueType
}

func linOperation(name string,
	def func(st funcGen.Stack[value.Value], a value.Value, b value.Value) (value.Value, error),
	f func(a, b *Linear) (value.Value, error),
	num func(a *Linear, fl float64) (value.Value, error)) func(st funcGen.Stack[value.Value], a value.Value, b value.Value) (value.Value, error) {

	return func(st funcGen.Stack[value.Value], a value.Value, b value.Value) (value.Value, error) {
		if ae, ok := a.(*Linear); ok {
			if be, ok := b.(*Linear); ok {
				// both are error values
				return f(ae, be)
			} else {
				// a is error value, b is'nt
				if bf, ok := b.ToFloat(); ok && num != nil {
					return num(ae, bf)
				} else {
					return nil, fmt.Errorf("%s operation not allowed with %v and %v ", name, a, b)
				}
			}
		} else {
			if be, ok := b.(*Linear); ok {
				// b is error value, a is'nt
				if af, ok := a.ToFloat(); ok && num != nil {
					return num(be, af)
				} else {
					return nil, fmt.Errorf("%s operation not allowed with %v and %v ", name, a, b)
				}
			} else {
				// no error value at all
				return def(st, a, b)
			}
		}
	}
}

func div(st funcGen.Stack[value.Value], a value.Value, b value.Value) (value.Value, error) {
	if aLin, ok := a.(*Linear); ok {
		if bLin, ok := b.(*Linear); ok {
			return aLin.Div(bLin), nil
		} else {
			if bFl, ok := b.ToFloat(); ok {
				return aLin.MulFloat(1 / bFl), nil
			}
		}
	} else if bLin, ok := b.(*Linear); ok {
		if aFl, ok := a.ToFloat(); ok {
			return bLin.Inv().MulFloat(aFl), nil
		}
	}

	return value.Div(st, a, b)
}

func sub(st funcGen.Stack[value.Value], a value.Value, b value.Value) (value.Value, error) {
	if aLin, ok := a.(*Linear); ok {
		if bLin, ok := b.(*Linear); ok {
			return aLin.Add(bLin.MulFloat(-1)), nil
		} else {
			if bFl, ok := b.ToFloat(); ok {
				return aLin.Add(NewConst(-bFl)), nil
			}
		}
	} else if bLin, ok := b.(*Linear); ok {
		if aFl, ok := a.ToFloat(); ok {
			return NewConst(aFl).Add(bLin.MulFloat(-1)), nil
		}
	}

	return value.Sub(st, a, b)
}

func exp(st funcGen.Stack[value.Value], a value.Value, b value.Value) (value.Value, error) {
	if aLin, ok := a.(*Linear); ok {
		if bInt, ok := b.(value.Int); ok {
			return aLin.Pow(int(bInt)), nil
		}
	}

	return value.Pow(st, a, b)
}

var parser = value.New().
	RegisterMethods(LinearValueType, linMethods()).
	AddStaticFunction("lin", funcGen.Function[value.Value]{
		Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			return &Linear{
				Numerator:   Polynomial{0, 1},
				Denominator: Polynomial{1},
			}, nil
		},
		Args:   0,
		IsPure: true,
	}.SetDescription("a linear system 's'")).
	AddStaticFunction("linConst", funcGen.Function[value.Value]{
		Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			if c, ok := stack.Get(0).ToFloat(); ok {
				return NewConst(c), nil
			}
			return nil, fmt.Errorf("linConst requires a float value")
		},
		Args:   1,
		IsPure: true,
	}.SetDescription("val", "converts a constant to a linear system")).
	AddStaticFunction("pid", funcGen.Function[value.Value]{
		Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			if kp, ok := stack.Get(0).ToFloat(); ok {
				if ti, ok := stack.Get(1).ToFloat(); ok {
					if td, ok := stack.Get(2).ToFloat(); ok {
						return PID(kp, ti, td), nil
					}
				}
			}
			return nil, fmt.Errorf("pid requires 3 float values")
		},
		Args:   3,
		IsPure: true,
	}.SetDescription("k_p", "T_I", "T_D", "a PID linear system")).
	AddOp("*", true, linOperation("*", value.Mul,
		func(a, b *Linear) (value.Value, error) {
			return a.Mul(b), nil
		},
		func(a *Linear, f float64) (value.Value, error) {
			return a.MulFloat(f), nil
		}),
	).
	AddOp("/", false, div).
	AddOp("-", false, sub).
	AddOp("^", false, exp).
	AddOp("+", true, linOperation("+", value.Add,
		func(a, b *Linear) (value.Value, error) {
			return a.Add(b), nil
		},
		func(a *Linear, f float64) (value.Value, error) {
			return a.Add(NewConst(f)), nil
		}),
	)

func linMethods() value.MethodMap {
	return value.MethodMap{
		"loop": funcGen.Function[value.Value]{
			Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
				if lin, ok := stack.Get(0).(*Linear); ok {
					return lin.Loop()
				}
				return nil, fmt.Errorf("loop requires a linear system")
			},
			Args:   0,
			IsPure: true,
		}.SetDescription("closes the loop"),
		"reduce": funcGen.Function[value.Value]{
			Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
				if lin, ok := stack.Get(0).(*Linear); ok {
					return lin.Reduce()
				}
				return nil, fmt.Errorf("loop requires a linear system")
			},
			Args:   0,
			IsPure: true,
		}.SetDescription("reduces the linear system"),
		"stringPoly": funcGen.Function[value.Value]{
			Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
				if lin, ok := stack.Get(0).(*Linear); ok {
					return value.String(lin.StringPoly(false)), nil
				}
				return nil, fmt.Errorf("stringPoly requires a linear system")
			},
			Args:   0,
			IsPure: true,
		}.SetDescription("reduces the linear system"),
	}
}
