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
	"fmt"
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

				ev := types.StorageEvent{}

				if err := json.Unmarshal([]byte(e), &ev); err != nil {
					log.V(logLevel).Errorf("%s:listen:> parse event from db err: %v", logPrefix, err)
					continue
				}

				switch ev.Channel {
				case "builder":

					response := []byte(fmt.Sprintf("{\"id\":\"%s\"}", ev.Entity))

					if ev.Operation == "insert" || ev.Operation == "update" {
						builder, err := bdm.Get(ev.Entity)
						if err != nil {
							log.V(logLevel).Errorf("%s:listen:> get builder err: %v", logPrefix, err)
							continue
						}

						response, err = v1.View().Builder().New(builder).ToJson()
						if err != nil {
							log.V(logLevel).Errorf("%s:listen:> convert struct to json err: %v", logPrefix, err)
							continue
						}
					}

					event := ""
					switch ev.Operation {
					case types.StorageInsertAction:
						event = "builder:connect"
					case types.StorageUpdateAction:
						event = "builder:update"
					case types.StorageDeleteAction:
						event = "builder:remove"
					default:
						continue
					}

					if err := m.pool.Broadcast(defaultChannel, event, response); err != nil {
						log.V(logLevel).Errorf("%s:listen:> send broadcast builder event err: %v", logPrefix, err)
						continue
					}

					continue

				case "build":

					response := []byte(fmt.Sprintf("{\"id\":\"%s\"}", ev.Entity))

					if ev.Operation == "insert" || ev.Operation == "update" {
						build, err := bm.Get(ev.Entity)
						if err != nil {
							log.V(logLevel).Errorf("%s:listen:> get builder err: %v", logPrefix, err)
							continue
						}

						response, err = v1.View().Build().New(build).ToJson()
						if err != nil {
							log.V(logLevel).Errorf("%s:listen:> convert struct to json err: %v", logPrefix, err)
							continue
						}
					}

					event := ""
					switch ev.Operation {
					case types.StorageInsertAction:
						event = "build:connect"
					case types.StorageUpdateAction:
						event = "build:update"
					case types.StorageDeleteAction:
						event = "build:remove"
					default:
						continue
					}

					if err := m.pool.Broadcast(defaultChannel, event, response); err != nil {
						log.V(logLevel).Errorf("%s:listen:> send broadcast build event err: %v", logPrefix, err)
						continue
					}

					continue
				}
			}
		}
	}
}
