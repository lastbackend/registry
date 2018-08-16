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
	"time"

	"github.com/lastbackend/registry/pkg/controller/envs"
	"github.com/lastbackend/registry/pkg/distribution"
	"github.com/lastbackend/registry/pkg/log"
)

const (
	logLevel                  = 3
	logPrefix                 = "runtime:build"
	delayForCheckUnfreezeTime = 1 * time.Minute
)

func Inspector(ctx context.Context) {
	log.V(logLevel).Infof("%s:> run inspector for build", logPrefix)

	for range time.Tick(delayForCheckUnfreezeTime) {
		var (
			bm = distribution.NewBuildModel(ctx, envs.Get().GetStorage())
		)

		if err := bm.Unfreeze(); err != nil {
			log.V(logLevel).Errorf("%s:inspector:> check and unfreeze dangling builds err: %v", logPrefix, err)
			continue
		}
	}
}
