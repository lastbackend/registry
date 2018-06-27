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

	v "github.com/lastbackend/registry/pkg/api/views"

	"github.com/lastbackend/registry/pkg/api/envs"
	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/util/http/utils"
	"github.com/lastbackend/registry/pkg/distribution/errors"
	"github.com/spf13/viper"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/distribution"
)

const logLevel = 2

func BuildList(w http.ResponseWriter, r *http.Request) {

	log.Debug("Handler: Builds: get build list")

	var (
		owner  = utils.Vars(r)[`owner`]
		name   = utils.Vars(r)[`name`]
		active = len(utils.Query(r, "active")) != 0
		bm     = distribution.NewBuildModel(r.Context(), envs.Get().GetStorage())
		rm     = distribution.NewRepoModel(r.Context(), envs.Get().GetStorage())
	)

	rps, err := rm.Get("")
	if err != nil {
		log.V(logLevel).Errorf("Handler: Builds: get repo info err: %s", err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if rps == nil {
		log.V(logLevel).Warnf("Handler: Builds: repo `%s/%s` not found", owner, name)
		errors.New("repo").NotFound().Http(w)
		return
	}

	builds, err := bm.List(rps, active)
	if err != nil {
		log.V(logLevel).Errorf("Handler: Build: get builds list err: %s", err)
		errors.HTTP.InternalServerError(w)
		return
	}

	response, err := v.V1().Build().NewList(builds).ToJson()
	if err != nil {
		log.V(logLevel).Errorf("Handler: Build: convert struct to json err: %v", err)
		errors.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(response); err != nil {
		log.Error("Error: write response", err)
		return
	}
}

func BuildGet(w http.ResponseWriter, r *http.Request) {

	log.Debug("Get build info handler")

	var (
		id    = utils.Vars(r)[`build`]
		owner = utils.Vars(r)[`owner`]
		name  = utils.Vars(r)[`name`]
		rm    = distribution.NewRepoModel(r.Context(), envs.Get().GetStorage())
		bm    = distribution.NewBuildModel(r.Context(), envs.Get().GetStorage())
	)

	rps, err := rm.Get("")
	if err != nil {
		log.V(logLevel).Errorf("Handler: Builds: get repo info err: %s", err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if rps == nil {
		log.V(logLevel).Warnf("Handler: Builds: repo `%s/%s` not found", owner, name)
		errors.New("repo").NotFound().Http(w)
		return
	}

	b, err := bm.Get(id)
	if err != nil {
		log.V(logLevel).Errorf("Handler: Build: get info err: %s", err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if b == nil {
		log.V(logLevel).Warnf("Handler: Build: build `%s` not found", id)
		errors.New("build").NotFound().Http(w)
		return
	}

	response, err := v.V1().Build().New(b).ToJson()
	if err != nil {
		log.V(logLevel).Errorf("Handler: Build: convert struct to json err: %v", err)
		errors.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(response); err != nil {
		log.Error("Error: write response", err)
		return
	}
}

func BuildLogs(w http.ResponseWriter, r *http.Request) {

	log.Debug("Get build logs handler")

	var (
		id    = utils.Vars(r)[`build`]
		owner = utils.Vars(r)[`owner`]
		name  = utils.Vars(r)[`name`]
		rm    = distribution.NewRepoModel(r.Context(), envs.Get().GetStorage())
		bm    = distribution.NewBuildModel(r.Context(), envs.Get().GetStorage())
	)

	rps, err := rm.Get("")
	if err != nil {
		log.V(logLevel).Errorf("Handler: Builds: get repo info err: %s", err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if rps == nil {
		log.V(logLevel).Warnf("Handler: Builds: repo `%s/%s` not found", owner, name)
		errors.New("repo").NotFound().Http(w)
		return
	}

	b, err := bm.Get(id)
	if err != nil {
		log.V(logLevel).Errorf("Handler: Build: get info err: %s", err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if b == nil {
		log.V(logLevel).Warnf("Handler: Build: build `%s` not found", id)
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

func BuildCancel(w http.ResponseWriter, r *http.Request) {

	var id = utils.Vars(r)[`build`]

	log.Debugf("Cancel build %s handler", id)

	var (
		owner = utils.Vars(r)[`owner`]
		name  = utils.Vars(r)[`name`]
		rm    = distribution.NewRepoModel(r.Context(), envs.Get().GetStorage())
		bm    = distribution.NewBuildModel(r.Context(), envs.Get().GetStorage())
	)

	rps, err := rm.Get("")
	if err != nil {
		log.V(logLevel).Errorf("Handler: Builds: get repo info err: %s", err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if rps == nil {
		log.V(logLevel).Warnf("Handler: Builds: repo `%s/%s` not found", owner, name)
		errors.New("repo").NotFound().Http(w)
		return
	}

	b, err := bm.Get(id)
	if err != nil {
		log.V(logLevel).Errorf("Handler: Build: get info err: %s", err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if b == nil {
		log.V(logLevel).Warnf("Handler: Build: build `%s` not found", id)
		errors.New("build").NotFound().Http(w)
		return
	}

	err = bm.Cancel(b)
	if err != nil {
		log.V(logLevel).Errorf("Handler: Build: cancel build err: %v", err)
		errors.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte{}); err != nil {
		log.Error("Error: write response", err)
		return
	}
}

func BuildCreate(w http.ResponseWriter, r *http.Request) {

	log.Debug("New build handler")

	var (
		owner = utils.Vars(r)[`owner`]
		repo  = utils.Vars(r)[`name`]
		bm    = distribution.NewBuildModel(r.Context(), envs.Get().GetStorage())
		rm    = distribution.NewRepoModel(r.Context(), envs.Get().GetStorage())
	)

	opts := new(types.BuildCreateOptions)
	if err := opts.DecodeAndValidate(r.Body); err != nil {
		log.V(logLevel).Errorf("Handler: Build: validation incoming data err: %v", err.Err())
		err.Http(w)
		return
	}

	rps, err := rm.Get("")
	if err != nil {
		log.V(logLevel).Errorf("Handler: Build: get repo info err: %s", err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if rps == nil {
		log.V(logLevel).Warnf("Handler: Build: repo `%s/%s` not found", owner, repo)
		errors.New("repo").NotFound().Http(w)
		return
	}

	if _, ok := rps.Tags[*opts.Tag]; !ok && !rps.Tags[*opts.Tag].AutoBuild {
		log.V(logLevel).Warn("Handler: Build: tag %s not found in build rules", *opts.Tag)
		errors.HTTP.NotFound(w)
		return
	}

	b, err := bm.Create(rps, rps.Tags[*opts.Tag].Name)
	if err != nil {
		log.V(logLevel).Errorf("Handler: Build: create build err: %v", err)
		errors.HTTP.InternalServerError(w)
		return
	}

	//if err := rm.UpdateTag(rps.Meta.ID, rps.Tags[*opts.Tag].Name); err != nil {
	//	log.V(logLevel).Errorf("Handler: Build: update repo tag err: %s", err)
	//	errors.HTTP.InternalServerError(w)
	//	return
	//}

	//if err := events.BuildProvisionRequest(envs.Get().GetRPC(), rps.Meta.ID, rps.Tags[*opts.Tag].Name); err != nil {
	//	log.V(logLevel).Errorf("Handler: Build: send event for provision build err: %s", err.Error())
	//	errors.HTTP.InternalServerError(w)
	//	return
	//}

	response, err := v.V1().Build().New(b).ToJson()
	if err != nil {
		log.V(logLevel).Errorf("Handler: Build: convert struct to json err: %v", err)
		errors.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(response); err != nil {
		log.Error("Error: write response", err)
		return
	}
}
