package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"log"
	"math/rand"
	"time"

	_ "image/png"

	eb "github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	gravity         = 0.0981
	startingGophers = 1000
	gophersPerClick = 1000
)

var (
	spriteSize  V2
	gopherImage *eb.Image
	//go:embed gopher.png
	gopherImageData []byte

	rng = func() *rand.Rand {
		src := rand.NewSource(time.Now().UnixNano())
		return rand.New(src)
	}()
	colors = []color.Color{
		color.White,
		color.RGBA{
			R: 255,
			A: 60,
		},
		color.RGBA{
			G: 255,
			A: 125,
		},
		color.RGBA{
			B: 255,
			A: 190,
		},
	}
)

func main() {
	eb.SetWindowResizingMode(eb.WindowResizingModeEnabled)

	gm := NewGophermark(startingGophers)
	if err := eb.RunGame(gm); err != nil {
		log.Fatal(err)
	}
}

func init() {
	img, _, err := image.Decode(bytes.NewReader(gopherImageData))
	if err != nil {
		log.Fatal(err)
	}

	gopherImage = eb.NewImageFromImage(img)

	{
		bounds := gopherImage.Bounds().Max
		spriteSize = V2{
			X: float32(bounds.X),
			Y: float32(bounds.Y),
		}
	}

}

type (
	Gophermark struct {
		Total   int
		Size    V2
		Gophers []Gopher
	}
	Gopher struct {
		Pos     V2
		Vel     V2
		Overlay color.Color
	}
	V2 struct {
		X, Y float32
	}
)

func NewGophermark(count int) *Gophermark {
	gm := new(Gophermark)
	gm.Total = count
	gm.Gophers = make([]Gopher, gm.Total)

	for i := range gm.Gophers {
		b := &gm.Gophers[i]
		initGopher(b, 0, 0)
	}

	return gm
}

func (gm *Gophermark) Update() error {
	sx := gm.Size.X
	sy := gm.Size.Y

	for i := range gm.Gophers {
		g := &gm.Gophers[i]

		g.Vel.Y += gravity
		g.Pos.X += g.Vel.X
		g.Pos.Y += g.Vel.Y

		switch {
		case g.Pos.Y >= sy-spriteSize.Y:
			g.Vel.Y *= 0.85 / 2
			if rng.Float32() > 0.5 {
				g.Vel.Y -= rng.Float32() * 8
			}
		case g.Pos.Y < 0:
			g.Vel.Y = -g.Vel.Y
		case g.Pos.X > sx-spriteSize.X:
			g.Vel.X = -abs(g.Vel.X)
		case g.Pos.X < 0:
			g.Vel.X = abs(g.Vel.X)
		}
	}

	mx, my := eb.CursorPosition()

	if eb.IsMouseButtonPressed(eb.MouseButtonLeft) {
		toAdd := make([]Gopher, gophersPerClick)
		for i := 0; i < gophersPerClick; i += 1 {
			initGopher(&toAdd[i], float32(mx), float32(my))
		}

		gm.Gophers = append(gm.Gophers, toAdd...)
		gm.Total += gophersPerClick
	}

	return nil
}

func (gm *Gophermark) Draw(sc *eb.Image) {
	sc.Fill(color.Black)

	for i := 0; i < len(gm.Gophers); i += 1 {
		b := &gm.Gophers[i]
		op := eb.DrawImageOptions{}
		op.ColorScale.ScaleWithColor(b.Overlay)
		op.GeoM.Translate(float64(b.Pos.X), float64(b.Pos.Y))
		sc.DrawImage(gopherImage, &op)
	}

	color := color.RGBA{0, 0, 0, 200}
	vector.DrawFilledRect(sc, 10, 10, 100, 50, color, false)

	{
		fps := eb.ActualFPS()
		tps := eb.ActualTPS()

		ebitenutil.DebugPrintAt(sc, fmt.Sprintf("fps: %.2f", fps), 10, 10)
		ebitenutil.DebugPrintAt(sc, fmt.Sprintf("tps: %.2f", tps), 10, 24)
		ebitenutil.DebugPrintAt(sc, fmt.Sprintf("gophers: %d", gm.Total), 10, 38)

		eb.SetWindowTitle(fmt.Sprintf("Ebitengen Gophermark, fps: %.2f, gophers: %d", fps, gm.Total))
	}
}

func (b *Gophermark) Layout(iw, ih int) (ow, oh int) {
	b.Size.X = float32(iw)
	b.Size.Y = float32(ih)
	return iw, ih
}

func initGopher(b *Gopher, x, y float32) {
	b.Overlay = colors[rng.Intn(len(colors))]
	b.Pos = V2{
		X: x,
		Y: y,
	}
	b.Vel = V2{
		X: rng.Float32(),
		Y: rng.Float32(),
	}
}

func abs(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}
