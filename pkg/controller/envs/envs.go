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

package envs

import (
	"github.com/lastbackend/registry/pkg/storage"
)

var e Env

type Env struct {
	storage storage.IStorage
}

func Get() *Env {
	return &e
}

func (c *Env) SetStorage(storage storage.IStorage) {
	c.storage = storage
}

func (c *Env) GetStorage() storage.IStorage {
	return c.storage
}
