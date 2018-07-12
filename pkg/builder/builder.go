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
	"github.com/spf13/viper"
)

// Daemon - run builder daemon
func Daemon() bool {

	log.Infof("Start builder service")

	sigs := make(chan os.Signal)

	ri, err := cri.New()
	if err != nil {
		log.Fatalf("Cannot initialize runtime: %v", err)
	}

	cfg := client.NewConfig()
	cfg.BearerToken = viper.GetString("token")
	cfg.TLS.Insecure = viper.GetBool("api.tls.insecure")
	cfg.TLS.CAFile = viper.GetString("api.tls.ca")
	cfg.TLS.CertFile = viper.GetString("api.tls.cert")
	cfg.TLS.KeyFile = viper.GetString("api.tls.key")

	endpoint := viper.GetString("api.endpoint")

	c, err := client.New(client.ClientHTTP, endpoint, cfg)
	if err != nil {
		log.Fatalf("Init client err: %s", err)
	}

	b := builder.New(
		ri,
		viper.GetString("builder.uuid"),
		viper.GetString("builder.docker.host"),
		viper.GetStringSlice("builder.extra_hosts"),
		viper.GetInt("builder.workers"),
		viper.GetString("builder.logs"),
	)

	envs.Get().SetHostname(viper.GetString("domain"))
	envs.Get().SetBuilder(b)
	envs.Get().SetClient(c)

	if err := b.Start(); err != nil {
		log.Fatalf("Start builder process err: %s", err)
	}

	go func() {
		opts := new(http.HttpOpts)
		opts.Insecure = viper.GetBool("builder.tls.insecure")
		opts.CertFile = viper.GetString("builder.tls.cert")
		opts.KeyFile = viper.GetString("builder.tls.key")

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
