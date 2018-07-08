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
	"time"
	"unsafe"
)

type BuildView struct{}

func (rv *BuildView) New(obj *types.Build) *Build {
	if obj == nil {
		return nil
	}
	i := new(Build)
	i.Meta = rv.ToBuildMeta(&obj.Meta)
	i.Spec = rv.ToBuildSpec(&obj.Spec)
	i.Status = rv.ToBuildStatus(&obj.Status)
	return i
}

func (obj *Build) ToJson() ([]byte, error) {
	return json.Marshal(obj)
}

func (rv *BuildView) NewList(list []*types.Build) *BuildList {
	if list == nil {
		return nil
	}
	il := make(BuildList, 0)
	for _, item := range list {
		il = append(il, rv.New(item))
	}
	return &il
}

func (obj *BuildList) ToJson() ([]byte, error) {
	if unsafe.Sizeof(obj) == 0 {
		return []byte{}, nil
	}
	return json.Marshal(obj)
}

func (rv *BuildView) ToBuildMeta(obj *types.BuildMeta) *BuildMeta {
	return &BuildMeta{
		ID:     obj.ID,
		Number: obj.Number,
	}
}

func (rv *BuildView) ToBuildSpec(obj *types.BuildSpec) *BuildSpec {
	return &BuildSpec{
		Source: BuildSources{
			Hub:    obj.Source.Hub,
			Owner:  obj.Source.Owner,
			Name:   obj.Source.Name,
			Branch: obj.Source.Branch,
		},
		Config: BuildConfig{
			Dockerfile: obj.Config.Dockerfile,
			Workdir:    obj.Config.Workdir,
			EnvVars:    obj.Config.EnvVars,
			Command:    obj.Config.Command,
		},
	}
}

func (rv *BuildView) ToBuildStatus(obj *types.BuildStatus) *BuildStatus {
	started := &time.Time{}
	if obj.Started.IsZero() {
		started = nil
	} else {
		started = &obj.Started
	}

	finished := &time.Time{}
	if obj.Finished.IsZero() {
		finished = nil
	} else {
		finished = &obj.Finished
	}

	return &BuildStatus{
		Size:       obj.Size,
		Step:       obj.Step,
		Message:    obj.Message,
		Status:     obj.Status,
		Done:       obj.Done,
		Processing: obj.Processing,
		Canceled:   obj.Canceled,
		Error:      obj.Error,
		Created:    obj.Created,
		Updated:    obj.Updated,
		Finished:   finished,
		Started:    started,
	}
}
