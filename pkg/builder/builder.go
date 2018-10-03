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
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/lastbackend/lastbackend/pkg/log"
	"github.com/lastbackend/lastbackend/pkg/runtime/cii/cii"
	"github.com/lastbackend/lastbackend/pkg/runtime/cri/cri"
	"github.com/lastbackend/registry/pkg/api/client"
	"github.com/lastbackend/registry/pkg/builder/builder"
	"github.com/lastbackend/registry/pkg/builder/envs"
	"github.com/lastbackend/registry/pkg/builder/http"
	"github.com/lastbackend/registry/pkg/util/blob"
	"github.com/lastbackend/registry/pkg/util/blob/azure"
	"github.com/lastbackend/registry/pkg/util/blob/s3"
	"github.com/lastbackend/registry/pkg/util/system"
	"github.com/spf13/viper"
)

// Daemon - run builder daemon
func Daemon() bool {

	sigs := make(chan os.Signal)

	log.New(viper.GetInt("verbose"))

	criDriver := viper.GetString("runtime.cri.type")
	_cri, err := cri.New(criDriver, viper.GetStringMap(fmt.Sprintf("runtime.%s", criDriver)))
	if err != nil {
		log.Errorf("Cannot initialize cri: %v", err)
	}

	ciiDriver := viper.GetString("runtime.cii.type")
	_cii, err := cii.New(ciiDriver, viper.GetStringMap(fmt.Sprintf("runtime.%s", ciiDriver)))
	if err != nil {
		log.Errorf("Cannot initialize cii: %v", err)
	}

	cfg := client.NewConfig()

	cfg.BearerToken = viper.GetString("secret.token")

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
	bo.DindHost = viper.GetString("builder.dind.host")
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
				viper.GetString("builder.blob_storage.region"),
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
			panic("unknown blog storage driver")
		}

		envs.Get().SetBlobStorage(blobStorage)
	}

	b := builder.New(_cri, _cii, bo)

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
