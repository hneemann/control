package graph

import (
	"fmt"
	"math"
)

type Rect struct {
	Min, Max Point
}

func (r Rect) String() string {
	return fmt.Sprintf("Rect(%v,%v)", r.Min, r.Max)
}

func NewRect(x0, x1, y0, y1 float64) Rect {
	return Rect{
		Min: Point{x0, y0},
		Max: Point{x1, y1},
	}
}

func (r Rect) Path() Path {
	return CloseablePointsPath{
		Points: PointsFromSlice(r.Min, Point{r.Max.X, r.Min.Y}, r.Max, Point{r.Min.X, r.Max.Y}),
		Closed: true,
	}
}

func (r Rect) Contains(p Point) bool {
	return p.X >= r.Min.X && p.X <= r.Max.X && p.Y >= r.Min.Y && p.Y <= r.Max.Y
}

type IntersectResult int

const (
	CompleteOutside IntersectResult = iota
	BothInside
	P0Outside
	P1Outside
	BothOutsidePartVisible
)

func (r Rect) Intersect(p0, p1 Point) (Point, Point, IntersectResult) {
	f0 := r.getFlags(p0)
	f1 := r.getFlags(p1)
	if f0|f1 == 0 {
		// both inside
		return p0, p1, BothInside
	}
	if f0&f1 != 0 {
		// completely outside, no intersection possible
		return p0, p1, CompleteOutside
	}
	if f0 == 0 {
		// p0 inside, p1 outside
		return p0, r.Cut(p0, p1), P1Outside
	} else if f1 == 0 {
		// p1 inside, p0 outside
		return r.Cut(p1, p0), p1, P0Outside
	}

	var s []float64
	s = appendCut(s, cutBothOutside(p0, p1, Point{r.Min.X, r.Min.Y}, Point{r.Min.X, r.Max.Y}))
	s = appendCut(s, cutBothOutside(p0, p1, Point{r.Max.X, r.Min.Y}, Point{r.Max.X, r.Max.Y}))
	s = appendCut(s, cutBothOutside(p0, p1, Point{r.Min.X, r.Min.Y}, Point{r.Max.X, r.Min.Y}))
	s = appendCut(s, cutBothOutside(p0, p1, Point{r.Min.X, r.Max.Y}, Point{r.Max.X, r.Max.Y}))

	if len(s) != 2 || math.Abs(s[0]-s[1]) < 1e-6 {
		return p0, p1, CompleteOutside
	}

	if s[0] > s[1] {
		s[0], s[1] = s[1], s[0]
	}

	return Point{
			X: p0.X + s[0]*(p1.X-p0.X),
			Y: p0.Y + s[0]*(p1.Y-p0.Y),
		}, Point{
			X: p0.X + s[1]*(p1.X-p0.X),
			Y: p0.Y + s[1]*(p1.Y-p0.Y),
		}, BothOutsidePartVisible

}

func (r Rect) Cut(inside Point, outside Point) Point {
	s := check(cutFirstInside(inside, outside, Point{r.Min.X, r.Min.Y}, Point{r.Min.X, r.Max.Y}), 2)
	s = check(cutFirstInside(inside, outside, Point{r.Max.X, r.Min.Y}, Point{r.Max.X, r.Max.Y}), s)
	s = check(cutFirstInside(inside, outside, Point{r.Min.X, r.Min.Y}, Point{r.Max.X, r.Min.Y}), s)
	s = check(cutFirstInside(inside, outside, Point{r.Min.X, r.Max.Y}, Point{r.Max.X, r.Max.Y}), s)
	if s > 1.5 {
		return inside
	}
	return Point{
		X: inside.X + s*(outside.X-inside.X),
		Y: inside.Y + s*(outside.Y-inside.Y),
	}
}

func cutFirstInside(i Point, o Point, c0, c1 Point) float64 {
	d := c0.X*(i.Y-o.Y) + c0.Y*(o.X-i.X) + c1.X*(o.Y-i.Y) + c1.Y*(i.X-o.X)
	if d == 0 {
		return -1
	}
	return -(c0.X*(c1.Y-i.Y) + c0.Y*(i.X-c1.X) + c1.X*i.Y - c1.Y*i.X) / d
}

