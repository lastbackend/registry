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

package controller

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/lastbackend/lastbackend/pkg/log"
	"github.com/lastbackend/registry/pkg/controller/envs"
	"github.com/lastbackend/registry/pkg/controller/runtime"
	"github.com/lastbackend/registry/pkg/controller/state"
	"github.com/lastbackend/registry/pkg/storage"
	"github.com/spf13/viper"
)

const (
	logPrefix = "controller:daemon"
)

func Daemon() bool {

	var (
		env  = envs.Get()
		sigs = make(chan os.Signal, 1)
		done = make(chan bool, 1)
	)

	log.New(viper.GetInt("verbose"))
	log.Infof("%s:> start controller", logPrefix)

	stg, err := storage.Get(viper.GetString("psql"))
	if err != nil {
		log.Fatalf("%s:> cannot initialize storage: %v", logPrefix, err)
	}
	env.SetStorage(stg)
	env.SetState(state.New())

	// Initialize Runtime
	r := runtime.New()
	go r.Loop()

	// Handle SIGINT and SIGTERM.
	signal.Notify(sigs, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			select {
			case <-sigs:
				r.Stop()
				done <- true
				return
			}
		}
	}()

	<-done

	log.Infof("%s:> handle SIGINT and SIGTERM.", logPrefix)

	return true
}
