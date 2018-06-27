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
	"github.com/lastbackend/registry/pkg/builder/envs"
	"github.com/lastbackend/registry/pkg/builder/handler"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/rpc"
	"github.com/spf13/viper"
)

type Runtime struct {
	id     string
	bridge *rpc.RPC
}

func New() *Runtime {

	log.Debug("Runtime initialise")

	var (
		r   = new(Runtime)
		err error
	)

	r.bridge = rpc.Register(types.KindBuilder, viper.GetString("builder.uuid"), viper.GetString("token"))
	if err != nil {
		log.Fatal(err)
	}
	r.bridge.SetURI(viper.GetString("amqp"))

	r.bridge.SetHandler(types.BuildTaskExecuteEventName, handler.BuildTaskExecuteHandler)
	r.bridge.SetHandler(types.BuildTaskCancelEventName, handler.BuildTaskCancelHandler)
	r.bridge.SetHandler(types.BuildTaskLogsEventName, handler.BuildLogsHandler)

	envs.Get().SetRPC(r.bridge)

	return r
}

func (r *Runtime) Start(ctx context.Context) {
	go func() {
		r.bridge.Listen()
	}()

	select {
	case <-ctx.Done():
	case <-r.bridge.Connected():
	}
}

func (r *Runtime) Stop() {
	r.bridge.Shutdown()
}
