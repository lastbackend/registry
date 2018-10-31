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
	"github.com/lastbackend/genesis/pkg/util/url"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/lastbackend/lastbackend/pkg/log"
	"github.com/lastbackend/lastbackend/pkg/runtime/cri"
	"github.com/lastbackend/registry/pkg/api/types/v1/request"
	"github.com/lastbackend/registry/pkg/builder/envs"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/util/cleaner"
	"github.com/lastbackend/registry/pkg/util/validator"

	lbt "github.com/lastbackend/lastbackend/pkg/distribution/types"
	lbcii "github.com/lastbackend/lastbackend/pkg/runtime/cii/cii"
)

const (
	logWorkerPrefix = "builder:worker"
)

const (
	errorBuildFailed  = "build process failed"
	errorUploadFailed = "push process failed"
)

type worker struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	lock       sync.RWMutex

	pid      string
	endpoint string
	dcid     string
	gcid     string
	step     string
	logDir   string

	task *types.Task

	stdout bool

	cri cri.CRI
}

type workerOpts struct {
	Stdout bool
}

// Create and configure new worker
func newWorker(ctx context.Context, id string, cri cri.CRI) *worker {
	log.Infof("%s:new:> create new worker", logWorkerPrefix)
	var w = new(worker)

	pid := id
	w.ctx, w.cancelFunc = context.WithCancel(ctx)
	w.pid = pid
	w.logDir = os.TempDir()
	w.cri = cri

	return w
}

// Start worker process
func (w *worker) run(t *types.Task, wo *workerOpts) error {

	log.Infof("%s:run:> run worker with pid %#v", logWorkerPrefix, w.pid)

	if wo == nil {
		wo = new(workerOpts)
		wo.Stdout = true
	}

	w.task = t
	w.stdout = wo.Stdout

	startTime := time.Now()

	log.Infof("%s:run:> worker %s start", logWorkerPrefix, w.pid)

	defer w.cleanup()

	// Configure environment for build and start docker dind
	err := w.configure()
	if err != nil && err != context.Canceled {
		log.Errorf("%s:start:> configure t %s err:  %s", logWorkerPrefix, w.pid, err)
		return err
	}

	// Running build process
	err = w.build()
	if err != nil && err != context.Canceled {
		log.Errorf("%s:start:> build t %s err:  %s", logWorkerPrefix, w.pid, err)
		// Upload logs to blob storage
		w.uploadLogs()
		return err
	}

	// Upload logs to blob storage
	w.uploadLogs()

	// Pushing docker image to docker registry
	err = w.push()
	if err != nil && err != context.Canceled {
		log.Errorf("%s:start:> push %s err:  %s", logWorkerPrefix, w.pid, err)
		return err
	}

	err = w.finish()
	if err != nil && err != context.Canceled {
		log.Errorf("%s:start:> finish %s err:  %s", logWorkerPrefix, w.pid, err)
		return err
	}

	log.Debugf("%s:run:> worker %s finish %v", logWorkerPrefix, w.pid, time.Since(startTime))

	return nil
}

