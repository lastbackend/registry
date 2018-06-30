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

package cluster

import (
	"net/http"

	"github.com/lastbackend/registry/pkg/builder/envs"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/builder/types/request"
)

const (
	logLevel  = 2
	logPrefix = "api:handler:build"
)

// BuildExecuteH - handler called build task execute
func BuildExecuteH(w http.ResponseWriter, r *http.Request) {

	log.Infof("%s:execute:> execute task for create image build", logPrefix)

	// request body struct
	rq := request.Request().Build().ExecuteOptions()
	if e := rq.DecodeAndValidate(r.Body); e != nil {
		log.V(logLevel).Errorf("%s:execute:> validation incoming data err: %v", logPrefix, e.Err())
		e.Http(w)
		return
	}

	job := new(types.BuildJob)
	job.ID = rq.ID
	job.Repo = rq.Repo
	job.Branch = rq.Branch
	job.LogUri = rq.LogUri

	job.Meta.ID = rq.Meta.ID
	job.Meta.LogsUri = rq.Meta.LogsUri

	job.Image.Host = rq.Image.Host
	job.Image.Name = rq.Image.Name
	job.Image.Owner = rq.Image.Owner
	job.Image.Tag = rq.Image.Tag
	job.Image.Token = rq.Image.Token

	job.Config.Dockerfile = rq.Config.Dockerfile
	job.Config.Workdir = rq.Config.Workdir
	job.Config.EnvVars = rq.Config.EnvVars

	if err := envs.Get().GetBuilder().NewTask(r.Context(), job); err != nil {
		log.Errorf("%s:execute:> create new task err: %v", logPrefix, err)
		return
	}

	return
}

// BuildCancelH - handler called build task cancel
func BuildCancelH(w http.ResponseWriter, r *http.Request) {

	log.Infof("%s:cancel:> cancel execute task", logPrefix)

	// request body struct
	rq := request.Request().Build().CancelOptions()
	if e := rq.DecodeAndValidate(r.Body); e != nil {
		log.V(logLevel).Errorf("%s:execute:> validation incoming data err: %v", logPrefix, e.Err())
		e.Http(w)
		return
	}

	log.Debugf("%s:cancel:> cancel task [%s]", rq.Task)

	if err := envs.Get().GetBuilder().CancelTask(r.Context(), rq.Task); err != nil {
		log.Errorf("%s:cancel:> cancel task err: %v", logPrefix, err)
		return
	}

	return
}

// BuildLogsCancelH - handler for get logs stream
func BuildLogsH(w http.ResponseWriter, r *http.Request) {

	log.Infof("%s:logs:> get task logs stream", logPrefix)

	// request body struct
	rq := request.Request().Build().LogsOptions()
	if e := rq.DecodeAndValidate(r.Body); e != nil {
		log.V(logLevel).Errorf("%s:execute:> validation incoming data err: %v", logPrefix, e.Err())
		e.Http(w)
		return
	}

	log.Debugf("%s:logs:> task [%s] task [%s] log to endpoint [%s] ", rq.Task, rq.URI)

	if err := envs.Get().GetBuilder().LogsTask(r.Context(), rq.Task, rq.URI); err != nil {
		log.Errorf("%s:logs:> get logs task err: %v", logPrefix, err)
		return
	}

	return
}
