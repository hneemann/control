package polynomial

import (
	"fmt"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
)

type BlockFactory struct {
	creator BlockFactoryFunc
	inputs  int
}

type BlockFactoryFunc func([]*float64) (BlockNextFunc, error)

type BlockNextFunc func(t, dt float64) float64

func Gain(g float64) BlockFactory {
	return BlockFactory{
		creator: func(args []*float64) (BlockNextFunc, error) {
			a := args[0]
			return func(_, _ float64) float64 {
				return *a * g
			}, nil
		},
		inputs: 1,
	}
}

func Const(c float64) BlockFactory {
	return BlockFactory{
		creator: func(args []*float64) (BlockNextFunc, error) {
			return func(_, _ float64) float64 {
				return c
			}, nil
		},
		inputs: 0,
	}
}

func Add() BlockFactory {
	return BlockFactory{
		creator: func(args []*float64) (BlockNextFunc, error) {
			a := args[0]
			b := args[1]
			return func(_, _ float64) float64 {
				return *a + *b
			}, nil
		},
		inputs: 2,
	}
}

func Limit(min, max float64) BlockFactory {
	return BlockFactory{
		creator: func(args []*float64) (BlockNextFunc, error) {
			in := args[0]
			return func(_, _ float64) float64 {
				if *in < min {
					return min
				} else if *in > max {
					return max
				}
				return *in
			}, nil
		},
		inputs: 1,
	}
}

func Sub() BlockFactory {
	return BlockFactory{
		creator: func(args []*float64) (BlockNextFunc, error) {
			a := args[0]
			b := args[1]
			return func(_, _ float64) float64 {
				return *a - *b
			}, nil
		},
		inputs: 2,
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
			return func(_, dt float64) float64 {
				y := c.Mul(x) + d**in

				a.Mul(xDot, x)
				xDot[n-1] += *in
				x.Add(dt, xDot)

				return y
			}, nil
		},
		inputs: 1,
	}
}

type SystemBlock struct {
	factory BlockFactory
	inputs  []string
	output  string
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
			return fmt.Errorf("invalid number of inputs in %v", block.factory)
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
			return fmt.Errorf("error creating block %v: %w", block.factory, err)
		}
		next = append(next, nextFunc)
	}

	s.outputs = outputs
	s.next = next
	s.values = values

	return nil
}

func (s *System) Run(tMax float64) value.Value {
	t := 0.0
	dt := 0.00001
	const skip = 1000

	nextValues := make([]float64, len(s.values))
	result := make([]*dataSet, len(s.outputs))
	maxn := int(tMax / dt / skip)
	for i := range result {
		result[i] = newDataSet(maxn, 2)
	}

	count := skip
	n := 0
	for t < tMax {

		count--
		if count == 0 {
			count = skip
			if n < maxn {
				for i, y := range s.values {
					result[i].set(n, 0, t)
					result[i].set(n, 1, y)
				}
			}
			n++
		}

		for i, next := range s.next {
			nextValues[i] = next(t, dt)
		}
		copy(s.values, nextValues)
		t += dt
	}

	rm := make(map[string]value.Value)
	for i, output := range s.outputs {
		rm[output] = result[i].toList()
	}

	return value.NewMap(value.RealMap(rm))
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

			fac, ok := m.Get("block")
			if !ok {
				return fmt.Errorf("block not found %w", err)
			}
			if f, ok := fac.(BlockFactoryValue); ok {
				sys.AddBlock(in, out[0], f.Value)
			} else {
				return fmt.Errorf("block not valid")
			}

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

	return sys.Run(tMax), nil
}

func getStringList(st funcGen.Stack[value.Value], m value.Map, key string) ([]string, error) {
	v, ok := m.Get(key)
	if !ok {
		return nil, fmt.Errorf("key %s not found", key)
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