// Configure environment for build and start docker dind
func (w *worker) configure() error {

	var (
		extraHosts = make([]string, 0)
		//dockerHost = defaultDockerHost
		rootCerts = make([]string, 0)
	)

	if w.ctx.Value("extraHosts") != nil {
		extraHosts = w.ctx.Value("extraHosts").([]string)
	}

	if w.ctx.Value("rootCerts") != nil {
		rootCerts = w.ctx.Value("rootCerts").([]string)
	}

	spec := lbt.ContainerManifest{
		Image:      "docker:dind",
		AutoRemove: true,
		ExtraHosts: extraHosts,
		Exec: lbt.SpecTemplateContainerExec{
			Command: []string{"--storage-driver=overlay"},
		},
		Security: lbt.SpecTemplateContainerSecurity{
			Privileged: true,
		},
		Labels:          map[string]string{lbt.ContainerTypeLBR: w.pid},
		PublishAllPorts: true,
	}

	// manual addition the CA certificates certificate to dind
	if len(rootCerts) != 0 {
		for _, cert := range rootCerts {
			items := strings.Split(cert, ":")

			if len(items) == 0 || len(items) < 2 {
				continue
			}

			hostPath := items[1]
			mode := "ro" // read only

			if len(items) > 2 {
				mode = items[2]
			}

			containerPath := fmt.Sprintf("/etc/docker/certs.d/%s/ca.crt", items[0])

			spec.Binds = append(spec.Binds, fmt.Sprintf("%s:%s:%s", hostPath, containerPath, mode))
		}
	}

	dcid, err := w.cri.Create(w.ctx, &spec)
	if err != nil {
		log.Errorf("%s:start:> create container with docker:dind err: %v", logWorkerPrefix, err)
		return err
	}

	if err := w.cri.Start(w.ctx, dcid); err != nil {
		log.Errorf("%s:start:> start container with docker:dind err: %v", logWorkerPrefix, err)
		return err
	}

	inspect, err := w.cri.Inspect(w.ctx, dcid)
	if err != nil {
		log.Errorf("%s:start:> Inspect docker:dind container err: %v", logWorkerPrefix, err)
		return err
	}
	if inspect == nil {
		err := fmt.Errorf("docker:dind does not exists")
		log.Errorf("%s:start:> container inspect err: %v", logWorkerPrefix, err)
		return err
	}
	if inspect.ExitCode != 0 {
		err := fmt.Errorf("docker:dind exit with status code %d", inspect.ExitCode)
		log.Errorf("%s:start:> container exit with err: %v", logWorkerPrefix, err)
		return err
	}

	w.endpoint = fmt.Sprintf("tcp://%s:2375", inspect.Network.IPAddress)
	w.dcid = dcid

	return nil
}

// Running build process
func (w *worker) build() error {

	w.step = types.BuildStepBuild
	w.sendEvent(event{step: types.BuildStepBuild})

	mspec := w.task.Spec

	var (
		image      = fmt.Sprintf("%s/%s/%s:%s", mspec.Image.Host, mspec.Image.Owner, mspec.Image.Name, mspec.Image.Tag)
		dockerfile = mspec.Config.Dockerfile
		gituri     = fmt.Sprintf("%s.git#%s", mspec.Source.Url, strings.ToLower(mspec.Source.Branch))
	)

	if len(mspec.Config.Context) != 0 && mspec.Config.Context != "/" {
		gituri = fmt.Sprintf("%s:%s", gituri, mspec.Config.Context)
	}

	log.Infof("%s:build:> running build image %s process for task %s", logWorkerPrefix, image, w.pid)

	if validator.IsValueInList(dockerfile, []string{"", " ", "/", "./", "../", "DockerFile", "/DockerFile"}) {
		dockerfile = "./DockerFile"
	}

	// TODO: change this logic to docker client [cli.ImageBuild]
	spec := &lbt.ContainerManifest{
		Image:  "docker:git",
		Labels: map[string]string{lbt.ContainerTypeLBR: w.pid},
		Envs:   []string{fmt.Sprintf("%s=%s", "DOCKER_HOST", w.endpoint)},
		Exec: lbt.SpecTemplateContainerExec{
			Command: []string{"build", "-f", dockerfile, "-t", image, gituri},
		},
	}

	cid, err := w.cri.Create(w.ctx, spec)
	switch err {
	case nil:
	case context.Canceled:
		log.Debugf("%s:build:> process canceled", logWorkerPrefix)
		w.sendEvent(event{step: types.BuildStepBuild, canceled: true})
		return nil
	default:
		log.Errorf("%s:build:> create container err: %v", logWorkerPrefix, err)
		w.sendEvent(event{step: types.BuildStepBuild, message: errorBuildFailed, error: true})
		return err
	}

	w.gcid = cid

	err = w.cri.Start(w.ctx, cid)
	switch err {
	case nil:
	case context.Canceled:
		log.Debugf("%s:build:> process canceled", logWorkerPrefix)
		w.sendEvent(event{step: types.BuildStepBuild, canceled: true})
		return nil
	default:
		log.Errorf("%s:build:> start container err: %v", logWorkerPrefix, err)
		w.sendEvent(event{step: types.BuildStepBuild, message: errorBuildFailed, error: true})
		return err
	}

	if w.stdout {
		go w.logging(os.Stdout)
	}

	ch := make(chan *lbt.Container)
	go w.cri.Subscribe(w.ctx, ch)

	for {
		select {
		case <-w.ctx.Done():
			log.Debugf("%s:build:> process canceled", logWorkerPrefix)
			w.sendEvent(event{step: types.BuildStepBuild, canceled: true})
			return nil
		case c := <-ch:

			if id, ok := c.Labels[lbt.ContainerTypeLBR]; !ok || id != w.pid {
				continue
			}

			if c.ExitCode != 0 {
				err := fmt.Errorf("container exited with %d code", c.ExitCode)
				log.Errorf("%s:build:> container exit with err %v", logWorkerPrefix, err)
				w.sendEvent(event{step: types.BuildStepBuild, message: errorBuildFailed, error: true})
				return err
			}

			if c.Status != "exited" {
				continue
			}

			return nil
		}
	}

	return nil
}

