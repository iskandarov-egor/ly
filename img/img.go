package img

import (
	"fmt"
	"os"
	"math"
	"sort"
	"strings"
	"image"
	"image/png"
	"ly/util/math32"
	"ly/colors"
	"ly/debug"
)

// 3 floats per pixel
type Image3 struct {
	W    int
	H    int
	ColorSpace int // colorspace id from colors module
	Data []float32
}

// 1 float per pixel
type Image1 struct {
	W    int
	H    int
	Data []float32
}

func NewImage3(w int, h int, colorSpace int) Image3 {
	return Image3{
		W:    w,
		H:    h,
		Data: make([]float32, 3*w*h),
		ColorSpace: colorSpace,
	}
}

func (im *Image3) SetGray(x, y int, val float32) {
	pos := 3*(y*im.W + x)
	im.Data[pos] = val
	im.Data[pos + 1] = val
	im.Data[pos + 2] = val
}

func (im Image3) GetImage1() Image1 {
	im1 := Image1{
		W: im.W,
		H: im.H,
		Data: make([]float32, im.W*im.H),
	}
	if im.ColorSpace != colors.XYZSpace {
		panic("aaa space")
	}

	j := 1
	for i := 0; i < im.W*im.H; i++ {
		im1.Data[i] = im.Data[j]
		j += 3
	}
	return im1
}

func (im *Image3) Log() {
	for i, val := range im.Data {
		im.Data[i] = math32.Log(val)
	}
	return
}

func (im *Image3) Normalize() (max float32) {
	for _, val := range im.Data {
		if max < val {
			max = val
		}
	}
	if max == 0 {
		return
	}
	for i, val := range im.Data {
		im.Data[i] = val/max
	}
	return
}

func (im *Image3) Mul(k float32) {
	for i, val := range im.Data {
		im.Data[i] = val*k
	}
}

func (im *Image3) Set(x, y int, X, Y, Z float32) {
	pos := 3*(y*im.W + x)
	im.Data[pos] = X
	im.Data[pos + 1] = Y
	im.Data[pos + 2] = Z
}

func (im *Image3) Stdout(title string) {
	header := fmt.Sprintf(" %s (%dx%d) ", title, im.W, im.H)
	w := 45
	wingsLen := w - len(header)
	if wingsLen > 0 {
		if wingsLen % 2 != 0 {
			header = header + "="
			wingsLen--
		}
		wing := strings.Repeat("=", wingsLen / 2)
		header = wing + header + wing
	}
	fmt.Println()
	fmt.Println(header)
	scale := float32(im.W)/float32(w)
	for y := 0; y < int(float32(w)/float32(im.W)*float32(im.H)/2); y++ {
		line := 3*int(float32(y)*scale*2)*im.W
		for x := 0; x < w; x++ {
			pos := line + 3*int(float32(x)*scale)
			val := im.Data[pos]
			if val > 0.5 {
				fmt.Print("X")
			} else {
				fmt.Print(" ")
			}
		}
		fmt.Println()
	}
	fmt.Println(strings.Repeat("=", w))
}

func topix(x float32) uint8 {
	if x > 1 {
		return 255
	}
	return uint8(255*x)
}

func LoadPng(path string) Image3 {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	decoded, err := png.Decode(file)
	if err != nil {
		panic(fmt.Sprintf("decode png: %v", err))
	}
	ii := 0

	rgba, ok := decoded.(*image.NRGBA)
	if ok {
		w := rgba.Stride / 4
		h := len(rgba.Pix) / w / 4
		img := NewImage3(w, h, colors.RGBSpace)
		
		for i := 0; i < len(rgba.Pix); i += 4 {
			r, g, b, a := rgba.Pix[i], rgba.Pix[i + 1], rgba.Pix[i + 2], rgba.Pix[i + 3]
			a = 255
			img.Data[ii] = float32(r)/float32(a)
			img.Data[ii + 1] = float32(g)/float32(a)
			img.Data[ii + 2] = float32(b)/float32(a)
			ii += 3
		}
		return img
	} else {
		rgba := decoded.(*image.RGBA)
		w := rgba.Stride / 4
		h := len(rgba.Pix) / w / 4
		img := NewImage3(w, h, colors.RGBSpace)
		
		for i := 0; i < len(rgba.Pix); i += 4 {
			r, g, b, a := rgba.Pix[i], rgba.Pix[i + 1], rgba.Pix[i + 2], rgba.Pix[i + 3]
			img.Data[ii] = float32(r) / float32(a)
			img.Data[ii + 1] = float32(g) / float32(a)
			img.Data[ii + 2] = float32(b) / float32(a)
			ii += 3
		}
		return img
	}
}

