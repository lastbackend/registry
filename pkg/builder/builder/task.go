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

	"github.com/lastbackend/registry/pkg/builder/envs"
	"github.com/lastbackend/registry/pkg/builder/logger"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/runtime/cri"
	"github.com/lastbackend/registry/pkg/runtime/cri/docker"
	"github.com/lastbackend/registry/pkg/util/blob"
	"github.com/lastbackend/registry/pkg/util/validator"
	"github.com/lastbackend/registry/pkg/api/types/v1/request"
	lbt "github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/util/generator"
)

const (
	logTaskPrefix = "builder:task"
)

const (
	errorBuildFailed  = "build process failed"
	errorUploadFailed = "push process failed"
)

const (
	defaultDockerHost = "172.17.0.1"
)

// The task is responsible for the process
// of image assembly, sending to the registry and logging
type task struct {
	ctx    context.Context
	cancel context.CancelFunc

	id       string
	builder  string
	step     string
	error    string
	dcid     string
	endpoint string

	canceled bool

	logger   *logger.Logger
	cri      cri.CRI
	manifest *types.BuildManifest
}

type taskEvent struct {
	step    string
	message string
	error   bool
}

// Creating a new task for incoming manifest and configure the environment for build process
// Here the running docker:dind for isolated storage of the image on the host
// until it is sent to the registry.
func NewTask(ctx context.Context, cri cri.CRI) (*task, error) {

	log.Infof("%s:new:> create new task", logTaskPrefix)

	var (
		builder = ctx.Value("builder").(string)
	)

	ct, cn := context.WithCancel(ctx)

	return &task{
		ctx:     ct,
		cancel:  cn,
		id:      generator.GetUUIDV4(),
		builder: builder,
		cri:     cri,
		logger:  logger.NewLogger(ctx),
	}, nil
}

func (t *task) Canceled() bool {
	return t.canceled
}

// Running build process
func (t *task) Start(manifest *types.BuildManifest) error {

	var (
		extraHosts = make([]string, 0)
		dockerHost = defaultDockerHost
	)

	if t.ctx.Value("extraHosts") != nil {
		extraHosts = t.ctx.Value("extraHosts").([]string)
	}

	if t.ctx.Value("dockerHost") != nil {
		dockerHost = t.ctx.Value("dockerHost").(string)
	}

	t.manifest = manifest

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
		Labels:          map[string]string{"LB": ""},
		PublishAllPorts: true,
	}

	dcid, err := t.cri.ContainerCreate(t.ctx, &spec)
	if err != nil {
		log.Errorf("%s:start:> create container with docker:dind err: %v", logTaskPrefix, err)
		return err
	}

	if err := t.cri.ContainerStart(t.ctx, dcid); err != nil {
		log.Errorf("%s:start:> start container with docker:dind err: %v", logTaskPrefix, err)
		return err
	}

	inspect, err := t.cri.ContainerInspect(t.ctx, dcid)
	if err != nil {
		log.Errorf("%s:start:> Inspect docker:dind container err: %v", logTaskPrefix, err)
		return err
	}
	if inspect == nil {
		err := fmt.Errorf("docker:dind does not exists")
		log.Errorf("%s:start:> container inspect err: %v", logTaskPrefix, err)
		return err
	}
	if inspect.ExitCode != 0 {
		err := fmt.Errorf("docker:dind exit with status code %d", inspect.ExitCode)
		log.Errorf("%s:start:> container exit with err: %v", logTaskPrefix, err)
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
			log.Errorf("%s:start:> cannot receive docker daemon connection port: %v", logTaskPrefix, err)
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

	t.endpoint = fmt.Sprintf("tcp://%s:%s", dockerHost, port)
	t.dcid = dcid

	defer func() {
		if err := t.finish(); err != nil && err != context.Canceled {
			log.Errorf("%s:start:> finish task %s err:  %s", logWorkerPrefix, t.id, err)
		}
	}()

	err = t.build()
	if err == nil && t.Canceled() {
		return nil
	}
	if err != nil && err != context.Canceled {
		log.Errorf("%s:start:> build t %s err:  %s", logWorkerPrefix, t.id, err)
		return err
	}

	err = t.push()
	if err == nil && t.Canceled() {
		return nil
	}
	if err != nil && err != context.Canceled {
		log.Errorf("%s:start:> push %s err:  %s", logWorkerPrefix, t.id, err)
		return err
	}

	return nil
}

