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
	"context"
	"io"
	"os"
	"sync"
	"time"

	"github.com/lastbackend/registry/pkg/api/types/v1"
	"github.com/lastbackend/registry/pkg/builder/envs"
	"github.com/lastbackend/registry/pkg/distribution/errors"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/runtime/cri"
	"github.com/spf13/viper"

	"github.com/lastbackend/registry/pkg/api/types/v1/request"
	vv1 "github.com/lastbackend/registry/pkg/api/types/v1/views"
	lbt "github.com/lastbackend/registry/pkg/distribution/types"
	"io/ioutil"
)

const (
	logBuilderPrefix = "builder"
)

// The main entity that is responsible for
// the environment and existence of workers
type Builder struct {
	sync.RWMutex

	ctx      context.Context
	cancel   context.CancelFunc
	hostname string
	cri      cri.CRI
	limit    int

	opts BuilderOpts

	tasks chan *types.Task
	done  chan bool

	workers map[*worker]bool
}

type BuilderOpts struct {
	Stdout     bool
	DockerHost string
	ExtraHosts []string
	Limit      int
	RootCerts  []string
}

// Preparing the builder environment for workers
func New(cri cri.CRI, opts *BuilderOpts) *Builder {

	log.Infof("%s:new:> create builder", logBuilderPrefix)

	var (
		b   = new(Builder)
		ctx = context.Background()
	)

	if opts == nil {
		opts = new(BuilderOpts)
	}

	b.limit = opts.Limit
	b.cri = cri
	b.opts = *opts

	b.done = make(chan bool)
	b.tasks = make(chan *types.Task)

	b.workers = make(map[*worker]bool, 0)

	ctx = context.WithValue(ctx, "dockerHost", opts.DockerHost)
	ctx = context.WithValue(ctx, "extraHosts", opts.ExtraHosts)
	ctx = context.WithValue(ctx, "rootCerts", opts.RootCerts)
	ctx = context.WithValue(ctx, "blob-storage", types.AzureBlobStorage{
		AccountName: viper.GetString("storage.azure.account"),
		AccountKey:  viper.GetString("storage.azure.key"),
	})

	b.ctx, b.cancel = context.WithCancel(ctx)

	return b
}

// Initializing the builder and preparing the necessary resources for correct operation
func (b *Builder) Start() error {

	log.Infof("%s:start:> start builder", logWorkerPrefix)

	if err := b.configure(); err != nil {
		log.Errorf("%s:start:> configure builder err: %v", logWorkerPrefix, err)
		return err
	}
	if err := b.restore(); err != nil {
		log.Errorf("%s:start:> restore builder err: %v", logWorkerPrefix, err)
		return err
	}

	if err := b.connect(); err != nil {
		log.Errorf("%s:start:> connect builder err: %v", logWorkerPrefix, err)
		return err
	}

	go b.manage()
	go b.status()
	// TODO: subscribe to docker

	return nil
}

// Proxy logs writer from task
func (b *Builder) BuildLogs(ctx context.Context, id string, stream io.Writer) error {
	log.Infof("%s:new_build:> get build logs for writer: %s", logWorkerPrefix, id)

	for w := range b.workers {
		if w.pid == id {
			log.Infof("%s:cancel:> worker process was found: %s", logWorkerPrefix, w.pid)
			return w.logs(stream)
		}
	}

	return errors.New("process build is not active")
}

// Interrupting the build process
func (b *Builder) BuildCancel(ctx context.Context, id string) error {

	log.Infof("%s:cancel:> cancel build: %s", logWorkerPrefix, id)

	for w := range b.workers {
		if w.pid == id {
			log.Infof("%s:cancel:> worker process was found: %s", logWorkerPrefix, w.pid)
			w.cancel()
			break
		}
	}

	return nil
}

// Spawn - finished or failed workers will be restarted until context is closed
// If worker failed - then wait some time until respawn
func (b *Builder) manage() error {

	log.Infof("%s:manage:> run manage workers", logWorkerPrefix)

	go func() {
		for {
			select {
			case <-b.ctx.Done():
				log.Debugf("%s:manage:> stop manage", logWorkerPrefix)
				return
			case t := <-b.tasks:

				log.Debugf("%s:manage:> create new worker", logWorkerPrefix)
				w := newWorker(b.ctx, t.Meta.ID, b.cri)

				b.Lock()
				b.workers[w] = true
				b.Unlock()

				go func() {
					opts := new(workerOpts)
					opts.Stdout = b.opts.Stdout

					if err := w.run(t, opts); err != nil {
						log.Errorf("%s:manage:> start worker for provision err: %v", logWorkerPrefix, err)
					}

					b.Lock()
					delete(b.workers, w)
					b.Unlock()
				}()
			}
		}
	}()

	for {
		select {
		case <-b.ctx.Done():
			log.Debugf("%s:manage:> stop manage", logWorkerPrefix)
			return nil
		default:
			if len(b.workers) >= b.limit {
				continue
			}

			log.Debugf("%s:manage:> get new manifest", logWorkerPrefix)

			manifest, err := envs.Get().GetClient().V1().Builder(envs.Get().GetHostname()).Manifest(b.ctx)
			if err != nil {
				if err.Error() != "Manifest not found" {
					log.Errorf("%s:manage:> get manifest err: %v", logWorkerPrefix, err)
				}
				select {
				// error delay
				case <-time.After(5 * time.Second):
				case <-b.ctx.Done():
					return nil
				}
				continue
			}

			if manifest != nil {
				log.Debugf("%s:manage:> create new task", logWorkerPrefix)
				b.tasks <- b.newTask(manifest)
			}

			<-time.After(5 * time.Second)
		}
	}
}

