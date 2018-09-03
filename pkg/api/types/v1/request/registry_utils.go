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
	"github.com/lastbackend/registry/pkg/distribution/errors"
	"io"
	"io/ioutil"
)

type RegistryRequest struct{}

func (RegistryRequest) UpdateOptions() *RegistryUpdateOptions {
	return new(RegistryUpdateOptions)
}

func (r *RegistryUpdateOptions) Validate() *errors.Err {
	return nil
}

func (r *RegistryUpdateOptions) DecodeAndValidate(reader io.Reader) *errors.Err {

	if reader == nil {
		err := errors.New("data body can not be null")
		return errors.New("registry").IncorrectJSON(err)
	}

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return errors.New("registry").Unknown(err)
	}

	err = json.Unmarshal(body, r)
	if err != nil {
		return errors.New("registry").IncorrectJSON(err)
	}

	if err := r.Validate(); err != nil {
		return err
	}

	return nil
}

func (r *RegistryUpdateOptions) ToJson() ([]byte, error) {
	return json.Marshal(r)
}
