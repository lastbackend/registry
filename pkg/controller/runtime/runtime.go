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
	"github.com/lastbackend/registry/pkg/controller/runtime/exporter"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/system"
)

const (
	logLevel  = 3
	logPrefix = "controller:runtime"
)

type Runtime struct {
	done   chan bool
	ctx    context.Context
	cancel context.CancelFunc

	active  bool
	process *system.Process

	builderCtrl  *builder.BuilderController
	buildCtrl    *build.BuildController
	exporterCtrl *exporter.ExporterController
}

func NewRuntime() *Runtime {
	r := new(Runtime)

	r.ctx, r.cancel = context.WithCancel(context.Background())

	r.done = make(chan bool)
	r.process = new(system.Process)

	_, err := r.process.Register(r.ctx, types.KindController)
	if err != nil {
		log.V(logLevel).Debugf("%s:register:> register controller", logPrefix)
		return nil
	}

	r.builderCtrl = builder.New()
	r.buildCtrl = build.New()
	r.exporterCtrl = exporter.New()

	return r
}

// Loop - runtime main loop watch
func (r *Runtime) Loop() {

	log.V(logLevel).Infof("%s:> runtime loop", logPrefix)

	var (
		lead = make(chan bool)
	)

	go func() {
		for {
			select {
			case <-r.ctx.Done():
				return
			case l := <-lead:
				{

					if l {
						if r.active {
							log.V(logLevel).Debugf("%s:loop:> is already marked as master -> skip", logPrefix)
							continue
						}

						log.V(logLevel).Debugf("%s:loop:> mark as master", logPrefix)

						r.active = true

						go r.builderCtrl.Start(r.ctx)
						go r.buildCtrl.Start(r.ctx)
						go r.exporterCtrl.Start(r.ctx)

						continue
					}

					if !r.active {
						log.V(logLevel).Debugf("%s:loop:> is already marked as slave -> skip", logPrefix)
						continue
					}

					log.V(logLevel).Debugf("%s:loop:> mark as slave", logPrefix)
					r.active = false

					r.builderCtrl.Stop()
					r.buildCtrl.Stop()
					r.exporterCtrl.Stop()
				}
			}
		}
	}()

	go r.process.HeartBeat(r.ctx, lead)
}

func (r *Runtime) Stop() {
	r.cancel()
}
