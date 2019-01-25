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
	"context"
	"github.com/lastbackend/lastbackend/pkg/distribution/errors"
	"github.com/lastbackend/registry/pkg/api/client/types"
	"github.com/lastbackend/registry/pkg/util/http/request"

	rv1 "github.com/lastbackend/registry/pkg/api/types/v1/request"
	vv1 "github.com/lastbackend/registry/pkg/api/types/v1/views"
)

type Client struct {
	client *request.RESTClient
}

func New(client *request.RESTClient) *Client {
	return &Client{client: client}
}

func (rc Client) Builder(args ...string) types.BuilderClientV1 {
	hostname := ""
	// Get any parameters passed to us out of the args variable into "real"
	// variables we created for them.
	for i := range args {
		switch i {
		case 0: // hostname
			hostname = args[0]
		default:
			panic("Wrong parameter count: (is allowed from 0 to 1)")
		}
	}
	return newBuilderClient(rc.client, hostname)
}

// Create build client with args
// Args[0] - image owner (optional), Args[1] - image name (optional)
func (rc Client) Image(args ...string) types.ImageClientV1 {

	owner := ""
	name := ""

	// Get any parameters passed to us out of the args variable into "real"
	// variables we created for them.
	for i := range args {
		switch i {
		case 0: // owner
			owner = args[0]
		case 1: // name
			name = args[1]
		default:
			panic("Wrong parameter count: (is allowed from 0 to 2)")
		}
	}

	return newImageClient(rc.client, owner, name)
}

func (rc Client) Get(ctx context.Context) (*vv1.Registry, error) {

	var s *vv1.Registry
	var e *errors.Http

	err := rc.client.Get("/registry").
		AddHeader("Content-Type", "application/json").
		JSON(&s, &e)

	if err != nil {
		return nil, err
	}
	if e != nil {
		return nil, errors.New(e.Message)
	}

	return s, nil
}

func (rc Client) Update(ctx context.Context, opts *rv1.RegistryUpdateOptions) (*vv1.Registry, error) {

	body, err := opts.ToJson()
	if err != nil {
		return nil, err
	}

	var s *vv1.Registry
	var e *errors.Http

	err = rc.client.Put("/registry").
		AddHeader("Content-Type", "application/json").
		Body(body).
		JSON(&s, &e)

	if err != nil {
		return nil, err
	}
	if e != nil {
		return nil, errors.New(e.Message)
	}

	return s, nil
}
