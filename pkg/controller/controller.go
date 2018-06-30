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

	"github.com/lastbackend/registry/pkg/controller/envs"
	"github.com/lastbackend/registry/pkg/controller/runtime"
	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/storage"
	"github.com/spf13/viper"
)

func Daemon() bool {

	var (
		env  = envs.Get()
		sigs = make(chan os.Signal)
		done = make(chan bool, 1)
	)

	log.Info("Start Status Controller")

	stg, err := storage.Get(viper.GetString("psql"))
	if err != nil {
		log.Fatalf("Cannot initialize storage: %v", err)
	}
	env.SetStorage(stg)

	// Initialize Runtime
	runtime.NewRuntime()
	runtime.Loop()

	// Handle SIGINT and SIGTERM.
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			select {
			case <-sigs:
				done <- true
				return
			}
		}
	}()

	<-done

	log.Info("Handle SIGINT and SIGTERM.")
	return true
}
