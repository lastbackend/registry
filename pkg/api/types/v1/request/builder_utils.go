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

type BuilderRequest struct{}

func (BuilderRequest) ConnectOptions() *BuilderInfoOptions {
	return new(BuilderInfoOptions)
}

func (b *BuilderInfoOptions) Validate() *errors.Err {
	switch true {
	case len(b.Hostname) == 0:
		return errors.New("builder").BadParameter("hostname")
	}
	return nil
}

func (b *BuilderInfoOptions) DecodeAndValidate(reader io.Reader) *errors.Err {

	if reader == nil {
		err := errors.New("data body can not be null")
		return errors.New("builder").IncorrectJSON(err)
	}

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return errors.New("builder").Unknown(err)
	}

	err = json.Unmarshal(body, b)
	if err != nil {
		return errors.New("builder").IncorrectJSON(err)
	}

	if err := b.Validate(); err != nil {
		return err
	}

	return nil
}

func (i *BuilderInfoOptions) ToJson() ([]byte, error) {
	return json.Marshal(i)
}

func (BuilderRequest) CreateManifestOptions() *BuilderCreateManifestOptions {
	return new(BuilderCreateManifestOptions)
}

func (b *BuilderCreateManifestOptions) Validate() *errors.Err {
	switch true {
	case len(b.TaskID) == 0:
		return errors.New("builder").BadParameter("task")
	}
	return nil
}

func (b *BuilderCreateManifestOptions) DecodeAndValidate(reader io.Reader) *errors.Err {

	if reader == nil {
		err := errors.New("data body can not be null")
		return errors.New("builder").IncorrectJSON(err)
	}

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return errors.New("builder").Unknown(err)
	}

	err = json.Unmarshal(body, b)
	if err != nil {
		return errors.New("builder").IncorrectJSON(err)
	}

	if err := b.Validate(); err != nil {
		return err
	}

	return nil
}

func (i *BuilderCreateManifestOptions) ToJson() ([]byte, error) {
	return json.Marshal(i)
}