// Running build process
func (t *task) build() error {

	t.step = types.BuildStepBuild
	t.sendEvent(taskEvent{step: types.BuildStepBuild})

	var (
		image      = fmt.Sprintf("%s/%s/%s:%s", t.manifest.Image.Host, t.manifest.Image.Owner, t.manifest.Image.Name, t.manifest.Image.Tag)
		dockerfile = t.manifest.Config.Dockerfile
		gituri     = fmt.Sprintf("%s.git#%s", t.manifest.Source.Url, strings.ToLower(t.manifest.Source.Branch))
	)

	if len(t.manifest.Config.Context) != 0 && t.manifest.Config.Context != "/" {
		gituri = fmt.Sprintf("%s:%s", gituri, t.manifest.Config.Context)
	}

	log.Infof("%s:build:> running build image %s process for manifest %s", logTaskPrefix, image, t.id)

	if validator.IsValueInList(dockerfile, []string{"", " ", "/", "./", "../", "DockerFile", "/DockerFile"}) {
		dockerfile = "./DockerFile"
	}

	// TODO: change this logic to docker client [cli.ImageBuild]
	spec := &lbt.SpecTemplateContainer{
		AutoRemove: true,
		Image: lbt.SpecTemplateContainerImage{
			Name: "docker:git",
		},
		Labels:  map[string]string{"LB": ""},
		EnvVars: []lbt.SpecTemplateContainerEnv{{Name: "DOCKER_HOST", Value: t.endpoint}},
		Exec: lbt.SpecTemplateContainerExec{
			Command: []string{"build", "-f", dockerfile, "-t", image, gituri},
		},
	}

	cid, err := t.cri.ContainerCreate(t.ctx, spec)
	switch err {
	case nil:
	case context.Canceled:
		log.Debugf("%s:build:> process canceled", logTaskPrefix)
		t.canceled = true
		t.sendEvent(taskEvent{step: types.BuildStepBuild})
		return nil
	default:
		log.Errorf("%s:build:> create container err: %v", logTaskPrefix, err)
		t.error = errorBuildFailed
		t.sendEvent(taskEvent{step: types.BuildStepBuild})
		return err
	}

	err = t.cri.ContainerStart(t.ctx, cid)
	switch err {
	case nil:
	case context.Canceled:
		log.Debugf("%s:build:> process canceled", logTaskPrefix)
		t.canceled = true
		t.sendEvent(taskEvent{step: types.BuildStepBuild})
		return nil
	default:
		log.Errorf("%s:build:> start container err: %v", logTaskPrefix, err)
		t.error = errorBuildFailed
		t.sendEvent(taskEvent{step: types.BuildStepBuild})
		return err
	}

	req, err := t.cri.ContainerLogs(t.ctx, cid, true, true, true)
	switch err {
	case nil:
	case context.Canceled:
		log.Debugf("%s:build:> process canceled", logTaskPrefix)
		t.canceled = true
		t.sendEvent(taskEvent{step: types.BuildStepBuild})
		return nil
	default:
		err := fmt.Errorf("running logs stream: %s", err)
		log.Errorf("%s:build:> logs container err: %v", logTaskPrefix, err)
		t.error = errorBuildFailed
		t.sendEvent(taskEvent{step: types.BuildStepBuild})
		return err
	}

	defer func() {
		if req != nil {
			if err := req.Close(); err != nil {
				log.Errorf("%s:build:> close log stream err: %s", err)
				return
			}
		}
	}()

	go func() {
		if err := t.writeLogFile(); err != nil {
			log.Warnf("%s:build:> write logs file err: %s", err)
			return
		}
	}()

	err = t.logger.Run(req, true)
	switch err {
	case nil:
	case context.Canceled:
		log.Debugf("%s:build:> process canceled", logTaskPrefix)
		t.canceled = true
		t.sendEvent(taskEvent{step: types.BuildStepBuild})
		return nil
	default:
		log.Warnf("%s:build:> write logs to stdout err: %v", logTaskPrefix, err)
	}

	inspect, err := t.cri.ContainerInspect(t.ctx, cid)
	switch err {
	case nil:
	case context.Canceled:
		log.Debugf("%s:build:> process canceled", logTaskPrefix)
		t.canceled = true
		t.sendEvent(taskEvent{step: types.BuildStepBuild})
		return nil
	default:
		log.Errorf("%s:build:> inspect container %v", logTaskPrefix, err)
		t.error = errorBuildFailed
		t.sendEvent(taskEvent{step: types.BuildStepBuild})
		return err
	}

	if inspect == nil {
		err := fmt.Errorf("docker:container daes not exists")
		log.Errorf("%s:build:> container inspect err: %v", logTaskPrefix, err)
		t.error = errorBuildFailed
		t.sendEvent(taskEvent{step: types.BuildStepBuild})
		return err
	}
	if err == nil && inspect.ExitCode != 0 {
		err := fmt.Errorf("container exited with %d code", inspect.ExitCode)
		log.Errorf("%s:build:> container exit with err %v", logTaskPrefix, err)
		t.error = errorBuildFailed
		t.sendEvent(taskEvent{step: types.BuildStepBuild})
		return err
	}

	return nil
}

