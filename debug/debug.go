package debug

import (
	"fmt"
)

var Flag bool
var Flag2 bool
var IX int
var IY int
var S int
var T int
var C float32
var D int
var INT bool

func IntPrint(f string, args ...interface{}) {
	if INT {
		fmt.Printf(f, args...)
		fmt.Println()
	}
}

func IntPrintln(args ...interface{}) {
	if INT {
		fmt.Printf("[%d] [%d] ", S, D)
		fmt.Println(args...)
	}
}

type RGB struct {R, G, B float32}

var Mark *(RGB)

type Metric struct {
	Avg float32
	Max float32
	Cnt int
}

type CellT struct {
	R Metric
}

type MatrixT struct {
	W int
	H int
	R []Metric
}

var Cell CellT
var Matrix MatrixT

func NewMatrix(w, h int) MatrixT {
	return MatrixT{
		W: w,
		H: h,
		R: make([]Metric, w*h),
	}
}

func ResetMatrix(w, h int) {
	Matrix = NewMatrix(w, h)
}

func (m *Metric) Up(val float32) {
	m.Avg += val
	m.Cnt++
	if val > m.Max {
		m.Max = val
	}
}

func (m *Metric) Finalize() Metric {
	m.Avg /= float32(m.Cnt)
	return *m
}

func CommitCell(x, y int) {
	p := y*Matrix.W + x
	Matrix.R[p] = Cell.R.Finalize()
	Cell = CellT{}
}

func (m *MatrixT) Rmap() []uint8 {
	ret := make([]uint8, m.W*m.H*4)
	var max float32 = 0
	for i, _ := range m.R {
		if max < m.R[i].Max {
			max = m.R[i].Max
		}
	}
	fmt.Println("MAX", max)
	i := 0
	for p := 0; p < len(ret); p += 4 {
		ret[p] = uint8(255 * m.R[i].Max / max)
		ret[p + 3] = 255
		i++
	}
	fmt.Println(ret)
	return ret
}

func Noop() {
}
