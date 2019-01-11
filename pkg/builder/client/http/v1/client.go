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
	"github.com/lastbackend/registry/pkg/builder/client/types"
	"github.com/lastbackend/registry/pkg/util/http/request"

	rv1 "github.com/lastbackend/registry/pkg/builder/types/v1/request"
)

type Client struct {
	client *request.RESTClient
}

func New(client *request.RESTClient) *Client {
	return &Client{client: client}
}

// Create build client with args
func (bc Client) Build(id string) types.BuildClientV1 {
	return newBuildClient(bc.client, id)
}

func (bc Client) Update(ctx context.Context, opts *rv1.BuilderUpdateManifestOptions) error {

	body, err := opts.ToJson()
	if err != nil {
		return err
	}

	var e *errors.Http

	err = bc.client.Put("/settings").
		AddHeader("Content-Type", "application/json").
		Body(body).
		JSON(nil, &e)

	if err != nil {
		return err
	}
	if e != nil {
		return errors.New(e.Message)
	}

	return nil
}