type ColorMap func(a,b,c float32) (q,w,e float32)

func (im Image3) GetNRGBA() image.NRGBA {
	nrgba := image.NRGBA{
		Pix: make([]uint8, im.W*im.H*4),
		Stride: 4*im.W,
		Rect: image.Rectangle{
			Min: image.Point{0, 0},
			Max: image.Point{im.W, im.H},
		},
	}
	np := 0
	for p := 0; p < im.W*im.H*3; p += 3 {
		r, g, b := im.Data[p], im.Data[p + 1], im.Data[p + 2]
		nrgba.Pix[np] = topix(r)
		nrgba.Pix[np + 1] = topix(g)
		nrgba.Pix[np + 2] = topix(b)
		nrgba.Pix[np + 3] = 255
		np += 4
	}
	return nrgba
}

func (im Image3) Map(fn ColorMap) Image3 {
	np := 0
	for p := 0; p < im.W*im.H*3; p += 3 {
		im.Data[p], im.Data[p + 1], im.Data[p + 2] =
			fn(im.Data[p], im.Data[p + 1], im.Data[p + 2])
		np += 4
	}
	return im
}

func (im *Image3) ChangeSpace(colorSpace int) *Image3 {
	convertFunc := colors.GetConvertFunc(im.ColorSpace, colorSpace)
	im.Map(ColorMap(convertFunc))
	im.ColorSpace = colorSpace
	return im
}

func (im Image3) Clone() (Image3) {
	ret := Image3{im.W, im.H, im.ColorSpace, make([]float32, len(im.Data))}
	copy(ret.Data, im.Data)
	return ret
}

func (im Image3) SavePng(path string) error {
	nrgba := image.NRGBA{
		Pix: make([]uint8, im.W*im.H*4),
		Stride: 4*im.W,
		Rect: image.Rectangle{
			Min: image.Point{0, 0},
			Max: image.Point{im.W, im.H},
		},
	}
	np := 0
	convert2srgb := colors.GetConvertFunc(im.ColorSpace, colors.SRGBSpace)
	for p := 0; p < im.W*im.H*3; p += 3 {
		r, g, b := im.Data[p], im.Data[p + 1], im.Data[p + 2]
		r, g, b = convert2srgb(r, g, b)
		nrgba.Pix[np] = topix(r)
		nrgba.Pix[np + 1] = topix(g)
		nrgba.Pix[np + 2] = topix(b)
		nrgba.Pix[np + 3] = 255
		np += 4
	}
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create png file: %s", err)
	}
	defer file.Close()
	err = png.Encode(file, &nrgba)
	if err != nil {
		return fmt.Errorf("write png file: $s", err)
	}
	return nil
}

