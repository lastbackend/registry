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
	"net/http"
	"strings"

	"github.com/lastbackend/registry/pkg/api/envs"
	"github.com/lastbackend/registry/pkg/api/types/v1"
	"github.com/lastbackend/registry/pkg/builder/client"
	rv1 "github.com/lastbackend/registry/pkg/builder/types/v1/request"
	"github.com/lastbackend/registry/pkg/distribution"
	"github.com/lastbackend/registry/pkg/distribution/errors"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/util/converter"
	"github.com/lastbackend/registry/pkg/util/http/utils"
	"io"
)

const (
	logLevel    = 2
	logPrefix   = "registry:api:handler:build"
	BUFFER_SIZE = 512
)

func BuildCreateH(w http.ResponseWriter, r *http.Request) {

	log.V(logLevel).Debugf("%s:build:> execute build image", logPrefix)

	var (
		im = distribution.NewImageModel(r.Context(), envs.Get().GetStorage())
		bm = distribution.NewBuildModel(r.Context(), envs.Get().GetStorage())
	)

	owner := utils.Vars(r)["owner"]
	name := utils.Vars(r)["name"]

	// request body struct
	rq := v1.Request().Build().BuildExecuteOptions()
	if e := rq.DecodeAndValidate(r.Body); e != nil {
		log.V(logLevel).Errorf("%s:build:> validation incoming data err: %v", logPrefix, e)
		e.Http(w)
		return
	}

	img, err := im.Get(owner, name)
	if err != nil {
		log.V(logLevel).Errorf("%s:build:> get image %s/%s err: %v", logPrefix, owner, name, err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if img == nil {
		log.V(logLevel).Warnf("%s:build:> image `%s/%s` not found", logPrefix, owner, name)
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

	opts.Spec.Context = rq.Context

	if len(rq.DockerFile) == 0 {
		opts.Spec.DockerFile = types.ImageDefaultContextLocation
	}

	opts.Spec.EnvVars = rq.EnvVars
	opts.Spec.Workdir = rq.Workdir
	opts.Spec.Command = rq.Command

	opts.Labels = rq.Labels

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

func BuildListH(w http.ResponseWriter, r *http.Request) {

	log.V(logLevel).Debugf("%s:list:> get builds list", logPrefix)

	var (
		im = distribution.NewImageModel(r.Context(), envs.Get().GetStorage())
		bm = distribution.NewBuildModel(r.Context(), envs.Get().GetStorage())

		owner  = utils.Vars(r)[`owner`]
		name   = utils.Vars(r)[`name`]
		active = r.URL.Query().Get("active")
	)

	img, err := im.Get(owner, name)
	if err != nil {
		log.V(logLevel).Errorf("%s:list:> get image %s/%s err: %v", logPrefix, owner, name, err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if img == nil {
		log.V(logLevel).Warnf("%s:list:> image `%s/%s` not found", logPrefix, owner, name)
		errors.New("image").NotFound().Http(w)
		return
	}

	opts := new(types.BuildListOptions)

	if len(active) != 0 {
		a, err := converter.ParseBool(active)
		if err != nil {
			log.V(logLevel).Errorf("%s:list:> parse active flag err: %v", logPrefix, err)
			errors.New("image").BadParameter("active").Http(w)
			return
		}
		opts.Active = &a
	}

	items, err := bm.List(img, opts)
	if err != nil {
		log.V(logLevel).Errorf("%s:list:> get builds list err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	response, err := v1.View().Build().NewList(items).ToJson()
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

func BuildInfoH(w http.ResponseWriter, r *http.Request) {

	log.V(logLevel).Debugf("%s:info:> get builds list", logPrefix)

	var (
		im = distribution.NewImageModel(r.Context(), envs.Get().GetStorage())
		bm = distribution.NewBuildModel(r.Context(), envs.Get().GetStorage())

		owner = utils.Vars(r)[`owner`]
		name  = utils.Vars(r)[`name`]
		bid   = utils.Vars(r)[`build`]
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

	build, err := bm.Get(bid)
	if err != nil {
		log.V(logLevel).Errorf("%s:info:> get build err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	response, err := v1.View().Build().New(build).ToJson()
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

func BuildCancelH(w http.ResponseWriter, r *http.Request) {

	bid := utils.Vars(r)["build"]

	log.V(logLevel).Debugf("%s:cancel:> cancel build %s process", logPrefix, bid)

	var (
		bdm = distribution.NewBuilderModel(r.Context(), envs.Get().GetStorage())
		bm  = distribution.NewBuildModel(r.Context(), envs.Get().GetStorage())
	)

	build, err := bm.Get(bid)
	if err != nil {
		log.V(logLevel).Errorf("%s:cancel:> get build by id err: %v", logPrefix, bid, err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if build == nil {
		log.V(logLevel).Warnf("%s:cancel:> build `%s` not found", logPrefix, bid)
		errors.New("build").NotFound().Http(w)
		return
	}

	builder, err := bdm.Get(build.Meta.Builder)
	if err != nil {
		log.V(logLevel).Errorf("%s:cancel:> get builder by name err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if builder == nil {
		log.V(logLevel).Warnf("%s:cancel:> builder %s not found", logPrefix, build.Meta.Builder)
		errors.New("build").NotFound().Http(w)
		return
	}

	cfg := client.NewConfig()
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
			log.V(logLevel).Errorf("%s:cancel:> create http client err: %v", logPrefix, err)
			errors.HTTP.InternalServerError(w)
			return
		}
	}

	err = httpcli.V1().Build(build.Meta.ID).Cancel(r.Context())
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
			log.V(logLevel).Errorf("%s:cancel:> get builder info from builder `%s` err: %v", logPrefix, endpoint, err)
			errors.HTTP.InternalServerError(w)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte{}); err != nil {
		log.V(logLevel).Errorf("%s:cancel:> write response err: %v", logPrefix, err)
		return
	}
}

func BuildLogsH(w http.ResponseWriter, r *http.Request) {

	bid := utils.Vars(r)["build"]

	log.V(logLevel).Debugf("%s:logs:> get logs build `%s`", logPrefix, bid)

	var (
		bdm = distribution.NewBuilderModel(r.Context(), envs.Get().GetStorage())
		bm  = distribution.NewBuildModel(r.Context(), envs.Get().GetStorage())
	)

	notify := w.(http.CloseNotifier).CloseNotify()
	done := make(chan bool, 1)

	pipe := func(writer io.Writer, reader io.ReadCloser) {

		var buffer = make([]byte, BUFFER_SIZE)

		for {
			select {
			case <-done:
				reader.Close()
				return
			default:

				n, err := reader.Read(buffer)
				if err != nil {
					log.Errorf("%s:logs:> read bytes from pipe err: %v", logPrefix, err)
					return
				}

				_, err = func(p []byte) (n int, err error) {

					n, err = writer.Write(p)
					if err != nil {
						log.Errorf("%s:logs:> write bytes to pipe err: %v", logPrefix, err)
						return n, err
					}

					if f, ok := writer.(http.Flusher); ok {
						f.Flush()
					}

					return n, nil
				}(buffer[0:n])

				if err != nil {
					log.Errorf("%s:logs:> written to pipe err: %v", logPrefix, err)
					return
				}

				for i := 0; i < n; i++ {
					buffer[i] = 0
				}
			}
		}
	}

	go func() {
		<-notify
		log.Debugf("%s:logs:> HTTP connection just closed.", logPrefix)
		done <- true
	}()

	build, err := bm.Get(bid)
	if err != nil {
		log.V(logLevel).Errorf("%s:logs:> get build by id %s err: %v", logPrefix, bid, err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if build == nil {
		log.V(logLevel).Warnf("%s:logs:> build `%s` not found", logPrefix, bid)
		errors.New("build").NotFound().Http(w)
		return
	}

	if build.Status.Done || build.Status.Error {
		read, write := io.Pipe()
		defer write.Close()
		go envs.Get().GetBlobStorage().ReadToWriter(build.Meta.ID, write)
		pipe(w, read)
		return
	}

	builder, err := bdm.Get(build.Meta.Builder)
	if err != nil {
		log.V(logLevel).Errorf("%s:logs:> get builder by name err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if builder == nil {
		log.V(logLevel).Warnf("%s:logs:> builder %s not found", logPrefix, build.Meta.Builder)
		errors.New("build").NotFound().Http(w)
		return
	}

	cfg := client.NewConfig()
	cfg.Timeout = 0
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
			log.V(logLevel).Errorf("%s:logs:> create http client err: %v", logPrefix, err)
			errors.HTTP.InternalServerError(w)
			return
		}
	}

	res, err := httpcli.V1().Build(build.Meta.ID).Logs(r.Context(), &rv1.BuildLogsOptions{Follow: true})
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
			log.V(logLevel).Errorf("%s:logs:> get builder info from builder `%s` err: %v", logPrefix, endpoint, err)
			errors.HTTP.InternalServerError(w)
			return
		}
	}

	pipe(w, res)

}

func BuildStatusUpdateH(w http.ResponseWriter, r *http.Request) {

	log.Debugf("%s:update_status:> set build status info handler", logPrefix)

	var (
		bm  = distribution.NewBuildModel(r.Context(), envs.Get().GetStorage())
		bid = utils.Vars(r)[`build`]
	)

	// request body struct
	rq := v1.Request().Build().BuildStatusOptions()
	if e := rq.DecodeAndValidate(r.Body); e != nil {
		log.V(logLevel).Errorf("%s:update_status:> validation incoming data err: %v", logPrefix, e)
		e.Http(w)
		return
	}

	build, err := bm.Get(bid)
	if err != nil {
		log.V(logLevel).Errorf("%s:update_status:> get build by id err: %v", logPrefix, bid, err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if build == nil {
		log.V(logLevel).Warnf("%s:update_status:> build `%s` not found", logPrefix, bid)
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

func BuildInfoUpdateH(w http.ResponseWriter, r *http.Request) {

	log.Debugf("%s:update_info:> set build info handler", logPrefix)

	var (
		bm  = distribution.NewBuildModel(r.Context(), envs.Get().GetStorage())
		bid = utils.Vars(r)[`build`]
	)

	// request body struct
	rq := v1.Request().Build().BuildInfoOptions()
	if e := rq.DecodeAndValidate(r.Body); e != nil {
		log.V(logLevel).Errorf("%s:update_info:> validation incoming data err: %v", logPrefix, e)
		e.Http(w)
		return
	}

	build, err := bm.Get(bid)
	if err != nil {
		log.V(logLevel).Errorf("%s:update_info:> get build by id %s err: %v", logPrefix, bid, err)
		errors.HTTP.InternalServerError(w)
		return
	}
	if build == nil {
		log.V(logLevel).Warnf("%s:update_info:> build `%s` not found", logPrefix, bid)
		errors.New("build").NotFound().Http(w)
		return
	}

	opts := new(types.BuildUpdateInfoOptions)
	opts.Size = rq.Size
	opts.Hash = rq.Hash

	if err := bm.UpdateInfo(build, opts); err != nil {
		log.V(logLevel).Errorf("%s:update_info:> update build err: %v", logPrefix, bid, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte{}); err != nil {
		log.V(logLevel).Errorf("%s:update_info:> write response err: %v", logPrefix, err)
		return
	}
}
