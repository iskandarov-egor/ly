package sampling

import (
	"sort"
	"fmt"
	"ly/img"
)

type Distribution1D struct {
	Pdf []float32
	Cdf []float32
}

/*
create a distribution from 0 to 1 that mimics the shape of @f
returns integral of @f
*/
func NewDistribution1D(f []float32) (Distribution1D, float32) {
	d := Distribution1D{
		Pdf: make([]float32, len(f)),
		Cdf: make([]float32, len(f) + 1),
	}
	d.Cdf[0] = 0
	for i := 1; i <= len(f); i++ {
		d.Cdf[i] = d.Cdf[i - 1] + f[i - 1]
	}
	I := d.Cdf[len(f)]
	if I == 0 {
		for i := 0; i < len(f); i++ {
			d.Cdf[i] = float32(i)/float32(len(f))
			d.Pdf[i] = 1
		}
		d.Cdf[len(f)] = 1
	} else {
		for i := 0; i < len(f); i++ {
			d.Cdf[i] /= I
			d.Pdf[i] = float32(len(f))*f[i]/I
		}
		d.Cdf[len(f)] = 1
	}
	return d, I
}

func (d Distribution1D) Sample(e float32) (x, pdf float32) {
	offset := -1 + sort.Search(len(d.Cdf), func (i int) bool {
		return d.Cdf[i] > e
	})
	if offset == len(d.Cdf) - 1 {
		panic("aaa")
	}
	if d.Cdf[offset + 1] - d.Cdf[offset] == 0 {
		panic("aaa")
	}
	delta := (e - d.Cdf[offset]) / (d.Cdf[offset + 1] - d.Cdf[offset])
	x = (float32(offset) + delta) / float32(len(d.Pdf))
	pdf = d.Pdf[offset]
	return
}

type Distribution2D struct {
	Marginal Distribution1D
	Conditional []Distribution1D
}

func NewDistribution2D(im img.Image1) Distribution2D {
	cond := make([]Distribution1D, im.H)
	marginal_f := make([]float32, im.H)
	for y := 0; y < im.H; y++ {
		cond[y], marginal_f[y] = NewDistribution1D(im.Data[im.W*y:im.W*(1 + y)])
	}
	marginal, _ := NewDistribution1D(marginal_f)
	return Distribution2D{
		Marginal: marginal,
		Conditional: cond,
	}
}

func (d *Distribution2D) Sample(e1, e2 float32) (x, y, pdf float32) {
	y, yPdf := d.Marginal.Sample(e1)
	x, xPdf := d.Conditional[int(y*float32(len(d.Conditional)))].Sample(e2)
	return x, y, xPdf*yPdf
}

func (d *Distribution2D) Pdf(x, y float32) float32 {
	yi := int(y * float32(len(d.Marginal.Pdf)))
	xi := int(x * float32(len(d.Conditional[0].Pdf)))
	if yi == len(d.Marginal.Pdf) {
		yi--
	}
	if xi == len(d.Conditional[0].Pdf) {
		xi--
	}
	return d.Marginal.Pdf[yi] * d.Conditional[yi].Pdf[xi]
}

func init() {
	_ = fmt.Print
}
