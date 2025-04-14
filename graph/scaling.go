package graph

import (
	"fmt"
	"math"
)

// CheckTextWidth returns true if a number with the given number of digits fits in the given space width
type CheckTextWidth func(width float64, digits int) bool

type Tick struct {
	Position float64
	Label    string
}

const expand = 0.02

type Axis func(minParent, maxParent float64, bounds Bounds, ctw CheckTextWidth) (func(v float64) float64, []Tick, Bounds)

func LinearAxis(minParent, maxParent float64, bounds Bounds, ctw CheckTextWidth) (func(v float64) float64, []Tick, Bounds) {
	delta := bounds.Width() * expand
	if delta < 1e-10 {
		delta = 1 + expand
	}
	eMin := bounds.Min - delta
	eMax := bounds.Max + delta
	l := linTickCreator{min: eMin, max: eMax}
	ticks := l.ticks(minParent, maxParent, ctw)
	return func(v float64) float64 {
		return (v-eMin)/(eMax-eMin)*(maxParent-minParent) + minParent
	}, ticks, Bounds{bounds.valid, eMin, eMax}
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

	for ctw(l.getPixels(maxParent-minParent), vks+l.getNks()) {
		l.inc()
	}
	l.dec()

	l.delta *= FINER[l.fineStep]

	format := fmt.Sprintf("%%.%df", l.getNks())

	tick := math.Ceil(l.min/l.delta) * l.delta
	ticks := []Tick{}
	for tick <= l.max {
		if math.Abs(tick) < 1e-10 {
			tick = 0
		}
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

func LogAxis(minParent, maxParent float64, bounds Bounds, ctw CheckTextWidth) (func(v float64) float64, []Tick, Bounds) {
	if bounds.Min <= 0 {
		return LinearAxis(minParent, maxParent, bounds, ctw)
	}

	logMin := math.Log10(bounds.Min)
	logMax := math.Log10(bounds.Max)

	if logMax-logMin < 1 {
		return LinearAxis(minParent, maxParent, bounds, ctw)
	}

	delta := (logMax - logMin) * expand
	logMin -= delta
	logMax += delta

	tr := func(v float64) float64 {
		f := (math.Log10(v) - logMin) / (logMax - logMin)
		return f*(maxParent-minParent) + minParent
	}
	ticks := CreateLogTicks(logMin, minParent, maxParent, tr, ctw)
	return tr, ticks, Bounds{bounds.valid, math.Pow(10, logMin), math.Pow(10, logMax)}
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

func CreateFixedStepAxis(step float64) Axis {
	return func(minParent, maxParent float64, bounds Bounds, ctw CheckTextWidth) (func(v float64) float64, []Tick, Bounds) {
		tr, ticks, b := LinearAxis(minParent, maxParent, bounds, ctw)

		width := b.Width()
		if width < step*2 || width > step*20 {
			return tr, ticks, b
		}

		delta := step
		var start float64
		for {
			start = math.Floor(b.Min/delta) * delta
			next := start + delta

			widthAvail := tr(next) - tr(start)
			if ctw(widthAvail, 3) {
				break
			}
			delta *= 2
		}

		ticks = []Tick{}
		pos := math.Ceil(b.Min/step) * step

		for pos > start {
			start += delta
		}

		for pos <= b.Max {
			if math.Abs(pos-start) < 1e-6 {
				ticks = append(ticks, Tick{pos, fmt.Sprintf("%g", pos)})
				start += delta
			} else {
				ticks = append(ticks, Tick{Position: pos})
			}
			pos += step
		}

		return tr, ticks, b
	}

}
