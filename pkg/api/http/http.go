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

package http

import (
	"github.com/gorilla/mux"
	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/util/http"
	"github.com/lastbackend/registry/pkg/util/http/cors"
	"github.com/lastbackend/registry/pkg/api/http/build"
	"github.com/lastbackend/registry/pkg/api/http/repo"
)

const logLevel = 2

// Extends routes variable
var Routes = make([]http.Route, 0)

func AddRoutes(r ...[]http.Route) {
	for i := range r {
		Routes = append(Routes, r[i]...)
	}
}

func init() {

	// Registry
	AddRoutes(repo.Routes)
	AddRoutes(build.Routes)

}

func Listen(host string, port int) error {

	log.V(logLevel).Debugf("HTTP: listen HTTP server on %s:%d", host, port)

	r := mux.NewRouter()
	r.Methods("OPTIONS").HandlerFunc(cors.Headers)

	var notFound http.MethodNotAllowedHandler
	r.NotFoundHandler = notFound

	var notAllowed http.MethodNotAllowedHandler
	r.MethodNotAllowedHandler = notAllowed

	for _, route := range Routes {
		log.V(logLevel).Debugf("HTTP: init route: %s", route.Path)
		r.Handle(route.Path, http.Handle(route.Handler, route.Middleware...)).Methods(route.Method)
	}

	return http.Listen(host, port, r)
}
