package spinec

import (
	"errors"
	"math"
	"unsafe"

	"github.com/adinfinit/spine-examples/cross-validate/gold"
)

// #cgo CFLAGS: -Ispine-c/include
//
// #include <stdlib.h>
// #include <spine-c/include/spine/spine.h>
//
// #include <spine-c/src/spine/Animation.c>
// #include <spine-c/src/spine/AnimationState.c>
// #include <spine-c/src/spine/AnimationStateData.c>
// #include <spine-c/src/spine/Array.c>
// #include <spine-c/src/spine/Atlas.c>
// #include <spine-c/src/spine/AtlasAttachmentLoader.c>
// #include <spine-c/src/spine/Attachment.c>
// #include <spine-c/src/spine/AttachmentLoader.c>
// #include <spine-c/src/spine/Bone.c>
// #include <spine-c/src/spine/BoneData.c>
// #include <spine-c/src/spine/BoundingBoxAttachment.c>
// #include <spine-c/src/spine/ClippingAttachment.c>
// #include <spine-c/src/spine/Color.c>
// #include <spine-c/src/spine/Event.c>
// #include <spine-c/src/spine/EventData.c>
// #include <spine-c/src/spine/IkConstraint.c>
// #include <spine-c/src/spine/IkConstraintData.c>
// #include <spine-c/src/spine/Json.c>
// #include <spine-c/src/spine/Json.h>
// #include <spine-c/src/spine/MeshAttachment.c>
// #include <spine-c/src/spine/PathAttachment.c>
// #include <spine-c/src/spine/PathConstraint.c>
// #include <spine-c/src/spine/PathConstraintData.c>
// #include <spine-c/src/spine/PointAttachment.c>
// #include <spine-c/src/spine/RegionAttachment.c>
// #include <spine-c/src/spine/Skeleton.c>
// #include <spine-c/src/spine/SkeletonBounds.c>
// #include <spine-c/src/spine/SkeletonClipping.c>
// #include <spine-c/src/spine/SkeletonData.c>
// #include <spine-c/src/spine/SkeletonJson.c>
// #include <spine-c/src/spine/Skin.c>
// #include <spine-c/src/spine/Slot.c>
// #include <spine-c/src/spine/SlotData.c>
// #include <spine-c/src/spine/TransformConstraint.c>
// #include <spine-c/src/spine/TransformConstraintData.c>
// #include <spine-c/src/spine/Triangulator.c>
// #include <spine-c/src/spine/VertexAttachment.c>
// #include <spine-c/src/spine/VertexEffect.c>
// #include <spine-c/src/spine/extension.c>
// #include <spine-c/src/spine/kvec.h>
//
// // alternative definition for _spUpdate, so that we can acces "type" field
//    typedef struct {
//       _spUpdateType updateType;
//       void* object;
//    } _spUpdate2;
import "C"

