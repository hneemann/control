package graph

import (
	"fmt"
	"math"
	"time"
)

// CheckTextWidth returns true if a number with the given number of digits fits in the given space width
type CheckTextWidth func(width float64, digits int) bool

type Tick struct {
	Position float64
	Label    string
}

type Axis func(minParent, maxParent float64, bounds Bounds, ctw CheckTextWidth, expand float64) (func(v float64) float64, []Tick, Bounds)

func LinearAxis(minParent, maxParent float64, bounds Bounds, ctw CheckTextWidth, expand float64) (func(v float64) float64, []Tick, Bounds) {
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

func LogAxis(minParent, maxParent float64, bounds Bounds, ctw CheckTextWidth, expand float64) (func(v float64) float64, []Tick, Bounds) {
	if bounds.Min <= 0 {
		return LinearAxis(minParent, maxParent, bounds, ctw, expand)
	}

	logMin := math.Log10(bounds.Min)
	logMax := math.Log10(bounds.Max)

	if logMax-logMin < 1 {
		return LinearAxis(minParent, maxParent, bounds, ctw, expand)
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
	return func(minParent, maxParent float64, bounds Bounds, ctw CheckTextWidth, expand float64) (func(v float64) float64, []Tick, Bounds) {
		tr, ticks, b := LinearAxis(minParent, maxParent, bounds, ctw, expand)

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

func CreateDateAxis(formatDate, formatMin string) Axis {
	return func(minParent, maxParent float64, bounds Bounds, ctw CheckTextWidth, expand float64) (func(v float64) float64, []Tick, Bounds) {
		delta := bounds.Width() * expand
		if delta < 1e-10 {
			delta = 1 + expand
		}
		eMin := bounds.Min - delta
		eMax := bounds.Max + delta
		tr := func(v float64) float64 {
			return (v-eMin)/(eMax-eMin)*(maxParent-minParent) + minParent
		}

		t := time.UnixMilli(int64(eMin) * 1000)
		index := 0
		for {
			i := incrementerList[index]
			t0 := i.norm(t)
			t1 := i.inc(t0)
			var digits int
			if i.showMinutes() {
				digits = len(t0.Format(formatMin))
			} else {
				digits = len(t0.Format(formatDate))
			}
			if ctw(tr(float64(t1.UnixMilli())/1000)-tr(float64(t0.UnixMilli())/1000), digits+2) || index == len(incrementerList)-1 {
				break
			}
			index++
		}

		textIncr := incrementerList[index]
		smallIncr := textIncr.getSmallTicks()

		smallTickTime := smallIncr.norm(t)
		for smallTickTime.Before(t) {
			smallTickTime = smallIncr.inc(smallTickTime)
		}
		textTickTime := textIncr.norm(t)
		for textTickTime.Before(t) {
			textTickTime = textIncr.inc(textTickTime)
		}

		t = time.UnixMilli(int64(eMax) * 1000)
		ticks := []Tick{}
		textTime := float64(textTickTime.UnixMilli()) / 1000
		for smallTickTime.Before(t) {
			tick := Tick{Position: float64(smallTickTime.UnixMilli()) / 1000}
			if math.Abs(tick.Position-textTime) < 100 {
				if textIncr.showMinutes() {
					tick.Label = textTickTime.Format(formatMin)
				} else {
					tick.Label = textTickTime.Format(formatDate)
				}
				textTickTime = textIncr.inc(textTickTime)
				textTime = float64(textTickTime.UnixMilli()) / 1000
			}
			ticks = append(ticks, tick)
			smallTickTime = smallIncr.inc(smallTickTime)
		}

		return tr, ticks, Bounds{bounds.valid, eMin, eMax}
	}
}

var incrementerList = []incrementer{
	minuteIncrementer(1),
	minuteIncrementer(5),
	minuteIncrementer(10),
	minuteIncrementer(15),
	minuteIncrementer(30),
	hourIncrementer(1),
	hourIncrementer(2),
	hourIncrementer(3),
	hourIncrementer(6),
	hourIncrementer(12),
	dayIncrementer(1),
	dayIncrementer(2),
	dayIncrementer(4),
	weekIncrementer(1),
	weekIncrementer(2),
	monthIncrementer(1),
	monthIncrementer(2),
	monthIncrementer(3),
	monthIncrementer(4),
	monthIncrementer(6),
	yearIncrementer(1),
	yearIncrementer(2),
	yearIncrementer(5),
	yearIncrementer(10),
	yearIncrementer(20),
}

type incrementer interface {
	inc(time.Time) time.Time
	norm(time.Time) time.Time
	showMinutes() bool
	getSmallTicks() incrementer
}

type minuteIncrementer int

func (m minuteIncrementer) showMinutes() bool {
	return true
}

func (m minuteIncrementer) getSmallTicks() incrementer {
	if m >= 15 {
		return minuteIncrementer(5)
	}
	return minuteIncrementer(1)
}

func (m minuteIncrementer) inc(t time.Time) time.Time {
	y := t.Year()
	mo := t.Month()
	d := t.Day()
	h := t.Hour()
	mi := t.Minute() + int(m)
	return time.Date(y, mo, d, h, mi, 0, 0, t.Location())
}

func (m minuteIncrementer) norm(t time.Time) time.Time {
	y := t.Year()
	mo := t.Month()
	d := t.Day()
	h := t.Hour()
	mi := t.Minute()
	mi = (mi / int(m)) * int(m)
	return time.Date(y, mo, d, h, mi, 0, 0, t.Location())
}

type hourIncrementer int

func (h hourIncrementer) getSmallTicks() incrementer {
	if h == 1 {
		return minuteIncrementer(10)
	}
	return hourIncrementer(1)
}

func (h hourIncrementer) showMinutes() bool {
	return true
}

func (h hourIncrementer) inc(t time.Time) time.Time {
	y := t.Year()
	mo := t.Month()
	d := t.Day()
	ho := t.Hour()
	return time.Date(y, mo, d, ho+int(h), 0, 0, 0, t.Location())
}

func (h hourIncrementer) norm(t time.Time) time.Time {
	y := t.Year()
	mo := t.Month()
	d := t.Day()
	ho := t.Hour()
	ho = (ho / int(h)) * int(h)
	return time.Date(y, mo, d, ho, 0, 0, 0, t.Location())
}

type dayIncrementer int

func (d dayIncrementer) getSmallTicks() incrementer {
	if d == 1 {
		return hourIncrementer(1)
	}
	return dayIncrementer(1)
}

func (d dayIncrementer) showMinutes() bool {
	return false
}

func (d dayIncrementer) inc(t time.Time) time.Time {
	y := t.Year()
	mo := t.Month()
	da := t.Day()
	ti := time.Date(y, mo, da+2*int(d), 0, 0, 0, 0, t.Location())
	if ti.Month() != mo && ti.Day() > 1 {
		return time.Date(y, mo+1, 1, 0, 0, 0, 0, t.Location())
	} else {
		return time.Date(y, mo, da+int(d), 0, 0, 0, 0, t.Location())
	}
}

func (d dayIncrementer) norm(t time.Time) time.Time {
	y := t.Year()
	mo := t.Month()
	da := t.Day()
	da = ((da-1)/int(d))*int(d) + 1
	return time.Date(y, mo, da, 0, 0, 0, 0, t.Location())
}

type weekIncrementer int

func (w weekIncrementer) getSmallTicks() incrementer {
	if w == 1 {
		return dayIncrementer(1)
	}
	return weekIncrementer(1)
}

func (w weekIncrementer) showMinutes() bool {
	return false
}

func (w weekIncrementer) inc(t time.Time) time.Time {
	y := t.Year()
	mo := t.Month()
	da := t.Day() + 7
	if da > 28 {
		da = 1
		mo += 1
	}
	return time.Date(y, mo, da, 0, 0, 0, 0, t.Location())
}

func (w weekIncrementer) norm(t time.Time) time.Time {
	y := t.Year()
	mo := t.Month()
	da := ((t.Day()-1)/7/int(w))*int(w) + 1
	if da > 28 {
		da = 1
		mo += 1
	}
	return time.Date(y, mo, da, 0, 0, 0, 0, t.Location())
}

type monthIncrementer int

func (m monthIncrementer) getSmallTicks() incrementer {
	if m == 1 {
		return weekIncrementer(1)
	}
	return monthIncrementer(1)
}

func (m monthIncrementer) showMinutes() bool {
	return false
}

func (m monthIncrementer) inc(t time.Time) time.Time {
	y := t.Year()
	mo := t.Month() + time.Month(m)
	return time.Date(y, mo, 1, 0, 0, 0, 0, t.Location())
}

func (m monthIncrementer) norm(t time.Time) time.Time {
	y := t.Year()
	mo := ((t.Month()-1)/time.Month(m))*time.Month(m) + 1
	return time.Date(y, mo, 1, 0, 0, 0, 0, t.Location())
}

type yearIncrementer int

func (y yearIncrementer) getSmallTicks() incrementer {
	if y == 1 {
		return monthIncrementer(1)
	}
	return yearIncrementer(1)
}

func (y yearIncrementer) showMinutes() bool {
	return false
}

func (y yearIncrementer) inc(t time.Time) time.Time {
	ye := t.Year() + int(y)
	return time.Date(ye, 1, 1, 0, 0, 0, 0, t.Location())
}

func (y yearIncrementer) norm(t time.Time) time.Time {
	ye := (t.Year() / int(y)) * int(y)
	return time.Date(ye, 1, 1, 0, 0, 0, 0, t.Location())
}
