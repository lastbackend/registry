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
)

// Service Build type for interface in interfaces folder
type BuildStorage struct {
	storage.Build
}

func (s *BuildStorage) Insert(ctx context.Context, build *types.Build) error {
	log.V(logLevel).Debugf("%s:build:insert> insert new build: %#v", logPrefix, build)
	return nil
}

func (s *BuildStorage) Get(ctx context.Context, id string) (*types.Build, error) {
	log.V(logLevel).Debugf("%s:build:get> get build: %s", logPrefix, id)
	return nil, nil
}

func (s *BuildStorage) List(ctx context.Context, repo string) (map[string]*types.Build, error) {
	log.V(logLevel).Debugf("%s:build:list> get builds list", logPrefix)
	return nil, nil
}

func (s *BuildStorage) Update(ctx context.Context, build *types.Build) error {
	log.V(logLevel).Debugf("%s:build:update> update build data", logPrefix)
	return nil
}

func newBuildStorage() *BuildStorage {
	s := new(BuildStorage)
	return s
}
