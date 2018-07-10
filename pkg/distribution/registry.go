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
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/storage"

		"github.com/spf13/viper"
)

const (
	logRegistryPrefix = "distribution:registry"
)

// Registry - distribution model
type Registry struct {
	context context.Context
	storage storage.Storage
}

// Info - get registry info
func (c *Registry) Get() (*types.Registry, error) {

	log.V(logLevel).Debugf("%s:get:> get info", logRegistryPrefix)

	registry := new(types.Registry)
	registry.Meta.Hostname = viper.GetString("domain")

	return registry, nil
}

// NewRegistryModel - return new registry model
func NewRegistryModel(ctx context.Context, stg storage.Storage) *Registry {
	return &Registry{ctx, stg}
}
