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
	"strings"

	"github.com/lastbackend/registry/pkg/distribution/errors"
	"github.com/lastbackend/registry/pkg/util/validator"
)

type ImageRequest struct{}

func (ImageRequest) CreateOptions() *ImageCreateOptions {
	return new(ImageCreateOptions)
}

func (i *ImageCreateOptions) Validate() *errors.Err {
	switch true {
	case len(i.Name) == 0:
		return errors.New("image").BadParameter("name")
	case len(i.Name) < 4 || len(i.Name) > 64:
		return errors.New("image").BadParameter("name")
	case !validator.IsImageName(strings.ToLower(i.Name)):
		return errors.New("image").BadParameter("name")
	case len(i.Description) > DEFAULT_DESCRIPTION_LIMIT:
		return errors.New("image").BadParameter("description")
	}
	return nil
}

func (i *ImageCreateOptions) DecodeAndValidate(reader io.Reader) *errors.Err {

	if reader == nil {
		err := errors.New("data body can not be null")
		return errors.New("image").IncorrectJSON(err)
	}

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return errors.New("image").Unknown(err)
	}

	err = json.Unmarshal(body, i)
	if err != nil {
		return errors.New("image").IncorrectJSON(err)
	}

	if err := i.Validate(); err != nil {
		return err
	}

	return nil
}

func (i *ImageCreateOptions) ToJson() ([]byte, error) {
	return json.Marshal(i)
}

func (ImageRequest) UpdateOptions() *ImageUpdateOptions {
	return new(ImageUpdateOptions)
}

func (b *ImageUpdateOptions) Validate() *errors.Err {
	switch true {
	case b.Description != nil && len(*b.Description) > DEFAULT_DESCRIPTION_LIMIT:
		return errors.New("image").BadParameter("description")
	}
	return nil
}

func (b *ImageUpdateOptions) DecodeAndValidate(reader io.Reader) *errors.Err {

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

func (b *ImageUpdateOptions) ToJson() ([]byte, error) {
	return json.Marshal(b)
}
