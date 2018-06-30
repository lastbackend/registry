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

package storage

import (
	"context"

	"github.com/lastbackend/registry/pkg/distribution/types"
)

type Build interface {
	Get(ctx context.Context, id string) (*types.Build, error)
	List(ctx context.Context, repo string) (map[string]*types.Build, error)
	Insert(ctx context.Context, build *types.Build) error
	Update(ctx context.Context, build *types.Build) error
}

type Repo interface {
	Get(ctx context.Context, owner, name string) (*types.Repo, error)
	List(ctx context.Context) (map[string]*types.Repo, error)
	Insert(ctx context.Context, repo *types.Repo) error
	Update(ctx context.Context, repo *types.Repo) error
	Remove(ctx context.Context, id string) error

	InsertTag(ctx context.Context, tag *types.RepoTag) error
	UpdateTag(ctx context.Context, repo, tag string) error
}
