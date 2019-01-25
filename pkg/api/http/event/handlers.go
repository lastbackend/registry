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

package event

import (
	"github.com/gorilla/websocket"
	"github.com/lastbackend/registry/pkg/api/envs"
	"net/http"

	"github.com/lastbackend/lastbackend/pkg/log"
)

const (
	logLevel  = 2
	logPrefix = "registry:api:handler:event"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func EventsSubscribeH(w http.ResponseWriter, r *http.Request) {

	log.V(logLevel).Debugf("%s:events_subscribe:> incoming connection", logPrefix)

	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	log.V(logLevel).Debugf("%s:events_subscribe:> upgrade connection", logPrefix)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.V(logLevel).Errorf("%s:events_subscribe:> upgrade connection err: %v", err)
		return
	}

	envs.Get().GetMonitor().Subscribe(conn)
}
