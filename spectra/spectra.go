package spectra

import (
	"fmt"
	"math"
	"ly/util/math32"
	"ly/colors"
)

type Spectr interface {
	Clone() Spectr
	XYZ() (x, y, z float32)
	RGB() (r, g, b float32)
	BSDF(Spectr)
	Mul(float32) Spectr
	SpectrMul(Spectr) Spectr
	SpectrDiv(Spectr) Spectr
	SpectrSub(Spectr) Spectr
	SpectrAdd(Spectr) Spectr
	Add(float32) Spectr
	Sqrt() Spectr
	IsBlack() bool
	Power() float32
}

type RGBSpectr struct {
	R, G, B float32
}

func NewRGBSpectr(r, g, b float32) *RGBSpectr {
	return &RGBSpectr{
		R: r, G: g, B: b,
	}
}

func (s *RGBSpectr) Power() float32 {
	// todo this is not true power
	return s.R + s.G + s.B
}

func (s *RGBSpectr) Clone() Spectr {
	ret := *s
	return &ret
}

func (s *RGBSpectr) IsBlack() bool {
	return s.R == 0 && s.G == 0 && s.B == 0
}

func (s *RGBSpectr) XYZ() (x, y, z float32) {
	return colors.Rgb2xyz(s.R, s.G, s.B)
}

func (s *RGBSpectr) RGB() (r, g, b float32) {
	return s.R, s.G, s.B
}

func (s *RGBSpectr) Sqrt() Spectr {
	s.R = math32.Sqrt(s.R)
	s.G = math32.Sqrt(s.G)
	s.B = math32.Sqrt(s.B)
	return s
}

func (s *RGBSpectr) Add(x float32) Spectr {
	s.R += x
	s.G += x
	s.B += x
	return s
}

func (s *RGBSpectr) SpectrSub(ss Spectr) Spectr {
	s2 := ss.(*RGBSpectr)
	s.R -= s2.R
	s.G -= s2.G
	s.B -= s2.B
	return s
}

func (s *RGBSpectr) SpectrMul(ss Spectr) Spectr {
	s2 := ss.(*RGBSpectr)
	s.R *= s2.R
	s.G *= s2.G
	s.B *= s2.B
	return s
}

func (s *RGBSpectr) SpectrDiv(ss Spectr) Spectr {
	s2 := ss.(*RGBSpectr)
	s.R /= s2.R
	s.G /= s2.G
	s.B /= s2.B
	return s
}

func (s *RGBSpectr) SpectrAdd(ss Spectr) Spectr {
	s2 := ss.(*RGBSpectr)
	if false {
		l, a, b := colors.Rgb2lab(s.R, s.G, s.B)
		ll, aa, bb := colors.Rgb2lab(s2.R, s2.G, s2.B)
		s.R, s.G, s.B = colors.Lab2rgb(l + ll, a + aa, b + bb)
	} else {
		s.R += s2.R
		s.G += s2.G
		s.B += s2.B
	}
	return s
}

func (s *RGBSpectr) BSDF(ss Spectr) {
	s2 := ss.(*RGBSpectr)
	s.R *= s2.R
	s.G *= s2.G
	s.B *= s2.B
	if (s2.R < 0 || s2.G < 0 || s2.B < 0) {
		fmt.Println(s2)
		panic("assertion error")
	}
}

const uselab bool = false

func (s *RGBSpectr) Mul(k float32) Spectr {
	if uselab {
		l, a, b := colors.Rgb2lab(s.R, s.G, s.B)
		l *= k
		s.R, s.G, s.B = colors.Lab2rgb(l, a, b)
	} else {
		s.R *= k
		s.G *= k
		s.B *= k
	}
	return s
}
/*

type GraySpectr struct {
	power float32 // power in watts inside visible wavelength range
}

func (s *GraySpectr) Clone() (ret Spectr) {
	return &GraySpectr{s.power}
}

func (s *GraySpectr) BSDF(s2 Spectr) {
	s.power *= s2.(*GraySpectr).power
}

func (s *GraySpectr) XYZ() (x, y, z float32) {
	x = 1e0*s.power/float32(len(XYZf.Values)/3)
	y = x
	z = y
	return
}

func (s *GraySpectr) Add(s2 Spectr) {
	g := s2.(*GraySpectr)
	s.power += g.power
}

func (s *GraySpectr) Mul(k float32) {
	s.power *= k
}
*/

type Spectr1 struct {
	Samples []float32
	FirstWave int
	Step float32
}
/*

func GrayBulb(watt float32) Spectr {
	return &GraySpectr{
		power: watt/10/(4*3.1415),
	}
}

func (s Spectr1) Add(s2 Spectr1) {
	for i, power := range s.Samples {
		s.Samples[i] = power + s2.Samples[i]
	}
}

func (s Spectr1) Mul(k float32) {
	for i, power := range s.Samples {
		s.Samples[i] = power*k
	}
}

func (s GraySpectr) RGB() (r, g, b float32) {
	x, y, z := s.XYZ()
	r = 3.2406*x - 1.5372*y - 0.4986*z
	g = -0.9689*x + 1.8758*y + 0.0415*z
	b = 0.0557*x - 0.2040*y + 1.0570*z
	gam := func(u float32) float32 {
		if u < 0.0031308 {
			return 12.92*u
		} else {
			return 1.055*float32(math.Pow(float64(u), 1/2.4)) - 0.055
		}
	}
	return gam(r), gam(g), gam(b)
}

func (s Spectr1) fillGray(val float32) {
	for i, _ := range s.Samples {
		s.Samples[i] = val
	}
}
*/

type SpectrSample struct {
	Wavelength float32 // in nanometers
	Power      float32 // in watts
}

