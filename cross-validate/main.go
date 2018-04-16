package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"text/tabwriter"

	"github.com/adinfinit/spine"
	"github.com/adinfinit/spine-examples/animation"
	"github.com/adinfinit/spine-examples/cross-validate/gold"
	"github.com/adinfinit/spine-examples/cross-validate/spinec"
)

func ReadSpineC(loc animation.Location) (gold.Skeleton, error) {
	atlas, err := ioutil.ReadFile(loc.Atlas)
	if err != nil {
		return gold.Skeleton{}, err
	}

	content, err := ioutil.ReadFile(loc.JSON)
	if err != nil {
		return gold.Skeleton{}, err
	}

	var x, y, scale, rotation float32
	scale = (float32)(*rootScale)

	gskeleton, err := spinec.Gold(loc.Dir, string(atlas), string(content), x, y, scale, rotation)
	if err != nil {
		return gold.Skeleton{}, err
	}

	return gskeleton, nil
}

func ReadSpineGo(loc animation.Location) (gold.Skeleton, error) {
	content, err := ioutil.ReadFile(loc.JSON)
	if err != nil {
		return gold.Skeleton{}, err
	}

	gskeleton, err := parseSpineGo(content)
	if err != nil {
		return gold.Skeleton{}, err
	}

	return gskeleton, nil
}

var (
	printFrames     = flag.Bool("frame", false, "print frame info")
	printBones      = flag.Bool("bone", false, "print bone info")
	printBoth       = flag.Bool("both", false, "print both")
	printConstraint = flag.Bool("constraint", false, "print constraint")

	selectAnimation = flag.String("animation-", "", "select animation")
	selectFrame     = flag.Int("frame-", -1, "select frame")
	selectBone      = flag.String("bone-", "", "select bone")

	rootScale = flag.Float64("scale", 1, "scaling factor")
)

func main() {
	flag.Parse()

	log.SetOutput(os.Stderr)
	for _, loc := range animation.LoadList("../animation") {
		fmt.Println()
		fmt.Println(loc.JSON)

		spinec, err := ReadSpineC(loc)
		if err != nil {
			log.Println("failed to read spine-c: ", err)
			continue
		}

		spinego, err := ReadSpineGo(loc)
		if err != nil {
			log.Println("failed to read spine-go: ", err)
			continue
		}

		diff := gold.DiffSkeletons(&spinec, &spinego)

		wf := new(tabwriter.Writer)
		wf.Init(os.Stdout, 4, 8, 4, ' ', 0)
		if len(diff.ResetBone) > 0 {
			fmt.Fprint(wf, "reset:\tC\tGo\n")
			for i, entry := range diff.ResetBone {
				fmt.Fprintf(wf, "%-d\t%v\t%v\n", i, entry[0], entry[1])
			}
		}
		if len(diff.UpdateOrder) > 0 {
			fmt.Fprint(wf, "order:\tC\tGo\n")
			for i, entry := range diff.UpdateOrder {
				fmt.Fprintf(wf, "%-d\t%v\t%v\n", i, entry[0], entry[1])
			}
		}
		wf.Flush()

		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 4, 8, 4, ' ', 0)

		printFrame := func(anim *gold.AnimationDiff, frameIndex int, frame *gold.FrameDiff) {
			var aframe, bframe *gold.Frame
			if *printBoth {
				if anim != nil && frameIndex >= 0 {
					aanim := spinec.FindAnimation(anim.Name)
					banim := spinego.FindAnimation(anim.Name)
					aframe = &aanim.Frame[frameIndex]
					bframe = &banim.Frame[frameIndex]
				} else {
					aframe = &spinec.Setup
					bframe = &spinego.Setup
				}
			}

			fmt.Fprintf(w, "  |\t%.1f\t%v\n", frame.Time, frame.Summary.LocalWorld())
			if *printBones && *printConstraint {
				fmt.Printf("%+v\n", aframe.TransfromConstraints)
				fmt.Printf("%+v\n", bframe.TransfromConstraints)
				fmt.Println()
			}

			if *printBones {
				for boneIndex, bone := range frame.Bones {
					if bone.Name != *selectBone && *selectBone != "" {
						continue
					}
					fmt.Fprintf(w, "\t%v\t%v\n", bone.Name, bone.LocalWorld())
					if *printBoth {
						abone := &aframe.Bones[boneIndex]
						bbone := &bframe.Bones[boneIndex]
						fmt.Fprintf(w, "\t\t%v\n", gold.LocalWorldCompare(abone, bbone))
					}
				}
			}
		}

		fmt.Fprintf(w, "Animation\tTime\tTX\t\tTY\t\tRo\t\tSX\t\tSY\t\tHX\t\tHY\t\tA\t\tB\t\tX\t\tC\t\tD\t\tY\t\n")
		fmt.Fprintf(w, "Setup\t-\t%v\n", diff.Setup.Summary.LocalWorld())
		if *printFrames {
			if "setup" == *selectAnimation || *selectAnimation == "" {
				printFrame(nil, -1, &diff.Setup)
			}
		}

		fmt.Fprintf(w, "Total\t-\t%v\n", diff.Summary.LocalWorld())
		for i := range diff.Animations {
			anim := &diff.Animations[i]
			if anim.Name != *selectAnimation && *selectAnimation != "" {
				continue
			}
			fmt.Fprintf(w, "%v\t-\t%v\n", anim.Name, anim.Summary.LocalWorld())
			if *printFrames {
				for frameIndex, frame := range anim.Frame {
					if frameIndex != *selectFrame && *selectFrame >= 0 {
						continue
					}
					printFrame(anim, frameIndex, &frame)
				}
			}
		}
		w.Flush()
	}
}

