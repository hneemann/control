package graph

import (
	"bytes"
	"fmt"
	"math"
	"time"
)

type DataContent struct {
	Name   string
	Unit   string
	Points Points
}

type Data struct {
	TimeIsDate  bool
	TimeUnit    string
	DataContent []DataContent
}

type pointErr struct {
	p   Point
	err error
}

type dataSync struct {
	name   string
	c      chan pointErr
	pe     pointErr
	isData bool
}

func (d *dataSync) update() error {
	d.pe, d.isData = <-d.c
	if d.pe.err != nil {
		return fmt.Errorf("error reading data for '%s': %w", d.name, d.pe.err)
	}
	return nil
}

type format interface {
	writeHeader(*bytes.Buffer, *Data)
	writeTime(*bytes.Buffer, float64)
	writeValue(*bytes.Buffer, float64)
	skipValue(*bytes.Buffer)
}

type dat struct{}

func (d dat) writeHeader(b *bytes.Buffer, data *Data) {
	if data.TimeIsDate {
		b.WriteString("#time is unix date\n")
	}

	b.WriteString("#time[" + data.TimeUnit + "]")
	for _, content := range data.DataContent {
		b.WriteString("\t" + content.Name + "[" + content.Unit + "]")
	}
}

func (d dat) writeTime(b *bytes.Buffer, t float64) {
	b.WriteString(fmt.Sprintf("\n%g", t))
}

func (d dat) writeValue(b *bytes.Buffer, v float64) {
	b.WriteString(fmt.Sprintf("\t%g", v))
}

func (d dat) skipValue(b *bytes.Buffer) {
	b.WriteString("\t-")
}

type csv struct {
	isDate bool
}

func (c csv) writeHeader(b *bytes.Buffer, data *Data) {
	if data.TimeIsDate {
		b.WriteString("\"date\",\"time\"")
	} else {
		b.WriteString("\"time[" + data.TimeUnit + "]\"")
	}
	for _, content := range data.DataContent {
		b.WriteString(",\"" + content.Name + "[" + content.Unit + "]\"")
	}
}

const (
	csvDateFormat = "2006-01-02"
	csvTimeFormat = "15:04:05"
)

func (c csv) writeTime(b *bytes.Buffer, t float64) {
	if c.isDate {
		sec := int64(math.Trunc(t))
		nsec := int64((t - float64(sec)) * 1e9)
		unix := time.Unix(sec, nsec)
		b.WriteString(fmt.Sprintf("\n\"%s\"", unix.Format(csvDateFormat)))
		b.WriteString(fmt.Sprintf(",\"%s\"", unix.Format(csvTimeFormat)))
	} else {
		b.WriteString(fmt.Sprintf("\n\"%g\"", t))
	}
}

func (c csv) writeValue(b *bytes.Buffer, f float64) {
	b.WriteString(fmt.Sprintf(",\"%g\"", f))
}

func (c csv) skipValue(b *bytes.Buffer) {
	b.WriteString(",\"\"")
}

func (d *Data) DatFile() ([]byte, error) {
	return d.writeFile(dat{})
}

func (d *Data) CsvFile() ([]byte, error) {
	return d.writeFile(csv{d.TimeIsDate})
}

func (d *Data) writeFile(f format) ([]byte, error) {
	var b bytes.Buffer

	f.writeHeader(&b, d)

	sy := make([]*dataSync, len(d.DataContent))
	stop := make(chan struct{})
	for i, content := range d.DataContent {
		c := make(chan pointErr)
		sy[i] = &dataSync{c: c, name: content.Name}
		go func() {
			for point, err := range content.Points {
				select {
				case c <- pointErr{point, err}:
					if err != nil {
						break
					}
				case <-stop:
					break
				}
			}
			close(c)
		}()
		// get first value
		err := sy[i].update()
		if err != nil {
			close(stop)
			return nil, err
		}
	}

	for {
		noData := true
		t := 0.0
		for _, s := range sy {
			if s.isData {
				if noData {
					t = s.pe.p.X
					noData = false
				} else {
					if s.pe.p.X < t {
						t = s.pe.p.X
					}
				}
			}
		}
		if noData {
			break
		}
		f.writeTime(&b, t)
		for _, s := range sy {
			if s.isData && t == s.pe.p.X {
				f.writeValue(&b, s.pe.p.Y)
				err := s.update()
				if err != nil {
					close(stop)
					return nil, err
				}
				if s.isData && s.pe.p.X < t {
					return nil, fmt.Errorf("data points '%s' are not in ascending order: %g < %g", s.name, s.pe.p.X, t)
				}
			} else {
				f.skipValue(&b)
			}
		}
	}

	return b.Bytes(), nil
}
