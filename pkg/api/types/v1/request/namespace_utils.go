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
	"github.com/lastbackend/registry/pkg/util/validator"
)

type NamespaceRequest struct{}

func (NamespaceRequest) CreateOptions() *NamespaceCreateOptions {
	return new(NamespaceCreateOptions)
}

func (n *NamespaceCreateOptions) Validate() *errors.Err {

	if err := n.NamespaceCreateOptions.Validate(); err != nil {
		return err
	}

	switch true {
	case n.Cluster == nil || !validator.IsClusterName(*n.Cluster):
		return errors.New("namespace").BadParameter("cluster")
	}

	return nil
}

func (n *NamespaceCreateOptions) DecodeAndValidate(reader io.Reader) *errors.Err {

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

func (n *NamespaceCreateOptions) ToJson() ([]byte, error) {
	return json.Marshal(n)
}

func (NamespaceRequest) UpdateOptions() *NamespaceUpdateOptions {
	return new(NamespaceUpdateOptions)
}

func (n *NamespaceUpdateOptions) Validate() *errors.Err {
	switch true {
	case n.Description != nil && len(*n.Description) > lbr.DEFAULT_DESCRIPTION_LIMIT:
		return errors.New("namespace").BadParameter("description")
		// TODO: check quotas data
	}
	return nil
}

func (n *NamespaceUpdateOptions) DecodeAndValidate(reader io.Reader) *errors.Err {

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

func (n *NamespaceUpdateOptions) ToJson() ([]byte, error) {
	return json.Marshal(n)
}

func (NamespaceRequest) RemoveOptions() *NamespaceRemoveOptions {
	return new(NamespaceRemoveOptions)
}

func (n *NamespaceRemoveOptions) Validate() *errors.Err {
	return nil
}

func (n *NamespaceRemoveOptions) ToJson() ([]byte, error) {
	return json.Marshal(n)
}
