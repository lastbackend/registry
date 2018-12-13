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
	"fmt"
	"net/http"
	"strings"

	"github.com/lastbackend/lastbackend/pkg/distribution/errors"
	"github.com/lastbackend/lastbackend/pkg/log"
	"github.com/lastbackend/registry/pkg/api/envs"
	"github.com/lastbackend/registry/pkg/api/types/v1"
	"github.com/lastbackend/registry/pkg/builder/client"
	"github.com/lastbackend/registry/pkg/builder/types/v1/request"
	"github.com/lastbackend/registry/pkg/distribution"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/util/http/utils"
	"github.com/spf13/viper"
)

const (
	logLevel  = 2
	logPrefix = "registry:api:handler:builder"
)

func BuilderListH(w http.ResponseWriter, r *http.Request) {

	log.V(logLevel).Debugf("%s:list:> get builders list", logPrefix)

	var (
		bm = distribution.NewBuilderModel(r.Context(), envs.Get().GetStorage())
	)

	builders, err := bm.List()
	if err != nil {
		log.V(logLevel).Errorf("%s:list:> get builder list err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	response, err := v1.View().Builder().NewList(builders).ToJson()
	if err != nil {
		log.V(logLevel).Errorf("%s:list:> convert struct to json err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(response); err != nil {
		log.Errorf("%s:list:> write response err: %v", logPrefix, err)
		return
	}
}

func BuilderUpdateH(w http.ResponseWriter, r *http.Request) {

	log.V(logLevel).Debugf("%s:update:> update builder data", logPrefix)

	// request body struct
	rq := v1.Request().Builder().UpdateOptions()
	if e := rq.DecodeAndValidate(r.Body); e != nil {
		log.V(logLevel).Errorf("%s:update:> validation incoming data err: %v", logPrefix, e)
		e.Http(w)
		return
	}

	var (
		bm       = distribution.NewBuilderModel(r.Context(), envs.Get().GetStorage())
		hostname = utils.Vars(r)["builder"]
	)

	builder, err := bm.Get(hostname)
	if err != nil {
		log.V(logLevel).Errorf("%s:update:> get builder info err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	if rq.WorkerLimit && rq.Workers > uint(builder.Status.Capacity.Memory/types.DEFAULT_MIN_WORKER_MEMORY) {
		errors.New("builder").BadParameter("workers").Http(w)
		return
	}

	if rq.WorkerLimit && rq.WorkerMemory > (builder.Status.Capacity.Memory/uint64(rq.Workers)) {
		errors.New("builder").BadParameter("worker.memory").Http(w)
		return
	}

	opts := new(types.BuilderUpdateOptions)
	opts.Limits = new(types.BuilderLimits)
	opts.Limits.Workers = rq.Workers
	opts.Limits.WorkerMemory = rq.WorkerMemory
	opts.Limits.WorkerLimit = rq.WorkerLimit

	if err := bm.Update(builder, opts); err != nil {
		log.V(logLevel).Errorf("%s:update:> update builder info err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	cfg := client.NewConfig()
	cfg.BearerToken = viper.GetString("secret.token")
	if builder.Spec.Network.SSL != nil {
		cfg.TLS = client.NewTLSConfig()
		cfg.TLS.CertData = builder.Spec.Network.SSL.Cert
		cfg.TLS.KeyData = builder.Spec.Network.SSL.Key
		cfg.TLS.CAData = builder.Spec.Network.SSL.CA
	}

	endpoint := fmt.Sprintf("%s:%d", builder.Spec.Network.IP, builder.Spec.Network.Port)

	httpcli, err := client.New(client.ClientHTTP, endpoint, cfg)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "failed to create tls config"):
			errors.HTTP.BadRequest(w, "SSL certificate is failed")
			return
		default:
			log.V(logLevel).Errorf("%s:update:> create http client err: %v", logPrefix, err)
			errors.HTTP.InternalServerError(w)
			return
		}
	}

	uopts := new(request.BuilderUpdateManifestOptions)
	uopts.Limits = new(request.BuilderLimitConfig)
	uopts.Limits.WorkerMemory = builder.Spec.Limits.WorkerMemory
	uopts.Limits.Workers = builder.Spec.Limits.Workers

	err = httpcli.V1().Update(r.Context(), uopts)
	if err != nil {
		switch {
		case err.Error() == "Unauthorized":
			errors.HTTP.BadRequest(w, "access token not set")
			return
		case strings.Contains(err.Error(), "x509: certificate signed by unknown authority"):
			errors.HTTP.BadRequest(w, "ssl certificate is failed")
			return
		case strings.Contains(err.Error(), "net/http: HTTP/1.x transport connection broken"):
			errors.HTTP.BadRequest(w, "tls transport connection broken")
			return
		default:
			log.V(logLevel).Errorf("%s:update:> update builder `%s` err: %v", logPrefix, endpoint, err)
			errors.HTTP.InternalServerError(w)
			return
		}
	}

	response, err := v1.View().Builder().New(builder).ToJson()
	if err != nil {
		log.V(logLevel).Errorf("%s:update:> convert struct to json err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(response); err != nil {
		log.Errorf("%s:update:> write response err: %v", logPrefix, err)
		return
	}
}

func BuilderConnectH(w http.ResponseWriter, r *http.Request) {

	log.V(logLevel).Debugf("%s:connect:> connect builder to system", logPrefix)

	var (
		bm       = distribution.NewBuilderModel(r.Context(), envs.Get().GetStorage())
		hostname = utils.Vars(r)["builder"]
	)

	// request body struct
	rq := v1.Request().Builder().ConnectOptions()
	if e := rq.DecodeAndValidate(r.Body); e != nil {
		log.V(logLevel).Errorf("%s:connect:> validation incoming data err: %v", logPrefix, e)
		e.Http(w)
		return
	}

	builder, err := bm.Get(hostname)
	if err != nil {
		log.V(logLevel).Errorf("%s:connect:> get builder info err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	if builder == nil {

		opts := new(types.BuilderCreateOptions)
		opts.Hostname = hostname
		opts.IP = rq.IP
		opts.Port = rq.Port
		opts.TLS = rq.TLS
		opts.Online = true

		if rq.SSL != nil {
			opts.SSL = new(types.SSL)
			opts.SSL.CA = rq.SSL.CA
			opts.SSL.Cert = rq.SSL.Cert
			opts.SSL.Key = rq.SSL.Key
		}

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
	opts.Hostname = &hostname
	opts.Online = &online
	opts.IP = &rq.IP
	opts.TLS = &rq.TLS
	opts.Port = &rq.Port

	if rq.SSL != nil {
		opts.SSL = new(types.SSL)
		opts.SSL.CA = rq.SSL.CA
		opts.SSL.Cert = rq.SSL.Cert
		opts.SSL.Key = rq.SSL.Key
	}

	if err := bm.Update(builder, opts); err != nil {
		log.V(logLevel).Errorf("%s:connect:> update builder info err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	response, err := v1.View().Builder().NewConfigManifest(builder).ToJson()
	if err != nil {
		log.V(logLevel).Errorf("%s:connect:> convert struct to json err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(response); err != nil {
		log.V(logLevel).Errorf("%s:connect:> write response err: %v", logPrefix, err)
		return
	}
}

func BuilderStatusH(w http.ResponseWriter, r *http.Request) {

	log.V(logLevel).Debugf("%s:status:> connect builder to system", logPrefix)

	var (
		bm       = distribution.NewBuilderModel(r.Context(), envs.Get().GetStorage())
		hostname = utils.Vars(r)["builder"]
	)

	// request body struct
	rq := v1.Request().Builder().StatusUpdateOptions()
	if e := rq.DecodeAndValidate(r.Body); e != nil {
		log.V(logLevel).Errorf("%s:status:> validation incoming data err: %v", logPrefix, e)
		e.Http(w)
		return
	}

	builder, err := bm.Get(hostname)
	if err != nil {
		log.V(logLevel).Errorf("%s:status:> get builder info err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	opts := new(types.BuilderUpdateOptions)
	online := true
	opts.Online = &online

	opts.Allocated = new(types.BuilderResources)
	opts.Allocated.Workers = rq.Allocated.Workers
	opts.Allocated.Memory = rq.Allocated.Memory
	opts.Allocated.Cpu = rq.Allocated.Cpu
	opts.Allocated.Storage = rq.Allocated.Storage

	opts.Capacity = new(types.BuilderResources)
	opts.Capacity.Workers = rq.Capacity.Workers
	opts.Capacity.Memory = rq.Capacity.Memory
	opts.Capacity.Cpu = rq.Capacity.Cpu
	opts.Capacity.Storage = rq.Capacity.Storage

	opts.Usage = new(types.BuilderResources)
	opts.Usage.Workers = rq.Usage.Workers
	opts.Usage.Memory = rq.Usage.Memory
	opts.Usage.Cpu = rq.Usage.Cpu
	opts.Usage.Storage = rq.Usage.Storage

	if err := bm.Update(builder, opts); err != nil {
		log.V(logLevel).Errorf("%s:status:> update builder info err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte{}); err != nil {
		log.Errorf("%s:status:> write response err: %v", logPrefix, err)
		return
	}
}

func BuilderGetManifestH(w http.ResponseWriter, r *http.Request) {

	log.V(logLevel).Debugf("%s:get_manifest:> create manifest for builder", logPrefix)

	var (
		bdm = distribution.NewBuilderModel(r.Context(), envs.Get().GetStorage())
		bid = utils.Vars(r)["builder"]
	)

	builder, err := bdm.Get(bid)
	if err != nil {
		log.V(logLevel).Errorf("%s:get_manifest:> get builder err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if builder == nil {
		log.V(logLevel).Warnf("%s:get_manifest:> builder `%s` not found", logPrefix, bid)
		errors.New("builder").NotFound().Http(w)
		return
	}

	build, err := bdm.FindBuild(builder)
	if err != nil {
		log.V(logLevel).Errorf("%s:get_manifest:> get build for builder %s err: %v", logPrefix, bid, err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if build == nil {
		log.V(logLevel).Warnf("%s:get_manifest:> manifest `%s` not found", logPrefix, bid)
		errors.New("manifest").NotFound().Http(w)
		return
	}

	response, err := v1.View().Builder().NewBuildManifest(build.NewBuildManifest()).ToJson()
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
