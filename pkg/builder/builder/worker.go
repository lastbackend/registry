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
	"os"
	"strings"
	"sync"
	"time"

	"github.com/lastbackend/registry/pkg/api/types/v1/request"
	"github.com/lastbackend/registry/pkg/builder/envs"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/runtime/cri"
	"github.com/lastbackend/registry/pkg/runtime/cri/docker"
	"github.com/lastbackend/registry/pkg/util/validator"

	lbt "github.com/lastbackend/registry/pkg/distribution/types"
	"net/http"
	"github.com/lastbackend/registry/pkg/util/cleaner"
)

const (
	logWorkerPrefix = "builder:worker"
)

const (
	errorBuildFailed  = "build process failed"
	errorUploadFailed = "push process failed"
)

const (
	defaultDockerHost = "172.17.0.1"
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
	logPath  string

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
	w.logPath = os.TempDir()
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
		return err
	}

	// Upload logs to blob storage
	go w.upload()

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
		dockerHost = defaultDockerHost
		rootCerts  = make([]string, 0)
	)

	if w.ctx.Value("extraHosts") != nil {
		extraHosts = w.ctx.Value("extraHosts").([]string)
	}

	if w.ctx.Value("dockerHost") != nil {
		dockerHost = w.ctx.Value("dockerHost").(string)
	}

	if w.ctx.Value("rootCerts") != nil {
		rootCerts = w.ctx.Value("rootCerts").([]string)
	}

	spec := lbt.SpecTemplateContainer{
		Image: lbt.SpecTemplateContainerImage{
			Name: "docker:dind",
		},
		AutoRemove: true,
		ExtraHosts: extraHosts,
		Exec: lbt.SpecTemplateContainerExec{
			Command: []string{"--storage-driver=overlay"},
		},
		Security: lbt.SpecTemplateContainerSecurity{
			Privileged: true,
		},
		Labels:          map[string]string{"LBR": w.pid},
		PublishAllPorts: true,
	}

	// manual addition the CA certificates certificate to dind
	if len(rootCerts) != 0 {
		for _, cert := range rootCerts {
			items := strings.Split(cert, ":")

			if len(items) == 0 || len(items) < 2 {
				continue
			}

			host := items[0]
			path := items[1]
			mode := "ro" // read only

			if len(items) > 2 {
				mode = items[2]
			}

			spec.Volumes = append(spec.Volumes, types.SpecTemplateContainerVolume{
				HostPath:      path,
				ContainerPath: fmt.Sprintf("/etc/docker/certs.d/%s/ca.crt", host),
				Mode:          mode,
			})
		}
	}

	dcid, err := w.cri.ContainerCreate(w.ctx, &spec)
	if err != nil {
		log.Errorf("%s:start:> create container with docker:dind err: %v", logWorkerPrefix, err)
		return err
	}

	if err := w.cri.ContainerStart(w.ctx, dcid); err != nil {
		log.Errorf("%s:start:> start container with docker:dind err: %v", logWorkerPrefix, err)
		return err
	}

	inspect, err := w.cri.ContainerInspect(w.ctx, dcid)
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

	var port = ""
	for p, binds := range inspect.Network.Ports {
		match := strings.Split(p, "/")

		if match[0] != "2375" {
			continue
		}

		if len(binds) == 0 {
			err := fmt.Errorf("there are no ports available")
			log.Errorf("%s:start:> cannot receive docker daemon connection port: %v", logWorkerPrefix, err)
			return err
		}

		for _, bind := range binds {
			if port != "" {
				break
			}
			if bind.HostPort != "" {
				port = bind.HostPort
			}
		}
	}

	w.endpoint = fmt.Sprintf("tcp://%s:%s", dockerHost, port)
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
	spec := &lbt.SpecTemplateContainer{
		Image: lbt.SpecTemplateContainerImage{
			Name: "docker:git",
		},
		Labels:  map[string]string{"LBR": w.pid},
		EnvVars: []lbt.SpecTemplateContainerEnv{{Name: "DOCKER_HOST", Value: w.endpoint}},
		Exec: lbt.SpecTemplateContainerExec{
			Command: []string{"build", "-f", dockerfile, "-t", image, gituri},
		},
	}

	cid, err := w.cri.ContainerCreate(w.ctx, spec)
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

	err = w.cri.ContainerStart(w.ctx, cid)
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

	ch := make(chan *types.Container)
	go w.cri.Subscribe(w.ctx, ch, &types.ContainerEventFilter{Image: "docker:git"})

	for {
		select {
		case <-w.ctx.Done():
			log.Debugf("%s:build:> process canceled", logWorkerPrefix)
			w.sendEvent(event{step: types.BuildStepBuild, canceled: true})
			return nil
		case c := <-ch:

			if c.ExitCode != 0 {
				err := fmt.Errorf("container exited with %d code", c.ExitCode)
				log.Errorf("%s:build:> container exit with err %v", logWorkerPrefix, err)
				w.sendEvent(event{step: types.BuildStepBuild, message: errorBuildFailed, error: true})
				return err
			}

			if c.Label != w.pid {
				continue
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
		registry  = mspec.Image.Host
		image     = fmt.Sprintf("%s/%s/%s", registry, mspec.Image.Owner, mspec.Image.Name)
		namespace = fmt.Sprintf("%s:%s", image, mspec.Image.Tag)
		auth      = mspec.Image.Auth
	)

	log.Infof("%s:push:> running push image %s process for task %s to registry %s", logWorkerPrefix, image, w.pid, registry)

	cli, err := docker.NewWithHost(w.endpoint)
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

	req, err := cli.ImagePush(w.ctx, &lbt.SpecTemplateContainerImage{Name: namespace, Auth: auth})
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

	defer func() {
		if req != nil {
			if err := req.Close(); err != nil {
				log.Errorf("%s:push:> close request writer err: %v", logWorkerPrefix, err)
				return
			}
		}
	}()

	result := new(struct {
		Progress    map[string]interface{} `json:"progressDetail"`
		ErrorDetail *struct {
			Message string `json:"message"`
			Error   string `json:"error"`
		} `json:"errorDetail,omitempty"`
		Aux struct {
			Tag    string `json:"Tag"`
			Digest string `json:"Digest"`
			Size   int    `json:"Limit"`
		} `json:"aux"`
	})

	if w.stdout {
		// Checking the writer for the existence of an error error
		err = func(stream io.ReadCloser, data interface{}) error {

			const DEFAULT_BUFFER_SIZE = 5e+6 //  5e+6 = 5MB

			readBytesLast := 0
			bufferLast := make([]byte, DEFAULT_BUFFER_SIZE)

			for {
				buffer := make([]byte, DEFAULT_BUFFER_SIZE)
				readBytes, err := stream.Read(buffer)
				if err != nil && err != io.EOF {
					log.Warnf("%s:push:> read bytes from reader err: %v", logWorkerPrefix, err)
				}
				if readBytes == 0 {
					if err := json.Unmarshal(bufferLast[:readBytesLast], &data); err != nil {
						log.Errorf("%s:push:> parse result writer err: %v", logWorkerPrefix, err)
						result = nil
						break
					}

					if result.ErrorDetail != nil {
						return fmt.Errorf("%s", result.ErrorDetail.Message)
					}

					break
				}

				bufferLast = make([]byte, DEFAULT_BUFFER_SIZE)

				readBytesLast = readBytes
				copy(bufferLast, buffer)

				os.Stdout.Write(buffer[:readBytes])
			}

			return nil
		}(req, result)

		switch err {
		case nil:
		case context.Canceled:
			log.Debugf("%s:build:> process canceled", logWorkerPrefix)
			w.sendEvent(event{step: types.BuildStepUpload, canceled: true})
			return nil
		default:
			log.Errorf("%s:push:> push image err: %v", logWorkerPrefix, err)
			w.sendEvent(event{step: types.BuildStepUpload, message: errorUploadFailed, error: true})
			return err
		}
	}

	info, _, err := cli.ImageInspect(w.ctx, namespace)
	switch err {
	case nil:
	case context.Canceled:
		log.Debugf("%s:build:> process canceled", logWorkerPrefix)
		w.sendEvent(event{step: types.BuildStepUpload, canceled: true})
		return nil
	default:
		log.Errorf("%s:push:> get image info err: %v", logWorkerPrefix, err)
		w.sendEvent(event{step: types.BuildStepUpload, message: errorUploadFailed, error: true})
		return err
	}

	w.sendInfo(info)

	return nil
}

func (w *worker) finish() error {
	log.Infof("%s:finish:> handler after task %s completion", logWorkerPrefix, w.pid)
	w.sendEvent(event{step: types.BuildStepDone})
	return nil
}

func (w *worker) upload() error {

	req, err := w.cri.ContainerLogs(w.ctx, w.gcid, true, true, true)
	switch err {
	case nil:
	case context.Canceled:
		return nil
	default:
		err := fmt.Errorf("running logs stream: %s", err)
		log.Errorf("%s:upload:> logs container err: %v", logWorkerPrefix, err)
		return err
	}
	defer func() {
		if req != nil {
			if err := req.Close(); err != nil {
				log.Errorf("%s:upload:> close log stream err: %s", err)
				return
			}
		}
	}()

	if envs.Get().GetBlobStorage() != nil {
		err = envs.Get().GetBlobStorage().Write(w.task.Meta.ID, cleaner.NewReader(req))
		if err != nil {
			log.Errorf("%s:upload:> write container logs to blob err: %v", logWorkerPrefix, err)
		}
	}

	return nil
}

func (w *worker) cancel() {
	w.cancelFunc()
}

func (w *worker) cleanup() {

	if err := w.cri.ContainerRemove(context.Background(), w.dcid, true, true); err != nil {
		log.Errorf("%s:cleanup:> remove %s container dind  err: %v", logWorkerPrefix, w.dcid, err)
	}

	if err := w.cri.ContainerRemove(context.Background(), w.gcid, true, true); err != nil {
		log.Errorf("%s:cleanup:> remove %s container git err: %v", logWorkerPrefix, w.gcid, err)
	}

}

// Get logs for build process
func (w *worker) logs(writer io.Writer) error {
	return w.logging(writer)
}

func (w *worker) logging(writer io.Writer) error {

	req, err := w.cri.ContainerLogs(w.ctx, w.gcid, true, true, true)
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
func (w *worker) sendInfo(info *lbt.ImageInfo) {
	log.Debugf("%s:send_info> send task status event %s", logWorkerPrefix, w.pid)

	mspec := w.task.Spec
	mmeta := w.task.Meta

	e := new(request.BuildSetImageInfoOptions)
	e.Size = info.Size
	e.Hash = info.ID
	e.VirtualSize = info.VirtualSize

	envs.Get().GetClient().V1().
		Image(mspec.Image.Owner, mspec.Image.Name).
		Build(mmeta.ID).
		SetImageInfo(w.ctx, e)
}
