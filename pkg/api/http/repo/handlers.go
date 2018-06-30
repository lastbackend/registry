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

package repo

import (
	"github.com/lastbackend/registry/pkg/distribution/errors"
	"github.com/lastbackend/registry/pkg/api/envs"
	v "github.com/lastbackend/registry/pkg/api/views"
	"github.com/lastbackend/registry/pkg/distribution"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/util/http/utils"
	"net/http"
)

const (
	logLevel  = 2
	logPrefix = "registry:api:handler:repo"
)

func RepoCreateH(w http.ResponseWriter, r *http.Request) {

	var (
		rm = distribution.NewRepoModel(r.Context(), envs.Get().GetStorage())
	)

	log.V(logLevel).Debugf("%s:create:> create new repo", logPrefix)

	// request body struct
	opts := new(types.RepoCreateOptions)
	if err := opts.DecodeAndValidate(r.Body); err != nil {
		log.V(logLevel).Errorf("%s:create:> validation incoming data err: %v", logPrefix, err)
		err.Http(w)
		return
	}

	image := opts.Spec.Image

	rps, err := rm.Get(image.Owner, image.Name)
	if err != nil {
		log.V(logLevel).Errorf("%s:create:> get repo err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if rps != nil {
		log.V(logLevel).Warnf("%s:create:> repo `%s` already exists", logPrefix, image.Name)
		errors.New("repo").NotUnique("name").Http(w)
		return
	}

	rps, err = rm.Create(opts)
	if err != nil {
		log.V(logLevel).Errorf("%s:create:> create repo err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	response, err := v.V1().Repo().New(rps).ToJson()
	if err != nil {
		log.V(logLevel).Errorf("%s:create:> convert struct to json err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(response); err != nil {
		log.V(logLevel).Errorf("%s:create:> write response err: %v", logPrefix, err)
		return
	}
}

func RepoInfoH(w http.ResponseWriter, r *http.Request) {

	owner := utils.Vars(r)[`owner`]
	name := utils.Vars(r)[`name`]

	log.V(logLevel).Debugf("%s:info:> get repo %s/%s info", logPrefix, owner, name)

	var (
		rm = distribution.NewRepoModel(r.Context(), envs.Get().GetStorage())
	)

	rps, err := rm.Get(owner, name)
	if err != nil {
		log.V(logLevel).Errorf("%s:info:> get repo info err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if rps == nil {
		log.V(logLevel).Warnf("%s:info:> repo `%s/%s` not found", logPrefix, owner, name)
		errors.New("repo").NotFound().Http(w)
		return
	}

	response, err := v.V1().Repo().New(rps).ToJson()
	if err != nil {
		log.V(logLevel).Errorf("%s:info:> convert struct to json err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(response); err != nil {
		log.V(logLevel).Errorf("%s:info:> write response err: %v", logPrefix, err)
		return
	}
}

func RepoListH(w http.ResponseWriter, r *http.Request) {

	log.V(logLevel).Debugf("%s:list:> get repos list", logPrefix)

	var (
		rm = distribution.NewRepoModel(r.Context(), envs.Get().GetStorage())
	)

	item, err := rm.List()
	if err != nil {
		log.V(logLevel).Errorf("%s:list:> get repos list err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	response, err := v.V1().Repo().NewList(item).ToJson()
	if err != nil {
		log.V(logLevel).Errorf("%s:list:> convert struct to json err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(response); err != nil {
		log.V(logLevel).Errorf("%s:list:> write response err: %v", logPrefix, err)
		return
	}
}

func RepoUpdateH(w http.ResponseWriter, r *http.Request) {

	log.V(logLevel).Debugf("%s:update:> update repo info", logPrefix)

	var (
		rm    = distribution.NewRepoModel(r.Context(), envs.Get().GetStorage())
		owner = utils.Vars(r)[`owner`]
		name  = utils.Vars(r)[`name`]
	)

	// request body struct
	opts := new(types.RepoUpdateOptions)
	if err := opts.DecodeAndValidate(r.Body); err != nil {
		log.V(logLevel).Errorf("%s:update:> validation incoming data err: %v", logPrefix, err)
		err.Http(w)
		return
	}

	rps, err := rm.Get(owner, name)
	if err != nil {
		log.V(logLevel).Errorf("%s:update:> get repo info err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if rps == nil {
		log.V(logLevel).Warnf("%s:update:> repo `%s/%s` not found", logPrefix, owner, name)
		errors.New("repo").NotFound().Http(w)
		return
	}

	if err := rm.Update(rps, opts); err != nil {
		log.V(logLevel).Errorf("%s:update:> update repo info err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	response, err := v.V1().Repo().New(rps).ToJson()
	if err != nil {
		log.V(logLevel).Errorf("%s:update:> convert struct to json err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(response); err != nil {
		log.V(logLevel).Errorf("%s:update:> write response err: %v", logPrefix, err)
		return
	}
}

func RepoRemoveH(w http.ResponseWriter, r *http.Request) {

	log.V(logLevel).Debugf("%s:remove:> remove repo", logPrefix)

	var (
		rm    = distribution.NewRepoModel(r.Context(), envs.Get().GetStorage())
		owner = utils.Vars(r)[`owner`]
		name  = utils.Vars(r)[`name`]
	)

	rps, err := rm.Get(owner, name)
	if err != nil {
		log.V(logLevel).Errorf("%s:remove:> get repo err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if rps == nil {
		log.V(logLevel).Warnf("%s:remove:> repo `%s/%s` not found", logPrefix, owner, name)
		errors.New("repo").NotFound().Http(w)
		return
	}

	if err := rm.Remove(rps.Meta.ID); err != nil {
		log.V(logLevel).Errorf("%s:remove:> remove repo err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte{}); err != nil {
		log.V(logLevel).Errorf("%s:remove:> write response err: %v", logPrefix, err)
		return
	}
}
