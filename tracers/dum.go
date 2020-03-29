package tracers

import (
	"fmt"
	"ly/geo"
	"ly/scene"
	"ly/spectra"
	"ly/debug"
	"ly/sampling"
	"ly/util/math32"
)

// simple dumb tracer without MIS sorcery
type DumTracer struct {
}

func NewDumTracer() DumTracer {
	return DumTracer{}
}

func (t DumTracer) Trace(ray geo.Ray, world *scene.Scene) spectra.Spectr {
	hit := world.CastRay(ray)
	if hit == nil {
		return spectra.NewRGBSpectr(0, 0, 0)
	} else if hit.Shading.Glow != nil {
		return hit.Shading.Glow.Clone()
	} else {
		var Lsum spectra.Spectr = spectra.NewRGBSpectr(0, 0, 0)
		
		nLightSamples := 1
		for _, light := range world.Lights {
			sampler := sampling.NewSampler2D(nLightSamples)
			for i := 0; i < nLightSamples; i++ {
				debug.S = i
				var L spectra.Spectr
				ok, pdf, L, source := light.SampleRadiance(hit.Point, sampler)
				if !ok {
					continue
				}
				shadowRay := geo.Ray{
					Origin: hit.Point,
					Direction: source.Sub(hit.Point),
				}
				// kostil
				kostil := shadowRay.Direction.Normalized().Mul(0.0001)
				shadowRay.Origin    = shadowRay.Origin.Add(kostil)
				shadowRay.Direction = shadowRay.Direction.Sub(kostil)
				hit2 := world.CastRay(shadowRay)
				if hit2 != nil && hit2.RayT < 0.999 { // kostil
					continue
				}
				dir := source.Sub(hit.Point)
				cosTheta := math32.Abs(dir.Normalized().Scalar(hit.ShadingNormal))
				L.Mul(cosTheta/pdf)
				material := hit.Shading.Material
				bsdf := material.BSDF(hit, dir, ray.Direction)
				L.BSDF(bsdf)
				Lsum.SpectrAdd(L)
			}
		}
		Lsum.Mul(1/float32(nLightSamples))

		return Lsum
	}
}

func init() {
	_ = fmt.Print
}
