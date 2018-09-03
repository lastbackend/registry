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

package request

const (
	DEFAULT_DESCRIPTION_LIMIT = 512
)

type Request struct{}

func New() *Request {
	return new(Request)
}

func (Request) Image() *ImageRequest {
	return new(ImageRequest)
}

func (Request) Build() *BuildRequest {
	return new(BuildRequest)
}

func (Request) Builder() *BuilderRequest {
	return new(BuilderRequest)
}

func (Request) Registry() *RegistryRequest {
	return new(RegistryRequest)
}
