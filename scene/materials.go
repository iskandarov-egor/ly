package scene

import (
	"fmt"
	"math"
	"sort"
	"math/rand"
	"ly/img"
	"ly/spectra"
	"ly/colors"
	"ly/geo"
	"ly/debug"
	"ly/sampling"
	"ly/util/pbrt"
	"ly/util/math32"
)

type Material interface {
	// dirIn - vector from the point to the light source
	// dirOut - vector from the eye to the point
	BSDF(hp *ShapeHitPoint, dirIn, dirOut geo.Vec3) (bsdf spectra.Spectr)
	PDF(hp *ShapeHitPoint, dirIn, dirOut geo.Vec3) float32
	// ray will be normalized
	BSDFSample(hp *ShapeHitPoint, dirOut geo.Vec3) (bsdf spectra.Spectr, ray geo.Ray,
		prob float32, specular bool)
	// true if BSDF() always returns 0
	BSDF0() bool
}

type FourierMaterial struct {
	Table *pbrt.FourierBSDFTable
}

// get spherical coordinates from vectors.
// @zenith - the zenith vector
// @x, y - target vectors
// returns the cosines of angles:
// - zenith angle of x
// - zenith angle of y
// - azimuth difference between x and y
// all 3 args must be normalized
func Spherical(zenith, x, y geo.Vec3) (xZenithCos, yZenithCos, azimuthDiffCos float32) {
	xZenithCos = zenith.Scalar(x)
	yZenithCos = zenith.Scalar(y)
	xProj := x.Sub(zenith.Mul(xZenithCos))
	yProj := y.Sub(zenith.Mul(yZenithCos))
	div := math32.Sqrt((xProj.X*xProj.X + xProj.Y*xProj.Y)*(yProj.X*yProj.X + yProj.Y*yProj.Y))
	azimuthDiffCos = math32.Clamp(xProj.Scalar(yProj)/div, -1, 1)
	return
}

func NewFourierMaterial(path string) *FourierMaterial {
	tab, err := pbrt.ReadFourierBSDF(path)
	if err != nil {
		panic(fmt.Sprintf("create fourier material: read table: %v", err))
	}
	return &FourierMaterial{
		Table: tab,
	}
}

func (m *FourierMaterial) PDF(hp *ShapeHitPoint, dirIn, dirOut geo.Vec3) float32 {
	panic("not impl")
}

func (m *FourierMaterial) BSDF0() bool {
	return false
}

func (m *FourierMaterial) BSDF(p *ShapeHitPoint, dirIn, dirOut geo.Vec3) (bsdf spectra.Spectr) {
	// bsdf table requires inverted directions
	dirIn, dirOut = dirIn.Normalized().Negated(), dirOut.Normalized().Negated()
	// compute zenith and azimuth angle cosines
	muI, muO, cosPhi := Spherical(p.Normal, dirIn, dirOut)

	tab := m.Table
	oi := sort.Search(len(tab.Mu), func(i int) bool { return tab.Mu[i] > muI }) - 1
	oo := sort.Search(len(tab.Mu), func(i int) bool { return tab.Mu[i] > muO }) - 1
	oi2 := oi + 1
	oo2 := oo + 1

	if oi < 0 || oo < 0 {
		panic(fmt.Sprintf("binary search fail for directions %g %g", muI, muO))
	}

	oi_inv_weight := 1/(tab.Mu[oi2] - tab.Mu[oi])
	oo_inv_weight := 1/(tab.Mu[oo2] - tab.Mu[oo])
	oi2_weight := (muI - tab.Mu[oi])*oi_inv_weight
	oi1_weight := (tab.Mu[oi2] - muI)*oi_inv_weight
	oo2_weight := (muO - tab.Mu[oo])*oo_inv_weight
	oo1_weight := (tab.Mu[oo2] - muO)*oo_inv_weight

	get := func(oo int, oi int) (Y, R, B float32) {
		pos := oo*int(tab.NMu) + oi
		offset := tab.AOffset[pos]
		order := tab.M[pos]
		Y = math32.Fourier(tab.A[offset:offset + order], cosPhi)
		R = math32.Fourier(tab.A[offset + order:offset + 2*order], cosPhi)
		B = math32.Fourier(tab.A[offset + 2*order:offset + 3*order], cosPhi)
		return
	}
	Y, R, B := get(oo, oi)
	Y12, R12, B12 := get(oo, oi2)
	Y21, R21, B21 := get(oo2, oi)
	Y22, R22, B22 := get(oo2, oi2)

	w11 := oi1_weight * oo1_weight
	w12 := oi1_weight * oo2_weight
	w21 := oi2_weight * oo1_weight
	w22 := oi2_weight * oo2_weight

	Y = Y*w11 + Y12*w12 + Y21*w21 + Y22 * w22
	B = B*w11 + B12*w12 + B21*w21 + B22 * w22
	R = R*w11 + R12*w12 + R21*w21 + R22 * w22
	G := 1.39829 * Y - 0.100913 * B - 0.297375 * R
	scale := float32(0)
	if muI != 0 {
		scale = 1 / math32.Abs(muI)
	}
	//scale = 1
	if R < 0 { R = 0 }
	if G < 0 { G = 0 }
	if B < 0 { B = 0 }
	R, G, B = scale * R, scale * G, scale * B
	return &spectra.RGBSpectr{R, G, B}
}

