package gui

import (
	"ly/geo"
	"ly/cameras"
	"ly/colors"
	"ly/films"
)

type Canvas struct {
	Server *Server
}

func (c *Canvas) DrawBox(box geo.Box, cam cameras.Camera, color colors.RGBColor, film *films.Film) {
	w, h := film.W, film.H
	line := func(p1, p2 geo.Vec3) {
		x1, y1 := cam.PlotDot(p1)
		x2, y2 := cam.PlotDot(p2)
		iy1 := int((0.5 - y1)*float32(h))
		ix1 := int(x1*float32(h)+0.5*float32(w))
		iy2 := int((0.5 - y2)*float32(h))
		ix2 := int(x2*float32(h)+0.5*float32(w))
		msg := LineMessage{
			X1: ix1,
			Y1: iy1,
			X2: ix2,
			Y2: iy2,
			R: color.R,
			G: color.G,
			B: color.B,
		}
		c.Server.QueueOut <- msg
	}
	center := box.Min.Add(box.Max).Mul(0.5)
	xstep := geo.Vec3{box.Max.X - center.X, 0, 0}
	ystep := geo.Vec3{0, box.Max.Y - center.Y, 0}
	zstep := geo.Vec3{0, 0, box.Max.Z - center.Z}
	// draw bottom square (-Z)
	// 2 3  +x
	// 1 4  0 +y
	floor := center.Sub(zstep)
	f1 := floor.Sub(xstep).Sub(ystep)
	f2 := floor.Sub(xstep).Add(ystep)
	f3 := floor.Add(xstep).Add(ystep)
	f4 := floor.Add(xstep).Sub(ystep)
	line(f1, f2)
	line(f2, f3)
	line(f3, f4)
	line(f4, f1)
	// draw top square (+Z)
	// 2 3  +x
	// 1 4  0 +y
	roof := center.Add(zstep)
	r1 := roof.Sub(xstep).Sub(ystep)
	r2 := roof.Sub(xstep).Add(ystep)
	r3 := roof.Add(xstep).Add(ystep)
	r4 := roof.Add(xstep).Sub(ystep)
	line(r1, r2)
	line(r2, r3)
	line(r3, r4)
	line(r4, r1)
	// connect
	line(f1, r1)
	line(f2, r2)
	line(f3, r3)
	line(f4, r4)
}
