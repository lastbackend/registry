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

package socket

import (
	"github.com/gorilla/websocket"
	"sync"
	"time"
	"errors"
	"net/http"
	"net/url"

	"github.com/lastbackend/registry/pkg/log"
)

var ErrNotConnected = errors.New("websocket: not connected")

type Socket struct {
	mutex sync.Mutex

	reqHeader   http.Header
	httpResp    *http.Response
	dialErr     error
	isConnected bool
	endpoint    string

	dialer *websocket.Dialer

	connOpts ConnectionOptions

	*websocket.Conn
}

func New() *Socket {
	return new(Socket)
}

type ConnectionOptions struct {
	Timeout time.Duration
	// HandshakeTimeout specifies the duration for the handshake to complete,
	// default to 2 seconds
	HandshakeTimeout time.Duration
	Reconnect        bool
}

// Close and try to reconnect
func (s *Socket) closeAndReconnect() {
	s.Close()
	go func() {
		s.connect()
	}()
}

// Close websocket connection without sending or waiting for a close frame
func (s *Socket) Close() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.Conn != nil {
		s.Conn.Close()
	}

	s.isConnected = false
}

func (s *Socket) connect() {

	for {

		wsConn, httpResp, err := s.dialer.Dial(s.endpoint, nil)

		s.mutex.Lock()

		s.Conn = wsConn
		s.dialErr = err
		s.isConnected = err == nil
		s.httpResp = httpResp

		s.mutex.Unlock()

		if err == nil {
			if !s.connOpts.Reconnect {
				log.Debugf("Dial: connection was successfully established with %s", s.endpoint)
			}
			break
		} else {
			if !s.connOpts.Reconnect {
				log.Error(err)
				log.Debugf("Dial: will try again in %s seconds.", s.connOpts.Timeout)
			}
		}

		time.Sleep(s.connOpts.Timeout)
	}
}

func (s *Socket) Upgrade(conn *websocket.Conn) {

}

func (s *Socket) Dial(endpoint string, reqHeader http.Header, opts *ConnectionOptions) {

	if endpoint == "" {
		log.Error("Dial: Endpoint cannot be empty")
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		log.Error("Endpoint:", err)
	}

	if u.Scheme != "ws" && u.Scheme != "wss" {
		log.Error("Endpoint: websocket endpoint's must start with ws or wss scheme")
	}

	s.endpoint = endpoint

	if opts != nil {

		if s.connOpts.Timeout == 0 {
			s.connOpts.Timeout = 2 * time.Second
		}

		if s.connOpts.HandshakeTimeout == 0 {
			s.connOpts.HandshakeTimeout = 2 * time.Second
		}

	}

	s.dialer = websocket.DefaultDialer
	s.dialer.HandshakeTimeout = s.connOpts.HandshakeTimeout

	go func() {
		s.connect()
	}()

	// wait on first attempt
	time.Sleep(s.connOpts.HandshakeTimeout)
}

func (s *Socket) ReadMessage() (messageType int, message []byte, err error) {
	err = ErrNotConnected
	if s.IsConnected() {
		messageType, message, err = s.Conn.ReadMessage()
		if err != nil {
			s.closeAndReconnect()
		}
	}

	return
}

func (s *Socket) WriteMessage(messageType int, data []byte) error {
	err := ErrNotConnected
	if s.IsConnected() {
		err = s.Conn.WriteMessage(messageType, data)
		if err != nil {
			s.closeAndReconnect()
		}
	}

	return err
}

func (s *Socket) GetHTTPResponse() *http.Response {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.httpResp
}

func (s *Socket) GetDialError() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.dialErr
}

func (s *Socket) IsConnected() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.isConnected
}