func (m *FourierMaterial) BSDFSample(hp *ShapeHitPoint, dirOut geo.Vec3) (bsdf spectra.Spectr, ray geo.Ray, prob float32, specular bool) {
	rnd := rand.Float32()
	rnd2 := rand.Float32()
	tab := m.Table
	muO := -dirOut.Normalized().Scalar(hp.Normal)
	if muO < 0 {
		// fuck this
		return
	}
	oo := sort.Search(len(tab.Mu), func(i int) bool { return tab.Mu[i] > muO }) - 1
	if oo < 0 {
		panic(fmt.Sprintf("binary search fail for direction %g", muO))
	}
	cdfOffset := (oo + 1)*len(tab.Mu)
	maxCdf := tab.Cdf[(oo + 2)*len(tab.Mu) - 1]
	rnd *= maxCdf // scale rnd
	oi := sort.Search(len(tab.Mu), func(i int) bool { return tab.Cdf[cdfOffset + i] > rnd }) - 1
	if oi < 0 {
		panic(fmt.Sprintf("binary search fail for cdf of direction %g", muO))
	}

	// prob to choose this oi
	dCdf := tab.Cdf[cdfOffset + oi + 1] - tab.Cdf[cdfOffset + oi]
	oiProb := (dCdf) / maxCdf
	// choose cosine within oi
	rnd = (rnd - tab.Cdf[cdfOffset + oi]) / dCdf
	muI := math32.Lerp(tab.Mu[oi], tab.Mu[oi + 1], rnd)
	// prob to choose this cosine
	cosProb := oiProb / (tab.Mu[oi + 1] - tab.Mu[oi])
	// choose azimuth angle
	azim := rnd2 * math.Pi * 2
	azimProb := float32(1/(2 * math.Pi))
	// overall prob
	// overall prob wrt solid angle is apparantly the same
	prob = cosProb * azimProb

	muI = -muI

	sinIn := math32.SafeSqrt(1 - muI*muI)
	z := muI
	x := sinIn*math32.Cos(azim)
	y := sinIn*math32.Sin(azim)

	bx, by := BasisAroundVector(hp.Normal)
	ray = geo.Ray{
		hp.Point,
		VectorFromBasis(bx, by, hp.Normal, x, y, z),
	}
	bsdf = m.BSDF(hp, ray.Direction, dirOut)
			
	return
}

type MatteMaterial struct {
	Texture img.Image3
	Roughness float32
	A, B float32
	IsTransparent bool
}

func New1ColorMatteMaterial(r, g, b, roughness float32, isTransparent bool) *MatteMaterial {
	im := img.NewImage3(1, 1, colors.RGBSpace)
	im.Data[0], im.Data[1], im.Data[2] = r, g, b
	return NewMatteMaterial(im, roughness, isTransparent)
}

func NewMatteMaterial(txt img.Image3, roughness float32, isTransparent bool) *MatteMaterial {
	sig := roughness
	A := 1 - 0.5*(sig*sig)/(sig*sig + 0.33)
	B := 0.45*(sig*sig)/(sig*sig + 0.09)
	return &MatteMaterial{txt, roughness, A, B, isTransparent}
}

func BasisAroundVector(z geo.Vec3) (x, y geo.Vec3) {
	if z.Y == 0 && z.X == 0 {
		x = geo.Vec3{z.Z, 0, 0}.Normalized()
	} else {
		x = geo.Vec3{-z.Y, z.X, 0}.Normalized()
	}
	y = x.Cross(z)
	return
}

func VectorFromBasis(X, Y, Z geo.Vec3, x, y, z float32) (geo.Vec3) {
	return X.Mul(x).Add(Y.Mul(y)).Add(Z.Mul(z))
}

func (m *MatteMaterial) PDF(hp *ShapeHitPoint, dirIn, dirOut geo.Vec3) float32 {
	cosIn := dirIn.Normalized().Scalar(hp.Normal)
	if m.IsTransparent {
		return math32.Abs(cosIn)/math.Pi/2
	}
	cosOut := dirOut.Normalized().Scalar(hp.Normal)
	if (cosIn > 0) == (cosOut > 0) {
		return 0
	}
	return math32.Abs(cosIn)/(math.Pi)
}

func (m *MatteMaterial) BSDF0() bool {
	return false
}
func (m *MatteMaterial) BSDF(hp *ShapeHitPoint, dirIn, dirOut geo.Vec3) (L spectra.Spectr) {
	dirIn = dirIn.Normalized()
	dirOut = dirOut.Normalized()
	cosZenIn := hp.Normal.Scalar(dirIn)
	cosZenOut := hp.Normal.Scalar(dirOut)

	var mul float32 = 1
	if m.IsTransparent {
		mul = 0.5
	} else if (cosZenIn > 0) == (cosZenOut > 0) {
		return spectra.NewRGBSpectr(0, 0, 0)
	}
	cosZenIn = hp.ShadingNormal.Scalar(dirIn)
	cosZenOut = hp.ShadingNormal.Scalar(dirOut)
	if math.IsNaN(float64(hp.U)) {
		panic("aaa")
	}
	r, g, b := m.Texture.AtUv(hp.U, hp.V)
	color := spectra.NewRGBSpectr(r, g, b)
	L = color
	if m.Roughness == 0 {
		L.Mul(mul/math.Pi)
	} else {
		cosThetaIn := cosZenIn
		cosThetaOut := cosZenOut
		if cosThetaIn < 0 {
			cosThetaIn = -cosThetaIn
		} else {
			cosThetaOut = -cosThetaOut
		}
		var sinAlpha, tanBeta float32
		if cosThetaIn > cosThetaOut {
			sinAlpha = math32.SafeSqrt(1 - cosThetaOut*cosThetaOut)
			tanBeta = math32.SafeSqrt(1 - cosThetaIn*cosThetaIn) / cosThetaIn
		} else {
			sinAlpha = math32.SafeSqrt(1 - cosThetaIn*cosThetaIn)
			tanBeta = math32.SafeSqrt(1 - cosThetaOut*cosThetaOut) / cosThetaOut
		}
		bx, by := BasisAroundVector(hp.ShadingNormal)
		projIn := dirIn.PlaneProj(hp.ShadingNormal).Normalized()
		projOut := dirOut.PlaneProj(hp.ShadingNormal).Normalized()
		cosAzimIn := projIn.Scalar(bx)
		cosAzimOut := projOut.Scalar(bx)
		sinAzimIn := projIn.Scalar(by)
		sinAzimOut := projOut.Scalar(by)
		cosDeltaAzim := cosAzimIn*cosAzimOut + sinAzimIn*sinAzimOut
		// oren-nayar formula
		mul *= (m.A + m.B*math32.Max(0, cosDeltaAzim)*sinAlpha*tanBeta)/math.Pi
		L.Mul(mul)
	}

	return
	/*
	if obj.BumpMap != nil {
		// f - bump map value at the point
		// s - surface without the bump map. lets find its derivatives
		u := hitPoint.U*float32(obj.BumpMap.W)
		v := hitPoint.V*float32(obj.BumpMap.W)
		f, dfdu, dfdv := obj.BumpMap.Derivative(u, v, 2.5) // TODO 0.5 kostil
		f *= 15
		dfdu *= 15
		dfdv *= 15
		dsdu :=
			hitPoint.Dpdu.Add(hitPoint.Normal.Mul(dfdu)).Add(hitPoint.Dndu.Mul(f))
		dsdv :=
			hitPoint.Dpdv.Add(hitPoint.Normal.Mul(dfdv)).Add(hitPoint.Dndv.Mul(f))
		newNormal := dsdu.Cross(dsdv).Normalized()
		_, _ = dsdu, dsdv
		hitPoint.Normal = newNormal
	}
	*/
}

