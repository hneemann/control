package polynomial

import (
	"fmt"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"strings"
	"unicode"
)

type BlockFactory struct {
	creator BlockFactoryFunc
	inputs  int
	name    string
}

type BlockFactoryFunc func([]*float64) (BlockNextFunc, error)

type BlockNextFunc func(t, dt float64) (float64, error)

func Gain(g float64) BlockFactory {
	return BlockFactory{
		creator: func(args []*float64) (BlockNextFunc, error) {
			a := args[0]
			return func(_, _ float64) (float64, error) {
				return *a * g, nil
			}, nil
		},
		inputs: 1,
		name:   fmt.Sprintf("Gain %f", g),
	}
}

func Const(c float64) BlockFactory {
	return BlockFactory{
		creator: func(args []*float64) (BlockNextFunc, error) {
			return func(_, _ float64) (float64, error) {
				return c, nil
			}, nil
		},
		inputs: 0,
		name:   fmt.Sprintf("Const %f", c),
	}
}

func Closure(c value.Closure) BlockFactory {
	return BlockFactory{
		creator: func(args []*float64) (BlockNextFunc, error) {
			st := funcGen.NewEmptyStack[value.Value]()
			vals := make([]value.Value, len(args))
			return func(_, _ float64) (float64, error) {
				for i, a := range args {
					vals[i] = value.Float(*a)
				}
				res, err := c.EvalSt(st, vals...)
				if err != nil {
					return 0, err
				}
				if f, ok := res.ToFloat(); ok {
					return f, nil
				}
				return 0, fmt.Errorf("invalid return value %v", res)
			}, nil
		},
		inputs: c.Args,
		name:   "function of input signals",
	}
}

func ClosureTime(c value.Closure) BlockFactory {
	return BlockFactory{
		creator: func(args []*float64) (BlockNextFunc, error) {
			st := funcGen.NewEmptyStack[value.Value]()
			return func(t, _ float64) (float64, error) {
				res, err := c.Eval(st, value.Float(t))
				if err != nil {
					return 0, err
				}
				if f, ok := res.ToFloat(); ok {
					return f, nil
				}
				return 0, fmt.Errorf("invalid return value %v", res)
			}, nil
		},
		inputs: 0,
		name:   "function of time",
	}
}
func Mul() BlockFactory {
	return BlockFactory{
		creator: func(args []*float64) (BlockNextFunc, error) {
			a := args[0]
			b := args[1]
			return func(_, _ float64) (float64, error) {
				return *a * *b, nil
			}, nil
		},
		inputs: 2,
		name:   "Mul",
	}
}

func Add() BlockFactory {
	return BlockFactory{
		creator: func(args []*float64) (BlockNextFunc, error) {
			a := args[0]
			b := args[1]
			return func(_, _ float64) (float64, error) {
				return *a + *b, nil
			}, nil
		},
		inputs: 2,
		name:   "Add",
	}
}

func AddMultiple(n int) BlockFactory {
	return BlockFactory{
		creator: func(args []*float64) (BlockNextFunc, error) {
			return func(_, _ float64) (float64, error) {
				var sum float64
				for i := 0; i < n; i++ {
					sum += *args[i]
				}
				return sum, nil
			}, nil
		},
		inputs: n,
		name:   "Add",
	}
}

func Limit(min, max float64) BlockFactory {
	return BlockFactory{
		creator: func(args []*float64) (BlockNextFunc, error) {
			in := args[0]
			return func(_, _ float64) (float64, error) {
				if *in < min {
					return min, nil
				} else if *in > max {
					return max, nil
				}
				return *in, nil
			}, nil
		},
		inputs: 1,
		name:   fmt.Sprintf("Limit %f-%f", min, max)}
}

func Sub() BlockFactory {
	return BlockFactory{
		creator: func(args []*float64) (BlockNextFunc, error) {
			a := args[0]
			b := args[1]
			return func(_, _ float64) (float64, error) {
				return *a - *b, nil
			}, nil
		},
		inputs: 2,
		name:   "Sub",
	}
}

func Integrate() BlockFactory {
	return BlockFactory{
		creator: func(args []*float64) (BlockNextFunc, error) {
			a := args[0]
			var sum float64
			return func(_, dt float64) (float64, error) {
				tf := sum
				sum += *a * dt
				return tf, nil
			}, nil
		},
		inputs: 1,
		name:   "Integrate",
	}
}

func BlockPID(kp, Ti, Td float64) (BlockFactory, error) {
	if Ti == 0 {
		return BlockFactory{}, fmt.Errorf("Ti must not be zero")
	}
	return BlockFactory{
		creator: func(args []*float64) (BlockNextFunc, error) {
			a := args[0]
			var sum float64
			var last float64
			return func(_, dt float64) (float64, error) {
				dif := (*a - last) / dt
				u := kp * (*a + sum/Ti + dif*Td)
				last = *a
				sum += *a * dt
				return u, nil
			}, nil
		},
		inputs: 1,
		name:   fmt.Sprintf("PID kp=%f, Ti=%f, Td=%f", kp, Ti, Td),
	}, nil
}

