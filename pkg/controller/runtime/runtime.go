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

package runtime

import (
	"context"
	"github.com/lastbackend/lastbackend/pkg/log"
	"github.com/lastbackend/registry/pkg/controller/runtime/build"
	"github.com/lastbackend/registry/pkg/controller/runtime/builder"
)

const (
	logLevel  = 3
	logPrefix = "controller:runtime"
)

type Runtime struct {
	done chan bool
	ctx  context.Context
}

func NewRuntime() *Runtime {
	r := new(Runtime)
	r.done = make(chan bool)
	r.ctx = context.Background()
	return r
}

func (r Runtime) Inspector() {
	log.V(logLevel).Infof("%s:> run runtime inspector", logPrefix)
	go builder.Inspector(r.ctx)
	go build.Inspector(r.ctx)
	<-r.done
}

func (r Runtime) Stop() {
	log.V(logLevel).Infof("%s:> stop runtime process", logPrefix)
	r.done <- true
}
