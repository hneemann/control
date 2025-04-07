package grParser

import (
	"errors"
	"fmt"
	"github.com/hneemann/control/graph"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
)

type Dummy[T any] struct {
	Value T
}

func (w Dummy[T]) ToList() (*value.List, bool) {
	return nil, false
}

func (w Dummy[T]) ToMap() (value.Map, bool) {
	return value.Map{}, false
}

func (w Dummy[T]) ToInt() (int, bool) {
	return 0, false
}

func (w Dummy[T]) ToFloat() (float64, bool) {
	return 0, false
}

func (w Dummy[T]) ToString(st funcGen.Stack[value.Value]) (string, error) {
	return fmt.Sprint(w.Value), nil
}

func (w Dummy[T]) ToBool() (bool, bool) {
	return false, false
}

func (w Dummy[T]) ToClosure() (funcGen.Function[value.Value], bool) {
	return funcGen.Function[value.Value]{}, false
}

const (
	PlotType     value.Type = 10
	PlotListType value.Type = 11
	PointType    value.Type = 12
)

type PlotValue struct {
	Dummy[*graph.Plot]
}

func (p PlotValue) GetType() value.Type {
	return PlotType
}

func (p PlotValue) add(pc value.Value) error {
	if c, ok := pc.(PlotContentValue); ok {
		p.Dummy.Value.AddContent(c.Value)
		return nil
	}
	return errors.New("value is not a plot content")
}

func createPlotMethods() value.MethodMap {
	return value.MethodMap{
		"add": value.MethodAtType(1, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if pc, ok := stack.Get(1).(PlotContentValue); ok {
				plot.Value.AddContent(pc.Value)
			} else {
				return nil, fmt.Errorf("add requires a plot content")
			}
			return plot, nil
		}).SetMethodDescription("plotContent", "Add a plot content to the plot"),
		"xLabel": value.MethodAtType(1, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if str, ok := stack.Get(1).(value.String); ok {
				plot.Value.XLabel = string(str)
			} else {
				return nil, fmt.Errorf("xLabel requires a string")
			}
			return plot, nil
		}).SetMethodDescription("label", "Sets the x-label"),
		"yLabel": value.MethodAtType(1, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if str, ok := stack.Get(1).(value.String); ok {
				plot.Value.YLabel = string(str)
			} else {
				return nil, fmt.Errorf("yLabel requires a string")
			}
			return plot, nil
		}).SetMethodDescription("label", "Sets the y-label"),
		"xBounds": value.MethodAtType(2, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if vmin, ok := stack.Get(1).ToFloat(); ok {
				if vmax, ok := stack.Get(2).ToFloat(); ok {
					plot.Value.XBounds = graph.NewBounds(vmin, vmax)
					return plot, nil
				}
			}
			return nil, fmt.Errorf("bounds need to be float values")
		}).SetMethodDescription("xMin", "xMax", "Sets the x-bounds"),
		"yBounds": value.MethodAtType(2, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if vmin, ok := stack.Get(1).ToFloat(); ok {
				if vmax, ok := stack.Get(2).ToFloat(); ok {
					plot.Value.YBounds = graph.NewBounds(vmin, vmax)
					return plot, nil
				}
			}
			return nil, fmt.Errorf("bounds need to be float values")
		}).SetMethodDescription("yMin", "yMax", "Sets the y-bounds"),
	}
}

type PlotContentValue struct {
	Dummy[graph.PlotContent]
}

func (p PlotContentValue) GetType() value.Type {
	return PlotListType
}

type PointValue struct {
	Dummy[graph.Point]
}

func (p PointValue) GetType() value.Type {
	return PointType
}

func Setup(fg *value.FunctionGenerator) {
	fg.RegisterMethods(PlotType, createPlotMethods())
	fg.AddStaticFunction("plot", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			p := PlotValue{Dummy[*graph.Plot]{&graph.Plot{}}}
			if list, ok := st.Get(0).ToList(); ok {
				slice, err := list.ToSlice(st)
				if err != nil {
					return nil, err
				}
				for _, pc := range slice {
					err = p.add(pc)
					if err != nil {
						return nil, err
					}
				}
			} else {
				err := p.add(st.Get(0))
				if err != nil {
					return nil, err
				}
			}
			return p, nil
		},
		Args:   1,
		IsPure: true,
	}.SetDescription("content", "Creates a new plot"))
	fg.AddStaticFunction("scatter", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			points, err := toPointsList(st)
			if err != nil {
				return nil, err
			}
			s := graph.Scatter{Points: points, Shape: graph.NewCrossMarker(4), Style: graph.Black}
			return PlotContentValue{Dummy[graph.PlotContent]{s}}, nil
		},
		Args:   1,
		IsPure: true,
	}.SetDescription("data", "Creates a new scatter dataset"))
	fg.AddStaticFunction("curve", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			points, err := toPointsList(st)
			if err != nil {
				return nil, err
			}
			path := graph.NewPath(false)
			for _, p := range points {
				path = path.Add(p)
			}
			s := graph.Curve{Path: path, Style: graph.Black}
			return PlotContentValue{Dummy[graph.PlotContent]{s}}, nil
		},
		Args:   1,
		IsPure: true,
	}.SetDescription("data", "Creates a new scatter dataset"))
}

func toPointsList(st funcGen.Stack[value.Value]) ([]graph.Point, error) {
	if list, ok := st.Get(0).ToList(); ok {
		var points []graph.Point
		err := list.Iterate(st, func(v value.Value) error {
			if vec, ok := v.ToList(); ok {
				slice, err := vec.ToSlice(st)
				if err != nil {
					return err
				}
				if len(slice) != 2 {
					return fmt.Errorf("list elements needs to contain two floats")
				}
				if x, ok := slice[0].ToFloat(); ok {
					if y, ok := slice[1].ToFloat(); ok {
						points = append(points, graph.Point{x, y})
					} else {
						return fmt.Errorf("list elements needs to contain two floats")
					}
				} else {
					return fmt.Errorf("list elements needs to contain two floats")
				}
			}
			return nil
		})
		return points, err
	}
	return nil, fmt.Errorf("scatter requires a list of points")
}