func Differentiate() BlockFactory {
	return BlockFactory{
		creator: func(args []*float64) (BlockNextFunc, error) {
			a := args[0]
			var last float64
			return func(_, dt float64) (float64, error) {
				dif := (*a - last) / dt
				last = *a
				return dif, nil
			}, nil
		},
		inputs: 1,
		name:   "Differentiate",
	}
}

func Delay(delayTime float64) BlockFactory {
	return BlockFactory{
		creator: func(args []*float64) (BlockNextFunc, error) {
			if delayTime <= 0 {
				return nil, fmt.Errorf("delay time must be greater than zero")
			}
			a := args[0]
			var buffer []float64
			i := 0
			n := 0
			return func(_, dt float64) (float64, error) {
				if n == 0 {
					n = int(delayTime / dt)
					buffer = make([]float64, n)
				}

				d := buffer[i]
				buffer[i] = *a
				i++
				if i >= n {
					i = 0
				}
				return d, nil
			}, nil
		},
		inputs: 1,
		name:   "Delay",
	}
}

func BlockLinear(lin *Linear) BlockFactory {
	return BlockFactory{
		creator: func(args []*float64) (BlockNextFunc, error) {
			in := args[0]
			a, c, d, err := lin.GetStateSpaceRepresentation()
			if err != nil {
				return nil, err
			}
			n := len(lin.Denominator) - 1
			x := make(Vector, n)
			xDot := make(Vector, n)
			if n > 0 {
				return func(_, dt float64) (float64, error) {
					y := c.Mul(x) + d**in

					a.Mul(xDot, x)
					xDot[n-1] += *in
					x.Add(dt, xDot)

					return y, nil
				}, nil
			} else {
				return func(_, dt float64) (float64, error) {
					return d * *in, nil
				}, nil
			}
		},
		inputs: 1,
		name:   fmt.Sprintf("Linear %v", lin),
	}
}

type SystemBlock struct {
	factory BlockFactory
	inputs  []string
	output  string
}

func (b SystemBlock) String() string {
	return fmt.Sprintf("%s: %v->%s", b.factory.name, b.inputs, b.output)
}

type System struct {
	blocks  []SystemBlock
	outputs []string
	next    []BlockNextFunc
	values  []float64
}

func NewSystem() *System {
	return &System{}
}

func (s *System) AddBlock(inputs []string, output string, factory BlockFactory) *System {
	s.blocks = append(s.blocks, SystemBlock{
		factory: factory,
		inputs:  inputs,
		output:  output,
	})
	return s
}

func (s *System) Initialize() error {
	var outputs []string
	var outputMap = make(map[string]int)
	for _, block := range s.blocks {
		if len(block.inputs) != block.factory.inputs {
			return fmt.Errorf("invalid number of inputs in '%v'", block)
		}

		if _, ok := outputMap[block.output]; ok {
			return fmt.Errorf("signal %s is created twice", block.output)
		}
		outputMap[block.output] = len(outputs)
		outputs = append(outputs, block.output)
	}
	if len(outputs) == 0 {
		return fmt.Errorf("no outputs defined")
	}

	for _, block := range s.blocks {
		for _, input := range block.inputs {
			if _, ok := outputMap[input]; !ok {
				return fmt.Errorf("input %s is not defined", input)
			}
		}
	}

	var next []BlockNextFunc
	var values = make([]float64, len(outputs))
	for _, block := range s.blocks {
		args := make([]*float64, block.factory.inputs)
		for i, input := range block.inputs {
			args[i] = &values[outputMap[input]]
		}
		nextFunc, err := block.factory.creator(args)
		if err != nil {
			return fmt.Errorf("error creating block '%v': %w", block, err)
		}
		next = append(next, nextFunc)
	}

	s.outputs = outputs
	s.next = next
	s.values = values

	return nil
}

func (s *System) Run(tMax, dt float64, pointsExported int) (*dataSet, error) {
	if pointsExported < 10 {
		pointsExported = 1000
	}

	if dt == 0 {
		dt = 1e-4
	}

	pointsCalculated := int(tMax / dt)
	skip := pointsCalculated / pointsExported
	if skip < 1 {
		skip = 1
	}

	t := 0.0

	nextValues := make([]float64, len(s.values))
	dataSetRows := pointsExported + 10

	resultData := newDataSet(dataSetRows, len(s.outputs)+1)

	counter := 0
	row := 0
	for {
		if counter == 0 || row < 10 {
			counter = skip
			resultData.set(row, 0, t)
			for i, y := range s.values {
				resultData.set(row, i+1, y)
			}
			row++
			if row >= dataSetRows {
				break
			}
		}
		counter--

		for i, next := range s.next {
			var err error
			nextValues[i], err = next(t, dt)
			if err != nil {
				return nil, fmt.Errorf("error in block %d: %w", i, err)
			}
		}
		copy(s.values, nextValues)
		t += dt
	}

	return resultData, nil
}

