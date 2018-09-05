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
	"io"
	"net/url"
	"strconv"

	"github.com/lastbackend/lastbackend/pkg/distribution/errors"
	"github.com/lastbackend/registry/pkg/util/http/request"

	rv1 "github.com/lastbackend/registry/pkg/api/types/v1/request"
	vv1 "github.com/lastbackend/registry/pkg/api/types/v1/views"
)

type BuildClient struct {
	client *request.RESTClient

	owner string
	name  string
	id    string
}

func (bc BuildClient) Create(ctx context.Context, opts *rv1.BuildCreateOptions) (*vv1.Build, error) {

	body, err := opts.ToJson()
	if err != nil {
		return nil, err
	}

	var s *vv1.Build
	var e *errors.Http

	err = bc.client.Post(fmt.Sprintf("/image/%s/%s/build", bc.owner, bc.name)).
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

func (bc BuildClient) Get(ctx context.Context) (*vv1.Build, error) {

	var s *vv1.Build
	var e *errors.Http

	u := url.URL{}
	u.Path = fmt.Sprintf("/image/%s/%s/build/%s", bc.owner, bc.name, bc.id)

	req := bc.client.Get(u.String())

	err := req.
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

func (bc BuildClient) List(ctx context.Context, opts *rv1.BuildListOptions) (*vv1.BuildList, error) {

	var s *vv1.BuildList
	var e *errors.Http

	u := url.URL{}
	u.Path = fmt.Sprintf("/image/%s/%s/build", bc.owner, bc.name)

	req := bc.client.Get(u.String())

	if opts != nil {
		if opts.Active != nil {
			req.Param("active", strconv.FormatBool(*opts.Active))
		}
	}

	err := req.
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

func (bc BuildClient) SetStatus(ctx context.Context, opts *rv1.BuildUpdateStatusOptions) error {

	body, err := opts.ToJson()
	if err != nil {
		return err
	}

	var e *errors.Http

	err = bc.client.Put(fmt.Sprintf("/image/%s/%s/build/%s/status", bc.owner, bc.name, bc.id)).
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

func (bc BuildClient) SetImageInfo(ctx context.Context, opts *rv1.BuildSetImageInfoOptions) error {

	body, err := opts.ToJson()
	if err != nil {
		return err
	}

	var e *errors.Http

	err = bc.client.Put(fmt.Sprintf("/image/%s/%s/build/%s/info", bc.owner, bc.name, bc.id)).
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

func (bc BuildClient) Logs(ctx context.Context, opts *rv1.BuildLogsOptions) (io.ReadCloser, error) {

	res := bc.client.Get(fmt.Sprintf("/image/%s/%s/build/%s/logs", bc.owner, bc.name, bc.id))

	if opts != nil {
		if opts.Follow {
			res.Param("follow", strconv.FormatBool(opts.Follow))
		}
	}

	return res.Stream()
}

func (bc BuildClient) Cancel(ctx context.Context) error {

	var e *errors.Http

	err := bc.client.Put(fmt.Sprintf("/image/%s/%s/build/%s/cancel", bc.owner, bc.name, bc.id)).
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

func newBuildClient(req *request.RESTClient, owner, name, id string) *BuildClient {
	return &BuildClient{client: req, owner: owner, name: name, id: id}
}