func Gold(dir string, atlasstr string, data string) (gold.Skeleton, error) {
	gskeleton := gold.Skeleton{}
	gskeleton.HasLocal = true
	gskeleton.HasAffineWorld = true
	gskeleton.HasAppliedWorld = true

	atlasdata := C.CString(atlasstr)
	defer C.free(unsafe.Pointer(atlasdata))

	atlasdir := C.CString(dir)
	defer C.free(unsafe.Pointer(atlasdir))

	atlas := C.spAtlas_create(atlasdata, C.int(len(atlasstr)), atlasdir, nil)
	defer C.spAtlas_dispose(atlas)

	json := C.spSkeletonJson_create(atlas)
	if json == nil {
		return gskeleton, errors.New("unable to create skeleton json")
	}
	defer C.spSkeletonJson_dispose(json)

	if json.error != nil {
		errmsg := C.GoString(json.error)
		return gskeleton, errors.New("unable to create skeleton json: " + errmsg)
	}

	jsondata := C.CString(data)
	defer C.free(unsafe.Pointer(jsondata))

	skeletondata := C.spSkeletonJson_readSkeletonData(json, jsondata)
	defer C.spSkeletonData_dispose(skeletondata)

	skeleton := C.spSkeleton_create(skeletondata)
	defer C.spSkeleton_dispose(skeleton)

	skeleton.flipY = 1

	C.spSkeleton_setToSetupPose(skeleton)
	C.spSkeleton_updateWorldTransform(skeleton)
	gskeleton.Setup = readFrame(0, skeleton)

	internal := (*C._spSkeleton)(unsafe.Pointer(skeleton))
	for i := 0; i < int(internal.updateCacheCount); i++ {
		update := (*C._spUpdate2)(unsafe.Pointer((uintptr(unsafe.Pointer(internal.updateCache)) + uintptr(i)*unsafe.Sizeof(C._spUpdate2{}))))
		switch update.updateType {
		case C.SP_UPDATE_BONE:
			bone := (*C.spBone)(update.object)
			gskeleton.UpdateOrder = append(gskeleton.UpdateOrder, "B:"+C.GoString(bone.data.name))
		case C.SP_UPDATE_IK_CONSTRAINT:
			constraint := (*C.spIkConstraint)(update.object)
			gskeleton.UpdateOrder = append(gskeleton.UpdateOrder, "I:"+C.GoString(constraint.data.name))
		case C.SP_UPDATE_PATH_CONSTRAINT:
			path := (*C.spPathConstraint)(update.object)
			gskeleton.UpdateOrder = append(gskeleton.UpdateOrder, "P:"+C.GoString(path.data.name))
		case C.SP_UPDATE_TRANSFORM_CONSTRAINT:
			transform := (*C.spTransformConstraint)(update.object)
			gskeleton.UpdateOrder = append(gskeleton.UpdateOrder, "T:"+C.GoString(transform.data.name))
		}
	}
	for i := 0; i < int(internal.updateCacheResetCount); i++ {
		bone := *(**C.spBone)(unsafe.Pointer((uintptr(unsafe.Pointer(internal.updateCacheReset)) + uintptr(i)*unsafe.Sizeof((*C.spBone)(nil)))))
		gskeleton.ResetBone = append(gskeleton.ResetBone, C.GoString(bone.data.name))
	}

	for i := 0; i < int(skeleton.transformConstraintsCount); i++ {
		constraint := *(**C.spTransformConstraint)(unsafe.Pointer((uintptr(unsafe.Pointer(skeleton.transformConstraints)) + uintptr(i)*unsafe.Sizeof((*C.spTransformConstraint)(nil)))))
		constraintdata := constraint.data
		gdata := gold.TransfromConstraintData{}

		gdata.Name = C.GoString(constraintdata.name)

		gdata.RotateMix = float32(constraintdata.rotateMix)
		gdata.TranslateMix = float32(constraintdata.translateMix)
		gdata.ScaleMix = float32(constraintdata.scaleMix)
		gdata.ShearMix = float32(constraintdata.shearMix)

		gdata.OffsetRotation = float32(constraintdata.offsetRotation) * math.Pi / 180
		gdata.OffsetX = float32(constraintdata.offsetX)
		gdata.OffsetY = float32(constraintdata.offsetY)
		gdata.OffsetScaleX = float32(constraintdata.offsetScaleX)
		gdata.OffsetScaleY = float32(constraintdata.offsetScaleY)
		gdata.OffsetShearY = float32(constraintdata.offsetShearY) * math.Pi / 180
		gdata.Relative = constraintdata.relative == 1
		gdata.Local = constraintdata.local == 1

		gskeleton.TransfromConstraints = append(gskeleton.TransfromConstraints, gdata)
	}

	for i := 0; i < int(skeletondata.animationsCount); i++ {
		animation := *(**C.spAnimation)(unsafe.Pointer((uintptr(unsafe.Pointer(skeletondata.animations)) + uintptr(i)*unsafe.Sizeof((*C.spAnimation)(nil)))))
		ganimation := gold.Animation{}
		ganimation.Name = C.GoString(animation.name)
		ganimation.Duration = float32(animation.duration)

		C.spSkeleton_setToSetupPose(skeleton)
		C.spSkeleton_updateWorldTransform(skeleton)

		prev := float32(0.0)
		for time := prev; time <= ganimation.Duration; time += gold.StepSize {
			C.spAnimation_apply(
				animation, skeleton,
				C.float(prev), C.float(time), 1,
				nil, nil, 1.0,
				C.SP_MIX_POSE_CURRENT, C.SP_MIX_DIRECTION_OUT)
			C.spSkeleton_updateWorldTransform(skeleton)
			prev = time

			ganimation.Frame = append(ganimation.Frame, readFrame(time, skeleton))
		}

		gskeleton.Animations = append(gskeleton.Animations, ganimation)
	}

	return gskeleton, nil
}