func valueToBlock(blockValue value.Value, in []string) (BlockFactory, error) {
	if lin, ok := blockValue.(*Linear); ok {
		return BlockLinear(lin), nil
	}
	if strVal, ok := blockValue.(value.String); ok {
		str := strings.ToLower(string(strVal))
		switch str {
		case "+":
			if len(in) > 2 {
				return AddMultiple(len(in)), nil
			} else {
				return Add(), nil
			}
		case "-":
			return Sub(), nil
		case "*":
			return Mul(), nil
		case "dif":
			return Differentiate(), nil
		case "int":
			return Integrate(), nil
		}
	}
	if c, ok := blockValue.ToFloat(); ok {
		if len(in) == 0 {
			return Const(c), nil
		} else {
			return Gain(c), nil
		}
	}

	if c, ok := blockValue.(value.Closure); ok {
		if len(in) == 0 {
			return ClosureTime(c), nil
		} else {
			return Closure(c), nil
		}
	}

	return BlockFactory{}, fmt.Errorf("invalid block type %v", blockValue)
}

func isIdent(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i, c := range s {
		if i == 0 {
			if !unicode.IsLetter(c) || c == '_' {
				return false
			}
		} else {
			if !(unicode.IsLetter(c) || unicode.IsDigit(c) || c == '_') {
				return false
			}
		}
	}
	return true
}

type NonLinDescription struct {
	blocks []SystemBlock
}

func NewNonLinSystem() NonLinDescription {
	return NonLinDescription{}
}

func (s NonLinDescription) ToList() (*value.List, bool) {
	return nil, false
}

func (s NonLinDescription) ToMap() (value.Map, bool) {
	return value.Map{}, false
}

func (s NonLinDescription) ToFloat() (float64, bool) {
	return 0, false
}

func (s NonLinDescription) ToString(_ funcGen.Stack[value.Value]) (string, error) {
	return "non linear system", nil
}

func (s NonLinDescription) GetType() value.Type {
	return NonLinValueType
}

