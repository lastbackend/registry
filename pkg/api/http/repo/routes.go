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

package repo

import (
	"github.com/lastbackend/registry/pkg/util/http"
	"github.com/lastbackend/registry/pkg/util/http/middleware"
)

var Routes = []http.Route{
	// Repo handlers
	{Path: "/repo", Method: http.MethodPost, Middleware: []http.Middleware{middleware.Context}, Handler: RepoCreateH},
	{Path: "/repo/{owner}", Method: http.MethodGet, Middleware: []http.Middleware{middleware.Context}, Handler: RepoListH},
	{Path: "/repo/{owner}/{name}", Method: http.MethodGet, Middleware: []http.Middleware{middleware.Context}, Handler: RepoInfoH},
	{Path: "/repo/{owner}/{name}", Method: http.MethodPut, Middleware: []http.Middleware{middleware.Context}, Handler: RepoUpdateH},
	{Path: "/repo/{owner}/{name}", Method: http.MethodDelete, Middleware: []http.Middleware{middleware.Context}, Handler: RepoRemoveH},
}
