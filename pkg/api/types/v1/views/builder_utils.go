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

type BuilderView struct{}

func (bv *BuilderView) New(obj *types.Builder) *Builder {
	if obj == nil {
		return nil
	}
	bl := new(Builder)
	bl.Meta = bv.ToBuilderMeta(obj.Meta)
	bl.Status = bv.ToBuilderStatus(obj.Status)
	bl.Spec = bv.ToBuilderSpec(obj.Spec)
	return bl
}

func (bv *BuilderView) ToBuilderMeta(meta types.BuilderMeta) BuilderMeta {
	return BuilderMeta{
		ID:       meta.ID,
		Hostname: meta.Hostname,
		Created:  meta.Created,
		Updated:  meta.Updated,
	}
}

func (bv *BuilderView) ToBuilderStatus(status types.BuilderStatus) BuilderStatus {
	return BuilderStatus{
		Insecure: status.Insecure,
		Online:   status.Online,
		TLS:      status.TLS,
		Capacity: BuilderResources{
			Workers: status.Capacity.Workers,
			Memory:  status.Capacity.Memory,
			Cpu:     status.Capacity.Cpu,
			Storage: status.Capacity.Storage,
		},
		Allocated: BuilderResources{
			Workers: status.Allocated.Workers,
			Memory:  status.Allocated.Memory,
			Cpu:     status.Allocated.Cpu,
			Storage: status.Allocated.Storage,
		},
	}
}

func (bv *BuilderView) ToBuilderSpec(spec types.BuilderSpec) BuilderSpec {
	bs := BuilderSpec{
		Network: BuilderSpecNetwork{
			IP:   spec.Network.IP,
			Port: spec.Network.Port,
			TLS:  spec.Network.TLS,
		},
		Limits: BuilderSpecLimits{
			WorkerLimit:  spec.Limits.WorkerLimit,
			Workers:      spec.Limits.Workers,
			WorkerMemory: spec.Limits.WorkerMemory,
		},
	}

	if bs.Network.SSL != nil {
		bs.Network.SSL = new(SSL)
		bs.Network.SSL.CA = spec.Network.SSL.CA
		bs.Network.SSL.ClientCert = spec.Network.SSL.Cert
		bs.Network.SSL.ClientKey = spec.Network.SSL.Key
	}

	return bs
}

func (obj Builder) ToJson() ([]byte, error) {
	if unsafe.Sizeof(obj) == 0 {
		return []byte{}, nil
	}
	return json.Marshal(obj)
}

func (bv *BuilderView) NewList(obj []*types.Builder) *BuilderList {
	if obj == nil {
		return nil
	}

	rl := make(BuilderList, 0)
	for _, v := range obj {
		rl = append(rl, bv.New(v))
	}

	return &rl
}

func (obj BuilderList) ToJson() ([]byte, error) {
	if unsafe.Sizeof(obj) == 0 {
		return []byte{}, nil
	}
	return json.Marshal(obj)
}

func (bv *BuilderView) NewBuildManifest(obj *types.Task) *BuildManifest {
	if obj == nil {
		return nil
	}

	manifest := new(BuildManifest)

	manifest.Meta.ID = obj.Meta.ID

	manifest.Spec.Image.Host = obj.Spec.Image.Host
	manifest.Spec.Image.Name = obj.Spec.Image.Name
	manifest.Spec.Image.Owner = obj.Spec.Image.Owner
	manifest.Spec.Image.Tag = obj.Spec.Image.Tag
	manifest.Spec.Image.Auth = obj.Spec.Image.Auth
	manifest.Spec.Source.Url = obj.Spec.Source.Url
	manifest.Spec.Source.Branch = obj.Spec.Source.Branch

	manifest.Spec.Config.Dockerfile = obj.Spec.Config.Dockerfile
	manifest.Spec.Config.Context = obj.Spec.Config.Context
	manifest.Spec.Config.Command = obj.Spec.Config.Command
	manifest.Spec.Config.Workdir = obj.Spec.Config.Workdir
	manifest.Spec.Config.EnvVars = obj.Spec.Config.EnvVars

	return manifest
}

func (obj BuildManifest) ToJson() ([]byte, error) {
	if unsafe.Sizeof(obj) == 0 {
		return []byte{}, nil
	}
	return json.Marshal(obj)
}

func (bv *BuilderView) NewConfigManifest(obj *types.Builder) *BuilderConfig {
	if obj == nil {
		return nil
	}
	bc := new(BuilderConfig)
	if obj.Spec.Limits.WorkerLimit {
		bc.Limits = new(BuilderLimitConfig)
		bc.Limits.WorkerLimit = obj.Spec.Limits.WorkerLimit
		bc.Limits.Workers = obj.Spec.Limits.Workers
		bc.Limits.WorkerMemory = obj.Spec.Limits.WorkerMemory
	}
	return bc
}

func (obj BuilderConfig) ToJson() ([]byte, error) {
	if unsafe.Sizeof(obj) == 0 {
		return []byte{}, nil
	}
	return json.Marshal(obj)
}
