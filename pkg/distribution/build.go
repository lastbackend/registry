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
	"strings"

	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/storage"
	"github.com/spf13/viper"
	"errors"
)

type IBuild interface {
	Create(repo *types.Repo, tag string) (*types.Build, error)
	Get(id string) (*types.Build, error)
	List(repo *types.Repo, active bool) (map[string]*types.Build, error)
	Update(build *types.Build) error
	Cancel(build *types.Build) error

	CreateJob(build *types.Build) (*types.BuildJob, error)
}

type Build struct {
	context context.Context
	storage storage.Storage
}

func (b Build) Create(repo *types.Repo, tag string) (*types.Build, error) {

	log.V(logLevel).Infof("%s:build:create> create new build", logPrefix)

	bld := new(types.Build)
	bld.Repo.Owner = repo.Meta.Owner
	bld.Repo.Name = repo.Meta.Name
	bld.Repo.ID = repo.Meta.ID
	bld.Repo.SelfLink = repo.Meta.SelfLink
	bld.State.Status = types.BuildStatusQueued

	bld.Sources.Hub = repo.Sources.Hub
	bld.Sources.Owner = repo.Sources.Owner
	bld.Sources.Name = repo.Sources.Name
	bld.Sources.Branch = repo.Sources.Branch

	bld.Image.Owner = strings.ToLower(repo.Meta.Owner)
	bld.Image.Name = strings.ToLower(repo.Meta.Name)
	bld.Image.Hub = viper.GetString("registry.uri")
	bld.Image.Tag = tag

	isExists := false

	for _, t := range repo.Tags {
		if t.Name == tag {
			isExists = true
			bld.Sources.Branch = t.Spec.Branch
			bld.Config.Dockerfile = t.Spec.FilePath
			bld.Config.EnvVars = t.Spec.EnvVars
			bld.Image.Tag = t.Name
			break
		}
	}

	if !isExists {
		bld.Sources.Branch = types.RepoDefaultBranch
		bld.Image.Tag = types.RepoDefaultTag
		bld.Image.Hub = viper.GetString("registry.uri")
	}

	if match := strings.Split(bld.Sources.Hub, "."); len(match) == 2 {
		// todo get last commit
	}

	if err := b.storage.Build().Insert(b.context, bld); err != nil {
		log.V(logLevel).Errorf("%s:build:create:> create new build err: %v", logPrefix, err)
		return nil, err
	}

	return bld, nil
}

func (b Build) Get(id string) (*types.Build, error) {
	log.V(logLevel).Infof("%s:build:get:> get build info", logPrefix)

	build, err := b.storage.Build().Get(b.context, id)
	if err != nil {
		log.V(logLevel).Errorf("%s:build:get:> get build info err: %v", logPrefix, err)
		return nil, err
	}

	return build, nil
}

func (b Build) List(repo *types.Repo, active bool) (map[string]*types.Build, error) {

	log.V(logLevel).Infof("%s:build:list:> get builds list for repo %s/%s active: %#v", logPrefix, repo.Meta.Owner, repo.Meta.Name, active)

	var (
		err    error
		builds = make(map[string]*types.Build, 0)
	)

	switch true {
	case repo != nil && !active:
		builds, err = b.storage.Build().List(b.context, repo.Meta.ID)
	case repo != nil && active:
		//builds, err = b.storage.Build().ListActiveByRepo(b.context, repo.Meta.ID)
	default:
		return nil, nil
	}
	if err != nil {
		log.V(logLevel).Errorf("%s:build:list:> get builds list err: %v", logPrefix, err)
		return nil, err
	}

	log.V(logLevel).Debugf("%s:build:list:> found builds count: %d", logPrefix, len(builds))

	return builds, nil
}

func (b Build) Update(build *types.Build) error {
	log.V(logLevel).Infof("%s:build:update:> update build data", logPrefix)

	if err := b.storage.Build().Update(b.context, build); err != nil {
		log.Errorf("%s:build:update:> set build info err: %v", logPrefix, err)
		return err
	}

	return nil
}

func (b Build) Cancel(build *types.Build) error {
	log.V(logLevel).Infof("%s:build:cancel:> cancel build process", logPrefix)

	if build.State.Status == types.BuildStatusQueued {
		build.MarkAsCanceled(types.EmptyString, types.EmptyString)
		if err := b.Update(build); err != nil {
			log.Errorf("%s:build:cancel:> update build err: %v", logPrefix, err)
			return err
		}
		return nil
	}

	return nil
}

func (b *Build) CreateJob(build *types.Build) (*types.BuildJob, error) {

	log.Debugf("Create job for build :%s", build.Meta.ID)

	job := build.NewBuildJob()
	if job == nil {
		err := errors.New("job create failed")
		log.Errorf("Build Controller: CreateJob: create job err: %s", err.Error())
		return nil, err
	}

	//registry, err := b.storage.Registry().Get(b.context, acc.Meta.ID, build.Image.Hub)
	//if err != nil {
	//	log.Errorf("Build Controller: CreateJob: get registry info err: %s", err.Error())
	//	return nil, err
	//}
	//if registry != nil {
	//	token, err := registry.Auth.Encode()
	//	if err != nil {
	//		log.Errorf("Build Controller: CreateJob: encode auth token err: %s", err.Error())
	//		return nil, err
	//	}
	//	job.Image.Token = token
	//}

	return job, nil
}

func NewBuildModel(ctx context.Context, stg storage.Storage) IBuild {
	return &Build{
		context: ctx,
		storage: stg,
	}
}
