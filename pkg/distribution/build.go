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
	"github.com/lastbackend/registry/pkg/storage"
)

type IBuild interface {
	Create(repo *types.Repo, tag string) (*types.Build, error)
	Get(id string) (*types.Build, error)
	List(repo *types.Repo, active bool) (map[string]*types.Build, error)
	Update(id string, build *types.Build) error
	Cancel(build *types.Build) error
}

type Build struct {
	context context.Context
	storage storage.Storage
}

func (b Build) Create(repo *types.Repo, tag string) (*types.Build, error) {
	return nil, nil
}

func (b Build) Get(id string) (*types.Build, error) {
	return nil, nil
}

func (b Build) List(repo *types.Repo, active bool) (map[string]*types.Build, error) {
	return nil, nil
}

func (b Build) Update(id string, build *types.Build) error {
	return nil
}

func (b Build) Cancel(build *types.Build) error {
	return nil
}


func NewBuildModel(ctx context.Context, stg storage.Storage) IBuild {
	return &Build{
		context: ctx,
		storage: stg,
	}
}
