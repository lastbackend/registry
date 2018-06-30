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

type RepoStorage struct {
	storage.Repo
}

func (s *RepoStorage) Get(ctx context.Context, owner, name string) (*types.Repo, error) {
	log.V(logLevel).Debugf("%s:repo:get> get repo `%s/%s`", logPrefix, owner, name)
	return nil, nil
}

func (s *RepoStorage) List(ctx context.Context) (map[string]*types.Repo, error) {
	log.V(logLevel).Debug("%s:repo:list> get repositories list", logPrefix)
	return nil, nil
}

func (s *RepoStorage) Insert(ctx context.Context, repo *types.Repo) error {
	log.V(logLevel).Debugf("%s:repo:insert> insert repo: %#v", logPrefix, repo)
	return nil
}

func (s *RepoStorage) Update(ctx context.Context, repo *types.Repo) error {
	log.V(logLevel).Debugf("%s:repo:update> update repo %#v", logPrefix, repo)
	return nil
}

func (s *RepoStorage) Remove(ctx context.Context, id string) error {
	log.V(logLevel).Debugf("%s:repo:remove> remove repo %s", logPrefix, id)
	return nil
}

func (s *RepoStorage) InsertTag(ctx context.Context, tag *types.RepoTag) error {
	log.V(logLevel).Debugf("%s:repo:insert_tag>: update build rules for repo %#v", logPrefix, tag.RepoID)
	return nil
}

func (s *RepoStorage) UpdateTag(ctx context.Context, repo, tag string) error {
	log.V(logLevel).Debugf("%s:repo:update_tag> update build rules for repo %#v", logPrefix, repo)
	return nil
}

func newRepoStorage() *RepoStorage {
	s := new(RepoStorage)
	return s
}
