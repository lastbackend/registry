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
	"sync"
	"time"

	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/runtime/cri"
	"github.com/lastbackend/registry/pkg/util/generator"
	"github.com/lastbackend/registry/pkg/util/stream"
	"github.com/lastbackend/registry/pkg/builder/envs"
	"github.com/lastbackend/registry/pkg/api/types/v1/request"
	)

const (
	logWorkerPrefix = "builder:worker"
)

type worker struct {
	ctx    context.Context
	cancel context.CancelFunc
	lock   sync.RWMutex

	id   string
	task *task
	cri  cri.CRI
}

// Create and configure new worker
func newWorker(ctx context.Context, cri cri.CRI) *worker {
	var w = new(worker)
	w.ctx, w.cancel = context.WithCancel(ctx)
	w.id = generator.GetUUIDV4()
	w.cri = cri
	return w
}

// Start worker process
func (w *worker) Run() error {

	log.Infof("%s:run:> run worker with id %#v", logWorkerPrefix, w.id)

	task, err := NewTask(w.ctx, w.cri)
	if err != nil {
		log.Errorf("%s:run:> create new task err:  %v", logWorkerPrefix, w.id, task.id, err)
		return err
	}

	opts := new(request.BuilderCreateManifestOptions)
	opts.TaskID = task.id

	manifest, err := envs.Get().GetClient().V1().Builder().GetManifest(w.ctx, envs.Get().GetHostname(), opts)
	if err != nil {
		log.Errorf("%s:spawn:> get manifest err: %v", logWorkerPrefix, err)
		return err
	}

	if manifest == nil {
		return nil
	}

	m := new(types.BuildManifest)

	m.Source.Url = manifest.Source.Url
	m.Source.Branch = manifest.Source.Branch

	m.Image.Host = manifest.Image.Host
	m.Image.Name = manifest.Image.Name
	m.Image.Owner = manifest.Image.Owner
	m.Image.Tag = manifest.Image.Tag
	m.Image.Auth = manifest.Image.Auth

	m.Config.Dockerfile = manifest.Config.Dockerfile
	m.Config.Context = manifest.Config.Context
	m.Config.Workdir = manifest.Config.Workdir
	m.Config.EnvVars = manifest.Config.EnvVars
	m.Config.Command = manifest.Config.Command

	w.task = task

	startTime := time.Now()

	log.Infof("%s:run:> worker %s start task %s", logWorkerPrefix, w.id, task.id)

	if err := task.Start(m); err != nil {
		log.Errorf("%s:run:> worker %s task %s start err: %s", w.id, task.id, err)
		return err
	}

	log.Debugf("%s:run:> worker %s task %s finish %v", logWorkerPrefix, w.id, task.id, time.Since(startTime))

	return nil
}

func (w *worker) Logs() error {
	endpoint := w.ctx.Value("endpoint").(string)
	return w.task.logger.Stream(stream.NewStream().AddSocketBackend(endpoint))
}

// Destroy worker process.
// When the process is complete, you must wait for all tasks to complete
func (w *worker) destroy() error {
	log.Infof("%s:destroy:> destroy worker process %s", logWorkerPrefix, w.id)
	w.task.stop()
	w.cancel()
	return nil
}
