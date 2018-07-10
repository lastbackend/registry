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
	"strconv"

	"github.com/lastbackend/registry/pkg/distribution/errors"
	"github.com/lastbackend/registry/pkg/api/client/http/request"
	rv1 "github.com/lastbackend/registry/pkg/api/types/v1/request"
	vv1 "github.com/lastbackend/registry/pkg/api/types/v1/views"
)

type ImageClient struct {
	client *request.RESTClient
	owner  string
	name   string
}

func (ic ImageClient) List(ctx context.Context) (*vv1.ImageList, error) {

	var s *vv1.ImageList
	var e *errors.Http

	err := ic.client.Get(fmt.Sprintf("/image")).
		AddHeader("Content-Type", "application/json").
		JSON(&s, &e)

	if err != nil {
		return nil, err
	}
	if e != nil {
		return nil, errors.New(e.Message)
	}

	if s == nil {
		list := make(vv1.ImageList, 0)
		s = &list
	}

	return s, nil
}

func (ic ImageClient) Create(ctx context.Context, opts *rv1.ImageCreateOptions) (*vv1.Image, error) {

	body, err := opts.ToJson()
	if err != nil {
		return nil, err
	}

	var s *vv1.Image
	var e *errors.Http

	err = ic.client.Post("/image").
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

func (ic ImageClient) Get(ctx context.Context) (*vv1.Image, error) {

	var s *vv1.Image
	var e *errors.Http

	err := ic.client.Get(fmt.Sprintf("/image/%s/%s", ic.owner, ic.name)).
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

func (ic ImageClient) Update(ctx context.Context, opts *rv1.ImageUpdateOptions) (*vv1.Image, error) {

	body, err := opts.ToJson()
	if err != nil {
		return nil, err
	}

	var s *vv1.Image
	var e *errors.Http

	err = ic.client.Put(fmt.Sprintf("/image/%s/%s", ic.owner, ic.name)).
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

func (ic ImageClient) Remove(ctx context.Context, opts *rv1.ImageRemoveOptions) error {

	req := ic.client.Delete(fmt.Sprintf("/image/%s/%s", ic.owner, ic.name)).
		AddHeader("Content-Entity", "application/json")

	if opts != nil {
		if opts.Force {
			req.Param("force", strconv.FormatBool(opts.Force))
		}
	}

	var e *errors.Http

	if err := req.JSON(nil, &e); err != nil {
		return err
	}
	if e != nil {
		return errors.New(e.Message)
	}

	return nil
}

func (ic ImageClient) BuildList(ctx context.Context) (*vv1.BuildList, error) {

	var s *vv1.BuildList
	var e *errors.Http

	err := ic.client.Get(fmt.Sprintf("/image/%s/%s/build", ic.owner, ic.name)).
		AddHeader("Content-Type", "application/json").
		JSON(&s, &e)

	if err != nil {
		return nil, err
	}
	if e != nil {
		return nil, errors.New(e.Message)
	}

	if s == nil {
		list := make(vv1.BuildList, 0)
		s = &list
	}

	return s, nil
}

func newImageClient(req *request.RESTClient, owner, name string) ImageClient {
	return ImageClient{client: req, owner: owner, name: name}
}
