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

	"github.com/lastbackend/registry/pkg/distribution/errors"
	vv1 "github.com/lastbackend/registry/pkg/api/types/v1/views"
	rv1 "github.com/lastbackend/registry/pkg/api/types/v1/request"
	"github.com/lastbackend/registry/pkg/api/client/http/request"
)

type BuilderClient struct {
	client *request.RESTClient
}

func (bc BuilderClient) Connect(ctx context.Context, hostname string) error {

	var e *errors.Http

	err := bc.client.Put(fmt.Sprintf("/builder/%s/connect", hostname)).
		AddHeader("Content-Type", "application/json").
		JSON(nil, &e)

	if err != nil {
		return err
	}
	if e != nil {
		return errors.New(e.Message)
	}

	return nil
}

func (bc BuilderClient) GetManifest(ctx context.Context, hostname string, opts *rv1.BuilderCreateManifestOptions) (*vv1.BuildManifest, error) {

	body, err := opts.ToJson()
	if err != nil {
		return nil, err
	}

	var s *vv1.BuildManifest
	var e *errors.Http

	err = bc.client.Post(fmt.Sprintf("/builder/%s/manifest", hostname)).
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

func newBuilderClient(req *request.RESTClient) BuilderClient {
	return BuilderClient{client: req}
}
