package polynomial

import (
	"fmt"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"strings"
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

func Closure(c funcGen.Function[value.Value]) BlockFactory {
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

func ClosureTime(c funcGen.Function[value.Value]) BlockFactory {
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
			return func(_, dt float64) (float64, error) {
				y := c.Mul(x) + d**in

				a.Mul(xDot, x)
				xDot[n-1] += *in
				x.Add(dt, xDot)

				return y, nil
			}, nil
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

func (s *System) Run(tMax float64) (value.Value, error) {
	const pointsExported = 1000
	const pointsInternal = 100000
	const skip = pointsInternal / pointsExported

	dt := tMax / pointsInternal
	t := 0.0

	nextValues := make([]float64, len(s.values))
	result := make([]*dataSet, len(s.outputs))
	dataSetRows := pointsExported + 10
	for i := range result {
		result[i] = newDataSet(dataSetRows, 2)
	}

	counter := 0
	row := 0
	for {

		if counter == 0 || row < 10 {
			counter = skip
			for i, y := range s.values {
				result[i].set(row, 0, t)
				result[i].set(row, 1, y)
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

	rm := make(map[string]value.Value)
	for i, output := range s.outputs {
		rm[output] = result[i].toList()
	}

	return value.NewMap(value.RealMap(rm)), nil
}

func SimulateBlock(st funcGen.Stack[value.Value], def *value.List, tMax float64) (value.Value, error) {
	sys := NewSystem()
	err := def.Iterate(st, func(v value.Value) error {
		if m, ok := v.(value.Map); ok {
			in, err := getStringList(st, m, "in")
			if err != nil {
				return fmt.Errorf("input not found: %w", err)
			}
			out, err := getStringList(st, m, "out")
			if err != nil {
				return fmt.Errorf("output not found %w", err)
			}
			if len(out) != 1 {
				return fmt.Errorf("output must be a single value")
			}

			blockValue, ok := m.Get("block")
			if !ok {
				return fmt.Errorf("block not found %w", err)
			}
			f, err := valueToBlock(blockValue, in)
			if err != nil {
				return fmt.Errorf("block not valid %w", err)
			}
			sys.AddBlock(in, out[0], f)
		} else {
			return fmt.Errorf("invalid block definition %v", v)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	err = sys.Initialize()
	if err != nil {
		return nil, err
	}

	return sys.Run(tMax)
}

func valueToBlock(blockValue value.Value, in []string) (BlockFactory, error) {
	if fac, ok := blockValue.(BlockFactoryValue); ok {
		return fac.Value, nil
	}
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
		case "dif":
			return Differentiate(), nil
		case "int":
			return Integrate(), nil
		}
	}
	if c, ok := blockValue.ToFloat(); ok {
		return Const(c), nil
	}

	if c, ok := blockValue.ToClosure(); ok {
		if len(in) == 0 {
			return ClosureTime(c), nil
		} else {
			return Closure(c), nil
		}
	}

	return BlockFactory{}, fmt.Errorf("invalid block type %v", blockValue)
}

func getStringList(st funcGen.Stack[value.Value], m value.Map, key string) ([]string, error) {
	v, ok := m.Get(key)
	if !ok {
		return []string{}, nil
	}
	if l, ok := v.(value.String); ok {
		return []string{string(l)}, nil
	}
	if l, ok := v.(*value.List); ok {
		var result []string
		err := l.Iterate(st, func(v value.Value) error {
			if s, ok := v.(value.String); ok {
				result = append(result, string(s))
				return nil
			}
			return fmt.Errorf("invalid string %v", v)
		})
		if err != nil {
			return nil, err
		}
		return result, nil
	}
	return nil, fmt.Errorf("invalid signal type", v)
}