// Running push process to registry
func (w *worker) push() error {

	w.step = types.BuildStepUpload
	w.sendEvent(event{step: types.BuildStepUpload})

	mspec := w.task.Spec

	var (
		registry = mspec.Image.Host
		name     = fmt.Sprintf("%s/%s/%s", registry, mspec.Image.Owner, mspec.Image.Name)
		tag      = mspec.Image.Tag
		auth     = mspec.Image.Auth
	)

	log.Infof("%s:push:> running push image %s process for task %s to registry %s", logWorkerPrefix, name, w.pid, registry)

	ciiDriver := viper.GetString("runtime.cii.type")
	opts := viper.GetStringMap(fmt.Sprintf("runtime.%s", ciiDriver))
	opts["host"] = w.endpoint

	cii, err := lbcii.New(ciiDriver, opts)
	switch err {
	case nil:
	case context.Canceled:
		log.Debugf("%s:build:> process canceled", logWorkerPrefix)
		w.sendEvent(event{step: types.BuildStepUpload, canceled: true})
		return nil
	default:
		log.Errorf("%s:push:> create docker extra_hosts client %v", logWorkerPrefix, err)
		w.sendEvent(event{step: types.BuildStepUpload, message: errorUploadFailed, error: true})
		return err
	}

	var writer io.Writer
	if w.stdout {
		writer = os.Stdout
	}

	img, err := cii.Push(w.ctx, &lbt.ImageManifest{Name: name, Tag: tag, Auth: auth}, writer)
	switch err {
	case nil:
	case context.Canceled:
		log.Debugf("%s:build:> process canceled", logWorkerPrefix)
		w.sendEvent(event{step: types.BuildStepUpload, canceled: true})
		return nil
	default:
		log.Errorf("%s:push:> running push process err: %v", logWorkerPrefix, err)
		w.sendEvent(event{step: types.BuildStepUpload, message: errorUploadFailed, error: true})
		return err
	}

	w.sendInfo(img)

	return nil
}

func (w *worker) finish() error {
	log.Infof("%s:finish:> handler after task %s completion", logWorkerPrefix, w.pid)
	w.sendEvent(event{step: types.BuildStepDone})
	return nil
}

func (w *worker) uploadLogs() error {

	if !strings.HasSuffix(w.logDir, string(os.PathSeparator)) {
		w.logDir += string(os.PathSeparator)
	}

	if err := os.MkdirAll(w.logDir, os.ModePerm); err != nil {
		log.Errorf("%s:upload_logs:> create directories [%s] err: %v", logWorkerPrefix, w.logDir, err)
		return err
	}

	filePath := fmt.Sprintf("%s%s", w.logDir, w.task.Meta.ID)

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		log.Errorf("%s:upload_logs:> open file [%s] err: %v", logWorkerPrefix, filePath, err)
		return err
	}
	// close fi on exit and check for its returned error
	defer func() {
		if err := file.Close(); err != nil {
			log.Errorf("%s:upload_logs:> close file [%s] err: %v", logWorkerPrefix, filePath, err)
		}
		err := os.Remove(filePath)
		if err != nil {
			log.Errorf("%s:upload_logs:> remove file [%s] err: %v", logWorkerPrefix, filePath, err)
		}
	}()

	w.logging(file)

	if envs.Get().GetBlobStorage() != nil {
		s := url.Decode(w.task.Spec.Source.Url)

		path := fmt.Sprintf("/%s/%s/%s/build/%s", s.Hub, s.Owner, s.Name, w.task.Meta.ID)
		err := envs.Get().GetBlobStorage().WriteFromFile(path, filePath)
		if err != nil {
			log.Errorf("%s:upload_logs:> write container logs to blob err: %v", logWorkerPrefix, err)
		}
	}

	return nil
}

