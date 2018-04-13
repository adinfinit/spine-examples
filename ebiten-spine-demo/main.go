package main

import (
	_ "image/png"
	"log"

	"github.com/adinfinit/spine"
	"github.com/adinfinit/spine-examples/animation"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

const (
	// Settings
	screenWidth  = 1024
	screenHeight = 512
)

var (
	characters     []*Character
	character      *Character
	characterIndex int
)

func main() {
skip:
	for _, loc := range animation.LoadList("../animation") {
		character, err := LoadCharacter(loc)
		if err != nil {
			log.Println(loc.Name, err)
			continue
		}

		for _, skin := range character.Skeleton.Data.Skins {
			for _, att := range skin.Attachments {
				if _, ismesh := att.(*spine.MeshAttachment); ismesh {
					log.Println(loc.Name, "Unsupported")
					continue skip
				}
			}
		}

		characters = append(characters, character)
	}

	character = characters[0]
	if err := ebiten.Run(update, screenWidth, screenHeight, 1, "Spine (Ebiten Demo)"); err != nil {
		panic(err)
	}
}

var wasPressed [256]bool

func justpressed(key ebiten.Key) bool {
	ispressed := ebiten.IsKeyPressed(key)
	if !wasPressed[key] && ispressed {
		wasPressed[key] = true
		return true
	}
	wasPressed[key] = ispressed
	return false
}

func update(screen *ebiten.Image) error {
	if justpressed(ebiten.KeyUp) {
		characterIndex = (characterIndex + len(characters) - 1) % len(characters)
		character = characters[characterIndex]
	}
	if justpressed(ebiten.KeyDown) {
		characterIndex = (characterIndex + len(characters) + 1) % len(characters)
		character = characters[characterIndex]
	}

	if justpressed(ebiten.KeyLeft) {
		character.NextAnimation(-1)
	}
	if justpressed(ebiten.KeyRight) {
		character.NextAnimation(1)
	}

	if justpressed(ebiten.KeyW) {
		character.NextSkin(-1)
	}
	if justpressed(ebiten.KeyS) {
		character.NextSkin(1)
	}

	character.Update(1/float64(ebiten.FPS), screenWidth/2, screenHeight-50)

	if ebiten.IsRunningSlowly() {
		return nil
	}

	screen.Clear()

	character.Draw(screen)

	ebitenutil.DebugPrint(screen, character.Description())

	return nil
}
