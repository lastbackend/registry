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

	"github.com/lastbackend/registry/pkg/runtime/cri"
	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/util/generator"
)

type worker struct {
	ctx  context.Context
	lock sync.RWMutex

	id  string
	cri cri.CRI

	tasks chan *task
}

// Create and configure new worker
func newWorker(ctx context.Context, cri cri.CRI) *worker {
	var w = new(worker)
	w.id = generator.GetUUIDV4()
	w.cri = cri
	w.ctx = ctx
	w.tasks = make(chan *task)
	return w
}

// Start worker process
func (w *worker) Run(queue chan chan *task) error {

	log.Infof("Worker: Run: run worker with id %s", w.id)

	// Allow to put tasks in queue of this worker.
	queue <- w.tasks

	select {
	case <-w.ctx.Done():
		log.Debugf("Worker: Run: worker stop: %s", w.id)
		close(w.tasks)
		return nil

	case task := <-w.tasks:

		log.Debugf("Worker: Run: worker %s task %s start", w.id, task.id)

		startTime := time.Now()

		if err := w.handle(task); err != nil {
			log.Errorf("Worker: Run: worker %s task %s start err: %s", w.id, task.id, err)
			return err
		}

		log.Debugf("Worker: Run: worker %s task %s finish %v", w.id, task.id, time.Since(startTime))

		return nil
	}
}

// The execution of a new task by the worker step by step
func (w *worker) handle(task *task) error {

	log.Infof("Worker: Handle: worker %s start task %s", w.id, task.job.ID)

	defer func() {
		if err := task.finish(); err != nil && err != context.Canceled {
			log.Errorf("Worker: Handle: worker %s finish task %s err:  %s", w.id, task.id, err)
		}
	}()

	err := task.build()
	if err == nil && task.Canceled() {
		return nil
	}
	if err != nil && err != context.Canceled {
		log.Errorf("Worker: Handle: worker %s build task %s err:  %s", w.id, task.id, err)
		return err
	}

	err = task.push()
	if err == nil && task.Canceled() {
		return nil
	}
	if err != nil && err != context.Canceled {
		log.Errorf("Worker: Handle: worker %s push task %s err:  %s", w.id, task.id, err)
		return err
	}

	return nil
}

// Destroy worker process.
// When the process is complete, you must wait for all tasks to complete
func (w *worker) destroy() error {
	log.Infof("Worker: Destroy: destroy worker process %s", w.id)
	return nil
}
