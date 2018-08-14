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

package builder

import (
	"github.com/lastbackend/registry/pkg/util/http"
)

var Routes = []http.Route{
	// Builder handlers
	{Path: "/builder", Method: http.MethodGet, Handler: BuilderListH},
	{Path: "/builder/{builder}", Method: http.MethodPut, Handler: BuilderUpdateH},
	{Path: "/builder/{builder}/connect", Method: http.MethodPut, Handler: BuilderConnectH},
	{Path: "/builder/{builder}/status", Method: http.MethodPut, Handler: BuilderStatusH},
	{Path: "/builder/{builder}/manifest", Method: http.MethodPost, Handler: BuilderCreateManifestH},
}
