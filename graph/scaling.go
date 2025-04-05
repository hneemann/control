package graph

import (
	"fmt"
	"math"
)

type CheckTextWidth func(width float64, vks, nks int) bool

type Tick struct {
	Position float64
	Label    string
}

type Axis interface {
	WithMinMax(min, max float64) Axis
	Bounds() (min, max float64, set bool)
	Create(minParent, maxParent float64, ctw CheckTextWidth) (func(v float64) float64, []Tick)
}

type LinearAxis struct {
	boundsSet bool
	min, max  float64
}

func NewLinearAxis(xMin, xMax float64) Axis {
	return LinearAxis{min: xMin, max: xMax, boundsSet: true}
}

func (l LinearAxis) WithMinMax(min, max float64) Axis {
	return LinearAxis{min: min, max: max, boundsSet: true}
}

func (l LinearAxis) Bounds() (min, max float64, set bool) {
	return l.min, l.max, l.boundsSet
}

func (l LinearAxis) Create(minParent, maxParent float64, ctw CheckTextWidth) (func(v float64) float64, []Tick) {
	ticks := CreateLinearTicks(l.min, l.max, minParent, maxParent, ctw)
	return func(v float64) float64 {
		return (v-l.min)/(l.max-l.min)*(maxParent-minParent) + minParent
	}, ticks
}

func CreateLinearTicks(min, max, parentMin, parentMax float64, ctw CheckTextWidth) []Tick {
	l := linTickCreator{min: min, max: max}
	return l.ticks(parentMin, parentMax, ctw)
}

type linTickCreator struct {
	min, max float64
	log      int
	fineStep int
	delta    float64
}

var FINER = []float64{1, 0.5, 0.25, 0.2}
var LOG_CORR = []int{0, 1, 2, 1}

func (l *linTickCreator) ticks(minParent, maxParent float64, ctw CheckTextWidth) []Tick {
	l.log = int(math.Log10(l.max - l.min))
	l.delta = exp10(l.log)
	l.fineStep = 0

	vks := int(math.Floor(math.Max(math.Log10(math.Abs(l.min)), math.Log10(math.Abs(l.max)))) + 1)
	if vks < 1 {
		vks = 1
	}
	if l.min < 0 {
		vks++
	}

	l.delta *= 10
	l.log++ // sicher zu klein starten!

	for ctw(l.getPixels(maxParent-minParent), vks, l.getNks()) {
		l.inc()
	}
	l.dec()

	l.delta *= FINER[l.fineStep]

	format := fmt.Sprintf("%%.%df", l.getNks())

	tick := math.Ceil(l.min/l.delta) * l.delta
	ticks := []Tick{}
	for tick <= l.max {
		ticks = append(ticks, Tick{tick, fmt.Sprintf(format, tick)})
		tick += l.delta
	}
	return ticks
}

func (l *linTickCreator) getNks() int {
	nks := LOG_CORR[l.fineStep] - l.log
	if nks < 0 {
		return 0
	}
	return nks
}

func (l *linTickCreator) getPixels(width float64) float64 {
	return width * (l.delta * FINER[l.fineStep]) / (l.max - l.min)
}

func (l *linTickCreator) inc() {
	l.fineStep++
	if l.fineStep == len(FINER) {
		l.delta /= 10
		l.log--
		l.fineStep = 0
	}
}

func (l *linTickCreator) dec() {
	l.fineStep--
	if l.fineStep < 0 {
		l.delta *= 10
		l.log++
		l.fineStep = len(FINER) - 1
	}
}

func exp10(log int) float64 {
	e10 := 1.0
	if log < 0 {
		for i := 0; i < -log; i++ {
			e10 /= 10
		}
	} else {
		for i := 0; i < log; i++ {
			e10 *= 10
		}
	}
	return e10
}

type LogAxis struct {
	boundsSet bool
	min, max  float64
}

func NewLogAxis(xMin, xMax float64) Axis {
	return LogAxis{min: xMin, max: xMax, boundsSet: true}
}

func (l LogAxis) WithMinMax(min, max float64) Axis {
	return LogAxis{min: min, max: max, boundsSet: true}
}

func (l LogAxis) Bounds() (min, max float64, set bool) {
	return l.min, l.max, l.boundsSet
}

func (l LogAxis) Create(minParent, maxParent float64, ctw CheckTextWidth) (func(v float64) float64, []Tick) {
	logMin := math.Log10(l.min)
	logMax := math.Log10(l.max)

	tr := func(v float64) float64 {
		f := (math.Log10(v) - logMin) / (logMax - logMin)
		return f*(maxParent-minParent) + minParent
	}
	ticks := CreateLogTicks(logMin, minParent, maxParent, tr, ctw)
	return tr, ticks
}

func CreateLogTicks(logMin, parentMin, parentMax float64, tr func(v float64) float64, ctw CheckTextWidth) []Tick {
	m := int(math.Floor(logMin))
	var ticks []Tick
	for {
		f := exp10(m)
		for i := 1; i < 10; i++ {
			t := Tick{
				Position: f * float64(i),
			}
			tv := tr(t.Position)
			if tv > parentMax {
				return ticks
			}
			if tv >= parentMin {
				if i == 1 {
					t.Label = fmt.Sprintf("%g", f)
				}
				ticks = append(ticks, t)
			}
		}
		m++
	}
}
