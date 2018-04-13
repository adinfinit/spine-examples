package gold

import (
	"fmt"
	"math"
)

const StepSize = 0.1

type Skeleton struct {
	Setup      Frame
	Animations []Animation

	ResetBone   []string
	UpdateOrder []string

	TransfromConstraints []TransfromConstraintData

	HasLocal        bool
	HasAppliedWorld bool
	HasAffineWorld  bool
}

func (skeleton *Skeleton) FindAnimation(name string) *Animation {
	for i := range skeleton.Animations {
		if skeleton.Animations[i].Name == name {
			return &skeleton.Animations[i]
		}
	}
	return nil
}

type Animation struct {
	Name     string
	Duration float32
	Frame    []Frame
}

type Frame struct {
	Time  float32
	Bones []Bone
	Slots []Slot

	TransfromConstraints []TransfromConstraint
}

type Bone struct {
	Name string

	X, Y, Rotation, ScaleX, ScaleY, ShearX, ShearY float32
	//AX, AY, ARotation, AScaleX, AScaleY, AShearX, AShearY float32

	A, B, WorldX float32
	C, D, WorldY float32
}

func LocalWorldCompare(a, b *Bone) string {
	r := ""
	r += fmt.Sprintf("%v\t%v\t", compare(a.X, b.X)...)
	r += fmt.Sprintf("%v\t%v\t", compare(a.Y, b.Y)...)
	r += fmt.Sprintf("%v\t%v\t", compare(a.Rotation, b.Rotation)...)
	r += fmt.Sprintf("%v\t%v\t", compare(a.ScaleX, b.ScaleX)...)
	r += fmt.Sprintf("%v\t%v\t", compare(a.ScaleY, b.ScaleY)...)
	r += fmt.Sprintf("%v\t%v\t", compare(a.ShearX, b.ShearX)...)
	r += fmt.Sprintf("%v\t%v\t", compare(a.ShearY, b.ShearY)...)
	r += fmt.Sprintf("%v\t%v\t", compare(a.A, b.A)...)
	r += fmt.Sprintf("%v\t%v\t", compare(a.B, b.B)...)
	r += fmt.Sprintf("%v\t%v\t", compare(a.WorldX, b.WorldX)...)
	r += fmt.Sprintf("%v\t%v\t", compare(a.C, b.C)...)
	r += fmt.Sprintf("%v\t%v\t", compare(a.D, b.D)...)
	r += fmt.Sprintf("%v\t%v", compare(a.WorldY, b.WorldY)...)
	return r
}

type Slot struct {
	Name string

	Attachment         string
	AttachmentVertices []float32
}

type TransfromConstraint struct {
	Name string

	RotateMix    float32
	TranslateMix float32
	ScaleMix     float32
	ShearMix     float32
}

type TransfromConstraintData struct {
	Name string

	RotateMix    float32
	TranslateMix float32
	ScaleMix     float32
	ShearMix     float32

	OffsetRotation float32
	OffsetX        float32
	OffsetY        float32
	OffsetScaleX   float32
	OffsetScaleY   float32
	OffsetShearY   float32
	Relative       bool
	Local          bool
}

type SkeletonDiff struct {
	ResetBone   [][2]string
	UpdateOrder [][2]string

	Setup      FrameDiff
	Summary    DiffSummary
	Animations []AnimationDiff
}

type AnimationDiff struct {
	Name    string
	Missing int
	Summary DiffSummary
	Frame   []FrameDiff
}

type FrameDiff struct {
	Time    float32
	Missing int
	Summary DiffSummary
	Bones   []Diff
}

type DiffSummary struct {
	Avg   Diff
	Min   Diff
	Max   Diff
	Count int
}

