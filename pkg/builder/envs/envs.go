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

package envs

import (
	"github.com/lastbackend/lastbackend/pkg/runtime/cri"
	"github.com/lastbackend/registry/pkg/api/client"
	"github.com/lastbackend/registry/pkg/builder/types"
	"github.com/lastbackend/registry/pkg/storage"
	"github.com/lastbackend/registry/pkg/util/blob"
)

var e Env

func Get() *Env {
	return &e
}

type Env struct {
	ip          string
	hostname    string
	builder     types.IBuilder
	client      client.IClient
	storage     storage.IStorage
	blobStorage blob.IBlobStorage
	cri         cri.CRI
}

func (env *Env) SetCri(cri cri.CRI) {
	env.cri = cri
}

func (env Env) GetCri() cri.CRI {
	return env.cri
}

func (env *Env) SetClient(c client.IClient) {
	env.client = c
}

func (env Env) GetClient() client.IClient {
	return env.client
}

func (env *Env) SetBuilder(b types.IBuilder) {
	env.builder = b
}

func (env Env) GetBuilder() types.IBuilder {
	return env.builder
}

func (env *Env) SetHostname(h string) {
	env.hostname = h
}

func (env *Env) SetBlobStorage(u blob.IBlobStorage) {
	env.blobStorage = u
}

func (env Env) GetBlobStorage() blob.IBlobStorage {
	return env.blobStorage
}

func (env Env) GetHostname() string {
	return env.hostname
}

func (env *Env) SetIP(ip string) {
	env.ip = ip
}

func (env Env) GetIP() string {
	return env.ip
}
