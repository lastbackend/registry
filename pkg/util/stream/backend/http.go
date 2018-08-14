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

package backend

import (
	"io"
	"net/http"
	"sync"
)

type Http struct {
	sync.RWMutex

	conn io.Writer

	write chan []byte

	err   chan error
	close chan bool

	end chan error

	attempt int
}

func NewHttpBackend(w io.Writer) IStreamBackend {

	var s = new(Http)

	s.write = make(chan []byte)

	s.err = make(chan error)
	s.close = make(chan bool)

	s.conn = w

	go s.manage()

	return s
}

func (h *Http) manage() {

	for {
		select {
		case m := <-h.write:
			_, err := func(p []byte) (n int, err error) {

				n, err = h.conn.Write(m)
				if err != nil {
					return n, err
				}

				if f, ok := h.conn.(http.Flusher); ok {
					f.Flush()
				}

				return n, nil

			}(m)

			if err != nil {
				h.err <- err
				return
			}
		}
	}

}

func (h *Http) send(data []byte) error {
	h.write <- data
	return nil
}

func (h *Http) disconnect() {
	h.close <- true
}

func (h *Http) Disconnect() {
	h.disconnect()
}

func (h *Http) End() error {
	return nil
}

func (h *Http) Write(chunk []byte) {
	h.write <- chunk
}