func parseSpineGo(content []byte) (gold.Skeleton, error) {
	gskeleton := gold.Skeleton{}
	gskeleton.HasLocal = true
	gskeleton.HasAffineWorld = true

	skeletondata, err := spine.ReadJSON(bytes.NewReader(content))
	if err != nil {
		return gold.Skeleton{}, err
	}

	skeleton := spine.NewSkeleton(skeletondata)

	skeleton.FlipY = true

	skeleton.SetToSetupPose()
	scale := float32(*rootScale)
	skeleton.Local.Scale.Set(scale, scale)
	skeleton.Update()

	gskeleton.Setup = readFrame(0, skeleton)

	for _, bone := range skeleton.ResetBones {
		gskeleton.ResetBone = append(gskeleton.ResetBone, bone.GetName())
	}

	for _, constraint := range skeleton.TransfromConstraints {
		constraintdata := constraint.Data
		gdata := gold.TransfromConstraintData{}

		gdata.Name = constraintdata.Name

		gdata.RotateMix = constraintdata.Mix.Rotate
		gdata.TranslateMix = constraintdata.Mix.Translate
		gdata.ScaleMix = constraintdata.Mix.Scale
		gdata.ShearMix = constraintdata.Mix.Shear

		gdata.OffsetRotation = constraintdata.Offset.Rotate
		gdata.OffsetX = constraintdata.Offset.Translate.X
		gdata.OffsetY = constraintdata.Offset.Translate.Y
		gdata.OffsetScaleX = constraintdata.Offset.Scale.X
		gdata.OffsetScaleY = constraintdata.Offset.Scale.Y
		gdata.OffsetShearY = constraintdata.Offset.Shear.Y
		gdata.Relative = constraintdata.Relative
		gdata.Local = constraintdata.Local

		gskeleton.TransfromConstraints = append(gskeleton.TransfromConstraints, gdata)
	}

	for _, updatable := range skeleton.UpdateOrder {
		switch updater := updatable.(type) {
		case *spine.Bone:
			gskeleton.UpdateOrder = append(gskeleton.UpdateOrder, "B:"+updater.GetName())
		case *spine.TransformConstraint:
			gskeleton.UpdateOrder = append(gskeleton.UpdateOrder, "T:"+updater.GetName())
		case *spine.IKConstraint:
			gskeleton.UpdateOrder = append(gskeleton.UpdateOrder, "I:"+updater.GetName())
		case *spine.PathConstraint:
			gskeleton.UpdateOrder = append(gskeleton.UpdateOrder, "P:"+updater.GetName())
		default:
			panic(updatable)
		}
	}

	for _, animation := range skeleton.Data.Animations {
		ganimation := gold.Animation{}
		ganimation.Name = animation.Name
		ganimation.Duration = animation.Duration

		skeleton.SetToSetupPose()
		skeleton.Update()

		prev := float32(0.0)
		for time := prev; time <= ganimation.Duration; time += gold.StepSize {
			animation.Apply(skeleton, time, true)
			skeleton.Update()
			ganimation.Frame = append(ganimation.Frame, readFrame(time, skeleton))
		}
		gskeleton.Animations = append(gskeleton.Animations, ganimation)
	}

	return gskeleton, nil
}

func readFrame(time float32, skeleton *spine.Skeleton) gold.Frame {
	frame := gold.Frame{}
	frame.Time = time
	for _, bone := range skeleton.Bones {
		gbone := gold.Bone{}
		gbone.Name = bone.Data.Name

		local := bone.Local.Combine(bone.Data.Local)

		gbone.X = local.Translate.X
		gbone.Y = local.Translate.Y
		gbone.Rotation = local.Rotate
		gbone.ScaleX = local.Scale.X
		gbone.ScaleY = local.Scale.Y
		gbone.ShearX = local.Shear.X
		gbone.ShearY = local.Shear.Y

		gbone.A, gbone.B, gbone.WorldX = bone.World.M00, bone.World.M01, bone.World.M02
		gbone.C, gbone.D, gbone.WorldY = bone.World.M10, bone.World.M11, bone.World.M12

		frame.Bones = append(frame.Bones, gbone)
	}

	for _, slot := range skeleton.Slots {
		gslot := gold.Slot{}
		gslot.Name = slot.Data.Name
		if slot.Attachment != nil {
			gslot.Attachment = slot.Attachment.GetName()
		}
		frame.Slots = append(frame.Slots, gslot)
	}

	for _, constraint := range skeleton.TransfromConstraints {
		gconstraint := gold.TransfromConstraint{}

		gconstraint.Name = constraint.Data.Name

		gconstraint.RotateMix = constraint.Mix.Rotate
		gconstraint.TranslateMix = constraint.Mix.Translate
		gconstraint.ScaleMix = constraint.Mix.Scale
		gconstraint.ShearMix = constraint.Mix.Shear

		frame.TransfromConstraints = append(frame.TransfromConstraints, gconstraint)
	}

	return frame
}
