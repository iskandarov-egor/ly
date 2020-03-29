package main

import (
	"ly/scene"
	"ly/testing"
	"ly/config"
	"ly/cameras"
	"ly/colors"
	"ly/geo"
	"ly/films"
	"ly/sampling"
	"ly/tracers"
	"ly/spectra"
	"ly/debug"
	"ly/img"
	"ly/gui"
	"ly/obj"
	"ly/util/math32"
	"ly/util/pbrt"
	"math/rand"
	"math"
	"sort"
	"os"
	"log"
	"flag"
	"fmt"
	"time"
	"runtime/pprof"
)

func recvCmd() {
	server := gui.NewServer()
	go server.Serve()

	world, cam := scene.Scene{}, &cameras.PerspectiveCamera{}
	tree := scene.MakeBVH(world.Shapes)
	world.Accelerator = tree
	tracer := tracers.NewPathTracer(0, 0)
	film := films.NewFilm(400, 400)
	//visualizeBVHTree(film, tree, world, cam)
	//im := film.ToImage()
	//im.SavePng("test/BVH.png", nil)
	//return

	for {
		log.Println("waiting for message...")
		msg := <- server.QueueIn
		switch msg := msg.(type) {
			case gui.RenderMessage:
				log.Println("got render msg", msg)
				var im img.Image3
				if msg.Area == nil {
					server.QueueOut <- gui.CanvasSizeMessage{
						W: film.W,
						H: film.H,
					}
					region := DrawRegion{
						x1: 0,
						y1: 0,
						x2: film.W,
						y2: film.H,
					}
					draw(&world, tracer, cam, film, 4, 10, region)
					im = film.ToImage()
				} else {
					a := msg.Area
					region := DrawRegion{
						x1: a.Left,
						y1: a.Top,
						x2: a.Right,
						y2: a.Bottom,
					}
					draw(&world, tracer, cam, film, 4, 10, region)
					//im = film.ToImageArea(a.Left, a.Top, a.Right, a.Bottom)
					panic("TODO")
				}
				im.Map(colors.Xyz2rgb)
				out := gui.ImageMessage{
					RGBA: im.GetNRGBA().Pix,
					W: im.W,
					H: im.H,
				}
				if msg.Area != nil {
					out.X = msg.Area.Left
					out.Y = msg.Area.Top
				} else {
					server.QueueOut <- gui.CanvasSizeMessage{
						W: im.W,
						H: im.H,
					}
				}
				log.Println("sending!")
				server.QueueOut <- out
			default:
				log.Println("got unk msg", msg)
		}
	}
}

func logProgress(progress float32, dt time.Duration) {
	fmtDuration := func(d time.Duration) string {
		h := int(d.Hours())
		d -= time.Duration(h) * time.Hour
		m := int(d.Minutes())
		d -= time.Duration(m) * time.Minute
		s := int(d.Seconds())
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	eta := time.Duration(float32(dt) * (1 - progress) / progress)
	log.Printf(
		"progress=%.4f time=%s eta=%s",
		progress, fmtDuration(dt), fmtDuration(eta),
	)
}

func renderFile(path string) error {
	world := scene.Scene{}
	conf, err := config.Load(path, &world)
	if err != nil {
		return fmt.Errorf("load scene file %q: %s", path, err)
	}
	options := conf.Options
	film := films.NewFilm(options.Profile.Width, options.Profile.Height)
	
	region := DrawRegion{y1: 0, y2: film.H, x1: 0, x2: film.W}
	if options.Region != nil {
		r := options.Region
		region = DrawRegion{x1: r[0], y1: r[1], x2: r[2], y2: r[3]}
	}
	drawing := startDrawing(
		&world,
		conf.Tracer,
		conf.Camera,
		film,
		options.Goroutines,
		options.Profile.PixelSamples,
		region,
	)

	startTime := time.Now()

	saveTicker := time.NewTicker(time.Duration(options.Profile.SaveInterval) * time.Second)
	logTicker := time.NewTicker(5 * time.Second)
	for done := false; !done; {
		select {
			case <-drawing.Done:
				done = true
			case <-logTicker.C:
				logProgress(drawing.GetProgress(), time.Since(startTime))
			case <-saveTicker.C:
				//drawing.Pause()
				im := film.ToImage()
				err := im.SavePng(options.Outfile)
				if err != nil {
					log.Printf("error saving intermediate image: %s", err)
				} else {
					log.Printf("saved intermediate image %s", options.Outfile)
				}
				//drawing.Unpause()
		}
	}
	im := film.ToImage()
	err = im.SavePng(options.Outfile)
	if err != nil {
		return fmt.Errorf("save result to %q: %s", options.Outfile, err)
	}
	return nil
}

func executeCmd() {
	flag.Parse()
	if len(flag.Args()) == 0 {
		log.Println("please provide scene config path")
		os.Exit(1)
	}
	for _, filename := range flag.Args() {
		err := renderFile(filename)
		if err != nil {
			log.Printf("render file %q: %s", filename, err)
			os.Exit(1)
		}
	}
}

func main() {
	log.SetOutput(os.Stdout)
	executeCmd()

	return
}

func init() {
	spectra.Noop()
	img.Noop()
	_ = rand.Intn
	_ = cameras.NewPerspectiveCamera
	_ = fmt.Println
	_ = math32.Sqrt
	_ = pprof.StartCPUProfile
	_ = time.Sleep
	_ = geo.Vec3{}
	pbrt.Noop()
	_ = math.Sin
	_ = sort.Search
	debug.Noop()
	sampling.Noop()
	testing.Noop()
	obj.Noop()
}
