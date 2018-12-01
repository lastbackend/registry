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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/lastbackend/lastbackend/pkg/distribution/errors"
	"github.com/lastbackend/lastbackend/pkg/log"
	"github.com/lastbackend/lastbackend/pkg/runtime/cii"
	"github.com/lastbackend/lastbackend/pkg/runtime/cri"
	"github.com/lastbackend/registry/pkg/api/types/v1"
	"github.com/lastbackend/registry/pkg/api/types/v1/request"
	"github.com/lastbackend/registry/pkg/builder/envs"
	"github.com/lastbackend/registry/pkg/builder/runtime"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/spf13/viper"

	lbt "github.com/lastbackend/lastbackend/pkg/distribution/types"
	vv1 "github.com/lastbackend/registry/pkg/api/types/v1/views"
	rbt "github.com/lastbackend/registry/pkg/builder/types"
)

const (
	logBuilderPrefix       = "builder"
	defaultWorkerMemory    = 512
	defaultWorkerInstances = 1
)

// The main entity that is responsible for
// the environment and existence of workers
type Builder struct {
	sync.RWMutex

	ctx      context.Context
	cancel   context.CancelFunc
	hostname string
	cri      cri.CRI
	cii      cii.CII

	limits limits

	opts BuilderOpts

	tasks chan *types.Task
	done  chan bool

	workers map[*worker]bool
}

type limits struct {
	Worker struct {
		Instances int
		Memory    int64
	}
}

type BuilderOpts struct {
	Stdout       bool
	DindHost     string
	ExtraHosts   []string
	WorkerLimit  int
	WorkerMemory int64
	RootCerts    []string
}

// Preparing the builder environment for workers
func New(cri cri.CRI, cii cii.CII, opts *BuilderOpts) *Builder {

	log.Infof("%s:new:> create builder", logBuilderPrefix)

	var (
		b   = new(Builder)
		ctx = context.Background()
	)

	if opts == nil {
		opts = new(BuilderOpts)
	}

	if opts.WorkerMemory != 0 {
		b.limits.Worker.Instances = opts.WorkerLimit
	} else {
		b.limits.Worker.Instances = defaultWorkerInstances
	}

	if opts.WorkerMemory != 0 {
		b.limits.Worker.Memory = opts.WorkerMemory
	} else {
		b.limits.Worker.Memory = defaultWorkerMemory
	}

	b.cri = cri
	b.cii = cii
	b.opts = *opts

	b.done = make(chan bool)
	b.tasks = make(chan *types.Task)

	b.workers = make(map[*worker]bool, 0)

	ctx = context.WithValue(ctx, "dockerHost", opts.DindHost)
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

// Shutdown builder process
func (b *Builder) Shutdown() {

	log.Infof("%s:shutdown:> shutdown builder process", logWorkerPrefix)

	b.done <- true
}

// Update builder manifest
func (b *Builder) Update(ctx context.Context, opts *rbt.BuilderManifest) error {

	log.Infof("%s:update:> update builder manifest", logWorkerPrefix)

	if opts.Limits != nil {
		if opts.Limits.WorkerLimit {
			b.limits.Worker.Instances = opts.Limits.Workers
			b.limits.Worker.Memory = opts.Limits.WorkerMemory
		} else {

			if b.opts.WorkerLimit != 0 {
				b.limits.Worker.Instances = b.opts.WorkerLimit
			} else {
				b.limits.Worker.Instances = defaultWorkerInstances
			}

			if b.opts.WorkerMemory != 0 {
				b.limits.Worker.Memory = b.opts.WorkerMemory
			} else {
				b.limits.Worker.Memory = defaultWorkerMemory
			}

		}
	}

	return nil
}

// Notification that the builder has completed its work
func (b *Builder) Done() <-chan bool {
	return b.done
}

func (b *Builder) ActiveWorkers() uint {
	return uint(len(b.workers))
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
					opts.Memory = b.limits.Worker.Memory

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
			if len(b.workers) >= b.limits.Worker.Instances {
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

	localImages, err := b.cii.List(b.ctx)
	if err != nil {
		return err
	}

	// Check image exists
	imageExists := func(name string) bool {
		for _, image := range localImages {
			for _, tag := range image.Meta.Tags {
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

		_, err := b.cii.Pull(b.ctx, &lbt.ImageManifest{Name: img}, os.Stdout)
		if err != nil {
			log.Errorf("%s:configure:> pull image err: %v", logWorkerPrefix, err)
			return err
		}

	}

	return nil
}

// Restore builder
func (b *Builder) restore() error {

	containers, err := b.cri.List(b.ctx, true)
	if err != nil {
		log.Errorf("Pods restore error: %v", err)
		return err
	}

	for _, c := range containers {

		info, err := b.cri.Inspect(b.ctx, c.ID)
		if err != nil {
			log.Errorf("inspect container err: %v", err)
			return err
		}

		if _, ok := c.Labels[lbt.ContainerTypeLBC]; !ok {
			continue
		}

		// TODO: here you need get builder state and implement the logic of workers recovery
		if err := b.cri.Remove(b.ctx, info.ID, true, true); err != nil {
			log.Errorf("remove container err: %v", err)
			return err
		}

	}

	return nil
}

// Create new build task
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

// Connect builder
func (b *Builder) connect() error {

	opts := v1.Request().Builder().ConnectOptions()
	opts.IP = envs.Get().GetIP()
	opts.Port = uint16(viper.GetInt("builder.port"))

	info := runtime.BuilderInfo()
	opts.System.Architecture = info.Architecture
	opts.System.OSName = info.OSName
	opts.System.OSType = info.OSType
	opts.System.Version = info.Version

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

	manifest, err := envs.Get().GetClient().V1().Builder(envs.Get().GetHostname()).Connect(b.ctx, opts)
	if err != nil {
		log.Errorf("%s:start:> send event connect builder err: %v", logWorkerPrefix, err)
		return err
	}

	buf, _ := json.Marshal(manifest)
	fmt.Println(string(buf))

	modify := false

	mopts := new(rbt.BuilderManifest)
	if manifest.Limits != nil {
		modify = true
		mopts.Limits = new(rbt.BuilderLimits)
		mopts.Limits.WorkerLimit = manifest.Limits.WorkerLimit
		mopts.Limits.WorkerMemory = int64(manifest.Limits.WorkerMemory)
		mopts.Limits.Workers = manifest.Limits.Workers
	}

	if modify {
		if err := b.Update(b.ctx, mopts); err != nil {
			log.Errorf("%s:start:> update builder manifest err: %v", logWorkerPrefix, err)
			return err
		}
	}

	return nil
}

// Set builder status
func (b *Builder) status() error {

	opts := v1.Request().Builder().StatusUpdateOptions()

	status := runtime.BuilderStatus()

	opts.Allocated.Cpu = status.Allocated.Cpu
	opts.Allocated.Memory = status.Allocated.Memory
	opts.Allocated.Workers = status.Allocated.Workers
	opts.Allocated.Storage = status.Allocated.Storage

	opts.Capacity.Cpu = status.Capacity.Cpu
	opts.Capacity.Memory = status.Capacity.Memory
	opts.Capacity.Workers = status.Capacity.Workers
	opts.Capacity.Storage = status.Capacity.Storage

	for {
		<-time.After(10 * time.Second)
		err := envs.Get().GetClient().V1().Builder(envs.Get().GetHostname()).SetStatus(b.ctx, opts)
		if err != nil {
			log.Errorf("%s:start:> send event status builder err: %v", logWorkerPrefix, err)
			return err
		}

	}
}
