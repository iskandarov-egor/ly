package films

import (
	"ly/img"
	"ly/colors"
	"ly/spectra"
	"fmt"
)

type Cell struct {
	x, y, z float32
	weight float32
}

type Film interface {
	AddSample(x, y int, L spectra.Spectr, weight float32)
	Width() int
	Height() int
	ToImage() img.Image3
}

type SimpleFilm struct {
	W    int
	H    int
	Cells []Cell
}

func NewFilm(w int, h int) *SimpleFilm {
	f := SimpleFilm{
		W:    w,
		H:    h,
		Cells: make([]Cell, w*h),
	}
	return &f
}

func (f *SimpleFilm) Width() int {
	return f.W
}

func (f *SimpleFilm) Height() int {
	return f.H
}

func (f *SimpleFilm) AddSample(x, y int, L spectra.Spectr, weight float32) {
	pos := (y*f.W + x)
	X, Y, Z := L.XYZ()
	f.Cells[pos].x += X*weight
	f.Cells[pos].y += Y*weight
	f.Cells[pos].z += Z*weight
	f.Cells[pos].weight += weight
}

func (f *SimpleFilm) ToImage() img.Image3 {
	im := img.NewImage3(f.W, f.H, colors.XYZSpace)
	ii := 0
	for i, _ := range f.Cells {
		inv := 1/float32(f.Cells[i].weight)
		im.Data[ii], im.Data[ii + 1], im.Data[ii + 2] =
			f.Cells[i].x*inv, f.Cells[i].y*inv, f.Cells[i].z*inv
		ii += 3
	}
	return im
}

func init() {
	_ = fmt.Println
}
