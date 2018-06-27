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

package v1

import (
	"encoding/json"
	"time"

	"github.com/lastbackend/registry/pkg/distribution/types"
)

type BuildView struct{}

func (bv *BuildView) New(obj *types.Build) *Build {
	b := new(Build)
	b.Meta = bv.ToBuildMeta(&obj.Meta)
	b.Repo = BuildRepo(obj.Repo.SelfLink)
	b.State = bv.ToBuildState(&obj.State)
	b.Stats = bv.ToBuildStats(&obj.Stats)
	b.Sources = bv.ToBuildSource(&obj.Sources)
	b.Image = bv.ToBuildImage(&obj.Image)
	return b
}

func (obj *Build) ToJson() ([]byte, error) {
	return json.Marshal(obj)
}

func (bv *BuildView) NewList(obj map[string]*types.Build) *BuildList {
	if obj == nil {
		return nil
	}

	b := make(BuildList, 0)
	for _, v := range obj {
		b = append(b, bv.New(v))
	}
	return &b
}

func (obj *BuildList) ToJson() ([]byte, error) {
	if obj == nil {
		obj = &BuildList{}
	}
	return json.Marshal(obj)
}

func (bv *BuildView) ToBuildMeta(obj *types.BuildMeta) BuildMeta {
	return BuildMeta{
		Number:   obj.Number,
		SelfLink: obj.SelfLink,
		Created:  obj.Created,
		Updated:  obj.Updated,
	}
}

func (bv *BuildView) ToBuildState(obj *types.BuildState) BuildState {

	started := &time.Time{}
	if obj.Started.IsZero() {
		started = nil
	} else {
		started = &obj.Started
	}

	return BuildState{
		Step:       obj.Step,
		Message:    obj.Message,
		Status:     obj.Status,
		Done:       obj.Done,
		Processing: obj.Processing,
		Canceled:   obj.Canceled,
		Error:      obj.Error,
		Finished:   obj.Finished,
		Started:    started,
	}
}

func (bv *BuildView) ToBuildStats(obj *types.BuildInfo) BuildStats {
	return BuildStats{}
}

func (bv *BuildView) ToBuildSource(obj *types.BuildSources) BuildSource {
	return BuildSource{
		Hub:    obj.Hub,
		Owner:  obj.Owner,
		Name:   obj.Name,
		Branch: obj.Branch,
		Commit: obj.Commit,
	}
}

func (bv *BuildView) ToBuildImage(obj *types.BuildImage) BuildImage {
	return BuildImage{
		Tag: obj.Tag,
	}
}
