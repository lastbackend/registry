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
	"fmt"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/storage"
	"github.com/lastbackend/registry/pkg/util/converter"
	"github.com/spf13/viper"
	"strings"
)

const (
	logPrefix = "registry:api:distribution"
)

type IRepo interface {
	Get(owner, name string) (*types.Repo, error)
	List() (map[string]*types.Repo, error)
	Create(opts *types.RepoCreateOptions) (*types.Repo, error)
	Update(repo *types.Repo, opts *types.RepoUpdateOptions) error
	Remove(id string) error

	UpdateTag(repo, tag string) error
}

type Repo struct {
	context context.Context
	storage storage.Storage
}

func (r Repo) Get(owner, name string) (*types.Repo, error) {
	log.V(logLevel).Debugf("%s:get:> get repo `%s/%s`", logPrefix, owner, name)

	rps, err := r.storage.Repo().Get(r.context, owner, name)
	if err != nil {
		log.V(logLevel).Debugf("%s:get:> get repo `%s/%s` err: %v", logPrefix, owner, name, err)
		return nil, err
	}
	if rps == nil {
		return rps, nil
	}
	return rps, nil
}

func (r Repo) List() (map[string]*types.Repo, error) {
	log.V(logLevel).Debugf("%s:list:> get repo list", logPrefix)

	list, err := r.storage.Repo().List(r.context)
	if err != nil {
		log.V(logLevel).Debugf("%s:list:> get repo list %s err: %v", logPrefix, err)
		return nil, err
	}

	return list, nil
}

func (r Repo) Create(opts *types.RepoCreateOptions) (*types.Repo, error) {

	ctx, err := r.storage.Begin(r.context)
	if err != nil {
		log.V(logLevel).Debugf("%s:create:> create storage context err: %v", logPrefix, err)
		return nil, err
	}

	log.V(logLevel).Debugf("%s:create:> create repo %#v", logPrefix, opts)

	rps := new(types.Repo)
	rps.Meta = types.RepoMeta{}
	rps.Meta.Owner = strings.ToLower(opts.Spec.Image.Owner)
	rps.Meta.Name = strings.ToLower(opts.Spec.Image.Name)
	rps.Meta.SelfLink = fmt.Sprintf("%s.%s", rps.Meta.Name, rps.Meta.Owner)

	rps.Meta.Labels = map[string]string{
		"technology": types.RepoDefaultTechnology,
	}

	if len(opts.Meta.Labels) != 0 {
		for k, v := range opts.Meta.Labels {
			rps.Meta.Labels[k] = v
		}
	}

	rps.Sources = types.RepoSources{
		Hub:    viper.GetString("git.uri"),
		Owner:  opts.Spec.Image.Owner,
		Name:   opts.Spec.Image.Name,
		Branch: types.RepoDefaultBranch,
	}

	if opts.Spec.Source != nil {

		rps.Remote = true

		if rps.Sources, err = sourceConfigure(*opts.Spec.Source); err != nil {
			log.V(logLevel).Errorf("%s:create:> configure source err: %v", logPrefix, err)
			r.storage.Rollback(ctx)
			return nil, err
		}

		rps.Tags = tagsConfigure(rps.Sources, opts.Spec.Rules)

	} else {
		rps.Remote = false
		rps.Tags = make(map[string]*types.RepoTag, 1)
		tag := new(types.RepoTag)
		tag.Name = types.RepoDefaultTag
		tag.Spec.Branch = types.RepoDefaultBranch
		tag.Spec.FilePath = types.RepoDefaultDockerfilePath
		rps.Tags[types.RepoDefaultTag] = tag
	}

	if err := r.storage.Repo().Insert(ctx, rps); err != nil {
		log.V(logLevel).Errorf("%s:create:> insert repo err: %v", logPrefix, err)
		r.storage.Rollback(ctx)
		return nil, err
	}

	for _, tag := range rps.Tags {
		tag.RepoID = rps.Meta.ID
		if err := r.storage.Repo().InsertTag(ctx, tag); err != nil {
			log.V(logLevel).Errorf("%s:create:> insert repo tag err: %v", logPrefix, err)
			r.storage.Rollback(ctx)
			return nil, err
		}
	}

	// For remote repo need create build
	if rps.Remote {

		bm := NewBuildModel(ctx, r.storage)

		for _, tag := range rps.Tags {

			build, err := bm.Create(rps, tag.Name)
			if err != nil {
				log.V(logLevel).Errorf("%s:create:> create build err: %v", logPrefix, err)
				r.storage.Rollback(ctx)
				return nil, err
			}

			rps.Tags[tag.Name].Build0 = new(types.BuildView)
			rps.Tags[tag.Name].Build0.ID = build.Meta.ID
			rps.Tags[tag.Name].Build0.Number = build.Meta.Number
			rps.Tags[tag.Name].Build0.Status = build.State.Status

			if err := r.storage.Repo().UpdateTag(ctx, tag.RepoID, tag.Name); err != nil {
				log.V(logLevel).Errorf("%s:create:> update repo tag err: %v", logPrefix, err)
				r.storage.Rollback(ctx)
				return nil, err
			}
		}

	}

	if _, err := r.storage.Commit(ctx); err != nil {
		log.V(logLevel).Debugf("%s:create:> commit storage context err: %v", logPrefix, err)
		return nil, err
	}

	return r.Get(rps.Meta.Owner, rps.Meta.Name)
}

