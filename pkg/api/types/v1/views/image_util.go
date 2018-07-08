//
// Last.Backend LLC CONFIDENTIAL
// __________________
//
// [2014] - [2018] Last.Backend LLC
// All Rights Reserved.
//
// NOTICE:  All information contained herein is, and remains
// the property of Last.Backend LLC and its suppliers,
// if any.  The intellectual and technical concepts contained
// herein are proprietary to Last.Backend LLC
// and its suppliers and may be covered by Russian Federation and Foreign Patents,
// patents in process, and are protected by trade secret or copyright law.
// Dissemination of this information or reproduction of this material
// is strictly forbidden unless prior written permission is obtained
// from Last.Backend LLC.
//

package views

import (
	"encoding/json"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"unsafe"
)

type ImageView struct{}

func (rv *ImageView) New(obj *types.Image) *Image {
	if obj == nil {
		return nil
	}
	i := new(Image)
	i.Meta = rv.ToImageMeta(&obj.Meta)
	i.Spec = rv.ToImageSpec(&obj.Spec)
	i.Status = rv.ToImageStatus(&obj.Status)
	return i
}

func (obj *Image) ToJson() ([]byte, error) {
	if unsafe.Sizeof(obj) == 0 {
		return []byte{}, nil
	}
	return json.Marshal(obj)
}

func (rv *ImageView) NewList(list []*types.Image) *ImageList {
	if list == nil {
		return nil
	}
	il := make(ImageList, 0)
	for _, item := range list {
		il = append(il, rv.New(item))
	}
	return &il
}

func (obj *ImageList) ToJson() ([]byte, error) {
	if unsafe.Sizeof(obj) == 0 {
		return []byte{}, nil
	}
	return json.Marshal(obj)
}

func (rv *ImageView) ToImageMeta(obj *types.ImageMeta) *ImageMeta {
	return &ImageMeta{
		Name:        obj.Name,
		Owner:       obj.Owner,
		Description: obj.Description,
	}
}

func (rv *ImageView) ToImageSpec(obj *types.ImageSpec) *ImageSpec {
	spec := ImageSpec{
		Private: obj.Private,
		TagList: make([]*ImageTag, 0),
	}

	for _, tag := range obj.TagList {

		it := new(ImageTag)
		it.Meta.Name = tag.Name
		it.Spec.DockerFile = tag.Spec.DockerFile
		it.Spec.Command = tag.Spec.Command
		it.Spec.EnvVars = tag.Spec.EnvVars
		it.Status.Disabled = tag.Disabled

		spec.TagList = append(spec.TagList, it)
	}

	return &spec
}

func (rv *ImageView) ToImageStatus(obj *types.ImageStatus) *ImageStatus {
	return &ImageStatus{
		Stats: ImageStats{
			BuildsQuantity: obj.Stats.BuildsQuantity,
			PullsQuantity:  obj.Stats.PullsQuantity,
			StarsQuantity:  obj.Stats.StarsQuantity,
			ViewsQuantity:  obj.Stats.ViewsQuantity,
		},
	}
}
