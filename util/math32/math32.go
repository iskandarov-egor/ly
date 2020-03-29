package math32

import (
	"math"
	"fmt"
)

func Max(x, y float32) float32 {
	if x > y {
		return x
	} else {
		return y
	}
}

func Min(x, y float32) float32 {
	if x < y {
		return x
	} else {
		return y
	}
}

func Max3(x, y, z float32) float32 {
	if x > y {
		if x > z {
			return x
		} else {
			return z
		}
	} else {
		if y > z {
			return y
		} else {
			return z
		}
	}
}

func Abs(x float32) float32 {
	if x < 0 {
		return -x
	} else {
		return x
	}
}

func Sin(x float32) float32 {
	return float32(math.Sin(float64(x)))
}

func Cos(x float32) float32 {
	return float32(math.Cos(float64(x)))
}

func Atan2(x, y float32) float32 {
	return float32(math.Atan2(float64(x), float64(y)))
}

func Log(x float32) float32 {
	return float32(math.Log(float64(x)))
}

func Cotan(x float32) float32 {
	return 1/float32(math.Tan(float64(x)))
}

func Acos(x float32) float32 {
	return float32(math.Acos(float64(x)))
}

func Square(x float32) float32 {
	return x*x
}

func Pow(x, y float32) float32 {
	return float32(math.Pow(float64(x), float64(y)))
}

func Exp(x float32) float32 {
	return float32(math.Exp(float64(x)))
}

func Lerp(x, y, k float32) float32 {
	return x + (y - x)*k
}

func Sqr(x float32) float32 {
	return x*x
}

func Sqrt(x float32) float32 {
	return float32(math.Sqrt(float64(x)))
}

func SafeSqrt(x float32) float32 {
	if x <= 0 {
		if x < -0.01 {
			panic(fmt.Sprintf("sqrt() of %v", x))
		}
		return 0
	} else {
		return Sqrt(x)
	}
}

//  len(coef)
//  ____
//  \
//   \
//   /    coef[k] * cos(kx)
//	/___
//  k = 0
func Fourier(coef []float32, cosx float32) (ret float32) {
	var value float64 // more accuracy with 64bits
	cosX := float64(cosx)
    cosKMinusOneX := cosX
	var cosKX float64 = 1.0
	for k := 0; k < len(coef); k++ {
		value += float64(coef[k]) * cosKX
		cosKPlusOneX := 2 * cosX * cosKX - cosKMinusOneX
		cosKMinusOneX = cosKX
		cosKX = cosKPlusOneX
	}
	return float32(value)
}

func Clamp(x, low, hi float32) float32 {
	if x < low {
		return low
	}
	if x > hi {
		return hi
	}
	return x
}
