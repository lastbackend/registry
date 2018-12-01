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
	"github.com/lastbackend/registry/pkg/builder/envs"
	"net/http"

	"github.com/lastbackend/lastbackend/pkg/log"
	"github.com/lastbackend/registry/pkg/builder/types"
	"github.com/lastbackend/registry/pkg/builder/types/v1"
)

const (
	logLevel  = 2
	logPrefix = "api:handler:build"
)

func BuilderUpdateH(w http.ResponseWriter, r *http.Request) {
	log.V(logLevel).Infof("%s:update:> update builder", logPrefix)

	// request body struct
	rq := v1.Request().Builder().BuilderUpdateManifestOptions()
	if e := rq.DecodeAndValidate(r.Body); e != nil {
		log.V(logLevel).Errorf("%s:update:> validation incoming data err: %v", logPrefix, e)
		e.Http(w)
		return
	}

	opts := new(types.BuilderManifest)

	if rq.Limits != nil {
		opts.Limits = new(types.BuilderLimits)
		opts.Limits.WorkerLimit = rq.Limits.WorkerLimit
		opts.Limits.WorkerMemory = rq.Limits.WorkerMemory
		opts.Limits.Workers = rq.Limits.Workers
	}

	log.V(logLevel).Debugf("%s:update:> have changes", logPrefix)
	if err := envs.Get().GetBuilder().Update(r.Context(), opts); err != nil {
		log.V(logLevel).Errorf("%s:update:> cancel build err: %v", logPrefix, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte{}); err != nil {
		log.V(logLevel).Errorf("%s:update:> write response err: %v", logPrefix, err)
		return
	}
}
