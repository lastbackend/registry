//
// Last.Backend LLC CONFIDENTIAL
// __________________
//
// [2014] - [2019] Last.Backend LLC
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

package monitor

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/lastbackend/lastbackend/pkg/log"
	"github.com/lastbackend/registry/pkg/api/types/v1"
	"github.com/lastbackend/registry/pkg/distribution"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/storage"
	"github.com/lastbackend/registry/pkg/util/eventer"
	"github.com/lastbackend/registry/pkg/util/socket"
)

const defaultChannel = "notify"

type Monitor struct {
	sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc

	pool *eventer.Controller

	storage storage.IStorage
}

type IMonitor interface {
	Stop()
	Subscribe(conn *websocket.Conn)
}

const (
	logLevel  = 5
	logPrefix = "monitor"
)

func New(storage storage.IStorage) *Monitor {
	var m = new(Monitor)

	m.ctx, m.cancel = context.WithCancel(context.Background())

	m.pool = eventer.New()

	m.storage = storage

	go m.listen()

	return m
}

func (m Monitor) Subscribe(conn *websocket.Conn) {

	log.Debugf("%s:subscribe: connection subscribe to events>", logPrefix)

	sock := socket.New(context.Background(), conn, m.pool.Leave, m.pool.Message)
	p := m.pool.Get(defaultChannel)

	if p == nil {
		p = m.pool.Add(defaultChannel, sock)
	} else {
		m.pool.Attach(p, sock)
	}

}

func (m Monitor) Stop() {
	m.cancel()
}

func (m Monitor) listen() {
	var bm = distribution.NewBuildModel(m.ctx, m.storage)
	var bdm = distribution.NewBuilderModel(m.ctx, m.storage)
	var evs = make(chan string)

	go func() {
		err := m.storage.Listen(m.ctx, "e_watch", evs)
		if err != nil {
			log.V(logLevel).Errorf("%s:listen:>  subscribe on events err: %v", logPrefix, err)
			return
		}
	}()

	for {
		select {
		case <-m.ctx.Done():
			return
		case e := <-evs:
			{

				event := types.Event{}

				if err := json.Unmarshal([]byte(e), &event); err != nil {
					log.V(logLevel).Errorf("%s:listen:> parse event from db err: %v", logPrefix, err)
					continue
				}

				switch event.Name {
				case "builder":

					builder, err := bdm.Get(event.Entity)
					if err != nil {
						log.V(logLevel).Errorf("%s:listen:> get builder err: %v", logPrefix, err)
						continue
					}

					response, err := v1.View().Builder().New(builder).ToJson()
					if err != nil {
						log.V(logLevel).Errorf("%s:listen:> convert struct to json err: %v", logPrefix, err)
						continue
					}

					if err := m.pool.Broadcast(defaultChannel, "update", event.Name, response); err != nil {
						log.V(logLevel).Errorf("%s:listen:> send broadcast builder event err: %v", logPrefix, err)
						continue
					}

					continue

				case "build":

					build, err := bm.Get(event.Entity)
					if err != nil {
						log.V(logLevel).Errorf("%s:listen:> get build err: %v", logPrefix, err)
						continue
					}

					response, err := v1.View().Build().New(build).ToJson()
					if err != nil {
						log.V(logLevel).Errorf("%s:listen:> convert struct to json err: %v", logPrefix, err)
						continue
					}

					if err := m.pool.Broadcast(defaultChannel, event.Operation, event.Name, response); err != nil {
						log.V(logLevel).Errorf("%s:listen:> send broadcast builder event err: %v", logPrefix, err)
						continue
					}

					continue
				}
			}
		}
	}
}