func (m *MatteMaterial) BSDFSample(hp *ShapeHitPoint, dirOut geo.Vec3) (bsdf spectra.Spectr, ray geo.Ray, prob float32, specular bool) {
	hemi := sampling.CosineSampleHemisphere().Normalized()
	prob = hemi.Z/(math.Pi)
	if m.IsTransparent {
		prob /= 2
		if rand.Float32() < 0.5 {
			hemi.Z = -hemi.Z
		}
	} else {
		if hp.Normal.Scalar(dirOut) > 0 {
			hemi.Z = -hemi.Z
		}
	}

	bx, by := BasisAroundVector(hp.Normal)
	ray = geo.Ray{
		hp.Point,
		VectorFromBasis(bx, by, hp.Normal, hemi.X, hemi.Y, hemi.Z),
	}
	bsdf = m.BSDF(hp, ray.Direction, dirOut)
	return
}

type MirrorMaterial struct {
	Color spectra.Spectr
}

func NewMirrorMaterial() *MirrorMaterial {
	return &MirrorMaterial{spectra.NewRGBSpectr(1, 1, 1)}
}

func (m *MirrorMaterial) PDF(hp *ShapeHitPoint, dirIn, dirOut geo.Vec3) float32 {
	return 0
}

func (m *MirrorMaterial) BSDF0() bool {
	return true
}
func (m *MirrorMaterial) BSDF(hp *ShapeHitPoint, dirIn, dirOut geo.Vec3) (L spectra.Spectr) {
	// zero probability that dirIn and dirOut are mirror-aligned
	L = &spectra.RGBSpectr{0, 0, 0}
	return
}

func (m *MirrorMaterial) BSDFSample(hp *ShapeHitPoint, dirOut geo.Vec3) (bsdf spectra.Spectr, ray geo.Ray, prob float32, specular bool) {
	proj := hp.Normal.Mul(dirOut.Scalar(hp.Normal)) // N normalized
	ray = geo.Ray{hp.Point, dirOut.Sub(proj.Mul(2)).Normalized()}
	bsdf = m.Color
	prob = 1
	specular = true
	return
}

type PortalMaterial struct {
	Bro *Mesh
}

func NewPortalMaterial(bro *Mesh) *PortalMaterial {
	return &PortalMaterial{bro}
}

func (m *PortalMaterial) BSDF(hp *ShapeHitPoint, dirIn, dirOut geo.Vec3) (L spectra.Spectr) {
	L = &spectra.RGBSpectr{0, 0, 0}
	return
}

func (m *PortalMaterial) BSDFSample(hp *ShapeHitPoint, dirOut geo.Vec3) (bsdf spectra.Spectr, ray geo.Ray, prob float32, specular bool) {
	newP, newDpdu, newDpdv := m.Bro.Uv2xyz(hp.U, hp.V)
	newDpdu, newDpdv = newDpdu.Normalized(), newDpdv.Normalized()
	newNorm := newDpdu.Cross(newDpdv).Normalized()
	dpduProj := dirOut.VectorProj(hp.Dpdu.Normalized())
	dpdvProj := dirOut.VectorProj(hp.Dpdv.Normalized())
	normProj := dirOut.VectorProj(hp.Normal)
	ray = geo.Ray{newP, newDpdu.Mul(dpduProj).Add(newDpdv.Mul(dpdvProj)).Add(newNorm.Mul(normProj)).Normalized()}
	prob = 1
	specular = true
	bsdf = spectra.NewRGBSpectr(1, 1, 1)
	return
}

