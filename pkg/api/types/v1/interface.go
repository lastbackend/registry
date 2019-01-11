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

package v1

import (
	"github.com/lastbackend/registry/pkg/api/types/v1/request"
	"github.com/lastbackend/registry/pkg/api/types/v1/views"
)

type IRequest interface {
	Image() *request.ImageRequest
	Build() *request.BuildRequest
	Builder() *request.BuilderRequest
	Registry() *request.RegistryRequest
	Event() *request.EventRequest
}

type IView interface {
	Build() *views.BuildView
	Builder() *views.BuilderView
	Image() *views.ImageView
	Registry() *views.RegistryView
}
