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
	logImagePrefix = "distribution:builder"
)

type IImage interface {
	Get(owner, name string) (*types.Image, error)
	List(opts *types.ImageListOptions) ([]*types.Image, error)
	Create(opts *types.ImageCreateOptions) (*types.Image, error)
	Update(image *types.Image, opts *types.ImageUpdateOptions) error
	Remove(image *types.Image) error
}

type Image struct {
	context context.Context
	storage storage.IStorage
}

func (i Image) Get(owner, name string) (*types.Image, error) {
	log.V(logLevel).Debugf("%s:get:> get image `%s/%s`", logImagePrefix, owner, name)

	image, err := i.storage.Image().Get(i.context, owner, name)
	if err != nil {
		log.V(logLevel).Debugf("%s:get:> get image `%s/%s` err: %v", logImagePrefix, owner, name, err)
		return nil, err
	}

	return image, nil
}

func (i Image) List(opts *types.ImageListOptions) ([]*types.Image, error) {
	log.V(logLevel).Debugf("%s:list:> get image list", logImagePrefix)

	if opts == nil {
		opts = new(types.ImageListOptions)
	}

	filter := storage.Filter().Image()
	if opts != nil {
		filter.Owner = &opts.Owner
	}

	list, err := i.storage.Image().List(i.context, filter)
	if err != nil {
		log.V(logLevel).Debugf("%s:list:> get image list err: %v", logImagePrefix, err)
		return nil, err
	}

	return list, nil
}

func (i Image) Create(opts *types.ImageCreateOptions) (*types.Image, error) {

	log.V(logLevel).Debugf("%s:create:> create image %#v", logImagePrefix, opts)

	if opts == nil {
		opts = new(types.ImageCreateOptions)
	}

	image := new(types.Image)
	image.Meta.Name = opts.Name
	image.Meta.Owner = opts.Owner

	image.Meta.Description = opts.Description
	image.Spec.Private = opts.Private

	if err := i.storage.Image().Insert(i.context, image); err != nil {
		log.V(logLevel).Errorf("%s:create:> insert image %s/%s err: %v", logImagePrefix, image.Meta.Owner, image.Meta.Name, err)
		return nil, err
	}

	return image, nil
}

func (i Image) Update(image *types.Image, opts *types.ImageUpdateOptions) error {

	if image == nil {
		return errors.New("invalid argument")
	}

	if opts == nil {
		opts = new(types.ImageUpdateOptions)
	}

	log.V(logLevel).Debugf("%s:update:> update image %s/%s -> %#v", logImagePrefix, image.Meta.Owner, image.Meta.Name, opts)

	if opts.Description != nil {
		image.Meta.Description = *opts.Description
	}

	if opts.Private != nil {
		image.Spec.Private = *opts.Private
	}

	if err := i.storage.Image().Update(i.context, image); err != nil {
		log.V(logLevel).Errorf("%s:update:> update image err: %v", logImagePrefix, err)
		return err
	}

	return nil
}

func (i Image) Remove(image *types.Image) error {

	if image == nil {
		return errors.New("invalid argument")
	}

	log.V(logLevel).Debugf("%s:remove:> remove image %s/%s", logImagePrefix, image.Meta.Owner, image.Meta.Name)

	if err := i.storage.Image().Remove(i.context, image); err != nil {
		log.V(logLevel).Debugf("%s:remove:> remove image `%s/%s` err: %v", logImagePrefix, image.Meta.Owner, image.Meta.Name, err)
		return err
	}

	return nil
}

func NewImageModel(ctx context.Context, stg storage.IStorage) IImage {
	return &Image{ctx, stg}
}
