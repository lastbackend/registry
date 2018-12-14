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
	"io"

	rv1 "github.com/lastbackend/registry/pkg/api/types/v1/request"
	vv1 "github.com/lastbackend/registry/pkg/api/types/v1/views"
)

type ClientV1 interface {
	Builder(args ...string) BuilderClientV1
	Image(args ...string) ImageClientV1

	Get(ctx context.Context) (*vv1.Registry, error)
	Update(ctx context.Context, opts *rv1.RegistryUpdateOptions) (*vv1.Registry, error)
}

type BuildClientV1 interface {
	Get(ctx context.Context) (*vv1.Build, error)
	List(ctx context.Context, opts *rv1.BuildListOptions) (*vv1.BuildList, error)
	Create(ctx context.Context, opts *rv1.BuildCreateOptions) (*vv1.Build, error)
	SetStatus(ctx context.Context, opts *rv1.BuildUpdateStatusOptions) error
	Logs(ctx context.Context, opts *rv1.BuildLogsOptions) (io.ReadCloser, error)
	Cancel(ctx context.Context) error
}

type BuilderClientV1 interface {
	List(ctx context.Context) (*vv1.BuilderList, error)
	Connect(ctx context.Context, opts *rv1.BuilderConnectOptions) (*vv1.BuilderConfig, error)
	Update(ctx context.Context, opts *rv1.BuilderUpdateOptions) (*vv1.Builder, error)
	SetStatus(ctx context.Context, opts *rv1.BuilderStatusUpdateOptions) error
	Manifest(ctx context.Context) (*vv1.BuildManifest, error)
}

type ImageClientV1 interface {
	Build(args ...string) BuildClientV1

	Get(ctx context.Context) (*vv1.Image, error)
	List(ctx context.Context) (*vv1.ImageList, error)
	Create(ctx context.Context, opts *rv1.ImageCreateOptions) (*vv1.Image, error)
	Update(ctx context.Context, opts *rv1.ImageUpdateOptions) (*vv1.Image, error)
	Remove(ctx context.Context, opts *rv1.ImageRemoveOptions) error
}
