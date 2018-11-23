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

package exporter

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/spf13/viper"
	"net/http"
	"time"

	"github.com/lastbackend/lastbackend/pkg/log"
	"github.com/lastbackend/registry/pkg/api/types/v1"
	"github.com/lastbackend/registry/pkg/api/types/v1/request"
	"github.com/lastbackend/registry/pkg/controller/envs"
	"github.com/lastbackend/registry/pkg/distribution"
	req "github.com/lastbackend/registry/pkg/util/http/request"
	"github.com/lastbackend/registry/pkg/util/url"
)

const (
	logLevel           = 3
	logPrefix          = "runtime:exporter_controller"
	delayForSendEvents = 1 * time.Minute
	delayTime          = 1 * time.Microsecond
)

type ExporterController struct {
	done chan bool
}

func New() *ExporterController {
	return new(ExporterController)
}

func (ec ExporterController) Start(ctx context.Context) {
	log.V(logLevel).Infof("%s:> start events exporter", logPrefix)
	go ec.loop(ctx)
	<-ec.done
}

func (ec ExporterController) Stop() {
	log.V(logLevel).Infof("%s:> stop events exporter", logPrefix)
	ec.done <- true
}

func (ec ExporterController) loop(ctx context.Context) {
	if !viper.IsSet("exporter.uri") {
		return
	}

	uri := viper.GetString("exporter.uri")
	timeout := viper.GetDuration("exporter.timeout")

	u, err := url.Parse(uri)
	if err != nil {
		log.V(logLevel).Errorf("%s:> invalid exporter uri", logPrefix)
		return
	}

	sm := distribution.NewSystemModel(ctx, envs.Get().GetStorage())
	sys, err := sm.Get()
	if err != nil {
		log.V(logLevel).Errorf("%s:> get system info err: %v", logPrefix, err)
		return
	}

	cfg := new(req.Config)
	cfg.BearerToken = sys.AccessToken

	endpoint := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
	if u.Host == "" {
		endpoint = fmt.Sprintf("http://%s", u.Host)
	}

	client, err := req.NewRESTClient(endpoint, cfg)
	if err != nil {
		log.V(logLevel).Errorf("%s:> create rest client for exporter err: %v", logPrefix, err)
		return
	}

	timeout = timeout * time.Second
	if timeout == 0 {
		timeout = delayTime
	}

	state := envs.Get().GetState()

	var lastEvent = sys.CtrlLastEvent

	for range time.Tick(delayForSendEvents) {

		select {
		case <-ec.done:
			return
		default:

			rq := v1.Request().Event().EventOptions()
			rq.Builds = make(map[string]request.BuildEvent, 0)

			var i = 0

			for _, b := range state.Build().List() {

				if lastEvent != nil && !lastEvent.IsZero() && b.Meta.Updated.Before(*lastEvent) {
					state.Build().Del(b.Meta.ID)
					continue
				}

				i++
				if i > 10 {
					break
				}

				be := request.BuildEvent{
					ID:         b.Meta.ID,
					Number:     b.Meta.Number,
					Branch:     b.Spec.Source.Branch,
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

				if b.Spec.Source.Commit != nil {
					be.Commit.Message = b.Spec.Source.Commit.Message
					be.Commit.Username = b.Spec.Source.Commit.Username
					be.Commit.Hash = b.Spec.Source.Commit.Hash
					be.Commit.Date = b.Spec.Source.Commit.Date
					be.Commit.Email = b.Spec.Source.Commit.Email
				}

				rq.Builds[b.Meta.ID] = be

				lastEvent = &b.Meta.Updated
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

			if lastEvent != nil && !lastEvent.IsZero() {
				err = sm.UpdateControllerLastEvent(sys, &types.SystemUpdateControllerLastEventOptions{LastEvent: *lastEvent})
				if err != nil {
					log.V(logLevel).Errorf("%s:> get system info err: %v", logPrefix, err)
					return
				}
			}

			for key := range rq.Builds {
				state.Build().Del(key)
			}

		}

	}

}