func cutBothOutside(p0 Point, p1 Point, c0, c1 Point) float64 {
	// s = - (c1x·(c2y - iy) + c1y·(ix - c2x) + c2x·iy - c2y·ix)/(c1x·(iy - oy) + c1y·(ox - ix) + c2x·(oy - iy) + c2y·(ix - ox))
	// t = (c1x·(iy - oy) + c1y·(ox - ix) + ix·oy - iy·ox)/(c1x·(iy - oy) + c1y·(ox - ix) + c2x·(oy - iy) + c2y·(ix - ox))
	d := c0.X*(p0.Y-p1.Y) + c0.Y*(p1.X-p0.X) + c1.X*(p1.Y-p0.Y) + c1.Y*(p0.X-p1.X)
	if d == 0 {
		return -1
	}

	t := (c0.X*(p0.Y-p1.Y) + c0.Y*(p1.X-p0.X) + p0.X*p1.Y - p0.Y*p1.X) / d
	if t < 0 || t > 1 {
		return -1
	}

	return -(c0.X*(c1.Y-p0.Y) + c0.Y*(p0.X-c1.X) + c1.X*p0.Y - c1.Y*p0.X) / d
}

func check(s float64, bestS float64) float64 {
	if s > 0 && s < bestS {
		return s
	} else {
		return bestS
	}
}

func appendCut(list []float64, s float64) []float64 {
	if s > 0 {
		return append(list, s)
	} else {
		return list
	}
}

func (r Rect) Width() float64 {
	return r.Max.X - r.Min.X
}

func (r Rect) Height() float64 {
	return r.Max.Y - r.Min.Y
}

func (r Rect) MaxDistance(p Point) float64 {
	d := math.Max(p.DistTo(r.Min), p.DistTo(r.Max))
	d = math.Max(d, p.DistTo(Point{X: r.Min.X, Y: r.Max.Y}))
	return math.Max(d, p.DistTo(Point{X: r.Max.X, Y: r.Min.Y}))
}

const nearDiv = 20

func (r Rect) IsInTopHalf(p Point) bool {
	return p.Y > r.Min.Y+(r.Height()/2)
}

func (r Rect) IsInLeftHalf(p Point) bool {
	return p.X < r.Min.X+(r.Width()/2)
}

func (r Rect) IsNearTop(p Point) bool {
	return math.Abs(r.Max.Y-p.Y) < r.Height()/nearDiv
}

func (r Rect) IsNearBottom(p Point) bool {
	return math.Abs(r.Min.Y-p.Y) < r.Height()/nearDiv
}

func (r Rect) IsNearLeft(p Point) bool {
	return math.Abs(r.Min.X-p.X) < r.Width()/nearDiv
}

func (r Rect) IsNearRight(p Point) bool {
	return math.Abs(r.Max.X-p.X) < r.Width()/nearDiv
}

func (r Rect) getFlags(p0 Point) int {
	flags := 0
	if p0.X < r.Min.X {
		flags |= 1
	} else if p0.X > r.Max.X {
		flags |= 2
	}
	if p0.Y < r.Min.Y {
		flags |= 4
	} else if p0.Y > r.Max.Y {
		flags |= 8
	}
	return flags

}

type interPath struct {
	p Path
	r Rect
}

func (i interPath) Iter(yield func(PathElement, error) bool) {
	var lastPoint Point
	var lastInside bool
	for pe, err := range i.p.Iter {
		inside := i.r.Contains(pe.Point)
		if pe.Mode == 'M' {
			if inside {
				if !yield(pe, err) {
					return
				}
			}
		} else {
			if lastInside && inside {
				if !yield(PathElement{Mode: 'L', Point: pe.Point}, err) {
					return
				}
			} else if !lastInside && inside {
				if !yield(PathElement{Mode: 'M', Point: i.r.Cut(pe.Point, lastPoint)}, err) {
					return
				}
				if !yield(PathElement{Mode: 'L', Point: pe.Point}, err) {
					return
				}
			} else if lastInside && !inside {
				if !yield(PathElement{Mode: 'L', Point: i.r.Cut(lastPoint, pe.Point)}, err) {
					return
				}
			} else {
				p0, p1, mode := i.r.Intersect(lastPoint, pe.Point)
				if mode == BothOutsidePartVisible {
					if !yield(PathElement{Mode: 'M', Point: p0}, err) {
						return
					}
					if !yield(PathElement{Mode: 'L', Point: p1}, err) {
						return
					}
				} else {
					// if an error occurs, and no other yield call has happened, the last
					// point including the error is yielded to avoid overlooking the error.
					if err != nil {
						if !yield(PathElement{Mode: 'M', Point: pe.Point}, err) {
							return
						} else {
							fmt.Println("Bug in error handling")
						}
					}
				}
			}
		}
		lastPoint = pe.Point
		lastInside = inside
	}
}

func (i interPath) IsClosed() bool {
	return i.p.IsClosed()
}

func (r Rect) IntersectPath(p Path) Path {
	return interPath{p, r}
}
