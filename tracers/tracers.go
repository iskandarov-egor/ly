package tracers

import (
	"fmt"
	"math/rand"
	"ly/geo"
	"ly/scene"
	"ly/spectra"
	"ly/debug"
	"ly/colors"
	"ly/util/math32"
)

type Tracer interface {
	Trace(ray geo.Ray, world *scene.Scene) spectra.Spectr
}

var PX1 int = -150
var PY1 int = 200
var PX2 int = 79
var PY2 int = 73

type GrayTracer struct {
}

func NewGrayTracer() GrayTracer {
	return GrayTracer{}
}

func (t GrayTracer) Trace(ray geo.Ray, world *scene.Scene) spectra.Spectr {
	hit := world.CastRay(ray)
	if hit == nil {
		return spectra.NewRGBSpectr(0, 0, 0)
	} else {
		if mat, ok := hit.Shading.Material.(*scene.MatteMaterial); ok {
			u := hit.U*float32(mat.Texture.W)
			v := hit.V*float32(mat.Texture.H)
			return spectra.NewRGBSpectr(mat.Texture.At(u, v))
		} else {
			return spectra.NewRGBSpectr(1, 1, 1)
		}
	}
}

type Inspector struct {
	Callback func(ray geo.Ray, world *scene.Scene) spectra.Spectr
}

func NewInspector() Inspector {
	return Inspector{}
}

func (t Inspector) Trace(ray geo.Ray, world *scene.Scene) spectra.Spectr {
	return t.Callback(ray, world)
}

func init() {
	debug.Noop()
	_ = rand.Intn
	_ = math32.Sin
	colors.Noop()
}

func main() {
	fmt.Println("vim-go")
}
