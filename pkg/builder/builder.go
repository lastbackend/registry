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
	"os"
	"os/signal"
	"syscall"

	"github.com/lastbackend/registry/pkg/api/client"
	"github.com/lastbackend/registry/pkg/builder/builder"
	"github.com/lastbackend/registry/pkg/builder/envs"
	"github.com/lastbackend/registry/pkg/builder/http"
	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/runtime/cri/cri"
	"github.com/lastbackend/registry/pkg/util/blob/s3"
	"github.com/lastbackend/registry/pkg/util/system"
	"github.com/spf13/viper"
	"github.com/lastbackend/registry/pkg/util/blob"
	"github.com/lastbackend/registry/pkg/util/blob/azure"
)

// Daemon - run builder daemon
func Daemon() bool {

	sigs := make(chan os.Signal)

	ri, err := cri.New()
	if err != nil {
		log.Fatalf("Cannot initialize runtime: %v", err)
	}

	cfg := client.NewConfig()

	if viper.IsSet("registry.tls") && !viper.GetBool("registry.tls.insecure") {
		cfg.TLS = client.NewTLSConfig()
		cfg.TLS.CertFile = viper.GetString("registry.tls.cert")
		cfg.TLS.KeyFile = viper.GetString("registry.tls.key")
		cfg.TLS.CAFile = viper.GetString("registry.tls.ca")
	}

	endpoint := viper.GetString("registry.uri")
	c, err := client.New(client.ClientHTTP, endpoint, cfg)
	if err != nil {
		log.Fatalf("Init client err: %s", err)
	}

	bo := new(builder.BuilderOpts)
	bo.DockerHost = viper.GetString("builder.docker.host")
	bo.ExtraHosts = viper.GetStringSlice("builder.extra_hosts")
	bo.Limit = viper.GetInt("builder.workers")
	bo.RootCerts = viper.GetStringSlice("builder.cacerts")

	if viper.IsSet("builder.logger") {
		bo.Stdout = viper.GetBool("builder.logger.stdout")
	}

	if viper.IsSet("builder.blob_storage") {
		var blobStorage blob.IBlobStorage
		switch viper.GetString("builder.blob_storage.type") {
		case "s3":
			blobStorage = s3.New(
				viper.GetString("builder.blob_storage.endpoint"),
				viper.GetString("builder.blob_storage.id"),
				viper.GetString("builder.blob_storage.secret"),
				viper.GetString("builder.blob_storage.bucket_name"),
				viper.GetBool("builder.blob_storage.ssl"),
			)
		case "azure":
			blobStorage = azure.New(
				viper.GetString("builder.blob_storage.endpoint"),
				viper.GetString("builder.blob_storage.account"),
				viper.GetString("builder.blob_storage.key"),
				viper.GetString("builder.blob_storage.container"),
				viper.GetBool("builder.blob_storage.ssl"),
			)
		default:

		}

		envs.Get().SetBlobStorage(blobStorage)
	}

	b := builder.New(ri, bo)

	if viper.IsSet("builder.ip") {
		envs.Get().SetIP(viper.GetString("builder.ip"))
	} else {
		ip, err := system.GetNodeIP()
		if err != nil {
			log.Errorf("get ip address err: %v", err)
			log.Fatalf("get ip err: %v", err)
		}
		envs.Get().SetIP(ip)
	}

	hostname, err := system.GetHostname()
	if err != nil {
		log.Fatalf("get hostname err: %v", err)
	}

	envs.Get().SetHostname(hostname)
	envs.Get().SetBuilder(b)
	envs.Get().SetClient(c)

	if err := b.Start(); err != nil {
		log.Fatalf("Start builder process err: %s", err)
	}

	go func() {
		opts := new(http.HttpOpts)
		opts.Insecure = viper.GetBool("builder.tls.insecure")
		opts.CertFile = viper.GetString("builder.tls.server_cert")
		opts.KeyFile = viper.GetString("builder.tls.server_key")
		opts.CaFile = viper.GetString("builder.tls.ca")

		if err := http.Listen(viper.GetString("builder.host"), viper.GetInt("builder.port"), opts); err != nil {
			log.Fatalf("Http server start error: %v", err)
		}
	}()

	// Handle SIGINT and SIGTERM.
	signal.Notify(sigs, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			select {
			case <-sigs:
				b.Shutdown()
				return
			}
		}
	}()

	<-b.Done()

	log.Info("Handle SIGINT and SIGTERM.")

	return true
}
