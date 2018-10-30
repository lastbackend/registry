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

package api

import (
	"github.com/lastbackend/registry/pkg/util/blob/config"
	"os"
	"os/signal"
	"syscall"

	"github.com/lastbackend/lastbackend/pkg/log"
	"github.com/lastbackend/registry/pkg/api/envs"
	"github.com/lastbackend/registry/pkg/api/http"
	"github.com/lastbackend/registry/pkg/storage"
	"github.com/lastbackend/registry/pkg/util/blob"
	"github.com/lastbackend/registry/pkg/util/blob/s3"
	"github.com/spf13/viper"
)

func Daemon() bool {

	var (
		sigs = make(chan os.Signal)
		done = make(chan bool, 1)
	)

	log.New(viper.GetInt("verbose"))
	log.Info("Start API server")

	stg, err := storage.Get(viper.GetString("psql"))
	if err != nil {
		log.Fatalf("Cannot initialize storage: %v", err)
	}

	envs.Get().SetStorage(stg)

	if viper.IsSet("api.blob_storage") {
		var blobStorage blob.IBlobStorage
		var cfg config.Config
		viper.UnmarshalKey("api.blob_storage", &cfg)

		switch viper.GetString("api.blob_storage.type") {
		case blob.DriverS3:
			blobStorage = s3.New(cfg)
		default:
			log.Fatalf("log driver not found")
		}
		envs.Get().SetBlobStorage(blobStorage)
	}

	go func() {
		opts := new(http.HttpOpts)
		if viper.IsSet("api.tls") {
			opts.Insecure = viper.GetBool("api.tls.insecure")
			opts.CertFile = viper.GetString("api.tls.cert")
			opts.KeyFile = viper.GetString("api.tls.key")
			opts.CaFile = viper.GetString("api.tls.ca")
		}

		if err := http.Listen(viper.GetString("api.host"), viper.GetInt("api.port"), opts); err != nil {
			log.Fatalf("Http server start error: %v", err)
		}
	}()

	// Handle SIGINT and SIGTERM.
	signal.Notify(sigs, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

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
