//
// Last.Backend LLC CONFIDENTIAL
// __________________
//
// [2014] - [2019] Last.Backend LLC
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

package system

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/lastbackend/lastbackend/pkg/log"
	"github.com/lastbackend/lastbackend/pkg/util/system"
	"github.com/lastbackend/registry/pkg/controller/envs"
	"github.com/lastbackend/registry/pkg/distribution"
	"github.com/lastbackend/registry/pkg/distribution/types"
)

// HeartBeat Interval
const heartBeatInterval = 5 // in seconds
const logPrefix = "system:process"
const logLevel = 7

type Process struct {
	process *types.Process
}

// Process register function
// The main purpose is to register process in the system
func (c *Process) Register(ctx context.Context, kind string) (*types.Process, error) {

	var err error

	c.process = new(types.Process)
	c.process.PID = system.GetPid()

	if c.process.Hostname, err = system.GetHostname(); err != nil {
		log.Errorf("%s:register:> get hostname: %v", logPrefix, err)
		return nil, err
	}

	return c.process, nil
}

// HeartBeat function - check controller for master mode
func (c *Process) HeartBeat(ctx context.Context, lead chan bool) {

	log.V(logLevel).Debugf("%s:heartbeat:> start heartbeat for: %s", logPrefix, c.process.Hostname)
	ticker := time.NewTicker(heartBeatInterval * time.Second)

	sm := distribution.NewSystemModel(context.Background(), envs.Get().GetStorage())

	sys, err := sm.Get()
	if err != nil {
		log.Errorf("%s:heartbeat:> remove dead controller err: %v", logPrefix, err)
		return
	}

	for range ticker.C {

		log.V(logLevel).Debugf("%s:> beat", logPrefix)

		opts := new(types.SystemUpdateControllerOptions)
		opts.Hostname = c.process.Hostname
		opts.Pid = c.process.PID

		err = sm.UpdateController(sys, opts)
		if err != nil {
			log.Errorf("%s:heartbeat:> register controller err: %v", logPrefix, err)
			return
		}

		match := strings.Split(sys.CtrlMaster, ":")
		if len(match) != 2 {
			log.Errorf("%s:heartbeat:> invalid controller argument err: %v", logPrefix, err)
			continue
		}

		pid, err := strconv.Atoi(match[0])
		if err != nil {
			log.Errorf("%s:heartbeat:> invalid convert pid argument err: %v", logPrefix, err)
			return
		}

		lead <- c.process.Hostname == match[1] && c.process.PID == pid
	}
}
