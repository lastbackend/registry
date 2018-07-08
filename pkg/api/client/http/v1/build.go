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
	"github.com/lastbackend/registry/pkg/api/client/http/request"
	rv1 "github.com/lastbackend/registry/pkg/api/types/v1/request"
	"github.com/lastbackend/registry/pkg/distribution/errors"
)

type BuildClient struct {
	client *request.RESTClient
}

func (bc BuildClient) SetStatus(ctx context.Context, task string, opts *rv1.BuildUpdateStatusOptions) error {

	body, err := opts.ToJson()
	if err != nil {
		return err
	}

	var e *errors.Http

	err = bc.client.Put(fmt.Sprintf("/build/task/%s/status", task)).
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

func (bc BuildClient) SetImageInfo(ctx context.Context, task string, opts *rv1.BuildUpdateImageInfoOptions) error {

	body, err := opts.ToJson()
	if err != nil {
		return err
	}

	var e *errors.Http

	err = bc.client.Put(fmt.Sprintf("/build/task/%s/info", task)).
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

func newBuildClient(req *request.RESTClient) BuildClient {
	return BuildClient{client: req}
}