// fresnel formula for conductors
// nIn - refractive index of conductor inside
// nOut - refractive index of dielectric outside
// k - absorbtion coefficient
// cos - cosine of the angle of incidence
func fresnelConductor(nIn, nOut, k spectra.Spectr, cos float32) spectra.Spectr {
    cos = math32.Clamp(cos, -1, 1);
	n := nOut.Clone().SpectrDiv(nIn)
	nk := k.Clone().SpectrDiv(nIn)

    cos2 := cos * cos;
    sin2 := 1 - cos2;
    n2 := n.SpectrMul(n)
    nk2 := nk.SpectrMul(nk)

	t0 := n2.Clone().SpectrSub(nk2).Add(-sin2)

	a2plusb2 := nk2.SpectrMul(n2).Mul(4).SpectrAdd(t0.Clone().SpectrMul(t0)).Sqrt()

	t1 := a2plusb2.Clone().Add(cos2)

	a := a2plusb2.Clone().SpectrAdd(t0).Mul(0.5).Sqrt()

	t2 := a.Mul(2 * cos)
	t1plust2 := t1.Clone().SpectrAdd(t2)
	Rs := t1.SpectrSub(t2).SpectrDiv(t1plust2)

	t3 := a2plusb2.Mul(cos2).Add(sin2 * sin2)
	t2.Mul(sin2)
	t2plust3 := t2.Clone().SpectrAdd(t3)
	Rp := t3.SpectrSub(t2).SpectrMul(Rs).SpectrDiv(t2plust3)

	return Rp.SpectrAdd(Rs).Mul(0.5)
}

//        |    /       
//        |1 /         
//        |/           
// ----------------    
// Fresnel returns reflectance of a material given the angle of incidence.
// @cosIncidence is the cosine of the angle of incidence (1)
// @cosIncidence > 0 when light is coming at the surface from the outside.
// @cosIncidence < 0 when light is coming at the surface from the inside.
type Fresnel func(cosIncidence float32) spectra.Spectr

// implements Fresnel type for a dielectric.
// @n is the refractive index of the dielectric.
func NewFresnelDielectric(n float32) Fresnel {
	return func(cos1 float32) spectra.Spectr {
		f := FresnelDielectric(n, cos1)
		return spectra.NewRGBSpectr(f, f, f)
	}
}

// implements Fresnel type for a conductor.
// @n is the refractive index
func NewFresnelConductor(n, k spectra.Spectr) Fresnel {
	return func(cos1 float32) spectra.Spectr {
		return fresnelConductor(spectr1, n, k, math32.Abs(cos1))
	}
}

//        |    /       returns reflectance of a dielectric given the angle of incidence.
// n1     |1 /         
//        |/           
// ----------------    @n is the refractive index of the dielectric.
// . . . /| . . . .    @cos1 is the cosine of the angle of incidence
// n2. ./2| . . . .    @cos1 > 0 when light is coming at the surface from the outside.
// . . / .| . . . .    @cos1 < 0 when light is coming at the surface from the inside.
func FresnelDielectric(n, cos1 float32) float32 {
	n1 := n
	if cos1 < 0 {
		cos1 = -cos1
		n1 = 1/n
	} else {
	}
	sin1 := math32.Sqrt(1 - cos1*cos1)
	sin2 := sin1 / n1
	if sin2 >= 1 {
		return 1
	}
	cos2 := math32.Sqrt(1 - sin2*sin2)
	r1 := (n1*cos1 - 1*cos2)/(n1*cos1 + 1*cos2)
	r2 := (1*cos1 - n1*cos2)/(1*cos1 + n1*cos2)
	f := (r1*r1 + r2*r2)/2
	return f
}

// Get Trowbridge Reitz NDF value
// @a2 - square of the alpha parameter (roughness)
// @cosO2 - square of the cosine of the angle of incidence
// @tanO2 - square of the tangent of the angle of incidence
func TrowbridgeReitzD(a2 float32, cosO2, tanO2 float32) float32 {
	cos4 := cosO2 * cosO2
	tmp := 1 / a2
	tmp2 := 1 +  tanO2 * tmp
	D := math.Pi*a2*cos4*tmp2*tmp2
	D = 1 / D
	return D
}

// Get Trowbridge Reitz masking-shadowing function value
// @a2 - square of the alpha parameter (roughness)
// @tanO2 - square of the tangent of the angle of incidence
func TrowbridgeReitzG(a2 float32, tanIn2 float32) float32 {
	lambda := (-1 + math32.Sqrt(1 + a2*tanIn2))/2
	G := 1/(1 + lambda)
	return G
}

// importance-sample a microfacet normal direction wh from Trowbridge Reitz distribution.
// copy pasted from pbrt
// TODO understand this code
// @alpha2 - square of the alpha roughness parameter
func TrowbridgeReitzSampleWh(alpha2 float32) geo.Vec3 {
	e1, e2 := rand.Float32(), rand.Float32()
	phi := (2 * math.Pi) * e2
	tanTheta2 := alpha2 * e1 / (1.0 - e1)
	cosTheta := 1 / math32.Sqrt(1 + tanTheta2)
	sinTheta := math32.Sqrt(math32.Max(0, 1 - cosTheta * cosTheta))
	wh := geo.Vec3{
		sinTheta * math32.Cos(phi),
		sinTheta * math32.Sin(phi),
		cosTheta,
	}
	return wh
}

var spectr1 spectra.Spectr = spectra.NewRGBSpectr(1, 1, 1)

type WeighedSumMaterial struct {
	Materials []Material
	Weights   []float32
}

func NewWeighedSumMaterial(materials []Material, weights []float32) *WeighedSumMaterial {
	var sum float32
	for _, w := range weights {
		sum += w
	}
	for i := range weights {
		weights[i] /= sum
	}
	return &WeighedSumMaterial{
		Materials: materials,
		Weights: weights,
	}
}

func (m *WeighedSumMaterial) BSDF0() bool {
	return false
}
func (m *WeighedSumMaterial) BSDF(hp *ShapeHitPoint, dirIn, dirOut geo.Vec3) (L spectra.Spectr) {
	L = spectra.NewRGBSpectr(0, 0, 0)
	for i, material := range m.Materials {
		L.SpectrAdd(material.BSDF(hp, dirIn, dirOut).Mul(m.Weights[i]))
	}
	return
}

