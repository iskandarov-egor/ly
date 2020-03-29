package testing


import (
	"ly/geo"
	"ly/img"
	"ly/sampling"
	"math"
	"math/rand"
	"fmt"
	"ly/scene"
	"ly/spectra"
	"ly/colors"
	"ly/util/math32"
)

func TestBulbPower() {
	//bulbPos := geo.Vec3{0, 0, 0}
	//sphereR := 10
	latStep := 3.1415/100
	n := 0
	for lat := -3.1415/2 + latStep; lat < 3.1415/2; lat += latStep {
		lonStep := latStep/math.Cos(lat)
		for lon := float64(0); lon < 3.1415*2; lon += lonStep {
			n++
		}
	}
	fmt.Println(n)
	samplesPerRad := 1/latStep
	fmt.Println(4*3.1415*samplesPerRad*samplesPerRad)
}

func testBoxIntersect1(box geo.Box) {
	center := box.Min.Add(box.Max).Mul(0.5)
	diag := box.Diagonal().Len()
	for i := 1; i < 360; i += 15 {
		for j := 1; j < 180; j += 15 {
			angleXY := float64(i)*3.1415/180
			angle2 := float64(j)*3.1415/180
			offset := geo.Vec3{
				diag*float32(math.Sin(angle2)),
				diag*float32(math.Cos(angleXY)*math.Cos(angle2)),
				diag*float32(math.Sin(angleXY)*math.Cos(angle2)),
			}
			// from outside in
			ray := geo.Ray{
				center.Add(offset),
				offset.Mul(-1),
			}
			isHit := box.Intersect(ray)
			if !isHit {
				fmt.Println("FAIL 1", angleXY, angle2)
			}

			// from inside out
			ray = geo.Ray{
				center,
				offset,
			}
			isHit = box.Intersect(ray)
			if !isHit {
				fmt.Println("FAIL 2", angleXY, angle2)
			}

			perp := geo.Vec3{
				1/offset.X,
				1/offset.Y,
				-2/offset.Z,
			}
			ray = geo.Ray{
				center.Add(offset),
				perp,
			}
			isHit = box.Intersect(ray)
			if isHit {
				fmt.Println("FAIL 3", ray)
			}
		}
	}
}

func TestBoxIntersect() {
	testBoxIntersect1(geo.Box{
		Min: geo.Vec3{-1, -1, -1},
		Max: geo.Vec3{1, 1, 1},
	})
	testBoxIntersect1(geo.Box{
		Min: geo.Vec3{-10, -10, -10},
		Max: geo.Vec3{-3, -3, -3},
	})
	testBoxIntersect1(geo.Box{
		Min: geo.Vec3{-1, -1, -1},
		Max: geo.Vec3{10, 1, 1},
	})
}

func testSpherical() {
	zen := geo.Vec3{0, 0, 1}
	r := geo.Vec3{1, 1, 1}.Normalized()
	r2 := geo.Vec3{-2, 1, 1}.Normalized()
	x := r.PlaneProj(zen).Normalized()
	y := r2.PlaneProj(zen)
	y = y.PlaneProj(x).Normalized()
	fmt.Println(x, y)
	for a1 := 0; a1 < 360; a1 += 10 {
		//for a2 := -90; a2 < 90; a2 += 10 {
		for a2 := -0; a2 < 10; a2 += 10 {
			aa1 := float32(a1)*3.141592/180
			aa2 := float32(a2)*3.141592/180
			t := x.Mul(math32.Cos(aa1)).Add(y.Mul(math32.Sin(aa1))).Add(zen.Mul(math32.Sin(aa2))).Normalized()
			q, w, e := scene.Spherical(zen, x, t)
			fmt.Println(a1, a2, 180/3.1415*math.Acos(float64(q)), 180/3.1415*math.Acos(float64(w)), 180/3.1415*math.Acos(float64(e)))
		}
	}
}

func TestBsdfEnergy(mat scene.Material) {
	angleStepIn := float32(1)
	angleStep := float32(0.5)
	var hp scene.ShapeHitPoint
	hp.Normal = geo.Vec3{0, 0, 1}
	dz := angleStep / 90 * math.Pi / 2
	da := angleStep / 360 * 2 * math.Pi
	achtung := false
	for zenIn := float32(0); zenIn < 90; zenIn += angleStepIn {
		var sum spectra.RGBSpectr
		_ = sum
		in := geo.Vec3{math32.Sin(zenIn * math.Pi / 180), 0, math32.Cos(zenIn * math.Pi / 180)}
		for zenOutI := float32(0); zenOutI < 90; zenOutI += angleStep {
			zenOut := zenOutI * math.Pi / 180
			for azimI := float32(0); azimI < 360; azimI += angleStep {
				azim := azimI * math.Pi / 180
				out := geo.Vec3{math32.Cos(azim), math32.Sin(azim), math32.Cotan(zenOut)}
				bsdf := mat.BSDF(&hp, in, out.Negated())
				//fmt.Println(in, out.Negated(), bsdf)
				if false {
					z1, z2, az := scene.Spherical(geo.Vec3{0, 0, 1}, in.Normalized(), out.Normalized())
					z1 = math32.Acos(z1) * 180 / math.Pi
					z2 = math32.Acos(z2) * 180 / math.Pi
					az = math32.Acos(az) * 180 / math.Pi
					fmt.Println(zenIn, zenOut, azim, "--", z1, z2, az, ";", in, out)
				}
				r := math32.Sin(zenOut)
				r2 := math32.Sin(zenOut + dz)
				ds := (r + r2) / 2 * da * dz
				bsdf.Mul(ds * math32.Cos(zenOut))
				sum.SpectrAdd(bsdf)
			}
		}

		var mark string
		if sum.R > 1.05 || sum.G > 1.05 || sum.B > 1.05 {
    		mark = fmt.Sprintf("AHTUNG")
			achtung = true
		}
		
		fmt.Println(zenIn, sum, mark)
	}
	if achtung {
		fmt.Println("AHTUNG")
	}
}

