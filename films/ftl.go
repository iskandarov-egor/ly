package films

import(
	"ly/spectra"
)

type FTLFilm struct {
	Fps    float32
	Frames []*SimpleFilm
}

// @fps is frames per second if the speed of light was 1 unit per second
func NewFTLFilm(w, h int, fps float32, nframes int) *FTLFilm {
	film := FTLFilm{
		Frames: make([]*SimpleFilm, nframes),
		Fps: fps,
	}
	for i := range film.Frames {
		film.Frames[i] = NewFilm(w, h)
	}
	return &film
}

func (f *FTLFilm) Width() int {
	return f.Frames[0].W
}

func (f *FTLFilm) Height() int {
	return f.Frames[0].H
}

func (f *FTLFilm) AddSample(x, y int, s *spectra.TimedSpectr, weight float32) {
	for i := 0; i < len(f.Frames); i++ {
		f.Frames[i].AddSample(x, y, s.Frames[i], weight)
	}
}
