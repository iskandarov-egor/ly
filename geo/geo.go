package geo

import (
	"fmt"
	"math"
	"unsafe"
	"ly/util/math32"
)

type Vec3 struct {
	X float32
	Y float32
	Z float32
}

var vec3tmp Vec3

type Axis int

const (
	AxisX Axis = 0
	AxisY Axis = 1
	AxisZ Axis = 2
)

var axisOffsets = []uintptr{unsafe.Offsetof(vec3tmp.X), unsafe.Offsetof(vec3tmp.Y), unsafe.Offsetof(vec3tmp.Z)}

func (v *Vec3) Axis(axis Axis) float32 {
	off := axisOffsets[axis]
	ptr := (*float32)(unsafe.Pointer(uintptr(unsafe.Pointer(v)) + off))
	return *ptr
	/*
	switch axis {
		case 0: return v.X
		case 1: return v.Y
		case 2: return v.Z
	}
	return 0
	*/
	//panic("panic")
}

func (v *Vec3) AxisP(axis Axis) *float32 {
	off := axisOffsets[axis]
	ptr := (*float32)(unsafe.Pointer(uintptr(unsafe.Pointer(v)) + off))
	return ptr
	/*
	switch axis {
		case 0: return &v.X
		case 1: return &v.Y
		case 2: return &v.Z
	}
	return nil
	*/
	//panic("panic")
}

func Vec3FromSpherical(zenith, azimuth float32) Vec3 {
	c := math32.Sin(zenith)
	return Vec3{
		math32.Cos(azimuth)*c,
		math32.Sin(azimuth)*c,
		c,
	}
}

// get spherical coords from vector.
// inverse of Vec3FromSpherical.
// @vec must be normalized
func SphericalFromVec3(vec Vec3) (zenithCos, azimuthSin, azimuthCos float32) {
	div := math32.Sqrt(vec.X*vec.X + vec.Y*vec.Y)
	zenithCos = vec.Z
	if div == 0 {
		return
	}
	azimuthCos = math32.Clamp(vec.X/div, -1, 1)
	azimuthSin = math32.Clamp(vec.Y/div, -1, 1)
	return
}

func (a Vec3) Add(b Vec3) Vec3 {
	return Vec3{a.X + b.X, a.Y + b.Y, a.Z + b.Z}
}

func (a Vec3) Sub(b Vec3) Vec3 {
	return Vec3{a.X - b.X, a.Y - b.Y, a.Z - b.Z}
}

func (v Vec3) Len() float32 {
	return float32(math.Sqrt(float64(v.X*v.X + v.Y*v.Y + v.Z*v.Z)))
}

