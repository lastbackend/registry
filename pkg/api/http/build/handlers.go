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
	"fmt"
	"io"
	"net/http"

	"github.com/lastbackend/registry/pkg/api/envs"
	"github.com/lastbackend/registry/pkg/distribution"
	"github.com/lastbackend/registry/pkg/distribution/errors"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/util/http/utils"
	"github.com/spf13/viper"

	v "github.com/lastbackend/registry/pkg/api/views"
)

const (
	logLevel  = 2
	logPrefix = "registry:api:handler:build"
)

func BuildListH(w http.ResponseWriter, r *http.Request) {

	log.V(logLevel).Debugf("%s:build:list:> get build list", logPrefix)

	var (
		owner  = utils.Vars(r)[`owner`]
		name   = utils.Vars(r)[`name`]
		active = len(utils.Query(r, "active")) != 0
		bm     = distribution.NewBuildModel(r.Context(), envs.Get().GetStorage())
		rm     = distribution.NewRepoModel(r.Context(), envs.Get().GetStorage())
	)

	rps, err := rm.Get(owner, name)
	if err != nil {
		log.V(logLevel).Errorf("%s:build:list:>  get repo info err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if rps == nil {
		log.V(logLevel).Warnf("%s:build:list:> repo `%s/%s` not found", logPrefix, owner, name)
		errors.New("repo").NotFound().Http(w)
		return
	}

	builds, err := bm.List(rps, active)
	if err != nil {
		log.V(logLevel).Errorf("%s:build:list:> get builds list err: %s", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	response, err := v.V1().Build().NewList(builds).ToJson()
	if err != nil {
		log.V(logLevel).Errorf("%s:build:list:> convert struct to json err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(response); err != nil {
		log.V(logLevel).Errorf("%s:build:list:> write response err: %v", err)
		return
	}
}

func BuildGetH(w http.ResponseWriter, r *http.Request) {

	log.V(logLevel).Debugf("%s:build:get:> get build info handler", logPrefix)

	var (
		build = utils.Vars(r)[`build`]
		owner = utils.Vars(r)[`owner`]
		name  = utils.Vars(r)[`name`]
		rm    = distribution.NewRepoModel(r.Context(), envs.Get().GetStorage())
		bm    = distribution.NewBuildModel(r.Context(), envs.Get().GetStorage())
	)

	rps, err := rm.Get(owner, name)
	if err != nil {
		log.V(logLevel).Errorf("%s:build:get:> get repo info err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if rps == nil {
		log.V(logLevel).Warnf("%s:build:get:> repo `%s/%s` not found", logPrefix, owner, name)
		errors.New("repo").NotFound().Http(w)
		return
	}

	b, err := bm.Get(build)
	if err != nil {
		log.V(logLevel).Errorf("%s:build:get:> get info err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if b == nil {
		log.V(logLevel).Warnf("%s:build:get:> build `%s` not found", logPrefix, build)
		errors.New("build").NotFound().Http(w)
		return
	}

	response, err := v.V1().Build().New(b).ToJson()
	if err != nil {
		log.V(logLevel).Errorf("%s:build:get:> convert struct to json err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(response); err != nil {
		log.V(logLevel).Errorf("%s:build:get:> write response err: %v", logPrefix, err)
		return
	}
}

func BuildLogsH(w http.ResponseWriter, r *http.Request) {

	log.V(logLevel).Debugf("%s:build:logs:>  get build logs handler", logPrefix)

	var (
		id    = utils.Vars(r)[`build`]
		owner = utils.Vars(r)[`owner`]
		name  = utils.Vars(r)[`name`]
		rm    = distribution.NewRepoModel(r.Context(), envs.Get().GetStorage())
		bm    = distribution.NewBuildModel(r.Context(), envs.Get().GetStorage())
	)

	rps, err := rm.Get(owner, name)
	if err != nil {
		log.V(logLevel).Errorf("%s:build:logs:> get repo info err: %s", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if rps == nil {
		log.V(logLevel).Warnf("%s:build:logs:> repo `%s/%s` not found", logPrefix, owner, name)
		errors.New("repo").NotFound().Http(w)
		return
	}

	b, err := bm.Get(id)
	if err != nil {
		log.V(logLevel).Errorf("%s:build:logs:> get info err: %s", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if b == nil {
		log.V(logLevel).Warnf("%s:build:logs:> build `%s` not found", logPrefix, id)
		errors.New("build").NotFound().Http(w)
		return
	}

	read, write := io.Pipe()

	// writing without a reader will deadlock so write in a goroutine
	go func() {
		defer write.Close()
		resp, err := http.Get(fmt.Sprintf("https://%s/logs/%s", viper.GetString("storage.azure.endpoint"), b.Meta.Task))
		if err != nil {
			return
		}
		defer resp.Body.Close()
		io.Copy(write, resp.Body)

	}()

	io.Copy(w, read)
}

func BuildCancelH(w http.ResponseWriter, r *http.Request) {

	var id = utils.Vars(r)[`build`]

	log.V(logLevel).Debugf("%s:build:cancel:> cancel build %s handler", logPrefix, id)

	var (
		owner = utils.Vars(r)[`owner`]
		name  = utils.Vars(r)[`name`]
		rm    = distribution.NewRepoModel(r.Context(), envs.Get().GetStorage())
		bm    = distribution.NewBuildModel(r.Context(), envs.Get().GetStorage())
	)

	rps, err := rm.Get(owner, name)
	if err != nil {
		log.V(logLevel).Errorf("%s:build:cancel:> get repo info err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if rps == nil {
		log.V(logLevel).Warnf("%s:build:cancel:> repo `%s/%s` not found", logPrefix, owner, name)
		errors.New("repo").NotFound().Http(w)
		return
	}

	b, err := bm.Get(id)
	if err != nil {
		log.V(logLevel).Errorf("%s:build:cancel:> get info err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if b == nil {
		log.V(logLevel).Warnf("%s:build:cancel:> build `%s` not found", logPrefix, id)
		errors.New("build").NotFound().Http(w)
		return
	}

	err = bm.Cancel(b)
	if err != nil {
		log.V(logLevel).Errorf("%s:build:cancel:> cancel build err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte{}); err != nil {
		log.V(logLevel).Errorf("%s:build:cancel:> write response err: %v", logPrefix, err)
		return
	}
}

func BuildCreateH(w http.ResponseWriter, r *http.Request) {

	log.V(logLevel).Debugf("New build handler", logPrefix)

	var (
		owner = utils.Vars(r)[`owner`]
		name  = utils.Vars(r)[`name`]
		bm    = distribution.NewBuildModel(r.Context(), envs.Get().GetStorage())
		rm    = distribution.NewRepoModel(r.Context(), envs.Get().GetStorage())
	)

	opts := new(types.BuildCreateOptions)
	if err := opts.DecodeAndValidate(r.Body); err != nil {
		log.V(logLevel).Errorf("%s:build:create:> validation incoming data err: %v", logPrefix, err)
		err.Http(w)
		return
	}

	rps, err := rm.Get(owner, name)
	if err != nil {
		log.V(logLevel).Errorf("%s:build:create:> get repo info err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if rps == nil {
		log.V(logLevel).Warnf("%s:build:create:> repo `%s/%s` not found", logPrefix, owner, name)
		errors.New("repo").NotFound().Http(w)
		return
	}

	if _, ok := rps.Tags[*opts.Tag]; !ok {
		log.V(logLevel).Warn("%s:build:create:> tag %s not found in build rules", logPrefix, *opts.Tag)
		errors.HTTP.NotFound(w)
		return
	}

	b, err := bm.Create(rps, rps.Tags[*opts.Tag].Name)
	if err != nil {
		log.V(logLevel).Errorf("%s:build:create:> create build err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	if err := rm.UpdateTag(rps.Meta.ID, rps.Tags[*opts.Tag].Name); err != nil {
		log.V(logLevel).Errorf("%s:build:create:> update repo tag err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	response, err := v.V1().Build().New(b).ToJson()
	if err != nil {
		log.V(logLevel).Errorf("%s:build:create:> convert struct to json err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(response); err != nil {
		log.V(logLevel).Errorf("%s:build:create:> write response err: %v", logPrefix, err)
		return
	}
}
