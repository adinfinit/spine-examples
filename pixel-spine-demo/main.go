package main

import (
	"log"
	"time"

	"github.com/adinfinit/spine-demo/animation"
	"github.com/golang/freetype/truetype"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"

	"golang.org/x/image/colornames"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
)

func main() {
	pixelgl.Run(run)
}
func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Pixel and Spine",
		Bounds: pixel.R(0, 0, 1024, 768),
		VSync:  true,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	face, err := loadFont(40)
	if err != nil {
		panic(err)
	}

	atlas := text.NewAtlas(face, text.ASCII)
	txt := text.New(pixel.V(0, 0), atlas)
	txt.Color = colornames.Black

	var characters []*Character
	for _, loc := range animation.LoadList("../animation") {
		character, err := LoadCharacter(loc)
		if err != nil {
			log.Println(loc.Name, err)
			continue
		}

		characters = append(characters, character)
	}

	var character *Character
	character = characters[0]
	characterIndex := 0

	last := time.Now()
	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()

		if win.JustPressed(pixelgl.KeyUp) {
			characterIndex = (characterIndex + len(characters) - 1) % len(characters)
			character = characters[characterIndex]
		}
		if win.JustPressed(pixelgl.KeyDown) {
			characterIndex = (characterIndex + len(characters) + 1) % len(characters)
			character = characters[characterIndex]
		}

		if win.JustPressed(pixelgl.KeyLeft) {
			character.NextAnimation(-1)
		}
		if win.JustPressed(pixelgl.KeyRight) {
			character.NextAnimation(1)
		}

		if win.JustPressed(pixelgl.KeyW) {
			character.NextSkin(-1)
		}
		if win.JustPressed(pixelgl.KeyS) {
			character.NextSkin(1)
		}
		win.Clear(colornames.Lightgray)

		center := win.Bounds().Center()
		center.Y = 50
		character.Update(dt, center)
		character.Draw(win)

		txt.Clear()
		txt.WriteString(character.Description())
		txt.Draw(win, pixel.IM.Moved(pixel.Vec{50, 50}))

		win.Update()
	}
}

func loadFont(size float64) (font.Face, error) {
	font, err := truetype.Parse(goregular.TTF)
	if err != nil {
		return nil, err
	}

	return truetype.NewFace(font, &truetype.Options{
		Size:              size,
		GlyphCacheEntries: 1,
	}), nil
}
