package graph

import (
	"fmt"
	"github.com/hneemann/parser2/value/export"
	"math"
	"time"
)

// CheckTextWidth returns true if a number with the given number of digits fits in the given width.
// The decimal point and a negative sign are counted as digits. So "-0.2" has 4 digits.
type CheckTextWidth func(width float64, digits int) bool

type Tick struct {
	// Position is the position of the tick in the axis coordinate space
	Position float64
	// Label is the label of the tick
	// If Label is empty, a short tick mark is drawn
	Label string
}

type Ticks []Tick

func (t Ticks) Characters() int {
	chars := 0
	for _, tick := range t {
		l := len(tick.Label)
		if l > chars {
			chars = l
		}
	}
	return chars
}

// Axis represents an axis.
type Axis struct {
	// Trans converts a value in the axis coordinate space to the parent coordinate space.
	// The parent coordinate space is typically the pixel space of the graph.
	Trans func(v float64) float64
	// Reverse converts a value in the parent coordinate space to the axis coordinate space.
	Reverse func(v float64) float64
	// Ticks are the ticks to be drawn on the axis.
	Ticks Ticks
	// Bounds are the bounds of the axis in the axis coordinate space.
	// They may be expanded beyond the data bounds to ensure that the axis labels are
	// not too close to the ends of the coordinate space.
	Bounds Bounds
	// LabelFormat is used to format the label of the axis.
	// It is used to deal with special units added by the scaling like dB.
	LabelFormat func(label string) string
	// IsLinear is true if the axis is linear.
	IsLinear bool
}

func noLabelFormat(label string) string {
	return label
}

// AxisFactory is a function that takes the parent coordinate space and the bounds of the axis and returns a function
// to convert values to the parent space, a slice of ticks, and the expanded bounds of the axis.
// The expanded bounds are required to ensure that the axis labels are not too close to the ends of the coordinate space.
type AxisFactory func(minParent, maxParent float64, bounds Bounds, ctw CheckTextWidth, expand float64) Axis

func LinearAxis(minParent, maxParent float64, bounds Bounds, ctw CheckTextWidth, expand float64) Axis {
	const max = 1e10
	delta := bounds.Width() * expand
	if delta > max {
		delta = max
	}
	eMin := bounds.Min - delta
	if eMin < -max {
		eMin = -max
	}
	eMax := bounds.Max + delta
	if eMax > max {
		eMax = max
	}
	l := linTickCreator{min: eMin, max: eMax}
	return Axis{
		Trans: func(v float64) float64 {
			return (v-eMin)/(eMax-eMin)*(maxParent-minParent) + minParent
		},
		Reverse: func(v float64) float64 {
			return (v-minParent)/(maxParent-minParent)*(eMax-eMin) + eMin
		},
		Ticks:       l.ticks(minParent, maxParent, ctw),
		Bounds:      Bounds{bounds.isSet, eMin, eMax},
		IsLinear:    true,
		LabelFormat: noLabelFormat,
	}
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

	const eps = 1e-10

	tick := math.Ceil(l.min/l.delta) * l.delta
	var ticks []Tick
	for tick <= l.max+eps {
		if math.Abs(tick) < eps {
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

func LogAxis(minParent, maxParent float64, bounds Bounds, ctw CheckTextWidth, expand float64) Axis {
	if bounds.Min <= 0 {
		return LinearAxis(minParent, maxParent, bounds, ctw, expand)
	}

	logMax := math.Log10(bounds.Max)
	if logMax > 10 {
		logMax = 10
	}
	logMin := math.Log10(bounds.Min)
	if logMin > logMax {
		logMin = logMax * 1e-3
	}

	if logMax-logMin < 1 {
		// drawing log axis marks makes little sense for less than one decade
		// fall back to plot logarithmic values on a linear axis
		return LogAxisSimple(minParent, maxParent, bounds, ctw, expand)
	}

	delta := (logMax - logMin) * expand
	logMin -= delta
	logMax += delta

	tr := func(v float64) float64 {
		f := (math.Log10(v) - logMin) / (logMax - logMin)
		return f*(maxParent-minParent) + minParent
	}
	rv := func(v float64) float64 {
		return math.Pow(10, (v-minParent)*(logMax-logMin)/(maxParent-minParent)+logMin)
	}
	return Axis{
		Trans:       tr,
		Reverse:     rv,
		Ticks:       createLogTicks(logMin, minParent, maxParent, tr),
		Bounds:      Bounds{bounds.isSet, math.Pow(10, logMin), math.Pow(10, logMax)},
		LabelFormat: noLabelFormat,
	}
}

func createLogTicks(logMin, parentMin, parentMax float64, tr func(v float64) float64) []Tick {
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
					t.Label = export.NewFormattedFloat(f, 6).Unicode()
				}
				ticks = append(ticks, t)
			}
		}
		m++
	}
}

