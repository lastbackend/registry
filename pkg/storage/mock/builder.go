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

	"github.com/lastbackend/lastbackend/pkg/log"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/storage/storage"
)

// Service Builder type for interface in interfaces folder
type BuilderStorage struct {
	storage.Builder
}

func (s *BuilderStorage) Get(ctx context.Context, id string) (*types.Builder, error) {
	log.V(logLevel).Debugf("%s:builder:get> get builder: %s", logPrefix, id)
	return nil, nil
}

func (s *BuilderStorage) Insert(ctx context.Context, builder *types.Builder) error {
	log.V(logLevel).Debugf("%s:builder:insert> insert new builder: %#v", logPrefix, builder)
	return nil
}

func (s *BuilderStorage) Update(ctx context.Context, builder *types.Builder) error {
	log.V(logLevel).Debugf("%s:builder:update> update builder data", logPrefix)
	return nil
}

func (s *BuilderStorage) MarkOffline(ctx context.Context) error {
	log.V(logLevel).Debugf("%s:builder:mark_offline> mark builders as offline", logPrefix)
	return nil
}

func newBuilderStorage() *BuilderStorage {
	s := new(BuilderStorage)
	return s
}
