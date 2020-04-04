package tracers


import (
	"ly/spectra"
	"ly/geo"
	"ly/debug"
	"math/rand"
	"ly/util/math32"
	"ly/scene"
)

type FTLTracer struct {
	minDepth int
	terminationProb float32
	NFrames int
	Fps float32
	TimeOffset float32
	LightDuration float32
	SkipFirstSegment bool
}

func NewFTLTracer(
	minDepth int,
	terminationProb float32,
	nFrames int,
	fps float32,
) *FTLTracer {
	if minDepth == 0 {
		minDepth = 3
	}
	if terminationProb == 0 {
		terminationProb = 0.3
	}
	return &FTLTracer{
		minDepth: minDepth,
		terminationProb: terminationProb,
		NFrames: nFrames,
		LightDuration: 1,
		Fps: fps,
		TimeOffset: 0,
		SkipFirstSegment: false,
	}
}

func (t FTLTracer) Trace(ray geo.Ray, world *scene.Scene) *spectra.TimedSpectr {
	ret := spectra.NewTimedSpectr(t.NFrames, spectra.NewRGBSpectr(0, 0, 0))
	beta := spectra.NewRGBSpectr(1, 1, 1) // current path throughput
	var pathLength float32
	for depth := 0; ; depth++ {
		debug.D = depth
		hit := world.CastRay(ray)
		if hit == nil {
			break
		}
		if depth > 0 || !t.SkipFirstSegment {
			pathLength += hit.Point.Sub(ray.Origin).Len()
		}
		if hit.Shading.Glow != nil {
			glow := hit.Shading.Glow.Clone()
			glow.BSDF(beta)
			startFrame := int((pathLength - t.TimeOffset) * t.Fps)
			if startFrame < 0 {
				startFrame = 0
			}
			endFrame := int((pathLength - t.TimeOffset + t.LightDuration) * t.Fps)
			if endFrame > t.NFrames {
				endFrame = t.NFrames
			}
			for i := startFrame; i < endFrame; i++ {
				ret.Frames[i].SpectrAdd(glow)
			}
		}
		if depth >= t.minDepth {
			roulette := rand.Float32()
			if roulette <= t.terminationProb {
				break
			}
			beta.Mul(1/(1 - t.terminationProb))
		}
		// create new ray
		var prob float32
		var bsdf spectra.Spectr
		material := hit.Shading.Material
		bsdf, ray, prob, _ = material.BSDFSample(hit, ray.Direction)
		if prob == 0 {
			break
		}
		ray.Origin = ray.Origin.Add(ray.Direction.Normalized().Mul(0.0001)) // kostil
		beta.BSDF(bsdf)
		cos := math32.Abs(ray.Direction.Scalar(hit.ShadingNormal))
		beta.Mul(cos/prob)
	}
	return ret
}
