package main

import (
	"ly/films"
	"ly/scene"
	"ly/cameras"
	"ly/gui"
	"ly/tracers"
	"ly/colors"
	"ly/geo"
	"ly/spectra"
	"fmt"
	"time"
)

func visualizeBVHTree(film *films.SimpleFilm, tree *scene.BVHNode, world scene.Scene, cam cameras.Camera) {
	server := gui.NewServer()
	go server.Serve()

	server.QueueOut <- gui.CanvasSizeMessage{
		W: film.Width(),
		H: film.Height(),
	}
	canvas := server.NewCanvas()

	colori := 0
	var walk func(node *scene.BVHNode, indent string)
	walk = func(node *scene.BVHNode, indent string) {
		fmt.Printf(indent + "%d shapes, box %v\n", len(node.Shapes), node.BoundingBox)
		if node.Left != nil {
			fmt.Printf(indent + "left:\n")
			walk(node.Left, indent + "  ")
			fmt.Printf(indent + "right:\n")
			walk(node.Right, indent + "  ")
		} else {
			canvas.DrawBox(node.BoundingBox, cam, colors.Distinct(colori), film)
			colori++
		}
	}
	inspector := tracers.Inspector{
		Callback: func(ray geo.Ray, world *scene.Scene) spectra.Spectr {
			id := scene.TestFindBVH(tree, ray)
			if id == -1 {
				return &spectra.RGBSpectr{}
			}
			color := colors.Distinct(id)
			return &spectra.RGBSpectr{color.R, color.G, color.B}
		},
	}
	region := DrawRegion{
		x1: 0,
		y1: 0,
		x2: film.Width(),
		y2: film.Height(),
	}
	draw(&world, inspector, cam, film, 4, 10, region)
	im := film.ToImage()

	msg := gui.ImageMessage{
		RGBA: im.GetNRGBA().Pix,
		W: im.W,
		H: im.H,
	}
	server.QueueOut <- msg
	walk(tree, "")
	time.Sleep(time.Second*200)
}
