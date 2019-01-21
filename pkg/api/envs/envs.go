//
// Last.Backend LLC CONFIDENTIAL
// __________________
//
// [2014] - [2019] Last.Backend LLC
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
	"github.com/lastbackend/registry/pkg/notifier"
	"github.com/lastbackend/registry/pkg/storage"
	"github.com/lastbackend/registry/pkg/util/blob"
)

var e Env

type Env struct {
	storage     storage.IStorage
	blobStorage blob.IBlobStorage
	notifier    notifier.INotifier
}

func Get() *Env {
	return &e
}

func (env *Env) SetStorage(storage storage.IStorage) {
	env.storage = storage
}

func (env *Env) GetStorage() storage.IStorage {
	return env.storage
}

func (env *Env) SetBlobStorage(u blob.IBlobStorage) {
	env.blobStorage = u
}

func (env Env) GetBlobStorage() blob.IBlobStorage {
	return env.blobStorage
}

func (env *Env) SetNotifier(u notifier.INotifier) {
	env.notifier = u
}

func (env Env) GetNotifier() notifier.INotifier {
	return env.notifier
}
