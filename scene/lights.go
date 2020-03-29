package scene

import (
	"ly/spectra"
	"ly/geo"
	"ly/debug"
	"ly/sampling"
	"ly/colors"
	"ly/img"
	"ly/util/math32"
	"fmt"
	"math"
)

type Light interface {
	SampleRadiance(dest geo.Vec3, sampler sampling.Sampler2D) (
		ok bool, prob float32, spectr spectra.Spectr, origin geo.Vec3)
	// return the probability of sampling a particular direction.
	// the probability is with respect to solid angle.
	// needed for multiple importance sampling
	PDF(origin, direction geo.Vec3) float32
	// return radiance along the ray
	// @ray is the ray from the viewer _towards_ the light
	GetRadiance(ray geo.Ray) spectra.Spectr
	Power() float32
}

type DirectionLight struct {
	Spectr spectra.Spectr
	Direction geo.Vec3
	Direction2 geo.Vec3
}

func NewDirectionLight(dir geo.Vec3, spectr spectra.Spectr, sceneRadius float32) *DirectionLight {
	return &DirectionLight{
		Spectr: spectr,
		Direction: dir.Normalized(),
		Direction2: dir.Normalized().Mul(sceneRadius),
	}
}

func (l *DirectionLight) Power() float32 {
	panic("not impl")
}

func (l *DirectionLight) PDF(origin, direction geo.Vec3) float32 {
	panic("not impl")
}

func (l *DirectionLight) SampleRadiance(dest geo.Vec3, sampler sampling.Sampler2D) (
	ok bool, prob float32, spectr spectra.Spectr, origin geo.Vec3,
) {
	return true, 1, l.Spectr.Clone(), dest.Sub(l.Direction2)
}

func (l *DirectionLight) GetRadiance(ray geo.Ray) spectra.Spectr {
	panic("not impl")
}

type AreaLight struct {
	Shape Shape
	Spectr spectra.Spectr
}

func NewAreaLight(shape Shape, spectr spectra.Spectr) *AreaLight {
	rect := AreaLight{
		Spectr: spectr,
		Shape: shape,
	}
	return &rect
}

func (l *AreaLight) PDF(origin, direction geo.Vec3) float32 {
	return l.Shape.SamplePdf(geo.Ray{origin, direction})
}

func (r *AreaLight) SampleRadiance(dest geo.Vec3, sampler sampling.Sampler2D) (
	ok bool, prob float32, spectr spectra.Spectr, origin geo.Vec3,
) {
	sample, _, _ := r.Shape.SamplePosition(sampler)
	dir := sample.Sub(dest)
	// prob with respect to solid angle
	probAngle := r.Shape.SamplePdf(geo.Ray{dest, dir})
	return true, probAngle, r.Spectr.Clone(), sample
}

func (l *AreaLight) GetRadiance(ray geo.Ray) spectra.Spectr {
	//fmt.Println(">")
	ok, hp := l.Shape.RayIntersection(ray)
	///fmt.Println("<")
	if !ok || hp.Normal.Scalar(ray.Direction) > 0 {
		//fmt.Println("a", ok)
		return spectra.NewRGBSpectr(0, 0, 0)
	}
		//fmt.Println("b")
	return l.Spectr.Clone()
}

func (l *AreaLight) Power() float32 {
	return l.Spectr.Power() * math.Pi * 2 * l.Shape.Area()
}

type InfiniteAreaLight struct {
	Texture img.Image3
	Distribution sampling.Distribution2D
	Scale float32
	Direction float32
	SceneRadius float32
}

func NewInfiniteAreaLight(texture img.Image3, scale float32) *InfiniteAreaLight {
	texture.Mul(scale)
	dist := texture.FitInRectangle(512, 512)
	dist.ChangeSpace(colors.XYZSpace)
	dist1 := dist.GetImage1()
	return &InfiniteAreaLight{
		Texture: texture,
		Distribution: sampling.NewDistribution2D(dist1),
		Direction: 0,
		SceneRadius: 1,
	}
}

func (l *InfiniteAreaLight) SetSceneRadius(radius float32) {
	l.SceneRadius = radius
}

func (l *InfiniteAreaLight) SampleRadiance(
	dest geo.Vec3,
	sampler sampling.Sampler2D,
) (
	ok bool,
	prob float32,
	spectr spectra.Spectr,
	origin geo.Vec3,
) {
	e1, e2 := sampler.Next()
	u, v, pdf := l.Distribution.Sample(e1, e2)
	spectr = spectra.NewRGBSpectr(l.Texture.At(
		u*float32(l.Texture.W - 1), (v)*float32(l.Texture.H - 1)))
	azimuthAngle := u*2*math.Pi + l.Direction
	zenithAngle := v*math.Pi
	zenithSin := math32.Sin(zenithAngle)
	prob = pdf/(2*math.Pi*math.Pi*zenithSin)
	origin = dest.Add(geo.Vec3{
		l.SceneRadius*zenithSin*math32.Cos(azimuthAngle),
		l.SceneRadius*zenithSin*math32.Sin(azimuthAngle),
		l.SceneRadius*math32.Cos(zenithAngle),
	})
	ok = true
	return
}

func (l *InfiniteAreaLight) PDF(origin, direction geo.Vec3) float32 {
	zenithCos, azimuthSin, azimuthCos := geo.SphericalFromVec3(direction.Normalized())
	azimuth := math32.Atan2(azimuthSin, azimuthCos) - l.Direction
	zenith := math32.Acos(zenithCos)
	for azimuth < 0 {
		azimuth += 2*math.Pi
	}
	u := azimuth / 2 / math.Pi
	v := zenith / math.Pi
	zenithSin := math32.Sqrt(1 - zenithCos*zenithCos)
	pdf := l.Distribution.Pdf(u, v)
	pdf = pdf/(2*math.Pi*math.Pi*zenithSin)
	return pdf
}

func (l *InfiniteAreaLight) GetRadiance(ray geo.Ray) spectra.Spectr {
	zenithCos, azimuthSin, azimuthCos := geo.SphericalFromVec3(ray.Direction.Normalized())
	azimuth := math32.Atan2(azimuthSin, azimuthCos) - l.Direction
	zenith := math32.Acos(zenithCos)
	for azimuth < 0 {
		azimuth += 2*math.Pi
	}
	u := azimuth / 2 / math.Pi
	v := zenith / math.Pi
	return spectra.NewRGBSpectr(l.Texture.At(
		u*float32(l.Texture.W - 1), (v)*float32(l.Texture.H - 1)))
}

func (l *InfiniteAreaLight) Power() float32 {
	var avg float32
	for y := 0; y < l.Texture.H; y++ {
		p := 3 * y * l.Texture.W
		for x := 0; x < l.Texture.W*3; x++ {
			avg += l.Texture.Data[p]
			p++
		}
	}
	// TODO this not true power, only a sum of rgb values
	avg = 3*avg/float32(l.Texture.H*l.Texture.W)
	return avg * math.Pi * l.SceneRadius * l.SceneRadius
}

func Noop() {}

func init() {
	debug.Noop()
	_ = fmt.Println
}
