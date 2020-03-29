package tracers

import (
	"fmt"
	"math"
	"math/rand"
	"ly/geo"
	"ly/scene"
	"ly/spectra"
	"ly/debug"
	"ly/sampling"
	"ly/util/math32"
)

type PathTracer struct {
	minDepth int
	terminationProb float32
}

func NewPathTracer(minDepth int, terminationProb float32) PathTracer {
	if minDepth == 0 {
		minDepth = 3
	}
	if terminationProb == 0 {
		terminationProb = 0.3
	}
	return PathTracer{minDepth: minDepth, terminationProb: terminationProb}
}

var IX = 50
var IY = 197

func (t PathTracer) Trace(ray geo.Ray, world *scene.Scene) (Lsum spectra.Spectr) {
	specularBounce := false
	Lsum = spectra.NewRGBSpectr(0, 0, 0)
	beta := spectra.NewRGBSpectr(1, 1, 1) // current path throughput
	sampler := sampling.NewUniform2D()
	for depth := 0; ; depth++ {
		debug.D = depth
		hit := world.CastRay(ray)
		if (depth == 0 || specularBounce) {
			if hit == nil {
				// need a separte list for area lights
				// because for example we may have 10k small triangle lights
				// and one sky light
				for _, light := range world.NonAreaLights {
					Lsum.SpectrAdd(light.GetRadiance(ray).SpectrMul(beta))
				}
				break
			}
			if hit.Shading.Glow != nil {
				glow := hit.Shading.Glow.Clone()
				glow.BSDF(beta)
				Lsum.SpectrAdd(glow)
			}
		}
		if hit == nil {
			break
		}
		if depth >= t.minDepth {
			roulette := rand.Float32()
			if roulette <= t.terminationProb {
				break
			}
			beta.Mul(1/(1 - t.terminationProb))
		}
		L := EstimateDirectIntegralOneLight(world, hit, ray.Direction, sampler, false)
		L.BSDF(beta)
		Lsum.SpectrAdd(L)
		// create new ray
		var prob float32
		var bsdf spectra.Spectr
		oldray := ray
		_ = oldray
		material := hit.Shading.Material
		bsdf, ray, prob, specularBounce = material.BSDFSample(hit, ray.Direction)
		if prob == 0 {
			// tupik!
			break
		}
		//ray.Origin.Add(hit.Normal.Mul(0.0001)) // kostil
		ray.Origin = ray.Origin.Add(ray.Direction.Normalized().Mul(0.0001)) // kostil
		beta.BSDF(bsdf)
		cos := math32.Abs(ray.Direction.Scalar(hit.ShadingNormal))
		if true || !specularBounce { // ??
			if depth == 1 {
				//break
			}
			beta.Mul(cos/prob)
		} else if (depth == 0) {
		}
	}
	return
}

func init() {
	_ = rand.Intn
	_ = math.Cos
	_ = fmt.Print
	debug.Noop()
}
