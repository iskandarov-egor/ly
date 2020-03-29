package cameras

import (
	"fmt"
	"math"
	"ly/geo"
)

type Camera interface {
	GenerateRay(x, y float32) geo.Ray
	PlotDot(geo.Vec3) (x, t float32)
}

type OrthoCamera struct {
	Position geo.Vec3
	Direction geo.Vec3
	Zoom float32
	up geo.Vec3
	right geo.Vec3
}

func NewOrthoCamera(pos geo.Vec3, dir geo.Vec3, zoom float32) *OrthoCamera {
	up, right := cameraFrame(dir)
	return &OrthoCamera{
		Position: pos,
		Direction: dir,
		Zoom: zoom,
		up: up,
		right: right,
	}
}

func cameraFrame(dir geo.Vec3) (up geo.Vec3, right geo.Vec3) {
	up = geo.Vec3{X: 0, Y: 0, Z: 1}
	if dir.X == 0 && dir.Y == 0 {
		up = geo.Vec3{0, 1, 0}
		right = geo.Vec3{1, 0, 0}
		return
	}
	up = up.PlaneProj(dir).Normalized()
	right = dir.Cross(up).Normalized()
	return
}

func (c *OrthoCamera) GenerateRay(x, y float32) geo.Ray {
	x /= c.Zoom
	y /= c.Zoom
	ray := geo.Ray{
		Origin: c.Position.Add(c.up.Mul(y).Add(c.right.Mul(x))),
		Direction: c.Direction,
	}
	return ray
}

func (c *OrthoCamera) PlotDot(dot geo.Vec3) (x, y float32) {
	panic("not impl")
}

type PerspectiveCamera struct {
	Position geo.Vec3
	Direction geo.Vec3
	Fov float32
	up geo.Vec3
	right geo.Vec3
}

func NewPerspectiveCamera(pos geo.Vec3, dir geo.Vec3, fov, zoom float32,
) *PerspectiveCamera {
	up, right := cameraFrame(dir)
	return &PerspectiveCamera{
		Position: pos,
		Direction: dir.Normalized().Mul(float32(0.5/math.Tan(float64(fov/2)))),
		Fov: fov,
		up: up.Mul(1/zoom),
		right: right.Mul(1/zoom),
	}
}

func (c *PerspectiveCamera) GenerateRay(x, y float32) geo.Ray {
	// obtain the point in world space that lies on the screen plane
	tmp := c.Position.Add(c.up.Mul(y).Add(c.right.Mul(x)))
	point := tmp.Add(c.Direction)
	// make a ray from camera position through z1
	ray := geo.Ray{
		Origin: c.Position,
		Direction: point.Sub(c.Position),
	}
	return ray
}

func (c *PerspectiveCamera) PlotDot(dot geo.Vec3) (x, y float32) {
	dot = dot.Sub(c.Position)
	z := dot.VectorProj(c.Direction)/c.Direction.Len()
	dotZ1 := dot.Div(z)
	x = dotZ1.VectorProj(c.right)
	y = dotZ1.VectorProj(c.up)
	return
}

// perspective camera that looks at a 1x1x1 cube in the Y direction
func New1x1Camera() *PerspectiveCamera {
	origin := geo.Vec3{
		0,
		-1.31-1,
		0,
	}
	return NewPerspectiveCamera(origin, geo.Vec3{0,1,0}, 0.27*2, 1)
}

func main() {
	fmt.Println("vim-go")
}
