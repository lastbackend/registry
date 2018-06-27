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

type IRepo interface {
	Get(id string) (*types.Repo, error)
	List() (map[string]*types.Repo, error)
	Create(opts *types.RepoCreateOptions) (*types.Repo, error)
	Update(repo *types.Repo, opts *types.RepoUpdateOptions) error
	Remove(id string) error
}

type Repo struct {
	context context.Context
	storage storage.Storage
}

func (r Repo) Get(id string) (*types.Repo, error) {
	return nil, nil
}

func (r Repo) List() (map[string]*types.Repo, error) {
	return nil, nil
}

func (r Repo) Create(opts *types.RepoCreateOptions) (*types.Repo, error) {
	return nil, nil
}

func (r Repo) Update(repo *types.Repo, opts *types.RepoUpdateOptions) error {
	return nil
}

func (r Repo) Remove(id string) error {
	return nil
}

func NewRepoModel(ctx context.Context, stg storage.Storage) IRepo {
	return &Repo{ctx, stg}
}