// LogAxisSimple creates a simple logarithmic axis.
// It maps the logarithm of the values to a linear axis.
// It does not try to create logarithmic tick marks.
// If the bounds are less than or equal to zero, it falls back to a linear axis.
func LogAxisSimple(minParent, maxParent float64, bounds Bounds, ctw CheckTextWidth, expand float64) Axis {
	if bounds.Min <= 0 {
		return LinearAxis(minParent, maxParent, bounds, ctw, expand)
	}

	la := LinearAxis(minParent, maxParent, NewBounds(math.Log10(bounds.Min), math.Log10(bounds.Max)), ctw, expand)
	la.Bounds = bounds
	tr := la.Trans
	la.Trans = func(v float64) float64 {
		return tr(math.Log10(v))
	}
	rev := la.Reverse
	la.Reverse = func(v float64) float64 {
		return math.Pow(10, rev(v))
	}
	la.LabelFormat = func(label string) string {
		if label != "" {
			return "log(" + label + ")"
		}
		return "log"
	}
	for i, t := range la.Ticks {
		la.Ticks[i].Position = math.Pow(10, t.Position)
	}
	return la
}

// CreateFixedStepAxis creates an axis with a fixed step size.
// If there are less than two or more than 20 tick marks in the available space,
// it falls back to a linear axis.
func CreateFixedStepAxis(step float64) AxisFactory {
	return func(minParent, maxParent float64, bounds Bounds, ctw CheckTextWidth, expand float64) Axis {
		linearDef := LinearAxis(minParent, maxParent, bounds, ctw, expand)

		width := linearDef.Bounds.Width()
		if width < step*2 || width > step*20 {
			return linearDef
		}

		delta := step
		var start float64
		for {
			start = math.Floor(linearDef.Bounds.Min/delta) * delta
			next := start + delta

			widthAvail := linearDef.Trans(next) - linearDef.Trans(start)
			if ctw(widthAvail, 3) {
				break
			}
			delta *= 2
		}

		pos := math.Ceil(linearDef.Bounds.Min/step) * step

		for pos > start {
			start += delta
		}

		var ticks Ticks
		for pos <= linearDef.Bounds.Max {
			if math.Abs(pos-start) < 1e-6 {
				ticks = append(ticks, Tick{pos, fmt.Sprintf("%g", pos)})
				start += delta
			} else {
				ticks = append(ticks, Tick{Position: pos})
			}
			pos += step
		}

		return Axis{
			Trans:       linearDef.Trans,
			Reverse:     linearDef.Reverse,
			Ticks:       ticks,
			Bounds:      linearDef.Bounds,
			LabelFormat: linearDef.LabelFormat,
			IsLinear:    true,
		}
	}
}

// DBAxis creates a dB axis.
// To draw the ticks in dB scale, it uses a linear axis.
func DBAxis(minParent, maxParent float64, bounds Bounds, ctw CheckTextWidth, expand float64) Axis {
	if bounds.Min <= 0 {
		return LinearAxis(minParent, maxParent, bounds, ctw, expand)
	}

	dBMax := 20 * math.Log10(bounds.Max)
	if dBMax > 400 {
		dBMax = 400
	}
	dBMin := 20 * math.Log10(bounds.Min)
	if dBMin > dBMax {
		dBMin = dBMax - 60
	}

	fixedAxis := CreateFixedStepAxis(20)(minParent, maxParent, Bounds{bounds.isSet, dBMin, dBMax}, ctw, expand)
	var ticks Ticks
	for _, t := range fixedAxis.Ticks {
		ticks = append(ticks, Tick{Position: math.Pow(10, t.Position/20), Label: t.Label})
	}
	labelFormat := func(label string) string {
		if label != "" {
			return label + " in dB"
		}
		return "dB"
	}
	return Axis{
		Trans: func(v float64) float64 {
			return fixedAxis.Trans(20 * math.Log10(v))
		},
		Reverse: func(v float64) float64 {
			return math.Pow(10, fixedAxis.Reverse(v)/20)
		},
		Ticks: ticks,
		Bounds: Bounds{
			isSet: true,
			Min:   math.Pow(10, fixedAxis.Bounds.Min/20),
			Max:   math.Pow(10, fixedAxis.Bounds.Max/20),
		},
		LabelFormat: labelFormat,
	}
}

// CreateDateAxis creates a date axis.
// The values are expected to be in seconds since the epoch.
func CreateDateAxis(formatDate, formatMin string) AxisFactory {
	return func(minParent, maxParent float64, bounds Bounds, ctw CheckTextWidth, expand float64) Axis {
		delta := bounds.Width() * expand
		if delta < 1e-10 {
			delta = 1 + expand
		}
		eMin := bounds.Min - delta
		eMax := bounds.Max + delta
		tr := func(v float64) float64 {
			return (v-eMin)/(eMax-eMin)*(maxParent-minParent) + minParent
		}
		rv := func(v float64) float64 {
			return (v-minParent)/(maxParent-minParent)*(eMax-eMin) + eMin
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
		var ticks []Tick
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

		return Axis{
			Trans:       tr,
			Reverse:     rv,
			Ticks:       ticks,
			Bounds:      Bounds{bounds.isSet, eMin, eMax},
			IsLinear:    true,
			LabelFormat: noLabelFormat,
		}
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
