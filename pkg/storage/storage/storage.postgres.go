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
	"github.com/lastbackend/registry/pkg/storage/types/filter"
)

type Build interface {
	Get(ctx context.Context, id string) (*types.Build, error)
	GetByTask(ctx context.Context, id string) (*types.Build, error)
	List(ctx context.Context, image string) ([]*types.Build, error)
	Insert(ctx context.Context, build *types.Build) error
	Update(ctx context.Context, build *types.Build) error

	Attach(ctx context.Context, builder *types.Builder) (*types.Build, error)
	Unfreeze(ctx context.Context) error
}

type Builder interface {
	Get(ctx context.Context, hostname string) (*types.Builder, error)
	Insert(ctx context.Context, builder *types.Builder) error
	Update(ctx context.Context, builder *types.Builder) error

	MarkOffline(ctx context.Context) error
}

type Image interface {
	Get(ctx context.Context, owner, name string) (*types.Image, error)
	List(ctx context.Context, f *filter.ImageFilter) ([]*types.Image, error)
	Insert(ctx context.Context, image *types.Image) error
	Update(ctx context.Context, image *types.Image) error
	Remove(ctx context.Context, image *types.Image) error
}
