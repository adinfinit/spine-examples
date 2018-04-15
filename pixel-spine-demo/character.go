package main

import (
	"fmt"
	"image"
	"os"
	"path/filepath"

	"github.com/adinfinit/spine"
	"github.com/adinfinit/spine-examples/animation"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"

	_ "image/png"
)

type Character struct {
	Time  float64
	Play  bool
	Speed float64

	// TODO: replace this with atlas
	ImagesPath string
	Images     map[string]*pixel.PictureData

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
	char.Images = make(map[string]*pixel.PictureData)

	char.Play = true
	char.DebugBones = false
	char.DebugCenter = false

	char.Speed = 1
	char.Skeleton = spine.NewSkeleton(data)
	char.Skeleton.Skin = char.Skeleton.Data.DefaultSkin
	char.Animation = char.Skeleton.Data.Animations[0]

	char.AnimationIndex = 0
	char.SkinIndex = 0

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
func (char *Character) Update(dt float64, center pixel.Vec) {
	if char.Play {
		char.Time += dt * char.Speed
	}

	char.Animation.Apply(char.Skeleton, float32(char.Time), true)

	char.Skeleton.Local.Translate.Set(float32(center.X), float32(center.Y))
	char.Skeleton.Update()
}

func (char *Character) GetImage(attachment, path string) *pixel.PictureData {
	if path != "" {
		attachment = path
	}
	if pd, ok := char.Images[attachment]; ok {
		return pd
	}
	fmt.Println("Loading " + attachment)

	fallback := func() *pixel.PictureData {
		fmt.Println("missing: ", attachment)

		m := image.NewRGBA(image.Rect(0, 0, 10, 10))
		for i := range m.Pix {
			m.Pix[i] = 0x80
		}
		pd := pixel.PictureDataFromImage(m)
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
	pd := pixel.PictureDataFromImage(m)

	char.Images[attachment] = pd

	return pd
}

func (char *Character) Draw(win pixel.Target) {
	for _, slot := range char.Skeleton.Order {
		bone := slot.Bone
		switch attachment := slot.Attachment.(type) {
		case nil:
		case *spine.RegionAttachment:
			final := bone.World.Mul(attachment.Local.Affine())
			xform := pixel.Matrix(final.Col64())

			pd := char.GetImage(attachment.Name, attachment.Path)
			sprite := pixel.NewSprite(pd, pd.Rect)
			sprite.DrawColorMask(win, xform, slot.Color)

		case *spine.MeshAttachment:
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
				}
			}

			var col pixel.RGBA
			col.R, col.G, col.B, col.A = slot.Color.RGBA64()
			for i := range *tridata {
				tri := &(*tridata)[i]
				tri.Color = col
				tri.Intensity = 1
			}

			batch := pixel.NewBatch(tridata, pd)
			batch.Draw(win)
		default:
			panic(fmt.Sprintf("unknown attachment %v", attachment))
		}
	}

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

}