func (r Repo) Update(repo *types.Repo, opts *types.RepoUpdateOptions) error {
	log.V(logLevel).Debugf("%s:update:> update repo %#v", logPrefix, opts)

	if opts.Spec.Rules != nil {
		tagCache := make(map[string]*types.RepoTag, len(repo.Tags))

		rules := *opts.Spec.Rules

		for _, rule := range rules {

			tag := new(types.RepoTag)
			tag.RepoID = repo.Meta.ID
			tag.Name = rule.Tag
			tag.Spec.Branch = rule.Branch
			tag.Spec.FilePath = rule.FilePath
			tag.Spec.EnvVars = *rule.Config.EnvVars

			tagCache[rule.Tag] = tag
		}

		for _, tag := range repo.Tags {
			item, ok := tagCache[tag.Name]

			// Disabled tag if not exists in request
			if !ok {
				repo.Tags[tag.Name].Disabled = true
				continue
			}

			// Mark as update if current tag not equal with request tag or
			// if equal but marked as disabled
			isEqual := compareBuildRule(tag, item)

			if !isEqual || (isEqual && tag.Disabled) {
				item.Disabled = false
				repo.Tags[tag.Name] = item
			}

			// Delete new tag from cache
			delete(tagCache, item.Name)
		}

		// if cache not empty what add all item (insert)
		for _, tag := range tagCache {
			repo.Tags[tag.Name] = tag
		}
	}

	if err := r.storage.Repo().Update(r.context, repo); err != nil {
		log.V(logLevel).Errorf("%s:update:> update repo err: %v", logPrefix, err)
		return err
	}

	return nil
}

func (r Repo) Remove(id string) error {
	log.V(logLevel).Debugf("%s:remove:> remove repo %s", logPrefix, id)

	if err := r.storage.Repo().Remove(r.context, id); err != nil {
		log.V(logLevel).Debugf("%s:remove:> remove repo `%s` err: %v", logPrefix, id, err)
		return err
	}
	return nil
}

func (r *Repo) UpdateTag(repo, tag string) error {
	log.V(logLevel).Info("%s:repo:update_tag:> update repo tag status build", logPrefix)

	err := r.storage.Repo().UpdateTag(r.context, repo, tag)
	if err != nil {
		log.V(logLevel).Debugf("%s:repo:update_tag:> update repo tag err: %v", logPrefix, err)
		return err
	}

	return nil
}

func NewRepoModel(ctx context.Context, stg storage.Storage) IRepo {
	return &Repo{ctx, stg}
}

func compareBuildRule(origin, item *types.RepoTag) bool {
	if origin.Spec.Branch != item.Spec.Branch {
		return false
	}
	if origin.Spec.FilePath != item.Spec.FilePath {
		return false
	}
	if origin.Name != item.Name {
		return false
	}
	if len(origin.Spec.EnvVars) != len(item.Spec.EnvVars) && !compare(origin.Spec.EnvVars, item.Spec.EnvVars) {
		return false
	}
	return true
}

func compare(left, right []string) bool {
	m := make(map[string]int)
	for _, y := range right {
		m[y]++
	}
	for _, x := range left {
		if m[x] > 0 {
			m[x]--
			continue
		}
		return false
	}
	return true
}

func sourceConfigure(opts types.RepoSourceOpts) (types.RepoSources, error) {

	sources, err := converter.GitUrlParse(opts.Url)
	if err != nil {
		return types.RepoSources{}, err
	}

	src := types.RepoSources{
		Hub:    sources.Hub,
		Owner:  sources.Owner,
		Name:   sources.Name,
		Branch: sources.Branch,
	}

	return src, nil
}

func tagsConfigure(sources types.RepoSources, rules types.RepoBuildRules) map[string]*types.RepoTag {

	tags := make(map[string]*types.RepoTag, len(rules))

	if len(rules) == 0 {
		rules = append(rules, types.RepoBuildRule{
			Branch:   sources.Branch,
			FilePath: types.RepoDefaultDockerfilePath,
			Tag:      types.RepoDefaultTag,
		})
	}

	for _, rule := range rules {

		tag := new(types.RepoTag)
		tag.Name = rule.Tag
		tag.Spec.Branch = rule.Branch
		tag.Spec.FilePath = rule.FilePath

		if rule.Config.EnvVars != nil {
			tag.Spec.EnvVars = *rule.Config.EnvVars
		}

		tags[rule.Tag] = tag
	}

	return tags
}
