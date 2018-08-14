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

package v1

import (
	"context"
	"fmt"

	"github.com/lastbackend/registry/pkg/builder/client/types"
	vv1 "github.com/lastbackend/registry/pkg/builder/types/v1/views"
	"github.com/lastbackend/registry/pkg/distribution/errors"
	"github.com/lastbackend/registry/pkg/util/http/request"
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

func (bc Client) Status(ctx context.Context) (*vv1.Builder, error) {

	var s *vv1.Builder
	var e *errors.Http

	err := bc.client.Get(fmt.Sprintf("/status")).
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