const (
	WaveStep = 1
	FirstWave = 390
	NSamples = 441
)

/*
type SampleSpectr struct {
	Samples []float32 // in watts
}

func (s SampleSpectr) Peak() (ret float32) {
	for _, power := range s.Samples {
		if power > ret {
			ret = power
		}
	}
	return
}

func (s *SampleSpectr) BSDF(spectr Spectr) {
	s2 := spectr.(*SampleSpectr)
	for i, _ := range s.Samples {
		if s2.Samples[i] > 1 {
			println("bsdf sample out of range", s2.Samples[i])
		}
		s.Samples[i] *= s2.Samples[i]
	}
}

func (s SampleSpectr) Clone() (ret Spectr) {
	s2 := &SampleSpectr{make([]float32, len(s.Samples))}
	for i, power := range s.Samples {
		s2.Samples[i] = power
	}
	return s2
}

func (s SampleSpectr) RGB() (r, g, b float32) {
	x, y, z := s.XYZ()
	r = 3.2406*x - 1.5372*y - 0.4986*z
	g = -0.9689*x + 1.8758*y + 0.0415*z
	b = 0.0557*x - 0.2040*y + 1.0570*z
	gam := func(u float32) float32 {
		if u < 0.0031308 {
			return 12.92*u
		} else {
			return 1.055*float32(math.Pow(float64(u), 1/2.4)) - 0.055
		}
	}
	return gam(r), gam(g), gam(b)
}

func (s SampleSpectr) XYZ() (x, y, z float32) {
	if float32(FirstWave) != XYZf.FirstWave {
		panic("XYZ table wavelength mismatch")
	}
	if WaveStep != 1 {
		panic("XYZ table wave step mismatch")
	}
	nm := 0
	for _, power := range s.Samples {
		x += XYZf.Values[nm]*power
		y += XYZf.Values[nm + 1]*power
		z += XYZf.Values[nm + 2]*power
		nm += 3
	}
	x /= XYZf.Integral
	y /= XYZf.Integral
	z /= XYZf.Integral
	return
}

func (s *SampleSpectr) Mul(k float32) {
	for i, _ := range s.Samples {
		s.Samples[i] *= k
	}
}

func (s *SampleSpectr) Add(arg Spectr) {
	s2 := arg.(*SampleSpectr)
	for i, _ := range s.Samples {
		s.Samples[i] += s2.Samples[i]
	}
}

func plank(l, T float32) float32 {
	h := float32(6.626e-34)
	c := float32(299792458)
	k := float32(1.380e-23)
	return 2*h*c*c/(l*l*l*l*l)/(math32.Exp(h*c/l/k/T) - 1)
}

// 25W for now
func Bulb(T float32) Spectr {
	T = 2700
	area := float32(6e-06)
	s := SampleSpectr{
		Samples: make([]float32, NSamples),
	}
	
	for i := 0; i < NSamples; i++ {
		wavelength := float32(FirstWave + i*WaveStep);
		s.Samples[i] = plank(wavelength*1e-9, T)*3.1415*area
	}
	return &s
}

func Black() Spectr {
	s := SampleSpectr{
		Samples: make([]float32, NSamples),
	}
	return &s
}

type SpectrBuilder struct {
	Spectr *SampleSpectr
	next int
	lastPower float32
}

func NewBuilder() SpectrBuilder {
	spectr := SampleSpectr{
		Samples: make([]float32, NSamples),
	}
	return SpectrBuilder{
		Spectr: &spectr,
		next: 0,
	}
}

// does not work when WaveStep is too large
func (b *SpectrBuilder) AddSample(wave int, power float32) {
	var i int = b.next
	for ; i <= wave - FirstWave; i += WaveStep {
		k := float32(i - b.next + 1)/float32(wave - FirstWave - b.next + 1)
		val := math32.Lerp(b.lastPower, power, k)
		b.Spectr.Samples[i] = val
	}
	b.next = i
	b.lastPower = power
}
*/

// this table describes how power depends on wavelength
// Units are Watts and Nanometers
// power outside Wavelength range is considered 0
type SpectrTable struct {
	Power []float32
	Wavelength []float32
}

func NewSpectrTable() SpectrTable {
	return SpectrTable{
		Power: []float32{},
		Wavelength: []float32{},
	}
}

func (s *SpectrTable) AppendSample(wavelen float32, power float32) {
	s.Power = append(s.Power, power)
	s.Wavelength = append(s.Wavelength, wavelen)
}

func (s SpectrTable) At(wavelen float32) {
	
}

func (s SpectrTable) GetXYZ() (X, Y, Z float32) {
	i := 0
	XYZi := 0
	for wave_i := XYZf.FirstWave; wave_i <= XYZf.LastWave; wave_i++ {
		wave := float32(wave_i)
		for i < len(s.Wavelength) && s.Wavelength[i] < wave {
			i++
		}
		var power float32 = 0
		if i > 0 && i < len(s.Wavelength) {
			lerp := (float32(wave)- s.Wavelength[i - 1]) /
				(s.Wavelength[i] - s.Wavelength[i - 1])
			power = math32.Lerp(s.Power[i - 1], s.Power[i], lerp)
		}
		X += XYZf.Values[XYZi]*power
		Y += XYZf.Values[XYZi + 1]*power
		Z += XYZf.Values[XYZi + 2]*power
		XYZi += 3
	}
	// we should multiply by (lastWave - firstWave) / N, but N == (lastWave - firstWave)
	return
}

func (s SpectrTable) MakeRGBSpectr() *RGBSpectr {
	X, Y, Z := s.GetXYZ()
	return NewRGBSpectr(colors.Xyz2rgb(X, Y, Z))
}

func Noop() {}

func init() {
	if false { fmt.Println(); math.Sqrt(1) }
}