func readFrame(time float32, skeleton *C.spSkeleton) gold.Frame {
	frame := gold.Frame{}
	frame.Time = time

	for i := 0; i < int(skeleton.bonesCount); i++ {
		gbone := gold.Bone{}
		bone := *(**C.spBone)(unsafe.Pointer((uintptr(unsafe.Pointer(skeleton.bones)) + uintptr(i)*unsafe.Sizeof((*C.spBone)(nil)))))

		gbone.Name = C.GoString(bone.data.name)
		gbone.X = float32(bone.x)
		gbone.Y = float32(bone.y)
		gbone.Rotation = float32(bone.rotation) * math.Pi / 180.0
		gbone.ScaleX = float32(bone.scaleX)
		gbone.ScaleY = float32(bone.scaleY)
		gbone.ShearX = float32(bone.shearX) * math.Pi / 180.0
		gbone.ShearY = float32(bone.shearY) * math.Pi / 180.0

		// gbone.AX = float32(bone.ax)
		// gbone.AY = float32(bone.ay)
		// gbone.ARotation = float32(bone.arotation) * math.Pi / 180.0
		// gbone.AScaleX = float32(bone.ascaleX)
		// gbone.AScaleY = float32(bone.ascaleY)
		// gbone.AShearX = float32(bone.ashearX) * math.Pi / 180.0
		// gbone.AShearY = float32(bone.ashearY) * math.Pi / 180.0

		gbone.A = float32(bone.a)
		gbone.B = float32(bone.b)
		gbone.WorldX = float32(bone.worldX)
		gbone.C = float32(bone.c)
		gbone.D = float32(bone.d)
		gbone.WorldY = float32(bone.worldY)

		frame.Bones = append(frame.Bones, gbone)
	}

	for i := 0; i < int(skeleton.slotsCount); i++ {
		slot := *(**C.spSlot)(unsafe.Pointer((uintptr(unsafe.Pointer(skeleton.slots)) + uintptr(i)*unsafe.Sizeof((*C.spSlot)(nil)))))
		gslot := gold.Slot{}

		gslot.Name = C.GoString(slot.data.name)
		if slot.attachment != nil {
			gslot.Attachment = C.GoString(slot.attachment.name)
		}
		gslot.AttachmentVertices = make([]float32, slot.attachmentVerticesCount)
		for k := range gslot.AttachmentVertices {
			value := *(*float32)(unsafe.Pointer((uintptr(unsafe.Pointer(slot.attachmentVertices)) + uintptr(k)*unsafe.Sizeof(C.float(0)))))
			gslot.AttachmentVertices[k] = value
		}
		frame.Slots = append(frame.Slots, gslot)
	}

	for i := 0; i < int(skeleton.transformConstraintsCount); i++ {
		constraint := *(**C.spTransformConstraint)(unsafe.Pointer((uintptr(unsafe.Pointer(skeleton.transformConstraints)) + uintptr(i)*unsafe.Sizeof((*C.spTransformConstraint)(nil)))))
		gconstraint := gold.TransfromConstraint{}

		gconstraint.Name = C.GoString(constraint.data.name)

		gconstraint.RotateMix = float32(constraint.rotateMix)
		gconstraint.TranslateMix = float32(constraint.translateMix)
		gconstraint.ScaleMix = float32(constraint.scaleMix)
		gconstraint.ShearMix = float32(constraint.shearMix)

		frame.TransfromConstraints = append(frame.TransfromConstraints, gconstraint)
	}

	return frame
}
