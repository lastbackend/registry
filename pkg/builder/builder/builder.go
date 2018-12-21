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
	"github.com/lastbackend/lastbackend/pkg/util/resource"
	"github.com/lastbackend/registry/pkg/api/types/v1"
	"github.com/lastbackend/registry/pkg/api/types/v1/request"
	"github.com/lastbackend/registry/pkg/builder/envs"
	"github.com/lastbackend/registry/pkg/builder/runtime"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/spf13/viper"

	lbt "github.com/lastbackend/lastbackend/pkg/distribution/types"
	vv1 "github.com/lastbackend/registry/pkg/api/types/v1/views"
)

const (
	logBuilderPrefix = "builder"
)

const (
	defaultSendStatusDelay = 5 * time.Second
	defaultWorkerInstances = 1
	defaultWorkerRAM       = "512MB"
	defaultWorkerCPU       = "0.5"
	defaultReserveMemory   = "512MB"
	minimumReserveMemory   = 512
	minimumRequiredMemory  = 1024
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

	reserveMemory int64

	limits *limits

	opts BuilderOpts

	tasks chan *types.Task
	done  chan bool

	workers map[*worker]bool
}

type BuilderOpts struct {
	Stdout     bool
	DindHost   string
	ExtraHosts []string
	RootCerts  []string
}