func (m *WeighedSumMaterial) PDF(hp *ShapeHitPoint, dirIn, dirOut geo.Vec3) (pdf float32) {
	//return m.Materials[0].PDF(normal, dirIn, dirOut)
	for _, material := range m.Materials {
		pdf += material.PDF(hp, dirIn, dirOut)
	}
	pdf /= float32(len(m.Materials))
	return
}

func (m *WeighedSumMaterial) BSDFSample(hp *ShapeHitPoint, dirOut geo.Vec3) (
	bsdf spectra.Spectr,
	ray geo.Ray,
	prob float32,
	specular bool,
) {
	// TODO choose according to weights?
	sampleI := rand.Intn(len(m.Materials))
	bsdf, ray, prob, specular = m.Materials[sampleI].BSDFSample(hp, dirOut)
	if prob == 0 {
		return
	}
	bsdf.Mul(m.Weights[sampleI])
	for i, material := range m.Materials {
		if i != sampleI {
			bsdf.SpectrAdd(material.BSDF(hp, ray.Direction, dirOut).Mul(m.Weights[i]))
			prob += material.PDF(hp, ray.Direction, dirOut)
		}
	}
	prob /= float32(len(m.Materials))
	return
}

func MetalPDF(alpha2 float32, hp *ShapeHitPoint, dirIn, dirOut geo.Vec3) float32 {
	if alpha2 == 0 {
		return 0
	}
	dirIn = dirIn.Normalized()
	dirOut = dirOut.Normalized()
	cosIn := dirIn.Scalar(hp.Normal)
	cosOut := dirOut.Scalar(hp.Normal)
	if (cosIn > 0) == (cosOut > 0) {
		return 0
	}
	wh := dirIn.Sub(dirOut).Normalized()
	cosH := hp.ShadingNormal.Scalar(wh)
	sinH2 := 1 - cosH*cosH
	D := TrowbridgeReitzD(alpha2, cosH*cosH, sinH2/(cosH*cosH))
	return D * math32.Abs(cosH) / (-4 * dirOut.Scalar(wh))
}

type PlasticMaterial struct {
	Ks, Kd float32
	KsKdRatio float32
	KsKdScale float32
	alpha2 float32 // square of the trowbridge reitz alpha parameter (roughness)
	n float32 // refractive index of dielectric
	matte *MatteMaterial
}

type MicrofacetMaterial struct {
	alpha2 float32 // square of the trowbridge reitz alpha parameter (roughness)
	TransmissionEnabled bool
	ReflectionEnabled bool
	TransmissionColor spectra.Spectr
	ReflectionColor spectra.Spectr
	n float32
	fresnel Fresnel
}

func NewMicrofacetMaterial(
	transmissionColor spectra.Spectr, // color filter for transmitted light
	reflectionColor   spectra.Spectr, // color filter for reflected light
	n                 float32, // refractive index. zero for metals.
	roughness         float32, // microfacet roughness
	fresnel           Fresnel,
) *MicrofacetMaterial {
	if roughness < 0.001 {
		roughness = 0
	} else {
		x := math32.Log(roughness)
		roughness = 1.62142 + 0.819955*x + 0.1734*x*x + 0.0171201*x*x*x + 0.000640711*x*x*x*x;
	}

	return &MicrofacetMaterial{
		fresnel: fresnel,
		alpha2: roughness*roughness,
		TransmissionEnabled: !transmissionColor.IsBlack(),
		ReflectionEnabled: !reflectionColor.IsBlack(),
		TransmissionColor: transmissionColor,
		ReflectionColor: reflectionColor,
		n: n,
	}
}

func NewDielectricMaterial(
	transmissionColor spectra.Spectr, // color filter for transmitted light
	reflectionColor   spectra.Spectr, // color filter for reflected light
	n                 float32, // refractive index. zero for metals.
	roughness         float32, // microfacet roughness
) *MicrofacetMaterial {
	return NewMicrofacetMaterial(
		transmissionColor,
		reflectionColor,
		n,
		roughness,
		NewFresnelDielectric(n),
	)
}

func NewMetalMaterial(
	n spectra.Spectr,
	k spectra.Spectr,
	roughness float32,
) *MicrofacetMaterial {
	return NewMicrofacetMaterial(
		spectra.NewRGBSpectr(0, 0, 0),
		spectra.NewRGBSpectr(1, 1, 1),
		0,
		roughness,
		NewFresnelConductor(n, k),
	)
}

