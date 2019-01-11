//
// Last.Backend LLC CONFIDENTIAL
// __________________
//
// [2014] - [2019] Last.Backend LLC
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
	"unsafe"

	"github.com/lastbackend/registry/pkg/distribution/types"
)

type ImageView struct{}

func (rv *ImageView) New(obj *types.Image) *Image {
	if obj == nil {
		return nil
	}
	i := new(Image)
	i.Meta = rv.ToImageMeta(&obj.Meta)
	i.TagList = rv.ToImageTags(obj.TagList)
	i.Status = rv.ToImageStatus(obj.Status)
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
		i := new(Image)
		i.Meta = rv.ToImageMeta(&item.Meta)
		il = append(il, i)
	}
	return &il
}

func (obj *ImageList) ToJson() ([]byte, error) {
	if unsafe.Sizeof(obj) == 0 {
		return []byte{}, nil
	}
	return json.Marshal(obj)
}

func (rv *ImageView) ToImageMeta(obj *types.ImageMeta) ImageMeta {
	im := ImageMeta{
		Name:        obj.Name,
		Owner:       obj.Owner,
		Description: obj.Description,
		Labels:      obj.Labels,
		Created:     obj.Created,
		Updated:     obj.Updated,
	}

	if obj.Labels == nil {
		im.Labels = make(map[string]string, 0)
	}

	return im
}

func (rv *ImageView) ToImageTags(obj map[string]*types.ImageTag) *ImageTags {
	var tl = ImageTags{}

	for _, tag := range obj {

		it := new(ImageTag)
		it.Meta.Name = tag.Name
		it.Spec.DockerFile = tag.Spec.DockerFile
		it.Spec.Context = tag.Spec.Context
		it.Spec.Command = tag.Spec.Command
		it.Spec.EnvVars = tag.Spec.EnvVars
		it.Status.Disabled = tag.Disabled

		tl = append(tl, it)
	}

	return &tl
}

func (rv *ImageView) ToImageStatus(obj types.ImageStatus) *ImageStatus {
	return &ImageStatus{
		Private: obj.Private,
		Stats: ImageStats{
			BuildsQuantity: obj.Stats.BuildsQuantity,
			PullsQuantity:  obj.Stats.PullsQuantity,
			StarsQuantity:  obj.Stats.StarsQuantity,
			ViewsQuantity:  obj.Stats.ViewsQuantity,
		},
	}
}
