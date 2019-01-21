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

package eventer

import (
	"fmt"
	"sync"

	"github.com/lastbackend/lastbackend/pkg/log"
	"github.com/lastbackend/registry/pkg/util/socket"
)

const (
	logLevel  = 3
	logPrefix = "wss:pool"
)

// Pool contains a connections used by the same id
type Pool struct {
	sync.Mutex

	ID string

	conns map[*socket.Socket]bool

	join  chan *socket.Socket
	leave chan *socket.Socket
	close chan *Pool

	ignore chan []byte

	broadcast chan []byte
}

// Listen broker channels to manage connections and broadcast data
func (p *Pool) Listen() {
	log.V(logLevel).Debugf("%s:listen:> listen broker channels to manage connections and broadcast data", logPrefix)

	go func() {
		for {
			select {
			case m := <-p.broadcast:
				log.V(logLevel).Debugf("%s:listen:> broker %s broadcast: %s", logPrefix, p.ID, string(m))

				for c := range p.conns {
					c.Write(m)
				}

			case c := <-p.join:
				log.V(logLevel).Debugf("%s:listen:> join connection to broker: %s", logPrefix, p.ID)
				p.conns[c] = true

			case c := <-p.leave:
				log.V(logLevel).Debugf("%s:listen:> leave connection from broker: %s", logPrefix, p.ID)

				delete(p.conns, c)

				if len(p.conns) == 0 {

					close(p.broadcast)
					close(p.join)
					close(p.leave)

					log.V(logLevel).Debugf("%s:listen:> broker closed successful", logPrefix)

					p.close <- p
					return
				}

			case m := <-p.ignore:
				log.V(logLevel).Debugf("%s:listen:> incoming data processing disabled: %s", logPrefix, string(m))
			}

		}
	}()

}

// Broadcast message to connections
func (p Pool) Broadcast(event, op, entity string, msg []byte) {
	log.V(logLevel).Debugf("%s:broadcast:> broadcast message to connections", logPrefix)
	p.broadcast <- []byte(fmt.Sprintf("{\"event\":\"%s\", \"operation\":\"%s\", \"entity\":\"%s\", \"payload\":%s}", event, op, entity, string(msg)))
}

// manage connection and attach it to broker
func (p Pool) Leave(s *socket.Socket) {
	log.V(logLevel).Debugf("%s:manage:> drop connection from broker", logPrefix)
	p.leave <- s
}

// Ping connection to stay it online
func (p Pool) Ping() {
	log.V(7).Debugf("%s:ping:> ping connection to stay it online", logPrefix)
	for c := range p.conns {
		c.Ping()
	}
}

// manage connection and attach it to pool
func (p Pool) manage(sock *socket.Socket) {
	p.join <- sock
}
