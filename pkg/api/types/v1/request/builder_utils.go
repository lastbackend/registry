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
	"github.com/lastbackend/lastbackend/pkg/distribution/errors"
	"io"
	"io/ioutil"
)

type BuilderRequest struct{}

func (BuilderRequest) UpdateOptions() *BuilderUpdateOptions {
	return new(BuilderUpdateOptions)
}

func (b BuilderUpdateOptions) Validate() *errors.Err {
	switch true {
	case b.WorkerLimit && b.WorkerMemory < 256:
		return errors.New("builder").BadParameter("worker.memory")
	case b.WorkerLimit && b.Workers <= 0:
		return errors.New("builder").BadParameter("workers")
	}
	return nil
}

func (b *BuilderUpdateOptions) DecodeAndValidate(reader io.Reader) *errors.Err {

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

func (b BuilderUpdateOptions) ToJson() ([]byte, error) {
	return json.Marshal(b)
}

func (BuilderRequest) ConnectOptions() *BuilderConnectOptions {
	return new(BuilderConnectOptions)
}

func (b BuilderConnectOptions) Validate() *errors.Err {
	return nil
}

func (b *BuilderConnectOptions) DecodeAndValidate(reader io.Reader) *errors.Err {

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

func (b BuilderConnectOptions) ToJson() ([]byte, error) {
	return json.Marshal(b)
}

func (BuilderRequest) CreateManifestOptions() *BuilderCreateManifestOptions {
	return new(BuilderCreateManifestOptions)
}

func (b BuilderCreateManifestOptions) Validate() *errors.Err {
	switch true {
	case len(b.PID) == 0:
		return errors.New("builder").BadParameter("pid")
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

func (b BuilderCreateManifestOptions) ToJson() ([]byte, error) {
	return json.Marshal(b)
}

func (BuilderRequest) StatusUpdateOptions() *BuilderStatusUpdateOptions {
	return new(BuilderStatusUpdateOptions)
}

func (b BuilderStatusUpdateOptions) Validate() *errors.Err {
	return nil
}

func (b *BuilderStatusUpdateOptions) DecodeAndValidate(reader io.Reader) *errors.Err {

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

func (b BuilderStatusUpdateOptions) ToJson() ([]byte, error) {
	return json.Marshal(b)
}
