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
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/lastbackend/lastbackend/pkg/log"
	"github.com/lastbackend/registry/pkg/api/types/v1"
	"github.com/lastbackend/registry/pkg/api/types/v1/request"
	"github.com/lastbackend/registry/pkg/controller/envs"
	"github.com/lastbackend/registry/pkg/controller/runtime/build"
	"github.com/lastbackend/registry/pkg/controller/runtime/builder"
	"github.com/lastbackend/registry/pkg/distribution"
	req "github.com/lastbackend/registry/pkg/util/http/request"
	"github.com/lastbackend/registry/pkg/util/url"
)

const (
	logLevel  = 3
	logPrefix = "controller:runtime"
	delayTime = 5 * time.Second
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

func (r Runtime) Watcher() {
	log.V(logLevel).Infof("%s:> run runtime watcher", logPrefix)
	go build.Watch(r.ctx)
	<-r.done
}

func (r Runtime) Stop() {
	log.V(logLevel).Infof("%s:> stop runtime process", logPrefix)
	r.done <- true
}

func (r Runtime) Exporter(uri string, timeout time.Duration) {
	log.V(logLevel).Infof("%s:> run runtime exporter for %s", logPrefix, uri)

	u, err := url.Parse(uri)
	if err != nil {
		log.V(logLevel).Errorf("%s:> invalid exporter uri", logPrefix)
		panic(err)
	}

	sm := distribution.NewSystemModel(r.ctx, envs.Get().GetStorage())
	system, err := sm.Get()
	if err != nil {
		log.V(logLevel).Errorf("%s:> get system info err: %v", logPrefix, err)
		return
	}

	cfg := new(req.Config)
	cfg.BearerToken = system.AccessToken

	endpoint := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
	if u.Host == "" {
		endpoint = fmt.Sprintf("http://%s", u.Host)
	}

	client, err := req.NewRESTClient(endpoint, cfg)
	if err != nil {
		log.V(logLevel).Errorf("%s:> create rest client for exporter err: %v", logPrefix, err)
		panic(err)
	}

	timeout = timeout * time.Second
	if timeout == 0 {
		timeout = delayTime
	}

	ticker := time.NewTicker(timeout)

	state := envs.Get().GetState()

	go func() {
		for range ticker.C {

			rq := v1.Request().Event().EventOptions()
			rq.Builds = make(map[string]request.BuildEvent, 0)

			var i = 0
			for key, b := range state.Build().List() {
				i++
				if i > 10 {
					break
				}
				rq.Builds[key] = request.BuildEvent{
					ID:         b.Meta.ID,
					Image:      fmt.Sprintf("%s/%s:%s", b.Spec.Image.Owner, b.Spec.Image.Name, b.Spec.Image.Tag),
					Source:     fmt.Sprintf("%s/%s/%s#%s", b.Spec.Source.Hub, b.Spec.Source.Owner, b.Spec.Source.Name, b.Spec.Source.Branch),
					Size:       b.Status.Size,
					Step:       b.Status.Step,
					Status:     b.Status.Status,
					Message:    b.Status.Message,
					Processing: b.Status.Processing,
					Done:       b.Status.Done,
					Error:      b.Status.Error,
					Canceled:   b.Status.Canceled,
					Finished:   &b.Status.Finished,
					Started:    &b.Status.Started,
				}
			}

			body, err := json.Marshal(rq)
			if err != nil {
				log.V(logLevel).Errorf("%s:> convert data to json err: %v", logPrefix, err)
				continue
			}

			res := client.Post(u.Path).Body(body).Do()
			data, err := res.Raw()
			if res.StatusCode() != http.StatusOK || err != nil {
				log.V(logLevel).Errorf("%s:> send events err: %v body: (%s)", logPrefix, res.Error(), data)
				continue
			}

			for b := range rq.Builds {
				state.Build().Del(b)
			}

		}
	}()

	<-r.done
}