func (m *MicrofacetMaterial) BSDF0() bool {
	return m.alpha2 == 0
}
func (m *MicrofacetMaterial) BSDF(hp *ShapeHitPoint, dirIn, dirOut geo.Vec3) (L spectra.Spectr) {
	if m.alpha2 == 0 {
		return &spectra.RGBSpectr{0, 0, 0}
	} else {
		dirIn = dirIn.Normalized()
		dirOut = dirOut.Normalized()
		cosIn := hp.Normal.Scalar(dirIn)
		cosOut := hp.Normal.Scalar(dirOut)
		transmissionCase := ((cosIn > 0) == (cosOut > 0))
		cosIn = hp.ShadingNormal.Scalar(dirIn)
		cosOut = hp.ShadingNormal.Scalar(dirOut)
		if cosIn == 0 || cosOut == 0 {
			return spectra.NewRGBSpectr(0, 0, 0)
		}
		effectiveN := m.n
		if cosOut > 0 {
			effectiveN = 1/m.n
		}

		/* find the microfacet normal 'wh' that fits dirIn and dirOut */
		var wh geo.Vec3
		var cosDirOutWh float32
		var cosDirInWh float32
		if transmissionCase {
			if (!m.TransmissionEnabled) {
				return spectra.NewRGBSpectr(0, 0, 0)
			}
			wh = dirIn.Mul(effectiveN).Sub(dirOut).Normalized()
			// we need wh such that dirOut.Scalar(wh) has the same sign as cosOut.
			// so we negate it when it is not.
			if (cosOut > 0) == (effectiveN < 1) {
				wh = wh.Negated()
			}
			cosDirOutWh = dirOut.Scalar(wh)
			cosDirInWh = dirIn.Scalar(wh)
			if (cosDirInWh > 0) != (cosDirOutWh > 0) {
				// it's impossible to find a refraction that coincides with the given
				// directions. strange that pbrt doesn't check for this
				return spectra.NewRGBSpectr(0, 0, 0)
			}
			if (cosDirOutWh > 0) != (cosOut > 0) {
				// dirOut is on different sides of the real surface and microfacet surface.
				panic("aaa")
			}
		} else {
			if (!m.ReflectionEnabled) {
				return spectra.NewRGBSpectr(0, 0, 0)
			}
			wh = dirIn.Sub(dirOut).Normalized()
			cosDirOutWh = dirOut.Scalar(wh)
			cosDirInWh = dirIn.Scalar(wh)
		}

		if cosDirOutWh == 0 {
			return spectra.NewRGBSpectr(0, 0, 0)
		}

		/* find values of D(wh) and G(wh) for the Torrance-Sparrow brdf */
		cosH2 := math32.Sqr(hp.ShadingNormal.Scalar(wh))
		tanIn2 := (1 - cosIn*cosIn)/(cosIn*cosIn)
		sinH2 := 1 - cosH2
		D := TrowbridgeReitzD(m.alpha2, cosH2, sinH2/cosH2)
		G := TrowbridgeReitzG(m.alpha2, tanIn2)

		F := m.fresnel(cosDirInWh)
		var f float32
		if transmissionCase {
			sqrtDenom := math32.Abs(cosDirOutWh) - math32.Abs(effectiveN * cosDirInWh)
			sqrtDenom = cosDirOutWh - (effectiveN) * cosDirInWh
			denom := sqrtDenom*sqrtDenom*cosOut*cosIn
			f = D*G*math32.Abs(cosDirInWh*cosDirOutWh/denom)
			F = spectra.NewRGBSpectr(1, 1, 1).SpectrSub(F)
			if cosIn < 0 {
				// light is leaving the body. apply transmission color.
				F.SpectrMul(m.TransmissionColor)
			}
		} else {
			f = math32.Abs(D*G/(4*cosIn*cosOut))
			if cosIn > 0 {
				// light is reflecting from the outer surface. apply reflection color.
				F.SpectrMul(m.ReflectionColor)
			}
		}

		F.Mul(f)
		return F
	}
}

//   normal ^   (dirIn | dirOut)     
//          |     ..         
//          |   ..           given the refracted vector @dirOut, compute the
//          | ..             original vector @dirIn. meaning of dirIn and dirOut is as described
//   1      |.               the Material interface. @cosOut is equal to dirOut.Scalar(normal).
//  --------------------     @normal is the normalized surface normal.
//   eta   .                 @eta is the refractive index under surface.
//       ..
//      ..
//     ..   (dirOut | dirIn)
func RefractAround(dirOut, normal geo.Vec3, cosOut, eta float32) (dirIn geo.Vec3, ok bool) {
	nInv := eta
	cos1 := cosOut
	if cosOut < 0 {
		cos1 = -cosOut
		nInv = 1/eta
	} else {
	}
	sin1 := math32.Sqrt(1 - cos1*cos1)
	sin2 := sin1 * nInv
	if sin2 >= 1 {
		return dirIn, false
	}
	vy := normal.Mul(cosOut)
	vx := dirOut.Sub(vy)
	cos2 := math32.Sqrt(1 - sin2*sin2)
	return vy.Normalized().Add(vx.Normalized().Mul(sin2/cos2)).Normalized(), true
}

