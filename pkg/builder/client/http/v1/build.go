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
	"strconv"

	"github.com/lastbackend/registry/pkg/builder/client/types"
	rv1 "github.com/lastbackend/registry/pkg/builder/types/v1/request"
	"github.com/lastbackend/registry/pkg/distribution/errors"
	"github.com/lastbackend/registry/pkg/util/http/request"
)

type BuildClient struct {
	client *request.RESTClient

	pid string
}

func (bc BuildClient) Logs(ctx context.Context, opts *rv1.BuildLogsOptions) (io.ReadCloser, error) {

	res := bc.client.Get(fmt.Sprintf("/build/%s/logs", bc.pid))

	if opts != nil {
		if opts.Follow {
			res.Param("follow", strconv.FormatBool(opts.Follow))
		}
	}

	return res.Stream()
}

func (bc BuildClient) Cancel(ctx context.Context) error {

	var e *errors.Http

	err := bc.client.Put(fmt.Sprintf("/build/%s/cancel", bc.pid)).
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

func newBuildClient(req *request.RESTClient, pid string) types.BuildClientV1 {
	return BuildClient{client: req, pid: pid}
}