func testDistr() {
	panic("aa seed")
	rand.Seed(1230)
	f := []float32{
		0, 2, 3, 0,
		1, 1, 4, 4,
		2, 2, 0, 1,
		1, 1, 1, 2,
	}
	r := []float32{
		0, 2, 3, 0,
		1, 0, 4, 4,
		2, 2, 0, 1,
		1, 1, 0, 2,
	}
	im1 := img.Image1{
		W: 4,
		H: 4,
		Data: f,
	}
	d := sampling.NewDistribution2D(im1)
	fmt.Println("MP", d.Marginal.Pdf)
	fmt.Println("MC", d.Marginal.Cdf)
	fmt.Println("0P", d.Conditional[0].Pdf)
	fmt.Println("0C", d.Conditional[0].Cdf)
	fmt.Println(d.Conditional[0].Pdf)
	fmt.Println(d.Conditional[1].Pdf)
	fmt.Println(d.Conditional[2].Pdf)
	fmt.Println(d.Conditional[3].Pdf)
	for i := 0; i < 16*10000; i++ {
		x, y, pdf := d.Sample(rand.Float32(), rand.Float32())
		xi := int(x)
		yi := int(y)
		r[yi*im1.W + xi]++
		if xi == 1 && yi == 2 {
			fmt.Println(4*(y - 2) + float32(int(4*(x - 1))*im1.W))
		}
		_ = pdf
	}
	fmt.Println(r)
}

func testDistr2() {
	im := img.LoadPng("files/textures/skylight-morn.png")
	im.Map(colors.Rgb2xyz)
	im.ColorSpace = colors.XYZSpace
	im = im.Scale(0.125, 0.125)
	txt := im.GetImage1()
	d := sampling.NewDistribution2D(txt)
	for i := 0; i < 1000000; i++ {
		x, y, pdf := d.Sample(rand.Float32(), rand.Float32())
		pdf2 := d.Pdf(x, y)
		if pdf != pdf2 {
			fmt.Println(x, y, pdf, pdf2)
			panic("aaa")
		}
	}
}

func testDistr3() {
	im := img.LoadPng("files/textures/skylight-morn.png")
	im.Map(colors.Rgb2xyz)
	im.ColorSpace = colors.XYZSpace
	im = im.Scale(0.125, 0.125)
	txt := im.GetImage1()
	d := sampling.NewDistribution2D(txt)
	pdfs := img.NewImage3(txt.W, txt.H, colors.RGBSpace)
	for i := 0; i < 1000000; i++ {
		x, y, pdf := d.Sample(rand.Float32(), rand.Float32())
		pdfs.Set(int(x*float32(pdfs.W)), int(y*float32(pdfs.H)), pdf, pdf, pdf)
	}
		sum := float32(0)
	for y := 0; y < pdfs.H; y++ {
	for x := 0; x < pdfs.W; x++ {
		sum += pdfs.Data[3*pdfs.W*y + 3*x]
	}
	}
		fmt.Println("Y", sum)
	pdfs.Normalize()
	pdfs.SavePng("d.png")
}

func testSpher() {
	for i := 0; i < 10; i++ {
		e1 := float32(i)/10
		for j := 0; j < 10; j++ {
			e2 := float32(j)/10
			azimuthAngle := e1*2*math.Pi // todo direction
			zenithAngle := e2*math.Pi
			zenithSin := math32.Sin(zenithAngle)
			dir := geo.Vec3{
				zenithSin*math32.Cos(azimuthAngle),
				zenithSin*math32.Sin(azimuthAngle),
				math32.Cos(zenithAngle),
			}.Normalized()

			zenithCos, azimuthSin, azimuthCos := geo.SphericalFromVec3(dir)
			azimuth := math32.Atan2(azimuthSin, azimuthCos)
			zenith := math32.Acos(zenithCos)
	if azimuth < 0 {
		azimuth += 2*math.Pi
		//todo
	}
			ee1 := azimuth / 2 / math.Pi
			ee2 := zenith / math.Pi
			if math32.Abs(e1 - ee1) < 0.0001 && math32.Abs(e2 - ee2) < 0.0001 {
				fmt.Println("O", e1, e2, ee1, ee2, dir)
			} else {
				fmt.Println("X", e1, e2, ee1, ee2, dir, azimuthCos, azimuthSin, zenithCos)
			}
		}
	}
}