func (m *MicrofacetMaterial) BSDFSample(hp *ShapeHitPoint, dirOut geo.Vec3) (bsdf spectra.Spectr, ray geo.Ray, prob float32, specular bool) {
	//     \  |  /         
	// n1   \ |1/       1 - angle between normal and ray corresponding to dirOut, 0..90
	//       \|/        2 - angle between normal and ray on the other side, 0..90
	// ----------------    
	//       /|
	// n2   /2|         n1 - refractive index on dirOut side
	//     /  |         n2 - refractive index on the other side
	dirOut = dirOut.Normalized()
	cosOut := dirOut.Scalar(hp.Normal)

	var wh geo.Vec3
	if m.alpha2 == 0 {
		wh = hp.ShadingNormal
	} else {
		wh = TrowbridgeReitzSampleWh(m.alpha2)
		bx, by := BasisAroundVector(hp.ShadingNormal)
		wh = VectorFromBasis(bx, by, hp.ShadingNormal, wh.X, wh.Y, wh.Z)
	}

	cosDirOutWh := dirOut.Scalar(wh)
	if cosDirOutWh == 0 {
		return
	}

	var dirIn geo.Vec3
	if (cosDirOutWh > 0) != (cosOut > 0) {
		// dirOut is on different sides of the real surface and microfacet surface.
		// must be because the shading normal is different from the real one.
		wh = wh.Negated()
		cosDirOutWh = -cosDirOutWh
	}
	var refSamplingProb float32
	if (!m.TransmissionEnabled) {
		refSamplingProb = 1
	} else if (!m.ReflectionEnabled) {
		refSamplingProb = 0
	} else {
		// prob equal to reflectance is a good prob to sample reflection
		refSamplingProb = FresnelDielectric(m.n, -cosDirOutWh)
	}
	reflectionCase := (rand.Float32() < refSamplingProb)
	if reflectionCase {
		// reflecion sampling case
		dirIn = dirOut.ReflectAround(wh, cosDirOutWh)
		if (dirIn.Scalar(hp.Normal) > 0) == (cosOut > 0) {
			// whoops, we reflected to the other side of the surface
			// e.g. "light leak error"
			return
		}
	} else {
		var ok bool
		dirIn, ok = RefractAround(dirOut, wh, cosDirOutWh, m.n)
		if !ok {
			return
		} else {
		}
		if (dirIn.Scalar(hp.Normal) > 0) != (cosOut > 0) {
			// whoops, we refracted to the same side of the surface
			// e.g. "dark spot error"
			return
		}
	}

	ray = geo.Ray{hp.Point, dirIn}

	if m.alpha2 == 0 {
		F := m.fresnel(dirIn.Scalar(wh))
		cosIn := dirIn.Scalar(hp.ShadingNormal)
		if reflectionCase {
			bsdf = F.Mul(1/math32.Abs(cosIn))
			prob = refSamplingProb
			if cosOut < 0 {
				// light is reflecting from the outer surface. apply reflection color.
				bsdf.SpectrMul(m.ReflectionColor)
			}
		} else {
			n := m.n
			if cosOut < 0 {
				n = 1/m.n
			}
			mul := n * n / math32.Abs(cosIn)
			bsdf = spectra.NewRGBSpectr(1, 1, 1).SpectrSub(F).Mul(mul)
			prob = 1 - refSamplingProb
			if cosOut < 0 {
				// light is leaving the body. apply transmission color.
				bsdf.SpectrMul(m.TransmissionColor)
			}
		}
		specular = true
		return
	}
	// possible optimization: BSDF() and PDF() do some redunant work
	bsdf = m.BSDF(hp, ray.Direction, dirOut)
	prob = m.PDF(hp, dirIn, dirOut)
	return
}

func (m *MicrofacetMaterial) PDF(hp *ShapeHitPoint, dirIn, dirOut geo.Vec3) float32 {
	if m.alpha2 == 0 {
		return 0
	}
	dirIn = dirIn.Normalized()
	dirOut = dirOut.Normalized()
	cosIn := dirIn.Scalar(hp.Normal)
	cosOut := dirOut.Scalar(hp.Normal)
	transmissionCase := ((cosIn > 0) == (cosOut > 0))
	effectiveN := m.n

	var wh geo.Vec3
	if transmissionCase {
		if !m.TransmissionEnabled {
			return 0
		}
		if cosOut > 0 {
			effectiveN = 1/m.n
		}
		wh = dirIn.Mul(effectiveN).Sub(dirOut).Normalized()
	} else {
		if !m.ReflectionEnabled {
			return 0
		}
		wh = dirIn.Sub(dirOut).Normalized()
	}
	cosDirOutWh := dirOut.Scalar(wh)
	if (cosDirOutWh > 0) != (cosOut > 0) {
		// dirOut is on different sides of the real surface and microfacet surface.
		wh = wh.Negated()
		cosDirOutWh = -cosDirOutWh
	} else {
		// TODO is this branch possible?
		//fmt.Println("NNEG", cosDirOutWh)
	}
	if cosDirOutWh == 0 {
		return 0
	}

	cosDirInWh := dirIn.Scalar(wh)

	var refSamplingProb float32
	if (!m.TransmissionEnabled) {
		refSamplingProb = 1
	} else if (!m.ReflectionEnabled) {
		refSamplingProb = 0
	} else {
		// prob equal to reflectance is a good prob to sample reflection
		refSamplingProb = FresnelDielectric(m.n, cosDirInWh)
	}

	cosH := hp.ShadingNormal.Scalar(wh)
	sinH2 := 1 - cosH*cosH
	D := TrowbridgeReitzD(m.alpha2, cosH*cosH, sinH2/(cosH*cosH))

	if transmissionCase {
		if (wh.Scalar(dirIn) > 0) != (cosDirOutWh > 0) {
			// refracted to the same side of the surface - impossible.
			// strange that pbrt doesn't check for this
			return 0
		}
		sqrtDenom := math32.Abs(cosDirOutWh) - math32.Abs(effectiveN * cosDirInWh)
		sqrtDenom = cosDirOutWh - effectiveN * cosDirInWh
		prob := D*math32.Abs((effectiveN * effectiveN * cosDirInWh) / (sqrtDenom * sqrtDenom))
		return prob * (1 - refSamplingProb)
	} else {
		// cosine may be positive because dirOut is on the "inner" side of the surface or
		// because the shading normal different from the real normal caused an
		// unfortunate microfacet to be sampled. in either case we just take the Abs(),
		// as pbrt seems to do.
		prob := D * math32.Abs(cosH) / math32.Abs(-4 * dirOut.Scalar(wh))
		return prob * refSamplingProb
	}
}

type LayeredMaterial struct {
	n float32
	Base Material
}

func NewLayeredMaterial(base Material, n float32) *LayeredMaterial {
	return &LayeredMaterial{
		n: n,
		//Base: New1ColorMatteMaterial(0, 1, 0, 0),
		Base: base,
	}
}