func (im *Image3) Equalize() {
	//colors := map[float32]int{}
	colorsum := map[float32]int{}
	println(0)
	levels := []float32{}
	for p := 0; p < im.W*im.H*3; p += 3 {
		x, y, z := im.Data[p], im.Data[p + 1], im.Data[p + 2]
		l, _, _ := colors.Xyz2lab(x, y, z)
		levels = append(levels, l)
	}
	sort.Slice(levels, func(i, j int) bool {
		return levels[i] < levels[j]
	})
	trHi := im.W*im.H/100 // kostil to drop outliers
	/*
	var hi int = trHi
	var i int
	for i = len(levels) - 1; i > 0 && hi > 0; i-- {
		if hi > int(levels[i]) {
			hi -= int(levels[i])
		}
	}
	for j := i + 1; j < len(levels); j++ {
		levels[j] = levels[i]
	}
	for j := im.W*im.H - trHi; j < len(levels); j++ {
		levels[j] = levels[im.W*im.H - trHi]
	}
	*/
	println(1)
	var color float32 = levels[0]
	for i, c := range levels {
		if c != color {
			color = c
			colorsum[c] = i - 1
		}
	}
	println(2)
	to := levels[len(levels) - trHi]
	fmt.Println(to)
	for p := 0; p < im.W*im.H*3; p += 3 {
		x, y, z := im.Data[p], im.Data[p + 1], im.Data[p + 2]
		l, a, b := colors.Xyz2lab(x, y, z)
		lt := colorsum[l]
		l = (float32(lt)/float32(im.W*im.H))
		if l > 1 {
			l = 1
		}
		im.Data[p], im.Data[p + 1], im.Data[p + 2] = colors.Lab2xyz(l, a, b)
	}
	println(3)
}

func (im *Image3) At(x, y float32) (r, g, b float32) {
	// 00 01
	// 10 11
	ix, iy := int(x), int(y)

	if ix >= im.W - 1 {
		if ix >= im.W*2 {
			fmt.Println("achtung, image sample x out of range", ix, im.W, x)
		}
		ix = im.W - 1
	} else if ix < 0 {
		if ix < -im.W {
			fmt.Println("achtung, image sample x out of range", ix, im.W, x)
		}
		ix = 0
	}
	if iy >= im.H - 1 {
		if iy >= im.H*2 {
			fmt.Println("achtung, image sample y out of range", iy, im.H)
		}
		iy = im.H - 1
	} else if iy < 0 {
		if iy < -im.H {
			fmt.Println("achtung, image sample y out of range", iy, im.H, debug.IX, debug.IY, debug.Flag)
		}
		iy = 0
	}
	i00 := (ix + iy*im.W)*3
	i01 := i00 + 3
	i10 := i00 + im.W * 3
	i11 := i10 + 3
	if iy + 1 >= im.H {
		i10 = i00
		i11 = i01
	}
	if ix + 1 >= im.W {
		i01 = i00
		i11 = i10
	}

	ky := y - float32(iy)
	kx := x - float32(ix)
	r = (1 - ky)*(im.Data[i00]*(1 - kx) + im.Data[i01]*kx) +
		ky*(im.Data[i10]*(1 - kx) + im.Data[i11]*kx)
	g = (1 - ky)*(im.Data[i00 + 1]*(1 - kx) + im.Data[i01 + 1]*kx) +
		ky*(im.Data[i10 + 1]*(1 - kx) + im.Data[i11 + 1]*kx)
	b = (1 - ky)*(im.Data[i00 + 2]*(1 - kx) + im.Data[i01 + 2]*kx) +
		ky*(im.Data[i10 + 2]*(1 - kx) + im.Data[i11 + 2]*kx)
	return
}

func (im *Image3) AtInt(x, y int) (r, g, b float32) {
	if x >= im.W {
		if x >= im.W*2 {
			fmt.Println("achtung, image sample x out of range", x, im.W)
		}
		x = im.W - 1
	}
	if y >= im.H {
		if y >= im.H*2 {
			fmt.Println("achtung, image sample x out of range", y, im.H)
		}
		y = im.H - 1
	}
	i := (x + y*im.W)*3
	r, g, b = im.Data[i], im.Data[i + 1], im.Data[i + 2]
	return
}

func (im *Image3) AtUv(u, v float32) (r, g, b float32) {
	u -= float32(int(u))
	v -= float32(int(v))
	if u < 0 {
		u++
	}
	if v < 0 {
		v++
	}
	return im.At(u*float32(im.W), v*float32(im.H))
}

