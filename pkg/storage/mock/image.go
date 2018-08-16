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

package mock

import (
	"context"

	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/storage/storage"
	"github.com/lastbackend/registry/pkg/storage/types/filter"
)

type ImageStorage struct {
	storage.Image
}

func (s *ImageStorage) Get(ctx context.Context, owner, name string) (*types.Image, error) {
	log.V(logLevel).Debugf("%s:image:get> get image `%s/%s`", logPrefix, owner, name)
	return nil, nil
}

func (s *ImageStorage) List(ctx context.Context, f *filter.ImageFilter) ([]*types.Image, error) {
	log.V(logLevel).Debug("%s:image:list> get repositories list", logPrefix)
	return nil, nil
}

func (s *ImageStorage) Insert(ctx context.Context, image *types.Image) error {
	log.V(logLevel).Debugf("%s:image:insert> insert image: %#v", logPrefix, image)
	return nil
}

func (s *ImageStorage) Update(ctx context.Context, image *types.Image) error {
	log.V(logLevel).Debugf("%s:image:update> update image %#v", logPrefix, image)
	return nil
}

func (s *ImageStorage) Remove(ctx context.Context, image *types.Image) error {
	log.V(logLevel).Debugf("%s:image:remove> remove image %#v", logPrefix, image)
	return nil
}

func newImageStorage() *ImageStorage {
	s := new(ImageStorage)
	return s
}
