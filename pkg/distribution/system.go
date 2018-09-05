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

package distribution

import (
	"context"
	"github.com/lastbackend/lastbackend/pkg/distribution/errors"
	"github.com/lastbackend/lastbackend/pkg/log"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/storage"
)

const (
	logSystemPrefix = "distribution:system"
)

type ISystem interface {
	Get() (*types.System, error)
	Update(system *types.System, opts *types.SystemUpdateOptions) error
}

type System struct {
	context context.Context
	storage storage.IStorage
}

func (i System) Get() (*types.System, error) {
	log.V(logLevel).Debugf("%s:get:> get system", logSystemPrefix)

	image, err := i.storage.System().Get(i.context)
	if err != nil {
		log.V(logLevel).Debugf("%s:get:> get system err: %v", logSystemPrefix, err)
		return nil, err
	}

	return image, nil
}

func (i System) Update(system *types.System, opts *types.SystemUpdateOptions) error {

	if system == nil {
		return errors.New("invalid argument")
	}

	if opts == nil {
		opts = new(types.SystemUpdateOptions)
	}

	log.V(logLevel).Debugf("%s:update:> update system %#v", logSystemPrefix, opts)

	if opts.AccessToken != nil {
		system.AccessToken = *opts.AccessToken
	}

	if opts.AuthServer != nil {
		system.AuthServer = *opts.AuthServer
	}

	if err := i.storage.System().Update(i.context, system); err != nil {
		log.V(logLevel).Errorf("%s:update:> update image err: %v", logSystemPrefix, err)
		return err
	}

	return nil
}

func NewSystemModel(ctx context.Context, stg storage.IStorage) ISystem {
	return &System{ctx, stg}
}
