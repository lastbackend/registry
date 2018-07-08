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

package build

import (
	"net/http"

	"github.com/lastbackend/registry/pkg/api/envs"
	"github.com/lastbackend/registry/pkg/api/types/v1"
	"github.com/lastbackend/registry/pkg/distribution"
	"github.com/lastbackend/registry/pkg/distribution/errors"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/util/http/utils"
	)

const (
	logLevel  = 2
	logPrefix = "registry:api:handler:build"
)

func BuildCreateH(w http.ResponseWriter, r *http.Request) {

	log.V(logLevel).Debugf("%s:build:> execute build image", logPrefix)

	var (
		im    = distribution.NewImageModel(r.Context(), envs.Get().GetStorage())
		bm    = distribution.NewBuildModel(r.Context(), envs.Get().GetStorage())
	)

	// request body struct
	rq := v1.Request().Build().BuildExecuteOptions()
	if e := rq.DecodeAndValidate(r.Body); e != nil {
		log.V(logLevel).Errorf("%s:build:> validation incoming data err: %v", logPrefix, e)
		e.Http(w)
		return
	}

	img, err := im.Get(rq.Owner, rq.Name)
	if err != nil {
		log.V(logLevel).Errorf("%s:build:> get image %s/%s err: %v", logPrefix, rq.Owner, rq.Name, err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if img == nil {
		log.V(logLevel).Warnf("%s:build:> image `%s/%s` not found", logPrefix, rq.Owner, rq.Name)
		errors.New("image").NotFound().Http(w)
		return
	}

	opts := new(types.BuildCreateOptions)

	opts.Source.Hub = rq.Source.Hub
	opts.Source.Owner = rq.Source.Owner
	opts.Source.Name = rq.Source.Name
	opts.Source.Branch = rq.Source.Branch
	opts.Source.Token = rq.Source.Token

	opts.Image.ID = img.Meta.ID
	opts.Image.Owner = img.Meta.Owner
	opts.Image.Name = img.Meta.Name
	opts.Image.Tag = rq.Tag
	opts.Image.Auth = rq.Auth

	opts.Spec.DockerFile = rq.DockerFile

	if len(rq.DockerFile) == 0 {
		opts.Spec.DockerFile = types.ImageDefaultDockerfilePath
	}

	opts.Spec.EnvVars = rq.EnvVars
	opts.Spec.Workdir = rq.Workdir
	opts.Spec.Command = rq.Command

	build, err := bm.Create(opts)
	if err != nil {
		log.V(logLevel).Errorf("%s:create:> create build err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	response, err := v1.View().Build().New(build).ToJson()
	if err != nil {
		log.V(logLevel).Errorf("%s:build:> convert struct to json err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(response); err != nil {
		log.V(logLevel).Errorf("%s:build:> write response err: %v", logPrefix, err)
		return
	}
}

func BuildCancelH(w http.ResponseWriter, r *http.Request) {

	log.V(logLevel).Debugf("%s:cancel:> cancel build process", logPrefix)

	var (
	//bm    = distribution.NewBuildModel(r.Context(), envs.Get().GetStorage())
	//bid = utils.Vars(r)[`build`]
	)

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte{}); err != nil {
		log.V(logLevel).Errorf("%s:cancel:> write response err: %v", logPrefix, err)
		return
	}
}

func BuildLogsH(w http.ResponseWriter, r *http.Request) {

	log.V(logLevel).Debugf("%s:logs:> cancel build process", logPrefix)

	var (
	//bm    = distribution.NewBuildModel(r.Context(), envs.Get().GetStorage())
	//bid  = utils.Vars(r)[`build`]
	)

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte{}); err != nil {
		log.V(logLevel).Errorf("%s:logs:> write response err: %v", logPrefix, err)
		return
	}
}

func BuildTaskStatusUpdateH(w http.ResponseWriter, r *http.Request) {

	log.Debugf("%s:update_status:> set build task status info handler", logPrefix)

	var (
		bm  = distribution.NewBuildModel(r.Context(), envs.Get().GetStorage())
		tid = utils.Vars(r)[`task`]
	)

	// request body struct
	rq := v1.Request().Build().BuildStatusOptions()
	if e := rq.DecodeAndValidate(r.Body); e != nil {
		log.V(logLevel).Errorf("%s:update_status:> validation incoming data err: %v", logPrefix, e)
		e.Http(w)
		return
	}

	build, err := bm.GetByTask(tid)
	if err != nil {
		log.V(logLevel).Errorf("%s:update_status:> get build by task err: %v", logPrefix, tid, err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if build == nil {
		log.V(logLevel).Warnf("%s:update_status:> build `%s` not found", logPrefix, tid)
		errors.New("build").NotFound().Http(w)
		return
	}

	opts := new(types.BuildUpdateStatusOptions)
	opts.Step = rq.Step
	opts.Message = rq.Message
	opts.Error = rq.Error
	opts.Canceled = rq.Canceled

	if err := bm.UpdateStatus(build, opts); err != nil {
		log.V(logLevel).Errorf("%s:update_status:> update build err: %v", logPrefix, build.Meta.ID, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte{}); err != nil {
		log.V(logLevel).Errorf("%s:update_status:> write response err: %v", logPrefix, err)
		return
	}
}

func BuildTaskInfoUpdateH(w http.ResponseWriter, r *http.Request) {

	log.Debugf("%s:update_info:> set build task info handler", logPrefix)

	var (
		bm  = distribution.NewBuildModel(r.Context(), envs.Get().GetStorage())
		tid = utils.Vars(r)[`task`]
	)

	// request body struct
	rq := v1.Request().Build().BuildInfoOptions()
	if e := rq.DecodeAndValidate(r.Body); e != nil {
		log.V(logLevel).Errorf("%s:update_info:> validation incoming data err: %v", logPrefix, e)
		e.Http(w)
		return
	}

	build, err := bm.GetByTask(tid)
	if err != nil {
		log.V(logLevel).Errorf("%s:update_info:> get build by task %s err: %v", logPrefix, tid, err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if build == nil {
		log.V(logLevel).Warnf("%s:update_info:> build `%s` not found", logPrefix, tid)
		errors.New("build").NotFound().Http(w)
		return
	}

	opts := new(types.BuildUpdateInfoOptions)
	opts.Size = rq.Size
	opts.Hash = rq.Hash

	if err := bm.UpdateInfo(build, opts); err != nil {
		log.V(logLevel).Errorf("%s:update_info:> update build err: %v", logPrefix, tid, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte{}); err != nil {
		log.V(logLevel).Errorf("%s:update_info:> write response err: %v", logPrefix, err)
		return
	}
}
