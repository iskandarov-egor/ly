package colors

import (
	"fmt"
	"math"
	"ly/util/math32"
)

const (
	RGBSpace = iota
	XYZSpace = iota
	LABSpace = iota
	SRGBSpace = iota
)

type ColorSpace int

type RGBColor struct {
	R, G, B float32
}

func NewRGB(r, g, b float32) RGBColor {
	return RGBColor{r, g, b}
}

var distinctColors []RGBColor

func init() {
	distinctColors = []RGBColor{
		RGBColor{0, 0, 1},
		RGBColor{0, 1, 0},
		RGBColor{1, 0, 0},
		RGBColor{0, 1, 1},
		RGBColor{1, 1, 0},
		RGBColor{1, 0, 1},

		RGBColor{0,   0.5, 1   },
		RGBColor{0,   1,   0.5 },
		RGBColor{1,   0.5, 0   },
		RGBColor{0.5, 1,   0   },
		RGBColor{1,   0,   0.5},
		RGBColor{0.5, 0,   1},

		RGBColor{0.5, 0.5, 1   },
		RGBColor{0.5, 1,  0.5 },
		RGBColor{1,   0.5, 0.5 },
		RGBColor{0.5, 1,   0.5 },
		RGBColor{1,   0.5, 0.5},
		RGBColor{0.5, 0.5, 1},
	}
}

func Distinct(idx int) RGBColor {
	return distinctColors[idx]
}

func Lab2xyz(l, a, b float32) (x, y, z float32) {
	var_Y := (l + 0.16) / 1.16
	var_X := a / 5 + var_Y
	var_Z := var_Y - b / 2

	if math32.Pow(var_Y, 3) > 0.008856 {
		var_Y = math32.Pow(var_Y, 3)
	} else {
		var_Y = (var_Y - float32(16) / 116) / 7.787
	}
	if math32.Pow(var_X, 3) > 0.008856 {
		var_X = math32.Pow(var_X, 3)
	} else {
		var_X = (var_X - float32(16) / 116) / 7.787
	}
	if math32.Pow(var_Z, 3) > 0.008856 {
		var_Z = math32.Pow(var_Z, 3)
	} else {
		var_Z = (var_Z - float32(16) / 116) / 7.787
	}

	x = 0.95047 * var_X     // ref_X = 0.95047     Observer= 2°, Illuminant= D65
	y = 1 * var_Y     // ref_Y = 1.00000
	z = 1.08883 * var_Z     // ref_Z = 1.08883
	return
}

func Xyz2lab(x, y, z float32) (l, a, b float32) {
	var_X := x / 0.95          // ref_X = 0.95047   Observer= 2°, Illuminant= D65
	var_Y := y / 1          // ref_Y = 1.000
	var_Z := z / 1.0888          // ref_Z = 1.08883

	if var_X > 0.008856 {
		var_X = math32.Pow(var_X, 1.0/3)
	} else {
		var_X = ( 7.787 * var_X ) + ( float32(16) / 116 )
	}
	if var_Y > 0.008856 {
		var_Y = math32.Pow(var_Y, 1.0/3)
	} else {
		var_Y = ( 7.787 * var_Y ) + ( float32(16) / 116 )
	}
	if var_Z > 0.008856 {
		var_Z = math32.Pow(var_Z, 1.0/3)
	} else {
		var_Z = ( 7.787 * var_Z ) + ( float32(16) / 116 )
	}

	l = ( 1.16 * var_Y ) - 0.16
	a = 5 * ( var_X - var_Y )
	b = 2 * ( var_Y - var_Z )
	return
}

func Xyz2srgb(x, y, z float32) (r, g, b float32) {
	r = 3.2406*x - 1.5372*y - 0.4986*z
	g = -0.9689*x + 1.8758*y + 0.0415*z
	b = 0.0557*x - 0.2040*y + 1.0570*z
	gam := func(u float32) float32 {
		var ret float32
		if u < 0.0031308 {
			ret = 12.92*u
		} else {
			ret = 1.055*float32(math.Pow(float64(u), 1/2.4)) - 0.055
		}
		if ret >= 0 {
			if ret > 1 {
				return 1
			}
			return ret
		} else {
			return 0
		}
	}
	return gam(r), gam(g), gam(b)
}

