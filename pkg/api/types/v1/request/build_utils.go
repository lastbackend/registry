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

type BuildRequest struct{}

func (BuildRequest) BuildExecuteOptions() *BuildCreateOptions {
	return new(BuildCreateOptions)
}

func (b *BuildCreateOptions) Validate() *errors.Err {
	switch true {
	case len(b.Tag) == 0:
		return errors.New("image").BadParameter("tag")
	case len(b.Source.Hub) == 0:
		return errors.New("source").BadParameter("hub")
	case len(b.Source.Owner) == 0:
		return errors.New("source").BadParameter("owner")
	case len(b.Source.Name) == 0:
		return errors.New("source").BadParameter("name")
	}
	return nil
}

func (b *BuildCreateOptions) DecodeAndValidate(reader io.Reader) *errors.Err {

	if reader == nil {
		err := errors.New("data body can not be null")
		return errors.New("image").IncorrectJSON(err)
	}

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return errors.New("image").Unknown(err)
	}

	err = json.Unmarshal(body, b)
	if err != nil {
		return errors.New("image").IncorrectJSON(err)
	}

	if err := b.Validate(); err != nil {
		return err
	}

	return nil
}

func (b *BuildCreateOptions) ToJson() ([]byte, error) {
	return json.Marshal(b)
}

func (BuildRequest) BuildStatusOptions() *BuildUpdateStatusOptions {
	return new(BuildUpdateStatusOptions)
}

func (b *BuildUpdateStatusOptions) Validate() *errors.Err {
	return nil
}

func (b *BuildUpdateStatusOptions) DecodeAndValidate(reader io.Reader) *errors.Err {

	if reader == nil {
		err := errors.New("data body can not be null")
		return errors.New("image").IncorrectJSON(err)
	}

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return errors.New("image").Unknown(err)
	}

	err = json.Unmarshal(body, b)
	if err != nil {
		return errors.New("image").IncorrectJSON(err)
	}

	if err := b.Validate(); err != nil {
		return err
	}

	return nil
}

func (b *BuildUpdateStatusOptions) ToJson() ([]byte, error) {
	return json.Marshal(b)
}