func (s *DiffSummary) LocalWorld() string {
	avg := s.Avg
	avg.Apply(&avg, func(a, b float32) float32 {
		if s.Count == 0 {
			return 0
		}
		return a / float32(s.Count)
	})

	// header TX
	r := ""
	r += fmt.Sprintf("%v\t%v\t", zero(s.Min.X), zero(s.Max.X))
	r += fmt.Sprintf("%v\t%v\t", zero(s.Min.Y), zero(s.Max.Y))
	r += fmt.Sprintf("%v\t%v\t", zero(s.Min.Rotation), zero(s.Max.Rotation))
	r += fmt.Sprintf("%v\t%v\t", zero(s.Min.ScaleX), zero(s.Max.ScaleX))
	r += fmt.Sprintf("%v\t%v\t", zero(s.Min.ScaleY), zero(s.Max.ScaleY))
	r += fmt.Sprintf("%v\t%v\t", zero(s.Min.ShearX), zero(s.Max.ShearX))
	r += fmt.Sprintf("%v\t%v\t", zero(s.Min.ShearY), zero(s.Max.ShearY))
	r += fmt.Sprintf("%v\t%v\t", zero(s.Min.A), zero(s.Max.A))
	r += fmt.Sprintf("%v\t%v\t", zero(s.Min.B), zero(s.Max.B))
	r += fmt.Sprintf("%v\t%v\t", zero(s.Min.WorldX), zero(s.Max.WorldX))
	r += fmt.Sprintf("%v\t%v\t", zero(s.Min.C), zero(s.Max.C))
	r += fmt.Sprintf("%v\t%v\t", zero(s.Min.D), zero(s.Max.D))
	r += fmt.Sprintf("%v\t%v", zero(s.Min.WorldY), zero(s.Max.WorldY))
	return r
}

func zero(v float32) string {
	eps := v
	if abs(eps) < 0.001 {
		return "."
	}
	return fmt.Sprintf("%.2f", v)
}

func compare(a, b float32) []interface{} {
	if abs(a-b) < 0.001 {
		return []interface{}{"", ""}
	}
	return []interface{}{fmt.Sprintf("%.2f", a), fmt.Sprintf("%.2f", b)}
}

func any(v float32) string {
	return fmt.Sprintf("%.2f", v)
}

type Diff Bone

func (bone *Diff) LocalWorld() string {
	r := ""
	r += fmt.Sprintf("%v\t\t", zero(bone.X))
	r += fmt.Sprintf("%v\t\t", zero(bone.Y))
	r += fmt.Sprintf("%v\t\t", zero(bone.Rotation))
	r += fmt.Sprintf("%v\t\t", zero(bone.ScaleX))
	r += fmt.Sprintf("%v\t\t", zero(bone.ScaleY))
	r += fmt.Sprintf("%v\t\t", zero(bone.ShearX))
	r += fmt.Sprintf("%v\t\t", zero(bone.ShearY))
	r += fmt.Sprintf("%v\t\t", zero(bone.A))
	r += fmt.Sprintf("%v\t\t", zero(bone.B))
	r += fmt.Sprintf("%v\t\t", zero(bone.WorldX))
	r += fmt.Sprintf("%v\t\t", zero(bone.C))
	r += fmt.Sprintf("%v\t\t", zero(bone.D))
	r += fmt.Sprintf("%v\t", zero(bone.WorldY))
	return r
}

func DiffSkeletons(a, b *Skeleton) SkeletonDiff {
	skeldiff := SkeletonDiff{}
	skeldiff.ResetBone = diffStrings(a.ResetBone, b.ResetBone)
	skeldiff.UpdateOrder = diffStrings(a.UpdateOrder, b.UpdateOrder)
	skeldiff.Setup = DiffFrames(&a.Setup, &b.Setup)
	for i := range a.Animations {
		aanim := &a.Animations[i]
		banim := b.FindAnimation(aanim.Name)
		if banim == nil {
			skeldiff.Animations = append(skeldiff.Animations, AnimationDiff{
				Name:    aanim.Name,
				Missing: len(aanim.Frame),
			})
			continue
		}

		diff := DiffAnimations(aanim, banim)
		skeldiff.Summary.Include(&diff.Summary)
		skeldiff.Animations = append(skeldiff.Animations, diff)
	}
	return skeldiff
}

func diffStrings(a, b []string) [][2]string {
	n := len(a)
	if n < len(b) {
		n = len(b)
	}

	xs := make([][2]string, n)
	diff := false
	for i := 0; i < n; i++ {
		xs[i] = [2]string{"???", "???"}
		if i < len(a) {
			xs[i][0] = a[i]
		}
		if i < len(b) {
			xs[i][1] = b[i]
		}
		if xs[i][0] != xs[i][1] {
			diff = true
		}
	}

	if !diff {
		return nil
	}
	return xs
}

