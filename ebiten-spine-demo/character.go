package main

import (
	"fmt"
	"image"
	"os"
	"path/filepath"

	"github.com/adinfinit/spine"
	"github.com/adinfinit/spine-examples/animation"

	"github.com/hajimehoshi/ebiten"

	_ "image/png"
)

type Character struct {
	Time  float64
	Play  bool
	Speed float64

	// TODO: replace this with atlas
	ImagesPath string
	Images     map[string]*ebiten.Image

	Skeleton  *spine.Skeleton
	Animation *spine.Animation

	SkinIndex      int
	AnimationIndex int

	DebugCenter bool
	DebugBones  bool
}

func LoadCharacter(loc animation.Location) (*Character, error) {
	rd, err := os.Open(loc.JSON)
	if err != nil {
		return nil, err
	}

	data, err := spine.ReadJSON(rd)
	if err != nil {
		return nil, err
	}
	data.Name = loc.Name

	char := &Character{}

	char.ImagesPath = loc.Images
	char.Images = make(map[string]*ebiten.Image)

	char.Play = true
	char.DebugBones = false
	char.DebugCenter = false

	char.Speed = 1
	char.Skeleton = spine.NewSkeleton(data)
	char.Skeleton.Skin = char.Skeleton.Data.DefaultSkin
	char.Animation = char.Skeleton.Data.Animations[0]

	char.AnimationIndex = 0
	char.SkinIndex = 0

	char.Skeleton.FlipY = true

	char.Skeleton.UpdateAttachments()
	char.Skeleton.Update()

	return char, nil
}

func (char *Character) Description() string {
	return char.Skeleton.Data.Name + " > " + char.Skeleton.Skin.Name + " > " + char.Animation.Name
}

func (char *Character) NextAnimation(offset int) {
	char.AnimationIndex += offset
	for char.AnimationIndex < 0 {
		char.AnimationIndex += len(char.Skeleton.Data.Animations)
	}
	char.AnimationIndex = char.AnimationIndex % len(char.Skeleton.Data.Animations)
	char.Animation = char.Skeleton.Data.Animations[char.AnimationIndex]
	char.Skeleton.Update()
}

func (char *Character) NextSkin(offset int) {
	char.SkinIndex += offset
	for char.SkinIndex < 0 {
		char.SkinIndex += len(char.Skeleton.Data.Skins)
	}
	char.SkinIndex = char.SkinIndex % len(char.Skeleton.Data.Skins)
	char.Skeleton.Skin = char.Skeleton.Data.Skins[char.SkinIndex]
	char.Skeleton.Update()
	char.Skeleton.UpdateAttachments()
}
func (char *Character) Update(dt float64, x, y float64) {
	if char.Play {
		char.Time += dt * char.Speed
	}

	char.Animation.Apply(char.Skeleton, float32(char.Time), true)
	char.Skeleton.Local.Translate.Set(float32(x), float32(y))
	char.Skeleton.Update()
}

func (char *Character) GetImage(attachment, path string) *ebiten.Image {
	if path != "" {
		attachment = path
	}
	if pd, ok := char.Images[attachment]; ok {
		return pd
	}
	fmt.Println("Loading " + attachment)

	fallback := func() *ebiten.Image {
		fmt.Println("missing: ", attachment)

		m := image.NewRGBA(image.Rect(0, 0, 10, 10))
		for i := range m.Pix {
			m.Pix[i] = 0x80
		}

		pd, _ := ebiten.NewImageFromImage(m, ebiten.FilterDefault)
		char.Images[attachment] = pd
		return pd
	}

	fullpath := filepath.Join(char.ImagesPath, attachment+".png")
	file, err := os.Open(fullpath)
	if err != nil {
		return fallback()
	}

	m, _, err := image.Decode(file)
	if err != nil {
		return fallback()
	}
	pd, _ := ebiten.NewImageFromImage(m, ebiten.FilterDefault)

	char.Images[attachment] = pd

	return pd
}

func (char *Character) Draw(target *ebiten.Image) {
	for _, slot := range char.Skeleton.Order {
		bone := slot.Bone
		switch attachment := slot.Attachment.(type) {
		case nil:
		case *spine.RegionAttachment:
			local := attachment.Local.Affine()
			final := bone.World.Mul(local)

			var geom ebiten.GeoM
			geom.SetElement(0, 0, float64(final.M00))
			geom.SetElement(0, 1, float64(final.M01))
			geom.SetElement(0, 2, float64(final.M02))
			geom.SetElement(1, 0, float64(final.M10))
			geom.SetElement(1, 1, float64(final.M11))
			geom.SetElement(1, 2, float64(final.M12))

			m := char.GetImage(attachment.Name, attachment.Path)
			box := m.Bounds()

			var flipped ebiten.GeoM
			flipped.Translate(-float64(box.Dx())*0.5, -float64(box.Dy())*0.5)
			flipped.Scale(1, -1)
			flipped.Concat(geom)

			target.DrawImage(m, &ebiten.DrawImageOptions{
				SourceRect: &box,
				GeoM:       flipped,
				ColorM:     ebiten.ScaleColor(attachment.Color.Float64()),
			})

		case *spine.MeshAttachment:
			/*
				pd := char.GetImage(attachment.Name, attachment.Path)
				size := pd.Bounds().Size()

				worldPosition := attachment.CalculateWorldVertices(char.Skeleton, slot)
				tridata := pixel.MakeTrianglesData(len(attachment.Triangles) * 3)
				for base, tri := range attachment.Triangles {
					for k, index := range tri {
						tri := &(*tridata)[base*3+k]
						tri.Position = worldPosition[index].V()
						uv := attachment.UV[index]
						uv.Y = 1 - uv.Y
						tri.Picture = uv.V()
						tri.Picture = tri.Picture.ScaledXY(size)
						tri.Intensity = 1
					}
				}

				batch := pixel.NewBatch(tridata, pd)
				batch.SetColorMask(attachment.Color)
				batch.Draw(win)
			*/
		default:
			panic(fmt.Sprintf("unknown attachment %v", attachment))
		}
	}
	/*
		imd := imdraw.New(nil)
		defer imd.Draw(win)

		if char.DebugBones {
			for _, bone := range char.Skeleton.Bones {
				h := float64(bone.Data.Length)
				if h < 10 {
					h = 10
				}

				imd.SetMatrix(pixel.Matrix(bone.World.Col64()))
				imd.Color = bone.Data.Color.WithAlpha(0.5)

				w := h * 0.1
				imd.Push(pixel.V(h, 0))
				imd.Push(pixel.V(w+w, -w))
				imd.Push(pixel.V(w+w, w))
				imd.Polygon(0)

				if bone.Parent != nil {
					imd.SetMatrix(pixel.IM)
					a := pixel.Vec(bone.World.Translation().V())
					b := pixel.Vec(bone.Parent.World.Transform(spine.Vector{bone.Parent.Data.Length, 0}).V())
					imd.Push(a)
					imd.Push(b)
					imd.Line(1)
				}
			}
		}

		if char.DebugCenter {
			imd.SetMatrix(pixel.Matrix(char.Skeleton.World().Col64()))
			imd.Color = pixel.RGB(0, 0, 1)
			imd.Push(pixel.Vec{})
			imd.Circle(20, 2)
		}
	*/
}
