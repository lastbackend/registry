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
	"github.com/lastbackend/registry/pkg/api/client/config"
	"github.com/lastbackend/registry/pkg/api/client/http/request"
	"github.com/lastbackend/registry/pkg/api/client/http/v1"
	"github.com/lastbackend/registry/pkg/api/client/types"
)

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

func (c Client) V1() types.ClientV1 {
	return v1.New(c.client)
}
