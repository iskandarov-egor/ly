package tracers

import (
	"fmt"
	"ly/geo"
	"ly/scene"
	"ly/spectra"
	"ly/debug"
	"math"
	"math/rand"
	"ly/sampling"
	"ly/util/math32"
)

type DirectTracer struct {
}

func NewDirectTracer() DirectTracer {
	return DirectTracer{}
}

func EstimateDirectLightContribution(
	world *scene.Scene,
	hit *scene.ShapeHitPoint,
	dirOut geo.Vec3,
	light scene.Light,
	sampler sampling.Sampler2D,
	allowSpecularBSDF bool,
) spectra.Spectr {
	Lsum := spectra.NewRGBSpectr(0, 0, 0)
	// MIS: sample the light
	if 0 == 0 {
	switch 1 {
		default:
		ok, pdf, L, source := light.SampleRadiance(hit.Point, sampler)
		if !ok || pdf == 0 {
			break
		}
		shadowRay := geo.Ray{
			Origin: hit.Point.Add(source.Sub(hit.Point).Normalized().Mul(0.00001)), // kostil
			Direction: source.Sub(hit.Point),
		}
		hit2 := world.CastRay(shadowRay)
		if hit2 != nil && hit2.RayT < 0.999 { // kostil
			break
		}
		dir := source.Sub(hit.Point)
		cosTheta := math32.Abs(dir.Normalized().Scalar(hit.ShadingNormal))

		pdf2 := hit.Shading.Material.PDF(hit, dir, dirOut)
		weight := (pdf*pdf) / (pdf*pdf + pdf2*pdf2) // power heuristic
		//weight = 1

		L.Mul(weight * cosTheta/pdf)
		material := hit.Shading.Material
		bsdf := material.BSDF(hit, dir, dirOut)
		L.BSDF(bsdf)
		Lsum.SpectrAdd(L)
	}}
	// MIS: sample the BSDF
	if 0 == 0 {
	switch 1 {
		default:
		bsdf, bsdfRay, pdf, specular := hit.Shading.Material.BSDFSample(hit, dirOut)
		if specular && !allowSpecularBSDF {
			break
		}
		if pdf == 0 {
			break
		}
		// kostil
		bsdfRay.Origin = bsdfRay.Origin.Add(bsdfRay.Direction.Normalized().Mul(0.00001))
		hit2 := world.CastRay(bsdfRay)
		var lightPdf float32
		var L spectra.Spectr
		if hit2 == nil {
			L = light.GetRadiance(bsdfRay)
		} else {
			if areaLight, ok := light.(*scene.AreaLight); ok {
				if areaLight.Shape != hit2.Shape {
					break
				}

				L = areaLight.Spectr.Clone()
			} else {
				// TODO
				break
			}
		}
		cosTheta := math32.Abs(bsdfRay.Direction.Normalized().Scalar(hit.ShadingNormal))
		var weight float32 = 1
		if !specular {
			lightPdf = light.PDF(bsdfRay.Origin, bsdfRay.Direction)
			weight = (pdf*pdf) / (pdf*pdf + lightPdf*lightPdf) // power heuristic
		}
		//weight = 1

		L.Mul(weight * cosTheta/pdf)
		L.BSDF(bsdf)
		Lsum.SpectrAdd(L)
	}}

	return Lsum
}

func EstimateDirectIntegralOneLight(
	world *scene.Scene,
	hp *scene.ShapeHitPoint,
	dirOut geo.Vec3,
	sampler sampling.Sampler2D,
	allowSpecularBSDF bool,
) spectra.Spectr {
	//light := world.Lights[rand.Intn(len(world.Lights))]
	light, prob := world.SampleLight()
	L := EstimateDirectLightContribution(world, hp, dirOut, light, sampler, allowSpecularBSDF)
	//return L.Mul(float32(len(world.Lights)))
	return L.Mul(1/prob)
}

/*
	light := world.Lights[rand.Intn(len(world.Lights))]
	L := EstimateDirectLightContribution(world, hp, dirOut, light, sampler, allowSpecularBSDF)
	return L.Mul(float32(len(world.Lights)))


*/

func (t DirectTracer) Trace(ray geo.Ray, world *scene.Scene) spectra.Spectr {
	hit := world.CastRay(ray)
	sampler := sampling.NewUniform2D()
	if hit == nil {
		return spectra.NewRGBSpectr(0, 0, 0)
	} else if hit.Shading.Glow != nil {
		return hit.Shading.Glow.Clone()
	} else {
		Lsum := EstimateDirectIntegralOneLight(world, hit, ray.Direction, sampler, true)
		return Lsum
	}
}

func init() {
	_ = fmt.Print
	_ = math.Sin
	_ = rand.Intn
	debug.Noop()
	sampling.Noop()
}
