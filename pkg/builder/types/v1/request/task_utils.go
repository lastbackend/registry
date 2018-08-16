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

	"github.com/lastbackend/registry/pkg/distribution/errors"
)

type TaskRequest struct{}

func (TaskRequest) ExecuteOptions() *BuildLogsOptions {
	return new(BuildLogsOptions)
}

func (b *BuildLogsOptions) Validate() *errors.Err {
	return nil
}

func (b *BuildLogsOptions) DecodeAndValidate(reader io.Reader) *errors.Err {

	if reader == nil {
		err := errors.New("data body can not be null")
		return errors.New("pid").IncorrectJSON(err)
	}

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return errors.New("pid").Unknown(err)
	}

	err = json.Unmarshal(body, b)
	if err != nil {
		return errors.New("pid").IncorrectJSON(err)
	}

	if err := b.Validate(); err != nil {
		return err
	}

	return nil
}
