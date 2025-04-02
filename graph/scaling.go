package graph

import (
	"fmt"
	"math"
)

type CheckTextWidth func(width float64, vks, nks int) bool

type Axis interface {
	Create(minParent, maxParent float64) (min, max float64, tr func(v float64) float64)
	Ticks(minParent, maxParent float64, ctw CheckTextWidth) []Tick
}

type Tick struct {
	Position float64
	Label    string
}

type LinearAxis struct {
	min, max float64
	log      int
	fineStep int
	delta    float64
}

func NewLinear(min, max float64) Axis {
	return &LinearAxis{min: min, max: max}
}

func (l *LinearAxis) Create(minParent, maxParent float64) (float64, float64, func(v float64) float64) {
	return l.min, l.max, func(v float64) float64 {
		return (v-l.min)/(l.max-l.min)*(maxParent-minParent) + minParent
	}
}

var FINER = []float64{1, 0.5, 0.25, 0.2}
var LOG_CORR = []int{0, 1, 2, 1}

func (l *LinearAxis) Ticks(minParent, maxParent float64, ctw CheckTextWidth) []Tick {
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

func (l *LinearAxis) getNks() int {
	nks := LOG_CORR[l.fineStep] - l.log
	if nks < 0 {
		return 0
	}
	return nks
}

func (l *LinearAxis) getPixels(width float64) float64 {
	return width * (l.delta * FINER[l.fineStep]) / (l.max - l.min)
}

func (l *LinearAxis) inc() {
	l.fineStep++
	if l.fineStep == len(FINER) {
		l.delta /= 10
		l.log--
		l.fineStep = 0
	}
}

func (l *LinearAxis) dec() {
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
