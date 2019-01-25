package eventer

import (
	"context"
	"sync"

	"github.com/lastbackend/registry/pkg/util/socket"
)

type Emitter struct {
	mu             sync.Mutex
	handlers       map[string][]HandleFunc
	defaultHandler HandleFunc
}

type HandleFunc func(ctx context.Context, event string, sock *socket.Socket, pools map[string]*Pool)

func newEmitter() *Emitter {
	e := new(Emitter)
	e.handlers = make(map[string][]HandleFunc, 0)
	e.defaultHandler = func(ctx context.Context, event string, sock *socket.Socket, pools map[string]*Pool) {}
	return e
}

func (e *Emitter) SetDefaultHandler(handler HandleFunc) {
	e.defaultHandler = handler
}

func (e *Emitter) AddHandler(name string, handlers ...HandleFunc) {
	if len(handlers) == 0 {
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if e.handlers == nil {
		e.handlers = make(map[string][]HandleFunc, 0)
	}

	h := e.handlers[name]

	if h == nil {
		h = make([]HandleFunc, 0)
	}

	e.handlers[name] = append(h, handlers...)
}

func (e *Emitter) Remove(name string) bool {
	if e.handlers == nil {
		return false
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if h := e.handlers[name]; h != nil {
		delete(e.handlers, name)
		return true
	}
	return false
}

func (e *Emitter) Clear() {
	e.handlers = make(map[string][]HandleFunc, 0)
}

func (e Emitter) Call(ctx context.Context, name string, sock *socket.Socket, pools map[string]*Pool) {
	if e.handlers == nil {
		return
	}

	if h := e.handlers[name]; h != nil && len(h) > 0 {
		for i := range h {
			l := h[i]
			if l != nil {
				l(ctx, name, sock, pools)
			}
		}
	} else {
		e.defaultHandler(ctx, name, sock, pools)
	}
}
