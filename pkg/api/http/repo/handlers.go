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
	"net/http"
	"github.com/lastbackend/registry/pkg/api/envs"
	"github.com/lastbackend/registry/pkg/distribution"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/util/http/utils"
	"github.com/lastbackend/lastbackend/pkg/distribution/errors"
	v "github.com/lastbackend/registry/pkg/api/views"
)

const logLevel = 2

func RepoCreateH(w http.ResponseWriter, r *http.Request) {

	if r.Context().Value("account") == nil {
		errors.HTTP.Unauthorized(w)
		return
	}

	var (
		rm = distribution.NewRepoModel(r.Context(), envs.Get().GetStorage())
	)

	log.V(logLevel).Debug("Handler: Repo: create repo")

	// request body struct
	opts := new(types.RepoCreateOptions)
	if err := opts.DecodeAndValidate(r.Body); err != nil {
		log.V(logLevel).Errorf("Handler: Repo: validation incoming data err: %s", err.Err())
		err.Http(w)
		return
	}

	rps, err := rm.Get("")
	if err != nil {
		log.V(logLevel).Errorf("Handler: Repo: get repo err: %v", err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if rps != nil {
		log.V(logLevel).Warnf("Handler: Repo: repo `%s` already exists", opts.Name)
		errors.New("repo").NotUnique("name").Http(w)
		return
	}

	rps, err = rm.Create(opts)
	if err != nil {
		log.V(logLevel).Errorf("Handler: Repo: create repo err: %v", err)
		errors.HTTP.InternalServerError(w)
		return
	}

	//for _, tag := range rps.Tags {
		//if err := events.BuildProvisionRequest(envs.Get().GetRPC(), rps.Meta.ID, tag.Name); err != nil {
		//	log.V(logLevel).Errorf("Handler: Repo: send event for provision build err: %s", err.Error())
		//	errors.HTTP.InternalServerError(w)
		//	return
		//}
	//}

	response, err := v.V1().Repo().New(rps).ToJson()
	if err != nil {
		log.V(logLevel).Errorf("Handler: Repo: convert struct to json err: %s", err)
		errors.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(response); err != nil {
		log.V(logLevel).Errorf("Handler: Repo: write response err: %s", err)
		return
	}
}

func RepoInfoH(w http.ResponseWriter, r *http.Request) {

	owner := utils.Vars(r)[`owner`]
	name := utils.Vars(r)[`name`]

	log.V(logLevel).Debugf("Handler: Repo: get repo %s/%s info", owner, name)

	if r.Context().Value("account") == nil {
		errors.HTTP.Unauthorized(w)
		return
	}

	if r.Context().Value("owner_account") == nil {
		errors.HTTP.Forbidden(w)
		return
	}

	var (
		rm = distribution.NewRepoModel(r.Context(), envs.Get().GetStorage())
	)

	rps, err := rm.Get("")
	if err != nil {
		log.V(logLevel).Errorf("Handler: Repo: get repo info err: %s", err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if rps == nil {
		log.V(logLevel).Warnf("Handler: Repo: repo `%s/%s` not found", owner, name)
		errors.New("repo").NotFound().Http(w)
		return
	}

	response, err := v.V1().Repo().New(rps).ToJson()
	if err != nil {
		log.V(logLevel).Errorf("Handler: Repo: convert struct to json err: %s", err)
		errors.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(response); err != nil {
		log.V(logLevel).Errorf("Handler: Repo: write response err: %s", err)
		return
	}
}

func RepoListH(w http.ResponseWriter, r *http.Request) {

	log.V(logLevel).Debug("Handler: Repo: get repos list")

	if r.Context().Value("account") == nil {
		errors.HTTP.Unauthorized(w)
		return
	}

	if r.Context().Value("owner_account") == nil {
		errors.HTTP.Unauthorized(w)
		return
	}

	var (
		rm  = distribution.NewRepoModel(r.Context(), envs.Get().GetStorage())
	)

	item, err := rm.List()
	if err != nil {
		log.V(logLevel).Errorf("Handler: Repo: get repos list err: %s", err)
		errors.HTTP.InternalServerError(w)
		return
	}

	response, err := v.V1().Repo().NewList(item).ToJson()
	if err != nil {
		log.V(logLevel).Errorf("Handler: Repo: convert struct to json err: %s", err)
		errors.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(response); err != nil {
		log.V(logLevel).Errorf("Handler: Repo: write response err: %s", err)
		return
	}
}

func RepoUpdateH(w http.ResponseWriter, r *http.Request) {

	log.V(logLevel).Debug("Handler: Repo: update repo info")

	var (
		rm    = distribution.NewRepoModel(r.Context(), envs.Get().GetStorage())
		owner = utils.Vars(r)[`owner`]
		name  = utils.Vars(r)[`name`]
	)

	// request body struct
	opts := new(types.RepoUpdateOptions)
	if err := opts.DecodeAndValidate(r.Body); err != nil {
		log.V(logLevel).Errorf("Handler: Repo: validation incoming data err: %s", err.Err())
		err.Http(w)
		return
	}

	rps, err := rm.Get("")
	if err != nil {
		log.V(logLevel).Errorf("Handler: Repo: get repo info err: %s", err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if rps == nil {
		log.V(logLevel).Warnf("Handler: Repo: repo `%s/%s` not found", owner, name)
		errors.New("repo").NotFound().Http(w)
		return
	}

	if err := rm.Update(rps, opts); err != nil {
		log.V(logLevel).Errorf("Handler: Repo: update repo info err: %s", err)
		errors.HTTP.InternalServerError(w)
		return
	}

	response, err := v.V1().Repo().New(rps).ToJson()
	if err != nil {
		log.V(logLevel).Errorf("Handler: Repo: convert struct to json err: %s", err)
		errors.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(response); err != nil {
		log.V(logLevel).Errorf("Handler: Repo: write response err: %s", err)
		return
	}
}

func RepoRemoveH(w http.ResponseWriter, r *http.Request) {

	log.V(logLevel).Debug("Handler: Repo: remove repo")

	var (
		rm    = distribution.NewRepoModel(r.Context(), envs.Get().GetStorage())
		owner = utils.Vars(r)[`owner`]
		name  = utils.Vars(r)[`name`]
	)

	rps, err := rm.Get("")
	if err != nil {
		log.V(logLevel).Errorf("Handler: Repo: get repo err: %s", err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if rps == nil {
		log.V(logLevel).Warnf("Handler: Repo: repo `%s/%s` not found", owner, name)
		errors.New("repo").NotFound().Http(w)
		return
	}

	if err := rm.Remove(rps.Meta.ID); err != nil {
		log.V(logLevel).Errorf("Handler: Repo: remove repo err: %s", err)
		errors.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte{}); err != nil {
		log.V(logLevel).Errorf("Handler: Repo: write response err: %s", err)
		return
	}
}
