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
	"errors"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/storage"
	"github.com/lastbackend/registry/pkg/storage/types/filter"
	"time"
)

const (
	logBuildPrefix = "distribution:build"
)

type IBuild interface {
	Get(id string) (*types.Build, error)
	List(image *types.Image, opts *types.BuildListOptions) ([]*types.Build, error)
	Create(opts *types.BuildCreateOptions) (*types.Build, error)
	UpdateStatus(build *types.Build, opts *types.BuildUpdateStatusOptions) error
	UpdateInfo(build *types.Build, opts *types.BuildUpdateInfoOptions) error
	Unfreeze() error
}

type Build struct {
	context context.Context
	storage storage.IStorage
}

func (b Build) Get(id string) (*types.Build, error) {

	log.V(logLevel).Infof("%s:build:get:> get build %s info", logBuildPrefix, id)

	build, err := b.storage.Build().Get(b.context, id)
	if err != nil {
		log.V(logLevel).Errorf("%s:build:get:> get build %s info err: %v", logBuildPrefix, id, err)
		return nil, err
	}

	return build, nil
}

func (b Build) List(image *types.Image, opts *types.BuildListOptions) ([]*types.Build, error) {

	if image == nil {
		return nil, errors.New("invalid argument")
	}

	log.V(logLevel).Infof("%s:build:list:> get builds list for image %s/%s", logBuildPrefix, image.Meta.Owner, image.Meta.Name)

	f := filter.NewFilter().Build()
	if opts.Active != nil {
		f.Active = opts.Active
	}

	builds, err := b.storage.Build().List(b.context, image.Meta.ID, f)
	if err != nil {
		log.V(logLevel).Errorf("%s:build:list:> get builds list err: %v", logBuildPrefix, err)
		return nil, err
	}

	log.V(logLevel).Debugf("%s:build:list:> found builds count: %d", logBuildPrefix, len(builds))

	return builds, nil
}

func (b Build) Create(opts *types.BuildCreateOptions) (*types.Build, error) {

	log.V(logLevel).Infof("%s:build:create> create new build %#v", logBuildPrefix, opts)

	if opts == nil {
		opts = new(types.BuildCreateOptions)
	}

	bld := new(types.Build)

	bld.Meta.Labels = opts.Labels

	bld.Status.Status = types.BuildStatusQueued

	bld.Spec.Source.Hub = opts.Source.Hub
	bld.Spec.Source.Owner = opts.Source.Owner
	bld.Spec.Source.Name = opts.Source.Name
	bld.Spec.Source.Branch = opts.Source.Branch

	if opts.Labels != nil {

		bld.Spec.Source.Commit = new(types.BuildCommit)

		if hash, ok := opts.Labels["commit_hash"]; ok {
			bld.Spec.Source.Commit.Hash = hash
		}
		if username, ok := opts.Labels["commit_username"]; ok {
			bld.Spec.Source.Commit.Username = username
		}
		if message, ok := opts.Labels["commit_message"]; ok {
			bld.Spec.Source.Commit.Message = message
		}
		if email, ok := opts.Labels["commit_email"]; ok {
			bld.Spec.Source.Commit.Email = email
		}
		if date, ok := opts.Labels["commit_date"]; ok {
			layout := "2006-01-02 15:04:05 -0700 MST"
			t, _ := time.Parse(layout, date)
			bld.Spec.Source.Commit.Date = t
		}

	}

	if len(opts.Source.Branch) == 0 {
		bld.Spec.Source.Branch = types.SourceDefaultBranch
	}

	bld.Spec.Source.Token = opts.Source.Token

	bld.Spec.Image.ID = opts.Image.ID
	bld.Spec.Image.Owner = opts.Image.Owner
	bld.Spec.Image.Name = opts.Image.Name
	bld.Spec.Image.Tag = opts.Image.Tag

	if len(opts.Image.Tag) == 0 {
		bld.Spec.Image.Tag = types.ImageDefaultTag
	}

	bld.Spec.Image.Auth = opts.Image.Auth

	bld.Spec.Config.Dockerfile = opts.Spec.DockerFile
	bld.Spec.Config.Context = opts.Spec.Context
	bld.Spec.Config.EnvVars = opts.Spec.EnvVars

	bld.Spec.Config.EnvVars = opts.Spec.EnvVars
	if opts.Spec.EnvVars == nil {
		bld.Spec.Config.EnvVars = make([]string, 0)
	}

	bld.Spec.Config.Workdir = opts.Spec.Workdir
	bld.Spec.Config.Command = opts.Spec.Command

	if err := b.storage.Build().Insert(b.context, bld); err != nil {
		log.V(logLevel).Errorf("%s:build:create> create new build err: %v", logBuildPrefix, err)
		return nil, err
	}

	return bld, nil
}

func (b Build) UpdateStatus(build *types.Build, opts *types.BuildUpdateStatusOptions) error {

	if build == nil {
		return errors.New("invalid argument")
	}

	if opts == nil {
		opts = new(types.BuildUpdateStatusOptions)
	}

	log.V(logLevel).Infof("%s:build:update_status:> update build %s data", logBuildPrefix, build.Meta.ID)

	switch true {
	case opts.Canceled:
		build.MarkAsCanceled(opts.Step, opts.Message)
	case opts.Error:
		build.MarkAsError(opts.Step, opts.Message)
		// BuildStepFetch -  not supported
	case opts.Step == types.BuildStepFetch:
		build.MarkAsFetching(opts.Step, opts.Message)
	case opts.Step == types.BuildStepBuild:
		build.MarkAsBuilding(opts.Step, opts.Message)
	case opts.Step == types.BuildStepUpload:
		build.MarkAsUploading(opts.Step, opts.Message)
	case opts.Step == types.BuildStepDone && !opts.Error:
		build.MarkAsDone(opts.Step, opts.Message)
	}

	if err := b.storage.Build().Update(b.context, build); err != nil {
		log.Errorf("%s:build:update_status:> update build status err: %v", logBuildPrefix, err)
		return err
	}

	return nil
}

func (b Build) UpdateInfo(build *types.Build, opts *types.BuildUpdateInfoOptions) error {

	if build == nil {
		return errors.New("invalid argument")
	}

	if opts == nil {
		opts = new(types.BuildUpdateInfoOptions)
	}

	log.V(logLevel).Infof("%s:build:update_info:> update build %s data", logBuildPrefix, build.Meta.ID)

	build.Status.Size = opts.Size
	build.Spec.Image.Hash = opts.Hash

	if err := b.storage.Build().Update(b.context, build); err != nil {
		log.Errorf("%s:build:update_info:> set build info err: %v", logBuildPrefix, err)
		return err
	}

	return nil
}

func (b Build) Unfreeze() error {

	log.V(logLevel).Infof("%s:build:unfreeze:> unfreeze dangling builds", logBuildPrefix)

	if err := b.storage.Build().Unfreeze(b.context); err != nil {
		log.Errorf("%s:build:unfreeze:> unfreeze builds err: %v", logBuildPrefix, err)
		return err
	}

	return nil
}

func NewBuildModel(ctx context.Context, stg storage.IStorage) IBuild {
	return &Build{
		context: ctx,
		storage: stg,
	}
}
