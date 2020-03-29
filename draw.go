package main

import (
	"ly/scene"
	"ly/cameras"
	"ly/films"
	"ly/tracers"
	"ly/debug"
	"ly/img"
	"ly/sampling"
	"ly/spectra"
	"ly/util/math32"
	"fmt"
	"math"
	"sync/atomic"
	"log"
	goDebug "runtime/debug"
	"math/rand"
	"sync"
)

type DrawRegion struct {
	x1, y1, x2, y2 int
}

// 0000000 00
// 1000000 32
// 0100000 16
// 1100000 48
// 0010000 08
// 1010000 40
// 0110000 24
// 1110000 56
func bitReversedSequence(dest []int) {
	var f func(dest []int, acc int, levelMask int)
	f = func (dest []int, acc int, levelMask int) {
		if len(dest) == 2 {
			dest[0] = acc
			dest[1] = levelMask + acc
			
			return
		}
		if len(dest) % 2 != 0 {
			panic("aaa")
		}
		f(dest[:len(dest)/2], acc, levelMask << 1)
		f(dest[len(dest)/2:], levelMask + acc, levelMask << 1)
	}
	f(dest, 0, 1)
}

type PixelTask struct{
	x, y int
}

type PixelOrder chan PixelTask

func bitReversedPixelOrder(region DrawRegion) PixelOrder {
	nStripes := 32
	if nStripes > region.y2 - region.y1 {
		// closest power of 2
		nStripes = int(math.Pow(2, float64(int(math.Log2(float64(region.y2 - region.y1))))))
	}
	stripeWidth := int(math.Ceil(float64(region.y2 - region.y1)/float64(nStripes)))
	stripeList := make([]int, nStripes)
	bitReversedSequence(stripeList)

	ret := make(chan PixelTask)
	
	go func() {
		for _, stripe := range stripeList {
			yTo := stripeWidth*(stripe + 1)
			if region.y2 - region.y1 < yTo {
				yTo = region.y2 - region.y1
			}
			for iy := stripeWidth*stripe; iy < yTo; iy++ {
				for ix := region.x1; ix < region.x2; ix++ {
					ret <- PixelTask{
						y: iy + region.y1,
						x: ix,
					}
				}
			}
		}
		close(ret)
	}()
	return ret
}

type Drawing struct {
	inProgressChan chan int
	Done chan struct{}
	progress uint32
}

func (d *Drawing) GetProgress() float32 {
	return math.Float32frombits(atomic.LoadUint32(&d.progress))
}

func (d *Drawing) Pause() {
	for i := 0; i < cap(d.inProgressChan); i++ {
		d.inProgressChan <- 1
	}
}

func (d *Drawing) Unpause() {
	for i := 0; i < cap(d.inProgressChan); i++ {
		<-d.inProgressChan
	}
}

func startDrawing(
	world *scene.Scene,
	tracer tracers.Tracer,
	cam cameras.Camera,
	film *films.Film,
	nGoroutines int,
	nPixelSamples int,
	region DrawRegion,
) *Drawing {
	w, h := film.W, film.H
	pxWidth := 1/float32(h)
	sampler := sampling.NewUniform2D()

	pixelChan := make(chan PixelTask, 1000)
	drawing := Drawing{
		inProgressChan: make(chan int, nGoroutines),
		Done: make(chan struct{}, 1),
	}

	var wg sync.WaitGroup
	
	for i := 0; i < nGoroutines; i++ {
		go func(i int) {
			for {
				//drawing.inProgressChan <- 1
				pix, ok := <-pixelChan
				if !ok {
					//<-drawing.inProgressChan
					break
				}
				y := (0.5 - float32(pix.y)/float32(h))
				x := (float32(pix.x) - 0.5*float32(w))/float32(h)
				debug.IX = pix.x
				debug.IY = pix.y
				debug.INT = (pix.x == 302 && pix.y == 310)
				for si := 0; si < nPixelSamples; si++ {
					debug.S = si
					offx, offy := sampler.Next()
					sx := x + pxWidth*(offx - 0.5)
					sy := y + pxWidth*(offy - 0.5)
					ray := cam.GenerateRay(sx, sy)
					var L spectra.Spectr
					func(){
						defer func() {
							if r := recover(); r != nil {
								L = spectra.NewRGBSpectr(1, 0, 0)
								log.Printf(
									"panic at pixel [%d, %d]: %s\n stack trace: %s\n",
									 pix.x, pix.y, r, goDebug.Stack())
							}
						}()
						L = tracer.Trace(ray, world)
					}()
					if debug.Mark != nil {
						L = spectra.NewRGBSpectr(debug.Mark.R, debug.Mark.G, debug.Mark.B)
						debug.Mark = nil
					}
					weight := 0.5 - math32.Abs((offx - 0.5)*(offy - 0.5))
					film.AddSample(pix.x, pix.y, L, weight)
				}
				//<-drawing.inProgressChan
			}
			wg.Done()
		}(i)
		wg.Add(1)
	}

	go func() {
		i := 0
		for task := range bitReversedPixelOrder(region) {
			pixelChan <- task
			i++
			progress := float32(i) / float32((region.y2 - region.y1)*(region.x2 - region.x1))
			atomic.StoreUint32(&drawing.progress, math.Float32bits(progress))
		}

		close(pixelChan)
		wg.Wait()
		drawing.Done <- struct{}{}
	}()

	return &drawing
}

func draw(
	world *scene.Scene,
	tracer tracers.Tracer,
	cam cameras.Camera,
	film *films.Film,
	nGoroutines int,
	nPixelSamples int,
	region DrawRegion,
) {
	drawing := startDrawing(world, tracer, cam, film, nGoroutines, nPixelSamples, region)
	_ = <- drawing.Done
	return
}

func putDot(cam cameras.Camera, im img.Image3, x, y float32) {
	w, h := im.W, im.H
	iy := int((0.5 - y)*float32(h))
	ix := int(x*float32(h)+0.5*float32(w))
	// image is not necessarily RGB
	im.Set(ix, iy, 1, 0, 0)
}

func init() {
	_ = fmt.Print
	_ = sync.Mutex{}
	_ = rand.Intn
}
