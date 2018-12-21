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

package build

import (
	"context"
	"encoding/json"
	"time"

	"github.com/lastbackend/lastbackend/pkg/log"
	"github.com/lastbackend/registry/pkg/controller/envs"
	"github.com/lastbackend/registry/pkg/distribution"
	"github.com/lastbackend/registry/pkg/distribution/types"
)

const (
	logLevel                  = 3
	logPrefix                 = "runtime:build_controller"
	delayForCheckUnfreezeTime = 5 * time.Minute
)

type BuildController struct {
	done chan bool
}

func New() *BuildController {
	bc := new(BuildController)
	bc.done = make(chan bool)
	go bc.watch(context.Background())
	return bc
}

func (bc BuildController) Start(ctx context.Context) {
	log.V(logLevel).Infof("%s:> start build controller", logPrefix)
	go bc.inspector(ctx)
	<-bc.done
}

func (bc BuildController) Stop() {
	log.V(logLevel).Infof("%s:> stop build controller", logPrefix)
	bc.done <- true
}

func (bc BuildController) inspector(ctx context.Context) {
	for range time.Tick(delayForCheckUnfreezeTime) {
		select {
		case <-bc.done:
			return
		default:
			var (
				bm = distribution.NewBuildModel(ctx, envs.Get().GetStorage())
			)

			if err := bm.Unfreeze(); err != nil {
				log.V(logLevel).Errorf("%s:inspector:> check and unfreeze dangling builds err: %v", logPrefix, err)
				continue
			}
		}
	}
}

func (bc BuildController) watch(ctx context.Context) {
	var bm = distribution.NewBuildModel(ctx, envs.Get().GetStorage())
	var evs = make(chan string)

	go func() {
		err := envs.Get().GetStorage().Listen(ctx, "e_watch_build", evs)
		if err != nil {
			log.Error(err)
			return
		}
	}()

	for {
		select {
		case <-bc.done:
			return
		case e := <-evs:
			{

				event := types.Event{}

				if err := json.Unmarshal([]byte(e), &event); err != nil {
					log.Errorf("%s:subscribe:> parse event from db err: %v", logPrefix, err)
					continue
				}

				build, err := bm.Get(event.Entity)
				if err != nil {
					log.Errorf("%s:subscribe:> get build err: %v", err)
					continue
				}

				switch event.Operation {
				case "insert":
					fallthrough
				case "update":
					envs.Get().GetState().Build().Set(build.Meta.ID, build)
				}
				break
			}
		}
	}

}