func (v Vec3) LenSquared() float32 {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

func (v Vec3) Mul(k float32) Vec3 {
	return Vec3{v.X * k, v.Y * k, v.Z * k}
}

func (v Vec3) Div(k float32) Vec3 {
	return Vec3{v.X / k, v.Y / k, v.Z / k}
}

func (v Vec3) Scalar(u Vec3) float32 {
	return v.X*u.X + v.Y*u.Y + v.Z*u.Z
}

// project v onto u
func (v Vec3) VectorProj(u Vec3) float32 {
	return v.Scalar(u)/u.Len()
}

func (v Vec3) Cross(u Vec3) (ret Vec3) {
	ret.X = v.Y*u.Z - v.Z*u.Y
	ret.Y = v.Z*u.X - v.X*u.Z
	ret.Z = v.X*u.Y - v.Y*u.X
	return
}

func (v Vec3) PlaneProj(norm Vec3) Vec3 {
	nproj := v.VectorProj(norm)
	return v.Sub(norm.Normalized().Mul(nproj))
}

func (v Vec3) Normalized() Vec3 {
	l := v.Len()
	if l == 0 {
		return v
	}
	return v.Div(l)
}

func (v Vec3) NormalizedSafe() Vec3 {
	l := v.Len()
	if l == 0 {
		return v
	} else {
		return v.Div(v.Len())
	}
}

func (v Vec3) Negated() Vec3 {
	return Vec3{0, 0, 0}.Sub(v)
}

func (v Vec3) String() string {
	return fmt.Sprintf("[%.3f, %.3f, %.3f]", v.X, v.Y, v.Z)
	//return fmt.Sprintf("[%.46f, %.46f, %.46f]", v.X, v.Y, v.Z)
}

//  a      ^ n     ret      return reflection of the vector
//  >>     |     >>         from the surface with normal n.
//    >>   |   >>           @n must be normalized.
//      >> | >>             @cosn must be equal to a.Scalar(n)/(a.Len()*n.Len())
//        >|>               
//  --------------------
func (a Vec3) ReflectAround(n Vec3, cosn float32) Vec3 {
	y := n.Mul(cosn)
	//x := a.Sub(y)
	return a.Sub(y.Mul(2)) // == x.Sub(y)
}

type Ray struct {
	Origin    Vec3
	Direction Vec3
}

func (r Ray) At(distance float32) Vec3 {
	return r.Origin.Add(r.Direction.Mul(distance))
}

// axis aligned box
type Box struct {
	Min Vec3
	Max Vec3
}

func (b Box) String() string {
	return fmt.Sprintf("%v..%v", b.Min, b.Max)
}

// makes an 'empty' box
func NewBox() Box {
	return Box{
		Min: Vec3{float32(math.Inf(+1)), float32(math.Inf(+1)), float32(math.Inf(+1))},
		Max: Vec3{float32(math.Inf(-1)), float32(math.Inf(-1)), float32(math.Inf(-1))},
	}
}

func (b1 *Box) Union(b2 Box) (ret Box) {
	ret.Min.X = math32.Min(b1.Min.X, b2.Min.X)
	ret.Min.Y = math32.Min(b1.Min.Y, b2.Min.Y)
	ret.Min.Z = math32.Min(b1.Min.Z, b2.Min.Z)
	ret.Max.X = math32.Max(b1.Max.X, b2.Max.X)
	ret.Max.Y = math32.Max(b1.Max.Y, b2.Max.Y)
	ret.Max.Z = math32.Max(b1.Max.Z, b2.Max.Z)
	return
}

func (b1 *Box) Include(v Vec3) {
	b1.Min.X = math32.Min(b1.Min.X, v.X)
	b1.Min.Y = math32.Min(b1.Min.Y, v.Y)
	b1.Min.Z = math32.Min(b1.Min.Z, v.Z)
	b1.Max.X = math32.Max(b1.Max.X, v.X)
	b1.Max.Y = math32.Max(b1.Max.Y, v.Y)
	b1.Max.Z = math32.Max(b1.Max.Z, v.Z)
}

func (b *Box) Diagonal() Vec3 {
	return b.Max.Sub(b.Min)
}

func isInf(x float32) bool {
	return x > 1e9 || x < -1e9
}

// TODO pass tmax to return quickly
func (b *Box) Intersect(ray Ray) (isHit bool) {
	/* == unoptimized hitSlab ==
	hitSlab := func(axis Axis) (tNear float32, tFar float32) {
		if math.IsInf(float64(inv[axis]), 0) {
			x := ray.Origin.Axis(axis)
			if x >= b.Min.Axis(axis) && x <= b.Max.Axis(axis) {
				return float32(math.Inf(-1)), float32(math.Inf(1))
			} else {
				return float32(math.Inf(1)), float32(math.Inf(-1))
			}
		}
		var near, far float32 = b.Min.Axis(axis), b.Max.Axis(axis)
		if neg[axis] {
			near, far = far, near
		}
		tNear = (near - ray.Origin.Axis(axis))*inv[axis]
		tFar = (far - ray.Origin.Axis(axis))*inv[axis]
		return tNear, tFar
	}
	*/
	hitSlabX := func() (tNear float32, tFar float32) {
		inv := 1/ray.Direction.X
		tNear = (b.Min.X - ray.Origin.X)*inv
		tFar = (b.Max.X - ray.Origin.X)*inv
		if ray.Direction.X < 0 {
			return tFar, tNear
		}
		return tNear, tFar
	}
	hitSlabY := func() (tNear float32, tFar float32) {
		inv := 1/ray.Direction.Y
		tNear = (b.Min.Y - ray.Origin.Y)*inv
		tFar = (b.Max.Y - ray.Origin.Y)*inv
		if ray.Direction.Y < 0 {
			return tFar, tNear
		}
		return tNear, tFar
	}
	hitSlabZ := func() (tNear float32, tFar float32) {
		inv := 1/ray.Direction.Z
		tNear = (b.Min.Z - ray.Origin.Z)*inv
		tFar = (b.Max.Z - ray.Origin.Z)*inv
		if ray.Direction.Z < 0 {
			return tFar, tNear
		}
		return tNear, tFar
	}
	near, far := hitSlabX()
	yNear, yFar := hitSlabY()
	//fmt.Println(near, far, yNear, yFar)
	if near > yFar || yNear > far {
		return false
	}
	if yNear > near {
		near = yNear
	}
	if yFar < far {
		far = yFar
	}
	zNear, zFar := hitSlabZ()
	if near > zFar || zNear > far {
		return false
	}
	//fmt.Println(near, far, zNear, zFar)
	if zFar > far {
		far = zFar
	}
	//fmt.Println(near, far)

	return far >= 0
}

func main() {
	fmt.Println("vim-go")
}
