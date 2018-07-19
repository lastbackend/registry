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

package builder

import (
	"github.com/lastbackend/registry/pkg/api/envs"
	"github.com/lastbackend/registry/pkg/api/types/v1"
	"github.com/lastbackend/registry/pkg/distribution"
	"github.com/lastbackend/registry/pkg/distribution/errors"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/util/http/utils"
	"net/http"
	)

const (
	logLevel  = 2
	logPrefix = "registry:api:handler:builder"
)

func BuilderConnectH(w http.ResponseWriter, r *http.Request) {

	log.V(logLevel).Debugf("%s:connect:> connect builder to system", logPrefix)

	var (
		bm  = distribution.NewBuilderModel(r.Context(), envs.Get().GetStorage())
		bid = utils.Vars(r)["builder"]
	)

	builder, err := bm.Get(bid)
	if err != nil {
		log.V(logLevel).Errorf("%s:connect:> get builder info err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	if builder == nil {

		opts := new(types.BuilderCreateOptions)
		opts.Hostname = bid
		opts.Online = true

		_, err := bm.Create(opts)
		if err != nil {
			log.V(logLevel).Errorf("%s:connect:> validation incoming data", logPrefix, err)
			errors.HTTP.InternalServerError(w)
			return
		}

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte{}); err != nil {
			log.Errorf("%s:connect:> write response err: %v", logPrefix, err)
			return
		}

		return
	}

	opts := new(types.BuilderUpdateOptions)
	online := true
	opts.Online = &online

	if err := bm.Update(builder, opts); err != nil {
		log.V(logLevel).Errorf("%s:connect:> update builder info err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte{}); err != nil {
		log.Errorf("%s:connect:> write response err: %v", logPrefix, err)
		return
	}
}

func BuilderCreateManifestH(w http.ResponseWriter, r *http.Request) {

	log.V(logLevel).Debugf("%s:create_manifest:> create manifest for builder", logPrefix)

	var (
		bdm = distribution.NewBuilderModel(r.Context(), envs.Get().GetStorage())
		blm = distribution.NewBuildModel(r.Context(), envs.Get().GetStorage())
		bid = utils.Vars(r)["builder"]
	)

	// request body struct
	rq := v1.Request().Builder().CreateManifestOptions()
	if e := rq.DecodeAndValidate(r.Body); e != nil {
		log.V(logLevel).Errorf("%s:create_manifest:> validation incoming data err: %v", logPrefix, e)
		e.Http(w)
		return
	}

	builder, err := bdm.Get(bid)
	if err != nil {
		log.V(logLevel).Errorf("%s:create_manifest:> get builder err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if builder == nil {
		log.V(logLevel).Warnf("%s:create_manifest:> builder `%s` not found", logPrefix, bid)
		errors.New("builder").NotFound().Http(w)
		return
	}

	builder_opts := new(types.BuilderUpdateOptions)
	online := true
	builder_opts.Online = &online

	if err := bdm.Update(builder, builder_opts); err != nil {
		log.V(logLevel).Errorf("%s:create_manifest:> update builder info err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	build, err := bdm.FindBuild(builder)
	if err != nil {
		log.V(logLevel).Errorf("%s:create_manifest:> get build for builder %s err: %v", logPrefix, bid, err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if build == nil {
		w.WriteHeader(http.StatusNotFound)
		if _, err := w.Write([]byte{}); err != nil {
			log.V(logLevel).Errorf("%s:create_manifest:> write response err: %v", logPrefix, err)
			return
		}
		return
	}

	opts := new(types.BuildUpdateTaskOptions)
	opts.TaskID = rq.TaskID

	err = blm.UpdateTask(build, opts)
	if err != nil {
		log.V(logLevel).Errorf("%s:create_manifest:> update build task err: %v", logPrefix, bid, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	response, err := v1.View().Builder().NewManifest(build.NewBuildManifest()).ToJson()
	if err != nil {
		log.V(logLevel).Errorf("%s:get_manifest:> convert struct to json err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(response); err != nil {
		log.V(logLevel).Errorf("%s:get_manifest:> write response err: %v", logPrefix, err)
		return
	}
}
