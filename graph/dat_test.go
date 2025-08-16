package graph

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func createPoints(count int, initial, increment float64, f func(float64) float64) Points {
	return func(yield func(Point, error) bool) {
		x := initial
		for i := 0; i < count; i++ {
			y := f(x)
			if !yield(Point{X: x, Y: y}, nil) {
				return
			}
			x += increment
		}
	}
}

func TestDataString(t *testing.T) {
	data := &Data{
		TimeUnit: "s",
	}
	data.Add(DataContent{
		Name: "square",
		Unit: "1",
		Points: createPoints(10, 0, 1, func(x float64) float64 {
			return x * x
		}),
	})
	data.Add(DataContent{
		Name: "square2",
		Unit: "1",
		Points: createPoints(10, 0, 1, func(x float64) float64 {
			return 2 * x * x
		}),
	})
	dataFile, err := data.DatFile()
	assert.NoError(t, err)

	assert.EqualValues(t, `#time[s]	square[1]	square2[1]
0	0	0
1	1	2
2	4	8
3	9	18
4	16	32
5	25	50
6	36	72
7	49	98
8	64	128
9	81	162`, string(dataFile))

	csvFile, err := data.CsvFile()
	assert.NoError(t, err)

	assert.EqualValues(t, `"time[s]","square[1]","square2[1]"
"0","0","0"
"1","1","2"
"2","4","8"
"3","9","18"
"4","16","32"
"5","25","50"
"6","36","72"
"7","49","98"
"8","64","128"
"9","81","162"`, string(csvFile))
}

func TestDataString2(t *testing.T) {
	data := &Data{
		TimeUnit: "s",
	}
	data.Add(DataContent{
		Name: "square",
		Unit: "1",
		Points: createPoints(5, 0, 2, func(x float64) float64 {
			return x * x
		}),
	})
	data.Add(DataContent{
		Name: "square2",
		Unit: "1",
		Points: createPoints(10, 0, 1, func(x float64) float64 {
			return 2 * x * x
		}),
	})
	dataFile, err := data.DatFile()
	assert.NoError(t, err)

	assert.EqualValues(t, `#time[s]	square[1]	square2[1]
0	0	0
1	-	2
2	4	8
3	-	18
4	16	32
5	-	50
6	36	72
7	-	98
8	64	128
9	-	162`, string(dataFile))

	csvFile, err := data.CsvFile()
	assert.NoError(t, err)

	assert.EqualValues(t, `"time[s]","square[1]","square2[1]"
"0","0","0"
"1","","2"
"2","4","8"
"3","","18"
"4","16","32"
"5","","50"
"6","36","72"
"7","","98"
"8","64","128"
"9","","162"`, string(csvFile))
}