// Configure builder
func (b *Builder) configure() error {
	// Check image exists
	imageExists := func(name string) bool {
		images, err := b.cri.ImageList(b.ctx)
		if err != nil {
			return false
		}

		for _, image := range images {
			for _, tag := range image.RepoTags {
				if tag == name+":latest" || tag == name {
					return true
				}
			}
		}

		log.Warnf("%s:configure:> image %s not found", logWorkerPrefix, name)

		return false
	}

	var images = []string{
		"docker:dind",
		"docker:git",
	}

	for _, img := range images {
		if imageExists(img) {
			continue
		}

		req, err := b.cri.ImagePull(b.ctx, &lbt.SpecTemplateContainerImage{
			Name: img,
		})
		if err != nil {
			log.Errorf("%s:configure:> pull image err: %v", logWorkerPrefix, err)
			return err
		}
		// TODO handle output in more beautiful way
		if _, err := io.Copy(os.Stdout, req); err != nil {
			log.Errorf("%s:configure:> copy writer to stdout err: %v", logWorkerPrefix, err)
		}

		if err := req.Close(); err != nil {
			log.Errorf("%s:configure:> close writer err: %v", logWorkerPrefix, err)
			return err
		}
	}

	return nil
}

func (b *Builder) restore() error {

	containers, err := b.cri.ContainerList(b.ctx, true)
	if err != nil {
		log.Errorf("Pods restore error: %v", err)
		return err
	}

	for _, c := range containers {

		info, err := b.cri.ContainerInspect(b.ctx, c.ID)
		if err != nil {
			log.Errorf("inspect container err: %v", err)
			return err
		}

		// TODO: here you need get builder state and implement the logic of workers recovery
		if err := b.cri.ContainerRemove(b.ctx, info.ID, true, true); err != nil {
			log.Errorf("remove container err: %v", err)
			return err
		}

	}

	return nil
}

func (b Builder) newTask(manifest *vv1.BuildManifest) *types.Task {
	t := new(types.Task)

	t.Meta.ID = manifest.Meta.ID

	t.Spec.Source.Url = manifest.Spec.Source.Url
	t.Spec.Source.Branch = manifest.Spec.Source.Branch

	t.Spec.Image.Host = manifest.Spec.Image.Host
	t.Spec.Image.Name = manifest.Spec.Image.Name
	t.Spec.Image.Owner = manifest.Spec.Image.Owner
	t.Spec.Image.Tag = manifest.Spec.Image.Tag
	t.Spec.Image.Auth = manifest.Spec.Image.Auth

	t.Spec.Config.Dockerfile = manifest.Spec.Config.Dockerfile
	t.Spec.Config.Context = manifest.Spec.Config.Context
	t.Spec.Config.Workdir = manifest.Spec.Config.Workdir
	t.Spec.Config.EnvVars = manifest.Spec.Config.EnvVars
	t.Spec.Config.Command = manifest.Spec.Config.Command

	return t
}

func (b *Builder) connect() error {
	opts := v1.Request().Builder().ConnectOptions()
	opts.IP = envs.Get().GetIP()
	opts.Port = uint16(viper.GetInt("builder.port"))

	if viper.IsSet("builder.tls") {
		opts.TLS = !viper.GetBool("builder.tls.insecure")

		if opts.TLS {
			caData, err := ioutil.ReadFile(viper.GetString("builder.tls.ca"))
			if err != nil {
				log.Errorf("%s:start:> read ca cert file err: %v", logWorkerPrefix, err)
				return err
			}

			certData, err := ioutil.ReadFile(viper.GetString("builder.tls.client_cert"))
			if err != nil {
				log.Errorf("%s:start:> read client cert file err: %v", logWorkerPrefix, err)
				return err
			}

			keyData, err := ioutil.ReadFile(viper.GetString("builder.tls.client_key"))
			if err != nil {
				log.Errorf("%s:start:> read client key file err: %v", logWorkerPrefix, err)
				return err
			}

			opts.SSL = new(request.SSL)
			opts.SSL.CA = caData
			opts.SSL.Key = keyData
			opts.SSL.Cert = certData
		}
	}

	if err := envs.Get().GetClient().V1().Builder(envs.Get().GetHostname()).Connect(b.ctx, opts); err != nil {
		log.Errorf("%s:start:> send event connect builder err: %v", logWorkerPrefix, err)
		return err
	}

	return nil
}

func (b *Builder) status() error {

	opts := v1.Request().Builder().StatusUpdateOptions()
	// TODO: send data builder info to api

	for {
		<-time.After(10 * time.Second)
		if err := envs.Get().GetClient().V1().Builder(envs.Get().GetHostname()).SetStatus(b.ctx, opts); err != nil {
			log.Errorf("%s:start:> send event status builder err: %v", logWorkerPrefix, err)
			return err
		}
	}

	return nil
}

// Shutdown builder process
func (b *Builder) Shutdown() {

	log.Infof("%s:shutdown:> shutdown builder process", logWorkerPrefix)

	b.done <- true
}

// Notification that the builder has completed its work
func (b *Builder) Done() <-chan bool {
	return b.done
}