func DiffAnimations(a, b *Animation) AnimationDiff {
	if a.Name != b.Name {
		panic("name mismatch")
	}
	n := len(a.Frame)
	if n > len(b.Frame) {
		n = len(b.Frame)
	}

	animdiff := AnimationDiff{}
	animdiff.Name = a.Name
	animdiff.Missing = len(a.Frame) - n + len(b.Frame) - n
	for i := 0; i < n; i++ {
		diff := DiffFrames(&a.Frame[i], &b.Frame[i])
		animdiff.Summary.Include(&diff.Summary)
		animdiff.Frame = append(animdiff.Frame, diff)
	}
	return animdiff
}

func DiffFrames(a, b *Frame) FrameDiff {
	n := len(a.Bones)
	if n > len(b.Bones) {
		n = len(b.Bones)
	}

	framediff := FrameDiff{}
	framediff.Time = a.Time
	framediff.Missing = len(a.Bones) - n + len(b.Bones) - n
	for i := 0; i < n; i++ {
		diff := DiffBones(&a.Bones[i], &b.Bones[i])
		framediff.Summary.Add(&diff)
		framediff.Bones = append(framediff.Bones, diff)
	}
	return framediff
}

func DiffBones(a, b *Bone) Diff {
	if a.Name != b.Name {
		panic("name mismatch")
	}
	r := Diff{}
	r.Name = a.Name
	r.X = diff(a.X, b.X)
	r.Y = diff(a.Y, b.Y)
	r.Rotation = diff(a.Rotation, b.Rotation)
	r.ScaleX = diff(a.ScaleX, b.ScaleX)
	r.ScaleY = diff(a.ScaleY, b.ScaleY)
	r.ShearX = diff(a.ShearX, b.ShearX)
	r.ShearY = diff(a.ShearY, b.ShearY)
	r.A = diff(a.A, b.A)
	r.B = diff(a.B, b.B)
	r.WorldX = diff(a.WorldX, b.WorldX)
	r.C = diff(a.C, b.C)
	r.D = diff(a.D, b.D)
	r.WorldY = diff(a.WorldY, b.WorldY)
	return r
}

func (a *Diff) Apply(b *Diff, op func(a, b float32) float32) {
	a.X = op(a.X, b.X)
	a.Y = op(a.Y, b.Y)
	a.Rotation = op(a.Rotation, b.Rotation)
	a.ScaleX = op(a.ScaleX, b.ScaleX)
	a.ScaleY = op(a.ScaleY, b.ScaleY)
	a.ShearX = op(a.ShearX, b.ShearX)
	a.ShearY = op(a.ShearY, b.ShearY)
	a.A = op(a.A, b.A)
	a.B = op(a.B, b.B)
	a.WorldX = op(a.WorldX, b.WorldX)
	a.C = op(a.C, b.C)
	a.D = op(a.D, b.D)
	a.WorldY = op(a.WorldY, b.WorldY)
}

func (s *DiffSummary) Include(d *DiffSummary) {
	s.Count += 1
	if s.Count == 1 {
		s.Min = d.Min
		s.Avg = d.Avg
		s.Max = d.Max
		return
	}
	s.Min.Apply(&d.Min, min)

	t := d.Avg
	t.Apply(&t, func(a, b float32) float32 {
		if d.Count == 0 {
			return 0
		}
		return a / float32(d.Count)
	})

	s.Avg.Apply(&t, add)
	s.Max.Apply(&d.Max, max)
}

func (s *DiffSummary) Add(d *Diff) {
	s.Count += 1
	if s.Count == 1 {
		s.Min = *d
		s.Avg = *d
		s.Max = *d
		return
	}
	s.Min.Apply(d, min)
	s.Avg.Apply(d, add)
	s.Max.Apply(d, max)
}

func abs(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

func min(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func add(a, b float32) float32 {
	return a + b
}

func max(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func diff(a, b float32) float32 {
	return a - b
}

func diffAngle(a, b float32) float32 {
	v := a - b
	for v >= math.Pi {
		v -= 2 * math.Pi
	}
	for v <= -math.Pi {
		v += 2 * math.Pi
	}
	return v
}

func diffRel(a, b float32) float32 {
	eps := abs(b - a)
	if eps < 0.001 {
		return 1
	}

	if a == 0 {
		return abs(b)
	} else if b == 0 {
		return abs(a)
	}
	x := float32(1.0)
	if a < 0 {
		x *= -1
		a = -a
	}
	if b < 0 {
		x *= -1
		b = -b
	}
	if a > b {
		return x * a / b
	} else {
		return x * b / a
	}
}