// Running push process to registry
func (t *task) push() error {

	t.step = types.BuildStepUpload
	t.sendEvent(taskEvent{step: types.BuildStepUpload})

	var (
		registry  = t.manifest.Image.Host
		image     = fmt.Sprintf("%s/%s/%s", registry, t.manifest.Image.Owner, t.manifest.Image.Name)
		namespace = fmt.Sprintf("%s:%s", image, t.manifest.Image.Tag)
		auth      = t.manifest.Image.Auth
	)

	log.Infof("%s:push:> running push image %s process for manifest %s to registry %s", logTaskPrefix, image, t.id, registry)

	cli, err := docker.NewWithHost(t.endpoint)
	switch err {
	case nil:
	case context.Canceled:
		log.Debugf("%s:build:> process canceled", logTaskPrefix)
		t.canceled = true
		t.sendEvent(taskEvent{step: types.BuildStepUpload})
		return nil
	default:
		log.Errorf("%s:push:> create docker client %v", logTaskPrefix, err)
		t.error = errorUploadFailed
		t.sendEvent(taskEvent{step: types.BuildStepUpload})
		return err
	}

	req, err := cli.ImagePush(t.ctx, &lbt.SpecTemplateContainerImage{Name: namespace, Auth: auth})
	switch err {
	case nil:
	case context.Canceled:
		log.Debugf("%s:build:> process canceled", logTaskPrefix)
		t.canceled = true
		t.sendEvent(taskEvent{step: types.BuildStepUpload})
		return nil
	default:
		log.Errorf("%s:push:> running push process err: %v", logTaskPrefix, err)
		t.error = errorUploadFailed
		t.sendEvent(taskEvent{step: types.BuildStepUpload})
		return err
	}

	defer func() {
		if req != nil {
			if err := req.Close(); err != nil {
				log.Errorf("%s:push:> close request stream err: %v", logTaskPrefix, err)
				return
			}
		}
	}()

	result := new(struct {
		Progress map[string]interface{} `json:"progressDetail"`
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

	// Сhecking the stream for the existence of an error error
	err = func(stream io.ReadCloser, data interface{}) error {

		const DEFAULT_BUFFER_SIZE = 5e+6 //  5e+6 = 5MB

		readBytesLast := 0
		bufferLast := make([]byte, DEFAULT_BUFFER_SIZE)

		for {
			buffer := make([]byte, DEFAULT_BUFFER_SIZE)
			readBytes, err := stream.Read(buffer)
			if err != nil && err != io.EOF {
				log.Warnf("%s:push:> read bytes from reader err: %v", logTaskPrefix, err)
			}
			if readBytes == 0 {
				if err := json.Unmarshal(bufferLast[:readBytesLast], &data); err != nil {
					log.Errorf("%s:push:> parse result stream err: %v", logTaskPrefix, err)
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
		log.Debugf("%s:build:> process canceled", logTaskPrefix)
		t.canceled = true
		t.sendEvent(taskEvent{step: types.BuildStepUpload})
		return nil
	default:
		log.Errorf("%s:push:> push image err: %v", logTaskPrefix, err)
		t.error = errorUploadFailed
		t.sendEvent(taskEvent{step: types.BuildStepUpload})
		return err
	}

	info, _, err := cli.ImageInspect(t.ctx, namespace)
	switch err {
	case nil:
	case context.Canceled:
		log.Debugf("%s:build:> process canceled", logTaskPrefix)
		t.canceled = true
		t.sendEvent(taskEvent{step: types.BuildStepUpload})
		return nil
	default:
		log.Errorf("%s:push:> get image info err: %v", logTaskPrefix, err)
		t.error = errorUploadFailed
		t.sendEvent(taskEvent{step: types.BuildStepUpload})
		return err
	}

	t.sendInfo(info)

	return nil
}

// Handler after task completion
func (t *task) finish() error {

	log.Infof("%s:finish:> handler after task %s completion", logTaskPrefix, t.id)

	defer func() {
		if err := t.cri.ContainerRemove(context.Background(), t.dcid, true, true); err != nil {
			log.Errorf("%s:finish:> cleanup err: %v", logTaskPrefix, err)
			return
		}
	}()

	var (
		filePath = fmt.Sprintf("%s%s", t.ctx.Value("logdir").(string), t.id)
		bs       = t.ctx.Value("blob-storage").(types.AzureBlobStorage)
	)

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		reader, err := os.Open(filePath)
		if err != nil {
			log.Errorf("%s:finish:> can not open log file for reading: %v", logTaskPrefix, err)
			return err
		}

		// TODO: change to the storage module from vendors
		if bs.AccountName != "" || bs.AccountKey != "" {
			cli := blob.NewClient(bs.AccountName, bs.AccountKey)
			if err := cli.Write(blob.CONTAINER_LOGS_NAME, t.id, reader); err != nil {
				log.Errorf("%s:finish:> write blob err: %v", logTaskPrefix, err)
			} else {
				err := os.Remove(filePath)
				if err != nil {
					log.Errorf("%s:finish:> remove file [%s] err: %v", logTaskPrefix, filePath, err)
					return err
				}
			}

		}
	}

	// Notify about completed task with status
	if len(t.error) != 0 {
		t.sendEvent(taskEvent{step: t.step, message: t.error, error: true})
	} else {
		t.sendEvent(taskEvent{step: types.BuildStepDone})
	}

	return nil
}

func (t *task) stop() {
	t.cancel()
}

// Handler write logs to file
func (t *task) writeLogFile() error {

	log.Infof("%s:write_to_file:> handler logger to file for task %s", logTaskPrefix, t.id)

	logdir := t.ctx.Value("logdir").(string)

	if !strings.HasSuffix(logdir, string(os.PathSeparator)) {
		logdir += "/"
	}

	filePath := fmt.Sprintf("%s%s", logdir, t.id)

	if err := os.MkdirAll(logdir, os.ModePerm); err != nil {
		log.Errorf("%s:write_to_file:> create directories [%s] err: %v", logTaskPrefix, logdir, err)
		return err
	}

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		log.Errorf("%s:write_to_file:> open file [%s] err: %v", logTaskPrefix, filePath, err)
		return err
	}

	// close fi on exit and check for its returned error
	defer func() {
		if err := file.Close(); err != nil {
			log.Errorf("%s:write_to_file:> close file [%s] err: %v", logTaskPrefix, filePath, err)
		}
	}()

	if err := t.logger.Pipe(file); err != nil {
		log.Errorf("%s:write_to_file:> write to file [%s] err: %v", logTaskPrefix, filePath, err)
		return err
	}

	return nil
}

// Send status build event to controller
func (t *task) sendEvent(event taskEvent) {

	log.Debugf("%s:send_event> send task status event %s", logTaskPrefix, t.id)

	e := new(request.BuildUpdateStatusOptions)
	e.Step = event.step
	e.Message = event.message
	e.Error = event.error
	e.Canceled = t.ctx.Err() == context.Canceled

	envs.Get().GetClient().V1().Build().SetStatus(t.ctx, t.id, e)
}

// Send status build event to controller
func (t *task) sendInfo(info *lbt.ImageInfo) {
	log.Debugf("%s:send_info> send task status event %s", logTaskPrefix, t.id)

	e := new(request.BuildUpdateImageInfoOptions)
	e.Size = info.Size
	e.Hash = info.ID
	e.VirtualSize = info.VirtualSize

	envs.Get().GetClient().V1().Build().SetImageInfo(t.ctx, t.id, e)
}
