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

func (bv *BuilderView) NewManifest(obj *types.BuildManifest) *BuildManifest {
	if obj == nil {
		return nil
	}

	manifest := new(BuildManifest)
	manifest.Image.Host = obj.Image.Host
	manifest.Image.Name = obj.Image.Name
	manifest.Image.Owner = obj.Image.Owner
	manifest.Image.Tag = obj.Image.Tag
	manifest.Image.Auth = obj.Image.Auth
	manifest.Source.Url = obj.Source.Url
	manifest.Source.Branch = obj.Source.Branch

	manifest.Config.Dockerfile = obj.Config.Dockerfile
	manifest.Config.Context = obj.Config.Context
	manifest.Config.Command = obj.Config.Command
	manifest.Config.Workdir = obj.Config.Workdir
	manifest.Config.EnvVars = obj.Config.EnvVars

	return manifest
}

func (obj BuildManifest) ToJson() ([]byte, error) {
	if unsafe.Sizeof(obj) == 0 {
		return []byte{}, nil
	}
	return json.Marshal(obj)
}