func Xyz2rgb(x, y, z float32) (r, g, b float32) {
	r = 3.2406*x - 1.5372*y - 0.4986*z
	g = -0.9689*x + 1.8758*y + 0.0415*z
	b = 0.0557*x - 0.2040*y + 1.0570*z
	return
}

func Srgb2xyz(r, g, b float32) (x, y, z float32) {
	gam := func(u float32) (ret float32) {
		if u < 0.04045 {
			ret = u/12.92
		} else {
			ret = float32(math.Pow(float64(u + 0.055)/1.055, 2.4))
		}
		return
	}
	r, g, b = gam(r), gam(g), gam(b)
	x = r*0.4124564 + g*0.3575761 + b*0.1804375
	y = r*0.2126729 + g*0.7151522 + b*0.0721750
	z = r*0.0193339 + g*0.1191920 + b*0.9503041
	return
}

func Rgb2xyz(r, g, b float32) (x, y, z float32) {
	x = r*0.4124564 + g*0.3575761 + b*0.1804375
	y = r*0.2126729 + g*0.7151522 + b*0.0721750
	z = r*0.0193339 + g*0.1191920 + b*0.9503041
	return
}

func Rgb2srgb(r, g, b float32) (x, y, z float32) {
	return SrgbGamma(r),  SrgbGamma(g),  SrgbGamma(b)
}

func SrgbGamma(u float32) float32 {
	var ret float32
	if u < 0.0031308 {
		ret = 12.92*u
	} else {
		ret = 1.055*float32(math.Pow(float64(u), 1/2.4)) - 0.055
	}
	if ret >= 0 {
		if ret > 1 {
			return 1
		}
		return ret
	} else {
		return 0
	}
}

func SrgbInvGamma(u float32) float32 {
	var ret float32
	if u < 0.04045 {
		ret = u/12.92
	} else {
		ret = float32(math.Pow(float64(u + 0.055)/1.055, 2.4))
	}
	if ret >= 0 {
		if ret > 1 {
			return 1
		}
		return ret
	} else {
		return 0
	}
}

func Lab2rgb(l, a, b float32) (r, g, bb float32) {
	return Xyz2rgb(Lab2xyz(l, a, b))
}

func Rgb2lab(r, g, b float32) (l, a, bb float32) {
	return Xyz2lab(Rgb2xyz(r, g, b))
}

type ConvertFunc3 func(c1, c2, c3 float32) (r1, r2, r3 float32)

func Noop3(c1, c2, c3 float32) (r1, r2, r3 float32) {
	return c1, c2, c3
}

func GetConvertFunc(colorSpace1, colorSpace2 int) ConvertFunc3 {
	switch colorSpace1 {
		case RGBSpace:
			switch colorSpace2 {
				case RGBSpace:
					return Noop3
				case XYZSpace:
					return Rgb2xyz
				case LABSpace:
					return Rgb2lab
				case SRGBSpace:
					return Rgb2srgb
			}
		case SRGBSpace:
			switch colorSpace2 {
				case SRGBSpace:
					return Noop3
			}
		case XYZSpace:
			switch colorSpace2 {
				case RGBSpace:
					return Rgb2xyz
				case XYZSpace:
					return Noop3
				case LABSpace:
					return Xyz2lab
				case SRGBSpace:
					return Xyz2srgb
			}
		case LABSpace:
			switch colorSpace2 {
				case RGBSpace:
					return Lab2rgb
				case XYZSpace:
					return Lab2xyz
				case LABSpace:
					return Noop3
			}
	}
	panic(fmt.Sprintf("no conversion function for color spaces %d -> %d", colorSpace1, colorSpace2))
}

//func Convert3(colorSpace1, colorSpace2 int, c1, c2, c3 float32) (r1, r2, r3 float32) {
//}

func Noop() {
	_ = fmt.Println
}
