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
	"os"
	"time"

	"github.com/lastbackend/registry/pkg/runtime/cri"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/util/stream"
	"github.com/spf13/viper"

	lbt "github.com/lastbackend/registry/pkg/distribution/types"
)

// The main entity that is responsible for
// the environment and existence of workers
type Builder struct {
	ctx        context.Context
	cancel     context.CancelFunc
	dockerHost string
	id         string
	cri        cri.CRI
	limit      int
	extraHosts []string
	logdir     string

	tasks   chan *task
	workers chan chan *task
	respawn chan *worker
	done    chan bool

	activeTask map[string]*task
}

// Preparing the builder environment for workers
func New(cri cri.CRI, id, dockerHost string, extraHosts []string, limit int, logdir string) *Builder {

	log.Info("Build: New: create builder")

	var (
		b   = new(Builder)
		ctx = context.Background()
	)

	b.ctx, b.cancel = context.WithCancel(ctx)
	b.limit = limit
	b.logdir = logdir
	b.id = id
	b.dockerHost = dockerHost
	b.extraHosts = extraHosts
	b.cri = cri

	b.tasks = make(chan *task, limit)
	b.workers = make(chan chan *task, limit)
	b.respawn = make(chan *worker, limit)
	b.done = make(chan bool)

	b.activeTask = make(map[string]*task, 0)

	return b
}

// Initializing the builder and preparing the necessary resources for correct operation
func (b *Builder) Start() error {

	log.Info("Build: Start: start builder")

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

		log.Warnf("Build: Start: image %s not found", name)

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
			log.Errorf("Build: Start: pull image err: %s", err)
			return err
		}
		// TODO handle output in more beautiful way
		if _, err := io.Copy(os.Stdout, req); err != nil {
			log.Errorf("Build: Start: copy stream to stdout err: %s", err)
		}

		if err := req.Close(); err != nil {
			log.Errorf("Build: Start: close stream err: %s", err)
			return err
		}
	}

	go b.spawn(b.ctx, b.limit)

	go func() {
		<-b.ctx.Done()
		close(b.tasks)
	}()

	go b.dispatch()

	// TODO: send event online
	//events.BuildOnlineEventRequest(envs.Get().GetRPC(), b.id)

	return nil
}

// Add new task for build to queue
func (b *Builder) NewTask(ctx context.Context, job *types.BuildJob) error {

	log.Infof("Build: NewBuild: create new build for job: %s", job.ID)

	ctx = context.WithValue(ctx, "logdir", b.logdir)
	ctx = context.WithValue(ctx, "blob-storage", types.AzureBlobStorage{
		AccountName: viper.GetString("storage.azure.account"),
		AccountKey:  viper.GetString("storage.azure.key"),
	})

	task, err := newTask(ctx, b.id, b.dockerHost, b.extraHosts, b.cri, job)
	if err != nil {
		log.Errorf("Build: NewBuild: create build err: %s", err)
		return err
	}

	b.activeTask[task.id] = task

	b.tasks <- task

	return nil
}

// Proxy logs stream from task
func (b *Builder) LogsTask(ctx context.Context, id, endpoint string) error {

	log.Infof("Build: NewBuild: get task logs for stream: %s", id)

	if task, ok := b.activeTask[id]; ok {
		go task.logger.Stream(stream.NewStream().AddSocketBackend(endpoint))
	}

	return nil
}

// Interrupting the build process
func (b *Builder) CancelTask(ctx context.Context, id string) error {

	log.Infof("Build: CancelBuild: cancel build: %s", id)

	if _, ok := b.activeTask[id]; !ok {
		err := fmt.Errorf("task %s is not active or was already finished", id)
		log.Warnf("Build: NewBuild:cancel build task err: %s", err)
		return err
	}

	b.activeTask[id].stop()
	delete(b.activeTask, id)

	return nil
}

// Dispatching of incoming tasks between workers
func (b *Builder) dispatch() {

	log.Info("Build: Dispatch: run dispatching tasks to workers")

	for task := range b.tasks {
		worker := <-b.workers
		worker <- task
	}
}

// Spawn - finished or failed workers will be restarted until context is closed
// If worker failed - then wait some time until respawn
func (b *Builder) spawn(ctx context.Context, workers int) {

	log.Info("Build: Spawn: run spawn workers")

	run := func(w *worker) {
		if err := w.Run(b.workers); err != nil {
			log.Errorf("Build: Spawn: start worker for provision err: %s", err)
			// TODO separate task fails and init fails
			select {
			// error delay
			case <-time.After(5 * time.Second):
			case <-ctx.Done():
			}
		}

		b.respawn <- w
	}

	log.Debugf("Build: Spawn: create %d workers", workers)

	for i := 0; i < workers; i++ {
		w := newWorker(ctx, b.cri)
		go run(w)
	}

	for {
		select {
		case <-ctx.Done():
			log.Debug("Build: Spawn: stop respawn")
			return
		case w := <-b.respawn:
			log.Debug("Build: Spawn: restarted worker")
			go run(w)
		}
	}
}

// Await checks that all workers finished
// then closes channel for registering new workers and respawns
func (b *Builder) await() {

	log.Info("Build: Await: run graceful shutdown for workers")

	alive := cap(b.respawn)

	log.Infof("Build: Await: graceful shutdown await [%d/%d]", len(b.respawn), alive)

	if len(b.respawn) == 0 {
		close(b.respawn)
		return
	}

	for range b.respawn {
		alive--
		if alive == 0 {
			close(b.workers)
			close(b.respawn)
		}
	}
}

// Shutdown builder process
func (b *Builder) Shutdown() {

	log.Info("Build: Shutdown: Shutdown builder")

	// TODO: send event offline
	//events.BuildOfflineEventRequest(envs.Get().GetRPC(), b.id)

	b.await()

	b.done <- true
}

// Notification that the builder has completed its work
func (b *Builder) Done() <-chan bool {
	return b.done
}
