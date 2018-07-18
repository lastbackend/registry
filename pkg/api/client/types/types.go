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

package types

import (
	"context"

	rv1 "github.com/lastbackend/registry/pkg/api/types/v1/request"
	vv1 "github.com/lastbackend/registry/pkg/api/types/v1/views"
)

type ClientV1 interface {
	Build() BuildClientV1
	Builder() BuilderClientV1
	Registry() RegistryClientV1
	Image(owner, name string) ImageClientV1
}

type BuildClientV1 interface {
	SetStatus(ctx context.Context, task string, opts *rv1.BuildUpdateStatusOptions) error
	SetImageInfo(ctx context.Context, task string, opts *rv1.BuildUpdateImageInfoOptions) error
	Create(ctx context.Context, opts *rv1.BuildCreateOptions) (*vv1.Build, error)
}

type BuilderClientV1 interface {
	Connect(ctx context.Context, hostname string) error
	Disconnect(ctx context.Context, hostname string) error
	GetManifest(ctx context.Context, hostname string, opts *rv1.BuilderCreateManifestOptions) (*vv1.BuildManifest, error)
}

type ImageClientV1 interface {
	Create(ctx context.Context, opts *rv1.ImageCreateOptions) (*vv1.Image, error)
	List(ctx context.Context) (*vv1.ImageList, error)
	Get(ctx context.Context) (*vv1.Image, error)
	Update(ctx context.Context, opts *rv1.ImageUpdateOptions) (*vv1.Image, error)
	Remove(ctx context.Context, opts *rv1.ImageRemoveOptions) error
	BuildList(ctx context.Context) (*vv1.BuildList, error)
}

type RegistryClientV1 interface {
	Get(ctx context.Context) (*vv1.Registry, error)
}
