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
	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/util/http/utils"
	"github.com/lastbackend/registry/pkg/util/stream"
)

const (
	logLevel  = 2
	logPrefix = "api:handler:build"
)

// BuildCancelH - handler called build cancel
func BuildCancelH(w http.ResponseWriter, r *http.Request) {

	build := utils.Vars(r)["build"]

	log.V(logLevel).Infof("%s:cancel:> cancel execute build with build %s", logPrefix, build)

	if err := envs.Get().GetBuilder().BuildCancel(r.Context(), build); err != nil {
		log.V(logLevel).Errorf("%s:cancel:> cancel build err: %v", logPrefix, err)
		return
	}

	return
}

// BuildLogsCancelH - handler for get build logs stream
func BuildLogsH(w http.ResponseWriter, r *http.Request) {

	buildid := utils.Vars(r)["build"]

	log.V(logLevel).Infof("%s:logs:> get logs stream for build with buildid %s", logPrefix, buildid)

	notify := w.(http.CloseNotifier).CloseNotify()
	done := make(chan bool, 1)

	s := stream.NewStream()

	go func() {
		<-notify
		log.Debugf("%s:logs:> HTTP connection just closed.", logPrefix)
		s.Close()
		done <- true
	}()

	err := envs.Get().GetBuilder().BuildLogs(r.Context(), buildid, w)
	if err != nil {
		log.V(logLevel).Errorf("%s:logs:> get logs build err: %v", logPrefix, err)
		return
	}

	<-done
}
