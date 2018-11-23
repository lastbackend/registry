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

package builder

import (
	"context"
	"time"

	"github.com/lastbackend/lastbackend/pkg/log"
	"github.com/lastbackend/registry/pkg/controller/envs"
	"github.com/lastbackend/registry/pkg/distribution"
)

const (
	logLevel                     = 3
	logPrefix                    = "runtime:builder_controller"
	delayForCheckOfflineBuilders = 30 * time.Second
)

type BuilderController struct {
	done chan bool
}

func New() *BuilderController {
	return new(BuilderController)
}

func (bc BuilderController) Start(ctx context.Context) {
	log.V(logLevel).Infof("%s:> run builder controller", logPrefix)
	go bc.inspector(ctx)
	<-bc.done
}

func (bc BuilderController) Stop() {
	log.V(logLevel).Infof("%s:> stop builder controller", logPrefix)
	bc.done <- true
}

func (bc BuilderController) inspector(ctx context.Context) {
	for range time.Tick(delayForCheckOfflineBuilders) {
		select {
		case <-bc.done:
			return
		default:

			var (
				bm = distribution.NewBuilderModel(ctx, envs.Get().GetStorage())
			)

			if err := bm.MarkOffline(); err != nil {
				log.V(logLevel).Errorf("%s:inspector:> check and mark BuilderController machine err: %v", logPrefix, err)
				continue
			}
		}
	}
}