func TestSnellLawNormalFinding() {
	sq := func(x float32) float32 {
		return x*x
	}
	fromSin := func(sin float32) geo.Vec3 {
		cos := math32.Sqrt(1 - sin*sin)
		return geo.Vec3{sin, cos, 0}
	}
	fromSin2 := func(sin float32) geo.Vec3 {
		cos := math32.Sqrt(1 - sin*sin)
		return geo.Vec3{sin, -cos, 0}
	}
	type Case struct{
		I, O geo.Vec3
		N float32
	}
	cases := map[string]Case{
		"glassInOk0": Case{
			fromSin(0.01),
			fromSin(0.005),
			0.5,
		},
		"glassInOk1": Case{
			fromSin(0.6),
			fromSin(0.3),
			0.5,
		},
		"glassInOk1Swap": Case{
			fromSin(0.3),
			fromSin(0.6),
			0.5,
		},
		"glassInOk2": Case{
			fromSin(0.98),
			fromSin(0.49),
			0.5,
		},
		"glassOutOk1": Case{
			fromSin(0.3),
			fromSin(0.6),
			2,
		},
		"glassOutRefEdge": Case{
			fromSin(0.4999),
			fromSin(0.4999*2),
			2,
		},
		"glassOutRef1": Case{
			fromSin(0.6),
			fromSin2(0.6),
			2,
		},
		"impossible1": Case{
			fromSin(0.6),
			//fromSin2(fromSin(0.6).Y).Negated().Add(fromSin(0.6).Mul(1.1)).Normalized(),
			fromSin(-0.3),
			30,
		},
	}
	casename := "glassOutRefEdge"
	casename = "glassInOk1"
	I := cases[casename].I
	O := cases[casename].O
	N := cases[casename].N
	

	Hdir := O.Sub(I.Mul(N))
	Hraw := Hdir.Normalized()
	IYraw := Hraw.Mul(I.Scalar(Hraw))
	IXraw := I.Sub(IYraw)
	OXraw := O.Sub(Hraw.Mul(O.Scalar(Hraw)))
	IYrawLen := IYraw.Len()
	HMagRHS := N*IYrawLen - O.Scalar(Hraw)

	sqr := (1 + (1 - N*N)/(N*N*IYrawLen*IYrawLen))
	sqrMin := 1 / (1 + N*N) / (N*N*IYrawLen*IYrawLen)
	sqrIyMin := N * N / (1 + N*N)
	mult := N*(-1 + math32.Sqrt(sqr))
	fmt.Println("Hdir", Hdir)
	fmt.Println("HdirLen", Hdir.Len())
	fmt.Println("HI HO signs", I.Scalar(Hraw) > 0, O.Scalar(Hraw) > 0)
	fmt.Println("I", I)
	fmt.Println("O", O)
	fmt.Println("N", N)
	fmt.Println("IXraw", IXraw)
	fmt.Println("OXraw", OXraw)
	fmt.Println("IYrawLen", IYrawLen)
	fmt.Println("IXraw", IXraw)
	fmt.Println("sqr", sqr)
	fmt.Println("sqr min", sqrMin)
	fmt.Println("sqr IY min", sqrIyMin, IYrawLen*IYrawLen)
	fmt.Println("mult", mult)
	fmt.Println("HMagRHS", HMagRHS)
	errVec := IXraw.Mul(N).Sub(OXraw)
	fmt.Println()
	fmt.Println("<errvec>", errVec)
	fmt.Println("<ok>", errVec.Len() < 0.001 * I.Len())

	OI := O.Scalar(I)
	NI := I.Mul(N)
	OY := O.Scalar(Hraw)
	fmt.Println("S1", math32.Sqr(O.Sub(NI).Normalized().Scalar(I)))
	fmt.Println("SL-3", (OI - N)*(OI - N)/O.Sub(NI).LenSquared())
	fmt.Println("SL-2", (OI - N)*(OI - N)/(1 - 2*N*OI + N*N))
	fmt.Println("SL", (OI*OI - 2*N*OI + N*N)/(1 - 2*N*OI + N*N))
	fmt.Println("SA", math32.Sqr(OI - N)/(1 - 2*N*OI + N*N))
	OI = 0
	fmt.Println("SLmin", (OI*OI - 2*N*OI + N*N)/(1 - 2*N*OI + N*N))
	OI = 1
	fmt.Println("SLmin", (OI*OI - 2*N*OI + N*N)/(1 - 2*N*OI + N*N))
	OI = 1/N
	fmt.Println("SLmin", (OI*OI - 2*N*OI + N*N)/(1 - 2*N*OI + N*N))
	OI = O.Scalar(I)
	OY2t := sq(1 - OI*N) / (sq(1 - OI*N) + N*N*(1 - OI*OI))
	fmt.Println("OY2", OY*OY, OY2t)
}

func Noop() {
	spectra.Noop()
}
