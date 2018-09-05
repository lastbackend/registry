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

package image

import (
	"net/http"

	"github.com/lastbackend/lastbackend/pkg/distribution/errors"
	"github.com/lastbackend/lastbackend/pkg/log"
	"github.com/lastbackend/registry/pkg/api/envs"
	"github.com/lastbackend/registry/pkg/api/types/v1"
	"github.com/lastbackend/registry/pkg/distribution"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/util/http/utils"
)

const (
	logLevel  = 2
	logPrefix = "registry:api:handler:image"
)

func ImageCreateH(w http.ResponseWriter, r *http.Request) {

	var (
		im = distribution.NewImageModel(r.Context(), envs.Get().GetStorage())
	)

	log.V(logLevel).Debugf("%s:create:> create new image", logPrefix)

	// request body struct
	rq := v1.Request().Image().CreateOptions()
	if e := rq.DecodeAndValidate(r.Body); e != nil {
		log.V(logLevel).Errorf("%s:create:> validation incoming data err: %v", logPrefix, e)
		e.Http(w)
		return
	}

	img, err := im.Get(rq.Owner, rq.Name)
	if err != nil {
		log.V(logLevel).Errorf("%s:create:> get image err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if img != nil {
		log.V(logLevel).Warnf("%s:create:> image `%s` already exists", logPrefix, rq.Name)
		errors.New("image").NotUnique("name").Http(w)
		return
	}

	opts := new(types.ImageCreateOptions)
	opts.Owner = rq.Owner
	opts.Name = rq.Name
	opts.Description = rq.Description
	opts.Private = rq.Private

	img, err = im.Create(opts)
	if err != nil {
		log.V(logLevel).Errorf("%s:create:> create image err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	response, err := v1.View().Image().New(img).ToJson()
	if err != nil {
		log.V(logLevel).Errorf("%s:create:> convert struct to json err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(response); err != nil {
		log.V(logLevel).Errorf("%s:create:> write response err: %v", logPrefix, err)
		return
	}
}

func ImageInfoH(w http.ResponseWriter, r *http.Request) {

	owner := utils.Vars(r)[`owner`]
	name := utils.Vars(r)[`name`]

	log.V(logLevel).Debugf("%s:info:> get image %s/%s info", logPrefix, owner, name)

	var (
		im = distribution.NewImageModel(r.Context(), envs.Get().GetStorage())
	)

	img, err := im.Get(owner, name)
	if err != nil {
		log.V(logLevel).Errorf("%s:info:> get image %s/%s err: %v", logPrefix, owner, name, err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if img == nil {
		log.V(logLevel).Warnf("%s:info:> image `%s/%s` not found", logPrefix, owner, name)
		errors.New("image").NotFound().Http(w)
		return
	}

	response, err := v1.View().Image().New(img).ToJson()
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

func ImageListH(w http.ResponseWriter, r *http.Request) {

	log.V(logLevel).Debugf("%s:list:> get repos list", logPrefix)

	var (
		im = distribution.NewImageModel(r.Context(), envs.Get().GetStorage())
	)

	opts := new(types.ImageListOptions)
	opts.Owner = utils.Vars(r)[`owner`]

	items, err := im.List(opts)
	if err != nil {
		log.V(logLevel).Errorf("%s:list:> get repos list err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	response, err := v1.View().Image().NewList(items).ToJson()
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

func ImageUpdateH(w http.ResponseWriter, r *http.Request) {

	log.V(logLevel).Debugf("%s:update:> update image info", logPrefix)

	var (
		im    = distribution.NewImageModel(r.Context(), envs.Get().GetStorage())
		owner = utils.Vars(r)[`owner`]
		name  = utils.Vars(r)[`name`]
	)

	// request body struct
	rq := v1.Request().Image().UpdateOptions()
	if e := rq.DecodeAndValidate(r.Body); e != nil {
		log.V(logLevel).Errorf("%s:update:> validation incoming data err: %v", logPrefix, e)
		e.Http(w)
		return
	}

	img, err := im.Get(owner, name)
	if err != nil {
		log.V(logLevel).Errorf("%s:update:> get image %s/%s err: %v", logPrefix, owner, name, err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if img == nil {
		log.V(logLevel).Warnf("%s:update:> image `%s/%s` not found", logPrefix, owner, name)
		errors.New("image").NotFound().Http(w)
		return
	}

	opts := new(types.ImageUpdateOptions)
	opts.Description = rq.Description
	opts.Private = rq.Private

	if err := im.Update(img, opts); err != nil {
		log.V(logLevel).Errorf("%s:update:> update image info err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	response, err := v1.View().Image().New(img).ToJson()
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

func ImageRemoveH(w http.ResponseWriter, r *http.Request) {

	log.V(logLevel).Debugf("%s:remove:> remove image", logPrefix)

	var (
		im    = distribution.NewImageModel(r.Context(), envs.Get().GetStorage())
		owner = utils.Vars(r)[`owner`]
		name  = utils.Vars(r)[`name`]
	)

	img, err := im.Get(owner, name)
	if err != nil {
		log.V(logLevel).Errorf("%s:remove:> get image %s/%s err: %v", logPrefix, owner, name, err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if img == nil {
		log.V(logLevel).Warnf("%s:remove:> image `%s/%s` not found", logPrefix, owner, name)
		errors.New("image").NotFound().Http(w)
		return
	}

	// TODO: remove physical image

	if err := im.Remove(img); err != nil {
		log.V(logLevel).Errorf("%s:remove:> remove image err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte{}); err != nil {
		log.V(logLevel).Errorf("%s:remove:> write response err: %v", logPrefix, err)
		return
	}
}
