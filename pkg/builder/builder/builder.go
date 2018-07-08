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
	"github.com/lastbackend/registry/pkg/builder/envs"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/runtime/cri"
	"github.com/spf13/viper"
	"io"
	"os"
	"sync"

	lbt "github.com/lastbackend/registry/pkg/distribution/types"
	"time"
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
	id       string
	hostname string
	cri      cri.CRI
	limit    int

	respawn chan *worker
	done    chan bool

	workers map[string]*worker
}

// Preparing the builder environment for workers
func New(cri cri.CRI, id, dockerHost string, extraHosts []string, limit int, logdir string) *Builder {

	log.Infof("%s:new:> create builder", logBuilderPrefix)

	var (
		b   = new(Builder)
		ctx = context.Background()
	)

	b.limit = limit
	b.id = id
	b.cri = cri

	b.respawn = make(chan *worker, limit)
	b.done = make(chan bool)

	b.workers = make(map[string]*worker, 0)

	ctx = context.WithValue(ctx, "builder", id)
	ctx = context.WithValue(ctx, "dockerHost", dockerHost)
	ctx = context.WithValue(ctx, "extraHosts", extraHosts)
	ctx = context.WithValue(ctx, "logdir", logdir)
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

	go b.spawn(b.ctx, b.limit)

	go func() {
		<-b.ctx.Done()
	}()

	if err := envs.Get().GetClient().V1().Builder().Connect(b.ctx, envs.Get().GetHostname()); err != nil {
		log.Errorf("%s:start:> send event connect builder err: %v", logWorkerPrefix, err)
		return err
	}

	return nil
}

// Proxy logs stream from task
func (b *Builder) BuildLogs(ctx context.Context, id, endpoint string) error {
	log.Infof("%s:new_build:> get task logs for stream: %s", logWorkerPrefix, id)
	w, ok := b.workers[id]
	if ok {
		go w.Logs()
	}
	return nil
}

// Interrupting the build process
func (b *Builder) BuildCancel(ctx context.Context, id string) error {

	log.Infof("%s:cancel:> cancel build: %s", logWorkerPrefix, id)

	worker, ok := b.workers[id]
	if ok {
		worker.destroy()
		b.Lock()
		delete(b.workers, id)
		b.Unlock()

	}
	return nil
}

// Spawn - finished or failed workers will be restarted until context is closed
// If worker failed - then wait some time until respawn
func (b *Builder) spawn(ctx context.Context, workers int) {

	log.Info("%s:spawn:> run spawn workers")

	run := func(w *worker) {

		if err := w.Run(); err != nil {
			log.Errorf("%s:spawn:> start worker for provision err: %v", logWorkerPrefix, err)
			select {
			// error delay
			case <-time.After(5 * time.Second):
			case <-ctx.Done():
			}
		}

		// respawn delay
		<-time.After(5 * time.Second)

		b.respawn <- w
	}

	log.Debugf("%s:spawn:> create %d workers", logWorkerPrefix, workers)

	for i := 0; i < workers; i++ {
		w := newWorker(ctx, b.cri)
		b.Lock()
		b.workers[w.id] = w
		b.Unlock()
		go run(w)
	}

	for {
		select {
		case <-ctx.Done():
			log.Debugf("%s:spawn:> stop respawn", logWorkerPrefix)
			return
		case w := <-b.respawn:
			log.Debugf("%s:spawn:> restarted worker", logWorkerPrefix)
			go run(w)
		}
	}
}

// Await checks that all workers finished
// then closes channel for registering new workers and respawns
func (b *Builder) await() {

	log.Infof("%s:await:> run graceful shutdown for workers", logWorkerPrefix)

	alive := cap(b.respawn)

	log.Infof("%s:await:> graceful shutdown await [%d/%d]", logWorkerPrefix, len(b.respawn), alive)

	if len(b.respawn) == 0 {
		close(b.respawn)
		return
	}

	for range b.respawn {
		alive--
		if alive == 0 {
			close(b.respawn)
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
			log.Errorf("%s:configure:> copy stream to stdout err: %v", logWorkerPrefix, err)
		}

		if err := req.Close(); err != nil {
			log.Errorf("%s:configure:> close stream err: %v", logWorkerPrefix, err)
			return err
		}
	}

	return nil
}

// Shutdown builder process
func (b *Builder) Shutdown() {

	log.Infof("%s:shutdown:> shutdown builder process", logWorkerPrefix)

	if err := envs.Get().GetClient().V1().Builder().Disconnect(b.ctx, envs.Get().GetHostname()); err != nil {
		log.Errorf("%s:shutdown:> send event offline err: %v", logWorkerPrefix, err)
	}

	b.await()

	b.done <- true
}

// Notification that the builder has completed its work
func (b *Builder) Done() <-chan bool {
	return b.done
}
