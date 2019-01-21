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

package notifier

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

type Notifier struct {
	sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc

	pool *eventer.Controller

	storage storage.IStorage
}

type INotifier interface {
	Stop()
	Attach(conn *websocket.Conn)
}

const (
	logLevel  = 5
	logPrefix = "notifier"
)

func New(storage storage.IStorage) *Notifier {
	var n = new(Notifier)

	n.ctx, n.cancel = context.WithCancel(context.Background())

	n.pool = eventer.New()

	n.storage = storage

	go n.listen()

	return n
}

func (n Notifier) Attach(conn *websocket.Conn) {
	sock := socket.New(context.Background(), conn, n.pool.Leave, n.pool.Message)
	p := n.pool.Get(defaultChannel)
	if p == nil {
		p = n.pool.Add(defaultChannel, sock)
	} else {
		n.pool.Attach(p, sock)
	}
}

func (n Notifier) Stop() {
	n.cancel()
}

func (n Notifier) listen() {
	var bm = distribution.NewBuildModel(n.ctx, n.storage)
	var bdm = distribution.NewBuilderModel(n.ctx, n.storage)
	var evs = make(chan string)

	go func() {
		err := n.storage.Listen(n.ctx, "e_watch", evs)
		if err != nil {
			log.V(logLevel).Errorf("%s:listen:>  subscribe on events err: %v", logPrefix, err)
			return
		}
	}()

	for {
		select {
		case <-n.ctx.Done():
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

					if err := n.pool.Broadcast(defaultChannel, "update", "builder", response); err != nil {
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

					if err := n.pool.Broadcast(defaultChannel, event.Operation, "build", response); err != nil {
						log.V(logLevel).Errorf("%s:listen:> send broadcast builder event err: %v", logPrefix, err)
						continue
					}

					continue
				}
			}
		}
	}
}
