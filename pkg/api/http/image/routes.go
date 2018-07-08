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

package image

import (
	"github.com/lastbackend/registry/pkg/util/http"
	"github.com/lastbackend/registry/pkg/util/http/middleware"
)

var Routes = []http.Route{
	// Image handlers
	{Path: "/image", Method: http.MethodPost, Middleware: []http.Middleware{middleware.Context}, Handler: ImageCreateH},
	{Path: "/image/{owner}", Method: http.MethodGet, Middleware: []http.Middleware{middleware.Context}, Handler: ImageListH},
	{Path: "/image/{owner}/{name}", Method: http.MethodGet, Middleware: []http.Middleware{middleware.Context}, Handler: ImageInfoH},
	{Path: "/image/{owner}/{name}", Method: http.MethodPut, Middleware: []http.Middleware{middleware.Context}, Handler: ImageUpdateH},
	{Path: "/image/{owner}/{name}", Method: http.MethodDelete, Middleware: []http.Middleware{middleware.Context}, Handler: ImageRemoveH},
	{Path: "/image/{owner}/{name}/build", Method: http.MethodGet, Middleware: []http.Middleware{middleware.Context}, Handler: ImageBuildListH},
}
