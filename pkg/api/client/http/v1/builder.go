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
	"github.com/lastbackend/registry/pkg/util/http/request"

	rv1 "github.com/lastbackend/registry/pkg/api/types/v1/request"
	vv1 "github.com/lastbackend/registry/pkg/api/types/v1/views"
)

type BuilderClient struct {
	client *request.RESTClient

	hostname string
}

func (bc BuilderClient) List(ctx context.Context) (*vv1.BuilderList, error) {

	var s *vv1.BuilderList
	var e *errors.Http

	err := bc.client.Get(fmt.Sprintf("/builder")).
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

func (bc BuilderClient) Connect(ctx context.Context, opts *rv1.BuilderConnectOptions) error {

	body, err := opts.ToJson()
	if err != nil {
		return err
	}

	var e *errors.Http

	err = bc.client.Put(fmt.Sprintf("/builder/%s/connect", bc.hostname)).
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

func (bc BuilderClient) SetStatus(ctx context.Context, opts *rv1.BuilderStatusUpdateOptions) error {

	body, err := opts.ToJson()
	if err != nil {
		return err
	}

	var e *errors.Http

	err = bc.client.Put(fmt.Sprintf("/builder/%s/status", bc.hostname)).
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

func (bc BuilderClient) Manifest(ctx context.Context) (*vv1.BuildManifest, error) {

	var s *vv1.BuildManifest
	var e *errors.Http

	err := bc.client.Get(fmt.Sprintf("/builder/%s/manifest", bc.hostname)).
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

func newBuilderClient(req *request.RESTClient, hostname string) *BuilderClient {
	return &BuilderClient{client: req, hostname: hostname}
}
