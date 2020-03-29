package sampling

import (
	"fmt"
	"math/rand"
	"math"
	"ly/geo"
	"ly/util/math32"
)

type Sampler2D interface {
	Next() (x, y float32)
}

type UniformSampler2D struct {}

func NewUniform2D() *UniformSampler2D {
	return &UniformSampler2D{}
}

func (s *UniformSampler2D) Next() (x, y float32) {
	x = rand.Float32()
	y = rand.Float32()
	return
}

type StratifiedSampler2D struct {
	w int
	cell float32
	x int
	y int
}

func NewSampler2D(nsamples int) *StratifiedSampler2D {
	w := int(math.Ceil(math.Sqrt(float64(nsamples))))
	return &StratifiedSampler2D{
		w: w,
		cell: 1/float32(w),
		x: 0,
		y: 0,
	}
}

func (s *StratifiedSampler2D) Next() (x float32, y float32) {
	x = (float32(s.x) + rand.Float32())*s.cell
	y = (float32(s.y) + rand.Float32())*s.cell
	s.x++
	if (s.x >= s.w) {
		s.y++
		s.x = 0
		if s.y > s.w {
			panic("assertion error")
		}
	}
	return
}

// sample hemisphere with cosine distribution.
// e.g. pdf with respect to solid angle = cos(zenith angle)
func CosineSampleHemisphere() (ret geo.Vec3) {
	// mapping square to disk with no sqrt! sorcery!
	e1, e2 := 1 - 2*rand.Float32(), 1 - 2*rand.Float32()
	var r, theta float32
	pi := float32(3.141592)
	if math32.Abs(e1) > math32.Abs(e2) {
		r = e1
		theta = pi/4*e2/e1
	} else {
		r = e2
		theta = pi/2 - pi/4*e1/e2
	}
	// map disk to hemishpere
	ret.X = r*math32.Cos(theta)
	ret.Y = r*math32.Sin(theta)
	ret.Z = math32.SafeSqrt(1 - ret.X*ret.X - ret.Y*ret.Y)
	return
}

func main() {
	fmt.Println("vim-go")
}

func Noop() {}