// derivative in the red channel
// todo pixel color values range from 0 to 1
// but x and y may have any range
// need to scale pixel color range accordingly
func (im *Image3) Derivative(x, y float32, delta float32) (r, drdx, drdy float32) {
	if x < 0 {
		panic(fmt.Sprintf("xW %v", x))
	}
	r, _, _ = im.At(x, y)
	if x + delta < float32(im.W) {
		rx, _, _ := im.At(x + delta, y)
		drdx = (rx - r)/delta
	}
	if y + delta < float32(im.H) {
		ry, _, _ := im.At(x, y + delta)
		drdy = (ry - r)/delta
	}
	return
}

func (im *Image3) Resize(w, h int) (im2 Image3) {
	im2 = NewImage3(w, h, im.ColorSpace)
	sx := float32(im.W) / float32(w)
	sy := float32(im.H) / float32(h)
	for y := 0; y < h; y++ {
		row := y*im2.W*3
		for x := 0; x < w; x++ {
			p := row + x*3
			r, g, b := im.At(float32(x)*sx, float32(y)*sy)
			im2.Data[p] = r
			im2.Data[p + 1] = g
			im2.Data[p + 2] = b
		}
	}
	return
}

// TODO implement a good downscaling method
func (im *Image3) Downscale(w, h int) (im2 Image3) {
	im2 = NewImage3(w, h, im.ColorSpace)
	sx := float32(w) / float32(im.W)
	sy := float32(h) / float32(im.H)
	add := func(ox, oy, x, y int, w float32) {
		if x >= im2.W || y >= im2.H {
			return
		}
		p := (y*im2.W + x)*3
		op := 3*(im.W*oy + ox)
		im2.Data[p] += w*im.Data[op]
		im2.Data[p + 1] += w*im.Data[op + 1]
		im2.Data[p + 2] += w*im.Data[op + 2]
	}
	for y := 0; y < im.H; y++ {
		p := 3*im.W*y
		ny := float32(y)*sy // new image y
		ny1 := float32(y + 1)*sy // new image y
		wy := float32(0)
		if int(ny) != int(ny1) {
			wy = (float32(int(ny1)) - ny)/sy
		}
		for x := 0; x < im.W; x++ {
			nx := float32(x)*sx
			nx1 := float32(x + 1)*sx
			if int(nx) != int(nx1) {
				w := (float32(int(nx1)) - nx)/sx
				if int(ny) != int(ny1) {
					add(x, y, int(nx), int(ny), w*wy)
					add(x, y, int(nx) + 1, int(ny), (1 - w)*wy)
					add(x, y, int(nx), int(ny) + 1, w*(1 - wy))
					add(x, y, int(nx) + 1, int(ny) + 1, (1 - w)*(1 - wy))
				} else {
					add(x, y, int(nx), int(ny), w)
					add(x, y, int(nx) + 1, int(ny), 1 - w)
				}
			} else {
				if int(ny) != int(ny1) {
					add(x, y, int(nx), int(ny), wy)
					add(x, y, int(nx), int(ny) + 1, 1 - wy)
				} else {
					add(x, y, int(nx), int(ny), 1)
				}
			}
			p += 3
		}
	}
	p := 0
	inv := 1/float32((1/sx)*(1/sy))
	for i := 0; i < im2.W*im2.H; i++ {
		im2.Data[p] *= inv
		im2.Data[p + 1] *= inv
		im2.Data[p + 2] *= inv
		p += 3
	}
	return im2
}

func (im Image3) Scale(w, h float32) (im2 Image3) {
	return im.Resize(int(w*float32(im.W)), int(h*float32(im.H)))
}

func (im Image3) FitInRectangle(w, h int) (im2 Image3) {
	if im.W <= w && im.H <= h {
		return im.Clone()
	}
	oratio := float32(im.W) / float32(im.H)
	fratio := float32(w) / float32(h)
	if oratio > fratio {
		return im.Downscale(w, int(float32(w)/oratio))
	} else {
		return im.Downscale(int(float32(h)*oratio), h)
	}
}

func Noop() {
	_ = math.Cos
	_ = math32.Pow
}

func main() {
	fmt.Println("vim-go")
}
