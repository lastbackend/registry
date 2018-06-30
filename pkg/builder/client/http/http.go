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
	"github.com/lastbackend/registry/pkg/builder/client/config"
	"github.com/lastbackend/registry/pkg/builder/client/http/request"
	"github.com/lastbackend/registry/pkg/builder/client/types"
)

// RestClient captures the set of operations for generically interacting with Last.Backend REST apis.
type RestClient interface {
	Do(verb, path string) *request.Request
	Post(path string) *request.Request
	Put(path string) *request.Request
	Get(path string) *request.Request
	Delete(path string) *request.Request
}

type Client struct {
	client *request.RESTClient
}

func New(endpoint string, cfg *config.Config) (*Client, error) {

	c := new(Client)

	if cfg == nil {
		c.client = request.DefaultRESTClient(endpoint)
		return c, nil
	}

	client, err := request.NewRESTClient(endpoint, cfg)
	if err != nil {
		return nil, err
	}

	c.client = client

	return c, nil
}

func (c Client) Event() types.EventClient {
	return newEventClient(c.client)
}

func (c Client) Build() types.BuildClient {
	return newBuildClient(c.client)
}
