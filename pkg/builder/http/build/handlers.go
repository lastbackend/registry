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

	"github.com/lastbackend/registry/pkg/builder/envs"
	"github.com/lastbackend/registry/pkg/builder/types/request"
	"github.com/lastbackend/registry/pkg/log"
)

const (
	logLevel  = 2
	logPrefix = "api:handler:build"
)

// BuildCancelH - handler called build task cancel
func BuildCancelH(w http.ResponseWriter, r *http.Request) {

	log.Infof("%s:cancel:> cancel execute task", logPrefix)

	// request body struct
	rq := request.Request().Build().CancelOptions()
	if e := rq.DecodeAndValidate(r.Body); e != nil {
		log.V(logLevel).Errorf("%s:execute:> validation incoming data err: %v", logPrefix, e)
		e.Http(w)
		return
	}

	log.Debugf("%s:cancel:> cancel task [%s]", rq.Task)

	if err := envs.Get().GetBuilder().BuildCancel(r.Context(), rq.Task); err != nil {
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
		log.V(logLevel).Errorf("%s:execute:> validation incoming data err: %v", logPrefix, e)
		e.Http(w)
		return
	}

	log.Debugf("%s:logs:> task [%s] task [%s] log to endpoint [%s] ", rq.Task, rq.URI)

	if err := envs.Get().GetBuilder().BuildLogs(r.Context(), rq.Task, rq.URI); err != nil {
		log.Errorf("%s:logs:> get logs task err: %v", logPrefix, err)
		return
	}

	return
}