func nonLinMethods() value.MethodMap {
	return value.MethodMap{
		"in": value.MethodAtType(2, func(s NonLinDescription, st funcGen.Stack[value.Value]) (value.Value, error) {
			val := st.Get(1)
			bf, err := valueToBlock(val, []string{})
			if err != nil {
				return nil, fmt.Errorf("only a function or a constant can be used as an input block: %w", err)
			}
			if out, err := st.Get(2).ToString(st); err == nil {
				return NonLinDescription{append(s.blocks, SystemBlock{
					factory: bf,
					inputs:  []string{},
					output:  out,
				})}, nil
			} else {
				return nil, fmt.Errorf("invalid output name %v", st.Get(2))
			}
		}).SetMethodDescription("input", "out", "Adds a constant or a function of time as an input to the system. "+
			"Most often used as a constant input signal 1 if a step response is desired."),
		"block": value.MethodAtType(3, func(s NonLinDescription, st funcGen.Stack[value.Value]) (value.Value, error) {
			inList, err := getInputList(st, st.Get(1))
			if err != nil {
				return nil, err
			}
			val := st.Get(2)
			bf, err := valueToBlock(val, inList)
			if err != nil {
				return nil, err
			}
			if out, err := st.Get(3).ToString(st); err == nil {
				return NonLinDescription{append(s.blocks, SystemBlock{
					factory: bf,
					inputs:  inList,
					output:  out,
				})}, nil
			} else {
				return nil, fmt.Errorf("invalid output name %v", st.Get(3))
			}
		}).SetMethodDescription("in", "type", "out", "Adds a block to the system. The type can be a string representing a basic "+
			"operation (\"+\", \"-\", \"*\", \"dif\", \"int\"), a float value for a constant gain, a Linear System, or a function (function of input signals). "+
			"If a function is used, and no input signals are specified, the function is assumed to be a function of time. Otherwise, "+
			"the number of input signals must match the number of arguments of the function."),
		"delay": value.MethodAtType(3, func(s NonLinDescription, st funcGen.Stack[value.Value]) (value.Value, error) {
			inList, err := getInputList(st, st.Get(1))
			if err != nil {
				return nil, err
			}
			if delay, ok := st.Get(2).ToFloat(); ok {
				if out, err := st.Get(3).ToString(st); err == nil {
					return NonLinDescription{append(s.blocks, SystemBlock{
						factory: Delay(delay),
						inputs:  inList,
						output:  out,
					})}, nil
				} else {
					return nil, fmt.Errorf("invalid output name %v", st.Get(3))
				}
			} else {
				return nil, fmt.Errorf("invalid delay value %v", st.Get(2))
			}
		}).SetMethodDescription("in", "delayTime", "out", "Adds a delay block to the system."),
		"limiter": value.MethodAtType(4, func(s NonLinDescription, st funcGen.Stack[value.Value]) (value.Value, error) {
			inList, err := getInputList(st, st.Get(1))
			if err != nil {
				return nil, err
			}
			if uMin, ok := st.Get(2).ToFloat(); ok {
				if uMax, ok := st.Get(3).ToFloat(); ok {
					if out, err := st.Get(4).ToString(st); err == nil {
						return NonLinDescription{append(s.blocks, SystemBlock{
							factory: Limit(uMin, uMax),
							inputs:  inList,
							output:  out,
						})}, nil
					} else {
						return nil, fmt.Errorf("invalid output name %v", st.Get(4))
					}
				} else {
					return nil, fmt.Errorf("invalid max value %v", st.Get(3))
				}
			} else {
				return nil, fmt.Errorf("invalid min value %v", st.Get(2))
			}
		}).SetMethodDescription("in", "min", "max", "out", "Adds a limiter block to the system. The output is limited to the range [min, max]."),
		"pid": value.MethodAtType(5, func(s NonLinDescription, st funcGen.Stack[value.Value]) (value.Value, error) {
			inList, err := getInputList(st, st.Get(1))
			if err != nil {
				return nil, err
			}
			if kp, ok := st.Get(2).ToFloat(); ok {
				if Ti, ok := st.Get(3).ToFloat(); ok {
					if Td, ok := st.Get(4).ToFloat(); ok {
						pid, err := BlockPID(kp, Ti, Td)
						if err != nil {
							return nil, fmt.Errorf("invalid PID parameters: %v", err)
						}
						if out, err := st.Get(5).ToString(st); err == nil {
							return NonLinDescription{append(s.blocks, SystemBlock{
								factory: pid,
								inputs:  inList,
								output:  out,
							})}, nil
						} else {
							return nil, fmt.Errorf("invalid output name %v", st.Get(5))
						}
					} else {
						return nil, fmt.Errorf("invalid Td value %v", st.Get(4))
					}
				} else {
					return nil, fmt.Errorf("invalid Ti value %v", st.Get(3))
				}
			} else {
				return nil, fmt.Errorf("invalid kp value %v", st.Get(2))
			}
		}).SetMethodDescription("in", "kp", "Ti", "Td", "out", "Adds a PID block to the system."),
		"simulate": value.MethodAtType(3, func(s NonLinDescription, st funcGen.Stack[value.Value]) (value.Value, error) {
			tMax, ok := st.Get(1).ToFloat()
			if !ok {
				return nil, fmt.Errorf("invalid tMax %v", st.Get(1))
			}

			dt := 0.0
			if dtVal, ok := st.GetOptional(2, value.Float(0)).ToFloat(); ok {
				dt = dtVal
			} else {
				return nil, fmt.Errorf("the second argument of simulate requires a float value")
			}

			points := 0
			if pointsVal, ok := st.GetOptional(3, value.Int(0)).(value.Int); ok {
				points = int(pointsVal)
			} else {
				return nil, fmt.Errorf("the third argument of simulate requires an int value")
			}

			sys := &System{blocks: s.blocks}
			err := sys.Initialize()
			if err != nil {
				return nil, err
			}

			resultData, err := sys.Run(tMax, dt, points)
			if err != nil {
				return nil, err
			}
			rm := make(map[string]value.Value)
			for i, name := range sys.outputs {
				rm[name] = resultData.toPointList(0, i+1)
			}

			return value.NewMap(value.RealMap(rm)), nil
		}).SetMethodDescription("tMax", "dt", "pointsExported", "Simulates the system and returns a map of output signals.").VarArgsMethod(1, 3),
	}
}

func getInputList(st funcGen.Stack[value.Value], v value.Value) ([]string, error) {
	if s, ok := v.(value.String); ok {
		if s == "" {
			return []string{}, nil
		}
		if isIdent(string(s)) {
			return []string{string(s)}, nil
		}
		return nil, fmt.Errorf("invalid signal name %v", v)
	}
	if l, ok := v.(*value.List); ok {
		var result []string
		for v, err := range l.Iterate(st) {
			if err != nil {
				return nil, err
			}
			if s, ok := v.(value.String); ok {
				if isIdent(string(s)) {
					result = append(result, string(s))
				} else {
					return nil, fmt.Errorf("invalid signal name %v", v)
				}
			} else {
				return nil, fmt.Errorf("not a string %v", v)
			}
		}
		return result, nil
	}
	return nil, fmt.Errorf("invalid signal type: %v", v)
}
