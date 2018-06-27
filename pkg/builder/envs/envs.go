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

package envs

import (
	"github.com/lastbackend/registry/pkg/builder/types"
	"github.com/lastbackend/registry/pkg/rpc"
	"github.com/lastbackend/registry/pkg/storage"
	"github.com/lastbackend/lastbackend/pkg/node/runtime/cri"
)

var e Env

func Get() *Env {
	return &e
}

type Env struct {
	builder types.IBuilder
	storage storage.Storage
	cri     cri.CRI
	rpc     *rpc.RPC
}

func (c *Env) SetCri(cri cri.CRI) {
	c.cri = cri
}

func (c *Env) GetCri() cri.CRI {
	return c.cri
}

func (c *Env) SetBuilder(b types.IBuilder) {
	c.builder = b
}

func (c *Env) GetBuilder() types.IBuilder {
	return c.builder
}

func (c *Env) SetRPC(rpc *rpc.RPC) {
	c.rpc = rpc
}

func (c *Env) GetRPC() *rpc.RPC {
	return c.rpc
}
