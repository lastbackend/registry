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

package distribution

import (
	"context"
	"github.com/lastbackend/lastbackend/pkg/distribution/errors"
	"github.com/lastbackend/lastbackend/pkg/log"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/storage"
)

const (
	logBuilderPrefix = "distribution:builder"
)

type IBuilder interface {
	Get(hostname string) (*types.Builder, error)
	List() ([]*types.Builder, error)
	Create(opts *types.BuilderCreateOptions) (*types.Builder, error)
	Update(builder *types.Builder, opts *types.BuilderUpdateOptions) error
	FindBuild(builder *types.Builder) (*types.Build, error)
	MarkOffline() error
}

type Builder struct {
	context context.Context
	storage storage.IStorage
}

func (b *Builder) Get(hostname string) (*types.Builder, error) {

	log.V(logLevel).Debugf("%s:get:> get by hostname %s", logBuilderPrefix, hostname)

	builder, err := b.storage.Builder().Get(b.context, hostname)
	if err != nil {
		log.V(logLevel).Errorf("%s:get:> get builder %s err: %v", logBuilderPrefix, hostname, err)
		return nil, err
	}

	return builder, nil
}

func (b *Builder) List() ([]*types.Builder, error) {

	log.V(logLevel).Debugf("%s:list:> get builders list", logBuilderPrefix)

	builder, err := b.storage.Builder().List(b.context, nil)
	if err != nil {
		log.V(logLevel).Errorf("%s:list:> get builders list err: %v", logBuilderPrefix, err)
		return nil, err
	}

	return builder, nil
}

func (b *Builder) Create(opts *types.BuilderCreateOptions) (*types.Builder, error) {

	log.V(logLevel).Debugf("%s:create:> create new builder %#v", logBuilderPrefix, opts)

	if opts == nil {
		opts = new(types.BuilderCreateOptions)
	}

	builder := new(types.Builder)
	builder.Meta.Hostname = opts.Hostname
	builder.Status.Online = true
	builder.Spec.Network.IP = opts.IP
	builder.Spec.Network.Port = opts.Port
	builder.Spec.Network.TLS = opts.TLS

	if opts.SSL != nil {
		builder.Spec.Network.SSL = new(types.SSL)
		builder.Spec.Network.SSL = opts.SSL
	}

	if err := b.storage.Builder().Insert(b.context, builder); err != nil {
		log.V(logLevel).Errorf("%s:create:> insert builder %s err: %v", logBuilderPrefix, builder.Meta.Hostname, err)
		return nil, err
	}

	return builder, nil
}

func (b *Builder) Update(builder *types.Builder, opts *types.BuilderUpdateOptions) error {

	if builder == nil {
		return errors.New("invalid argument")
	}

	if opts == nil {
		opts = new(types.BuilderUpdateOptions)
	}

	log.V(logLevel).Debugf("%s:update:> update builder %s -> %#v", logBuilderPrefix, builder.Meta.Hostname, opts)

	if opts.Hostname != nil {
		builder.Meta.Hostname = *opts.Hostname
	}

	if opts.IP != nil {
		builder.Spec.Network.IP = *opts.IP
	}

	if opts.Port != nil {
		builder.Spec.Network.Port = *opts.Port
	}

	if opts.Online != nil {
		builder.Status.Online = *opts.Online
	}

	if opts.TLS != nil {
		builder.Spec.Network.TLS = *opts.TLS
	}

	if opts.SSL != nil {
		builder.Spec.Network.SSL = new(types.SSL)
		builder.Spec.Network.SSL = opts.SSL
	}

	if opts.Allocated != nil {
		builder.Status.Allocated.Workers = opts.Allocated.Workers
		builder.Status.Allocated.RAM = opts.Allocated.RAM
		builder.Status.Allocated.CPU = opts.Allocated.CPU
		builder.Status.Allocated.Storage = opts.Allocated.Storage
	}

	if opts.Capacity != nil {
		builder.Status.Capacity.Workers = opts.Capacity.Workers
		builder.Status.Capacity.RAM = opts.Capacity.RAM
		builder.Status.Capacity.CPU = opts.Capacity.CPU
		builder.Status.Capacity.Storage = opts.Capacity.Storage
	}

	if opts.Usage != nil {
		builder.Status.Usage.Workers = opts.Usage.Workers
		builder.Status.Usage.RAM = opts.Usage.RAM
		builder.Status.Usage.CPU = opts.Usage.CPU
		builder.Status.Usage.Storage = opts.Usage.Storage
	}

	if err := b.storage.Builder().Update(b.context, builder); err != nil {
		log.V(logLevel).Errorf("%s:update:> update builder %s err: %v", logBuilderPrefix, builder.Meta.Hostname, err)
		return err
	}

	return nil
}

func (b *Builder) FindBuild(builder *types.Builder) (*types.Build, error) {

	if builder == nil {
		return nil, errors.New("invalid argument")
	}

	log.V(logLevel).Debugf("%s:find_build:> get build for builder  %s", logBuilderPrefix, builder.Meta.Hostname)

	build, err := b.storage.Build().Attach(b.context, builder)
	if err != nil {
		log.V(logLevel).Errorf("%s:find_build:> get build for builder %s err: %v", logBuilderPrefix, builder.Meta.Hostname, err)
		return nil, err
	}

	if build == nil {
		return nil, nil
	}

	return build, nil
}

func (b *Builder) MarkOffline() error {

	log.V(logLevel).Debugf("%s:mark_offline:> find and mark offline builders", logBuilderPrefix)

	if err := b.storage.Builder().MarkOffline(b.context); err != nil {
		log.V(logLevel).Errorf("%s:mark_offline:> find and mark offline builders err: %v", logBuilderPrefix, err)
		return err
	}

	return nil
}

func NewBuilderModel(ctx context.Context, stg storage.IStorage) IBuilder {
	return &Builder{
		context: ctx,
		storage: stg,
	}
}