func (m *LayeredMaterial) BSDFSample(hp *ShapeHitPoint, dirOut geo.Vec3) (bsdf spectra.Spectr, ray geo.Ray, prob float32, specular bool) {
	dirOut = dirOut.Normalized()
	cosOut := hp.Normal.Scalar(dirOut)
	normal := hp.Normal
	if cosOut > 0 {
		cosOut = -cosOut
		normal = normal.Negated()
	}

	F := FresnelDielectric(m.n, -cosOut)

	if rand.Float32() < F {
		cosOutShading := hp.ShadingNormal.Scalar(dirOut)
		dirIn := dirOut.ReflectAround(hp.ShadingNormal, cosOutShading)
		if (dirIn.Scalar(hp.Normal) > 0) == (cosOut > 0) {
			// whoops, we reflected to the other side of the surface
			// e.g. "light leak error"
			return
		}
		ray = geo.Ray{hp.Point, dirIn}
		prob = F
		specular = true

		b := F/math32.Abs(dirIn.Scalar(hp.ShadingNormal))
		bsdf = spectra.NewRGBSpectr(b, b, b)
	} else {
		hemi := sampling.CosineSampleHemisphere().Normalized()
		prob = (1 - F)*hemi.Z/(math.Pi)
		bx, by := BasisAroundVector(normal)
		ray = geo.Ray{
			hp.Point,
			VectorFromBasis(bx, by, normal, hemi.X, hemi.Y, hemi.Z),
		}
		bsdf = m.BSDF(hp, ray.Direction, dirOut)
	}
	return
}

func (m *LayeredMaterial) BSDF0() bool {
	return false
}
func (m *LayeredMaterial) BSDF(hp *ShapeHitPoint, dirIn, dirOut geo.Vec3) (L spectra.Spectr) {
	dirOut = dirOut.Normalized()
	dirIn = dirIn.Normalized()
	cosOut := hp.Normal.Scalar(dirOut)
	cosIn := hp.Normal.Scalar(dirIn)
	if (cosOut > 0) == (cosIn > 0) {
		return spectra.NewRGBSpectr(0, 0, 0)
	}
	normal := hp.ShadingNormal
	if cosOut > 0 {
		cosOut = -cosOut
		cosIn = -cosIn
		normal = normal.Negated()
	}
	if cosOut == 0 {
		return
	}

	tIn := 1 - FresnelDielectric(m.n, cosIn)
	tOut := 1 - FresnelDielectric(m.n, -cosOut)
	dirOutT, ok := RefractAround(dirOut, normal, cosOut, m.n)
	if !ok {
		if cosOut < -0.0001 {
			panic("aaa")
		}
		return
	}
	dirInT, ok := RefractAround(dirIn, normal, cosIn, 1/m.n)
	if !ok {
		if cosIn > 0.0001 {
			panic("aaa")
		}
		return
	}
	
	bsdf := m.Base.BSDF(hp, dirInT, dirOutT)
	bsdf = bsdf.Mul(tIn*tOut)
	return bsdf
}

func (m *LayeredMaterial) PDF(hp *ShapeHitPoint, dirIn, dirOut geo.Vec3) float32 {
	dirOut = dirOut.Normalized()
	dirIn = dirIn.Normalized()
	cosOut := hp.Normal.Scalar(dirOut)
	cosIn := hp.Normal.Scalar(dirIn)
	if (cosOut > 0) == (cosIn > 0) {
		return 0
	}

	return math32.Abs(cosIn) / math.Pi
}

type BlendMapMaterial struct {
	Black Material
	White Material
	Map   img.Image3
}

func NewBlendMapMaterial(black, white Material, themap img.Image3) *BlendMapMaterial {
	return &BlendMapMaterial{
		Black: black,
		White: white,
		Map: themap,
	}
}

func (m *BlendMapMaterial) BSDF0() bool {
	return false
}
func (m *BlendMapMaterial) BSDF(hp *ShapeHitPoint, dirIn, dirOut geo.Vec3) (L spectra.Spectr) {
	ratio, _, _ := m.Map.AtUv(hp.U, hp.V)
	L = m.Black.BSDF(hp, dirIn, dirOut).Mul(1 - ratio)
	L.SpectrAdd(m.White.BSDF(hp, dirIn, dirOut).Mul(ratio))
	return
}

func (m *BlendMapMaterial) PDF(hp *ShapeHitPoint, dirIn, dirOut geo.Vec3) (pdf float32) {
	//return m.Materials[0].PDF(normal, dirIn, dirOut)
	ratio, _, _ := m.Map.AtUv(hp.U, hp.V)
	pdf += m.Black.PDF(hp, dirIn, dirOut) * (1 - ratio)
	pdf += m.White.PDF(hp, dirIn, dirOut) * ratio
	return
}

func (m *BlendMapMaterial) BSDFSample(hp *ShapeHitPoint, dirOut geo.Vec3) (
	bsdf spectra.Spectr,
	ray geo.Ray,
	prob float32,
	specular bool,
) {
	ratio, _, _ := m.Map.AtUv(hp.U, hp.V)
	if rand.Float32() < ratio {
		bsdf, ray, prob, specular = m.White.BSDFSample(hp, dirOut)
		if prob == 0 {
			return
		}
		bsdf.Mul(ratio)
		bsdf.SpectrAdd(m.Black.BSDF(hp, ray.Direction, dirOut).Mul(1 - ratio))
		prob += m.Black.PDF(hp, ray.Direction, dirOut) * (1 - ratio)
	} else {
		bsdf, ray, prob, specular = m.Black.BSDFSample(hp, dirOut)
		if prob == 0 {
			return
		}
		bsdf.Mul(1 - ratio)
		bsdf.SpectrAdd(m.White.BSDF(hp, ray.Direction, dirOut).Mul(ratio))
		prob += m.White.PDF(hp, ray.Direction, dirOut) * ratio
	}
	return
}

func main() {
	_ = math32.Cos
	debug.Noop()
	fmt.Println("vim-go")
}
