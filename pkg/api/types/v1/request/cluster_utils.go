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

package request

import (
	"encoding/json"
	"io"
	"io/ioutil"

	// lastbackend
	lbr "github.com/lastbackend/registry/pkg/api/types/v1/request"
	"github.com/lastbackend/registry/pkg/distribution/errors"
)

type ClusterRequest struct{}

func (ClusterRequest) CreateOptions() *ClusterCreateOptions {
	return new(ClusterCreateOptions)
}

func (c *ClusterCreateOptions) Validate() *errors.Err {

	switch true {
	case c.Name == nil:
		return errors.New("namespace").BadParameter("name")
	case c.Token == nil:
		return errors.New("namespace").BadParameter("token")
	case c.Endpoint == nil:
		return errors.New("namespace").BadParameter("endpoint")
	}

	return nil
}

func (c *ClusterCreateOptions) DecodeAndValidate(reader io.Reader) *errors.Err {

	if reader == nil {
		err := errors.New("data body can not be null")
		return errors.New("namespace").IncorrectJSON(err)
	}

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return errors.New("namespace").Unknown(err)
	}

	err = json.Unmarshal(body, c)
	if err != nil {
		return errors.New("namespace").IncorrectJSON(err)
	}

	if err := c.Validate(); err != nil {
		return err
	}

	return nil
}

func (c *ClusterCreateOptions) ToJson() ([]byte, error) {
	return json.Marshal(c)
}

func (ClusterRequest) UpdateOptions() *ClusterUpdateOptions {
	return new(ClusterUpdateOptions)
}

func (n *ClusterUpdateOptions) Validate() *errors.Err {
	switch true {
	case n.Description != nil && len(*n.Description) > lbr.DEFAULT_DESCRIPTION_LIMIT:
		return errors.New("namespace").BadParameter("description")
		// TODO: check quotas data
	}
	return nil
}

func (n *ClusterUpdateOptions) DecodeAndValidate(reader io.Reader) *errors.Err {

	if reader == nil {
		err := errors.New("data body can not be null")
		return errors.New("namespace").IncorrectJSON(err)
	}

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return errors.New("namespace").Unknown(err)
	}

	err = json.Unmarshal(body, n)
	if err != nil {
		return errors.New("namespace").IncorrectJSON(err)
	}

	if err := n.Validate(); err != nil {
		return err
	}

	return nil
}

func (n *ClusterUpdateOptions) ToJson() ([]byte, error) {
	return json.Marshal(n)
}

func (ClusterRequest) RemoveOptions() *ClusterRemoveOptions {
	return new(ClusterRemoveOptions)
}

func (n *ClusterRemoveOptions) Validate() *errors.Err {
	return nil
}

func (n *ClusterRemoveOptions) ToJson() ([]byte, error) {
	return json.Marshal(n)
}