type limits struct {
	workerInstances int
	workerRAM       int64
	workerCPU       int64
	workerStorage   int64
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

	b.cri = cri
	b.cii = cii
	b.opts = *opts

	b.done = make(chan bool)
	b.tasks = make(chan *types.Task)

	b.workers = make(map[*worker]bool, 0)

	if err := b.SetReserveMemory(defaultReserveMemory); err != nil {
		log.Errorf("%s:> init reserve number err: %v", logWorkerPrefix, err)
	}

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

	status := runtime.BuilderStatus()

	if status.Capacity.RAM < minimumRequiredMemory {
		return fmt.Errorf("recommended system ram more then %dMB", minimumRequiredMemory)
	}

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

// Notification that the builder has completed its work
func (b *Builder) Done() <-chan bool {
	return b.done
}

func (b *Builder) ActiveWorkers() int {
	return len(b.workers)
}

func (b *Builder) SetReserveMemory(memory string) error {

	log.Infof("%s:set_reserve_memory:> set reserve memory", logWorkerPrefix)

	if len(memory) == 0 {
		return fmt.Errorf("bad reserve memory parameter")
	}

	rm, err := resource.DecodeMemoryResource(memory)
	if err != nil {
		log.Infof("%s:set_reserve_memory:> invalid reserve memory parameter err: err", logWorkerPrefix, err)
		panic(err)
	}
	b.reserveMemory = rm / 1024 / 1024

	if b.reserveMemory < minimumReserveMemory {
		return fmt.Errorf("recommended reserved ram more then %dMB", minimumReserveMemory)
	}

	return nil
}

func (b *Builder) SetWorkerLimits(instances int, ram, cpu string) error {

	log.Infof("%s:set_worker_limits:> set worker limits", logWorkerPrefix)

	if instances <= 0 {
		return fmt.Errorf("bad workers instances parameter")
	}

	if len(ram) == 0 {
		ram = defaultWorkerRAM
	}

	if len(cpu) == 0 {
		cpu = defaultWorkerCPU
	}

	b.limits = new(limits)
	b.limits.workerInstances = instances

	rm, err := resource.DecodeMemoryResource(ram)
	if err != nil {
		log.Infof("%s:set_worker_limits:> invalid resource worker ram parameter err: err", logWorkerPrefix, err)
		return err
	}
	b.limits.workerRAM = rm

	cp, err := resource.DecodeCpuResource(cpu)
	if err != nil {
		log.Infof("%s:set_worker_limits:> invalid resource worker cpu parameter err: err", logWorkerPrefix, err)
		return err
	}

	b.limits.workerCPU = cp

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

					opts.RAM = types.DEFAULT_MIN_WORKER_MEMORY

					if b.limits != nil && b.limits.workerRAM != 0 {
						opts.RAM = b.limits.workerRAM
					}

					opts.CPU = types.DEFAULT_MIN_WORKER_CPU

					if b.limits != nil && b.limits.workerCPU != 0 {
						opts.CPU = b.limits.workerCPU
					}

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

	instances := defaultWorkerInstances
	if b.limits != nil && b.limits.workerInstances != 0 {
		instances = int(b.limits.workerInstances)
	}

	for {

		select {
		case <-b.ctx.Done():
			log.Debugf("%s:manage:> stop manage", logWorkerPrefix)
			return nil
		default:

			if len(b.workers) >= instances {
				<-time.After(10 * time.Second)
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

	log.Infof("%s:manage:> run configure builder", logWorkerPrefix)

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

	log.Infof("%s:manage:> restore builder processes", logWorkerPrefix)

	containers, err := b.cri.List(b.ctx, true)
	if err != nil {
		log.Errorf("Container restore error: %v", err)
		return err
	}

	for _, c := range containers {

		info, err := b.cri.Inspect(b.ctx, c.ID)
		if err != nil {
			log.Errorf("inspect container err: %v", err)
			return err
		}

		if _, ok := c.Labels[lbt.ContainerTypeLBR]; !ok {
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

	if b.limits != nil {
		opts.Resource.Workers = b.limits.workerInstances
		opts.Resource.RAM = b.limits.workerRAM
		opts.Resource.Storage = b.limits.workerStorage
	} else {
		r := runtime.BuilderCapacity()
		opts.Resource.Workers = r.Workers
		opts.Resource.RAM = r.RAM
		opts.Resource.CPU = r.CPU
		opts.Resource.Storage = r.Storage
	}

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

	err := envs.Get().GetClient().V1().Builder(envs.Get().GetHostname()).Connect(b.ctx, opts)
	if err != nil {
		log.Errorf("%s:start:> send event connect builder err: %v", logWorkerPrefix, err)
		return err
	}

	return nil
}

// Set builder status
func (b Builder) status() error {

	opts := v1.Request().Builder().StatusUpdateOptions()

	for {

		select {
		case <-b.ctx.Done():
			log.Debugf("%s:status:> stop send status", logWorkerPrefix)
			return nil
		default:

			status := runtime.BuilderStatus()

			opts.Capacity.Workers = status.Capacity.Workers
			opts.Capacity.RAM = status.Capacity.RAM - b.reserveMemory
			opts.Capacity.Storage = status.Capacity.Storage
			opts.Capacity.CPU = status.Capacity.CPU

			if b.limits != nil {
				opts.Allocated.Workers = b.limits.workerInstances
				opts.Allocated.RAM = int64(b.limits.workerInstances) * b.limits.workerRAM
				opts.Allocated.Storage = int64(b.limits.workerInstances) * b.limits.workerStorage
				opts.Allocated.CPU = int64(b.limits.workerInstances) * b.limits.workerCPU
			} else {
				opts.Allocated.Workers = status.Capacity.Workers
				opts.Allocated.RAM = status.Capacity.RAM
				opts.Allocated.Storage = status.Capacity.Storage
				opts.Allocated.CPU = status.Capacity.CPU
			}

			u := b.getUsage()
			opts.Usage.Workers = u.Workers
			opts.Usage.RAM = u.RAM
			opts.Usage.Storage = u.Storage
			opts.Usage.CPU = u.CPU

			<-time.After(defaultSendStatusDelay)

			err := envs.Get().GetClient().V1().Builder(envs.Get().GetHostname()).SetStatus(b.ctx, opts)
			if err != nil {
				log.Errorf("%s:start:> send event status builder err: %v", logWorkerPrefix, err)
				continue
			}
		}

	}
}

func (b Builder) getUsage() *types.BuilderResources {
	br := new(types.BuilderResources)
	br.Workers = len(b.workers)

	br.RAM = int64(br.Workers) * types.DEFAULT_MIN_WORKER_MEMORY
	br.CPU = int64(br.Workers) * types.DEFAULT_MIN_WORKER_CPU
	br.Storage = 0 //uint64(br.Workers) * types.DEFAULT_MIN_WORKER_Storage

	if b.limits != nil {

		if b.limits.workerRAM != 0 {
			br.RAM = int64(br.Workers) * b.limits.workerRAM
		}

		if b.limits.workerCPU != 0 {
			br.CPU = int64(br.Workers) * b.limits.workerCPU
		}

		if b.limits.workerStorage != 0 {
			br.Storage = 0 //uint64(br.Workers) * uint64(b.limits.WorkerStorage)
		}

	}

	return br
}
