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
	logPrefix                    = "runtime:builder"
	delayForCheckOfflineBuilders = 30 * time.Second
)

func Inspector(ctx context.Context) {
	log.V(logLevel).Infof("%s:> run inspector for builder", logPrefix)

	for range time.Tick(delayForCheckOfflineBuilders) {
		var (
			bm = distribution.NewBuilderModel(ctx, envs.Get().GetStorage())
		)

		if err := bm.MarkOffline(); err != nil {
			log.V(logLevel).Errorf("%s:inspector:> check and mark builder machine err: %v", logPrefix, err)
			continue
		}
	}
}