func (w *worker) cancel() {
	w.cancelFunc()
}

func (w *worker) cleanup() {

	if err := w.cri.Remove(context.Background(), w.dcid, true, true); err != nil {
		log.Errorf("%s:cleanup:> remove %s container dind  err: %v", logWorkerPrefix, w.dcid, err)
	}

	if err := w.cri.Remove(context.Background(), w.gcid, true, true); err != nil {
		log.Errorf("%s:cleanup:> remove %s container git err: %v", logWorkerPrefix, w.gcid, err)
	}

}

// Get logs for build process
func (w *worker) logs(writer io.Writer) error {
	return w.logging(writer)
}

func (w *worker) logging(writer io.Writer) error {

	req, err := w.cri.Logs(w.ctx, w.gcid, true, true, true)
	switch err {
	case nil:
	case context.Canceled:
		return nil
	default:
		err := fmt.Errorf("running logs stream: %s", err)
		log.Errorf("%s:logging:> logs container err: %v", logWorkerPrefix, err)
		return err
	}
	defer func() {
		if req != nil {
			if err := req.Close(); err != nil {
				log.Errorf("%s:logging:> close log stream err: %s", err)
				return
			}
		}
	}()

	var (
		buffer = make([]byte, 2048)
	)

	for {
		select {
		case <-w.ctx.Done():
			return nil
		default:

			readBytes, err := cleaner.NewReader(req).Read(buffer)
			switch err {
			case nil:
			case context.Canceled:
				return nil
			default:
				if err != io.EOF {
					log.Errorf("%s:logging:> read data from stream err: %v", logWorkerPrefix, err)
					return err
				}
			}

			if readBytes == 0 {
				return nil
			}

			_, err = writer.Write(buffer[0:readBytes])
			if err != nil {
				if err == context.Canceled {
					return nil
				}
				log.Errorf("%s:logging:> write stream data err: %v", logWorkerPrefix, err)
				continue
			}

			if f, ok := writer.(http.Flusher); ok {
				f.Flush()
			}

			for i := 0; i < readBytes; i++ {
				buffer[i] = 0
			}
		}
	}

	return nil
}

type event struct {
	step     string
	message  string
	error    bool
	canceled bool
}

// Send status build event to controller
func (w *worker) sendEvent(event event) {

	log.Debugf("%s:send_event> send task status event %s", logWorkerPrefix, w.pid)

	mspec := w.task.Spec
	mmeta := w.task.Meta

	e := new(request.BuildUpdateStatusOptions)
	e.Step = event.step
	e.Message = event.message
	e.Error = event.error
	e.Canceled = w.ctx.Err() == context.Canceled

	envs.Get().GetClient().V1().
		Image(mspec.Image.Owner, mspec.Image.Name).
		Build(mmeta.ID).
		SetStatus(w.ctx, e)
}

// Send status build event to controller
func (w *worker) sendInfo(info *lbt.Image) {
	log.Debugf("%s:send_info> send task status event %s", logWorkerPrefix, w.pid)

	mspec := w.task.Spec
	mmeta := w.task.Meta

	e := new(request.BuildSetImageInfoOptions)
	e.Size = info.Status.Size
	e.Hash = info.Meta.ID
	e.VirtualSize = info.Status.VirtualSize

	envs.Get().GetClient().V1().
		Image(mspec.Image.Owner, mspec.Image.Name).
		Build(mmeta.ID).
		SetImageInfo(w.ctx, e)
}
