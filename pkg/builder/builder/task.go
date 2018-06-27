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
	"github.com/lastbackend/registry/pkg/events"
	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/util/blob"
	"github.com/lastbackend/registry/pkg/util/generator"
	"github.com/lastbackend/registry/pkg/util/validator"
	lbt "github.com/lastbackend/lastbackend/pkg/distribution/types"
	"github.com/lastbackend/lastbackend/pkg/node/runtime/cri"
	"github.com/lastbackend/lastbackend/pkg/node/runtime/cri/docker"
)

const (
	errorBuildFailed  = "build process failed"
	errorUploadFailed = "push process failed"
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

	logger *logger.Logger
	cri    cri.CRI
	job    *types.BuildJob
}

type taskEvent struct {
	step    string
	message string
	error   bool
}

// Creating a new task for incoming job and prepare the environment for build process
// Here the running docker:dind for isolated storage of the image on the host
// until it is sent to the registry.
func newTask(ctx context.Context, builder, dockerHost string, extraHosts []string, cri cri.CRI, job *types.BuildJob) (*task, error) {

	log.Infof("Task: New: create new task for job %s", job.ID)

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

	dcid, err := cri.ContainerCreate(ctx, &spec)
	if err != nil {
		log.Errorf("Task: New: create container with docker:dind err: %s", err)
		return nil, err
	}

	if err := cri.ContainerStart(ctx, dcid); err != nil {
		log.Errorf("Task: New: start container with docker:dind err: %s", err)
		return nil, err
	}

	inspect, err := cri.ContainerInspect(ctx, dcid)
	if err != nil {
		log.Errorf("Task: New: Inspect docker:dind container err: %s", err)
		return nil, err
	}
	if inspect == nil {
		err := fmt.Errorf("docker:dind does not exists")
		log.Errorf("Task: New: container inspect err: %s", err)
		return nil, err
	}
	if inspect.ExitCode != 0 {
		err := fmt.Errorf("docker:dind exit with status code %d", inspect.ExitCode)
		log.Errorf("Task: New: container exit with err: %s", err)
		return nil, err
	}

	var port = ""
	for p, binds := range inspect.Network.Ports {
		match := strings.Split(p, "/")

		if match[0] != "2375" {
			continue
		}

		if len(binds) == 0 {
			err := fmt.Errorf("there are no ports available")
			log.Errorf("Task: New: cannot receive docker daemon connection port: %s", err)
			return nil, err
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

	ct, cn := context.WithCancel(ctx)

	if len(dockerHost) == 0 {
		dockerHost = "172.17.0.1"
	}

	return &task{
		ctx:      ct,
		cancel:   cn,
		id:       generator.GetUUIDV4(),
		builder:  builder,
		endpoint: fmt.Sprintf("tcp://%s:%s", dockerHost, port),
		dcid:     dcid,
		cri:      cri,
		job:      job,
		logger:   logger.NewLogger(ctx),
	}, nil
}

func (t *task) Canceled() bool {
	return t.canceled
}

// Running build process
func (t *task) build() error {

	t.step = types.BuildStepBuild
	t.sendEvent(taskEvent{step: types.BuildStepBuild})

	var (
		image      = fmt.Sprintf("%s/%s/%s:%s", t.job.Image.Host, t.job.Image.Owner, t.job.Image.Name, t.job.Image.Tag)
		dockerfile = t.job.Config.Dockerfile
		gituri     = fmt.Sprintf("%s.git#%s", t.job.Repo, strings.ToLower(t.job.Branch))
	)

	log.Infof("Task: Build: running build image %s process for job %s", image, t.job.ID)

	if validator.IsValueInList(dockerfile, []string{"", " ", "/", "./", "../", "Dockerfile", "/Dockerfile"}) {
		dockerfile = "./Dockerfile"
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
		log.Debugf("Task: Build: process canceled")
		t.canceled = true
		t.sendEvent(taskEvent{step: types.BuildStepBuild})
		return nil
	default:
		log.Errorf("Task: Build: create container err: %s", err)
		t.error = errorBuildFailed
		t.sendEvent(taskEvent{step: types.BuildStepBuild})
		return err
	}

	err = t.cri.ContainerStart(t.ctx, cid)
	switch err {
	case nil:
	case context.Canceled:
		log.Debugf("Task: Build: process canceled")
		t.canceled = true
		t.sendEvent(taskEvent{step: types.BuildStepBuild})
		return nil
	default:
		log.Errorf("Task: Build: start container err: %s", err)
		t.error = errorBuildFailed
		t.sendEvent(taskEvent{step: types.BuildStepBuild})
		return err
	}

	req, err := t.cri.ContainerLogs(t.ctx, cid, true, true, true)
	switch err {
	case nil:
	case context.Canceled:
		log.Debugf("Task: Build: process canceled")
		t.canceled = true
		t.sendEvent(taskEvent{step: types.BuildStepBuild})
		return nil
	default:
		err := fmt.Errorf("running logs stream: %s", err)
		log.Errorf("Task: Build: logs container err: %s", err)
		t.error = errorBuildFailed
		t.sendEvent(taskEvent{step: types.BuildStepBuild})
		return err
	}

	defer func() {
		if req != nil {
			if err := req.Close(); err != nil {
				log.Errorf("Task: Build: close log stream err: %s", err)
				return
			}
		}
	}()

	go func() {
		if err := t.writeLogFile(); err != nil {
			log.Warnf("Task: Build: write logs file err: %s", err)
			return
		}
	}()

	err = t.logger.Run(req, true)
	switch err {
	case nil:
	case context.Canceled:
		log.Debugf("Task: Build: process canceled")
		t.canceled = true
		t.sendEvent(taskEvent{step: types.BuildStepBuild})
		return nil
	default:
		log.Warnf("Task: Build: write logs to stdout err: %s", err)
	}

	inspect, err := t.cri.ContainerInspect(t.ctx, cid)
	switch err {
	case nil:
	case context.Canceled:
		log.Debugf("Task: Build: process canceled")
		t.canceled = true
		t.sendEvent(taskEvent{step: types.BuildStepBuild})
		return nil
	default:
		log.Errorf("Task: Build: inspect container %s", err)
		t.error = errorBuildFailed
		t.sendEvent(taskEvent{step: types.BuildStepBuild})
		return err
	}

	// todo check canceled context
	if inspect == nil {
		err := fmt.Errorf("docker:container daes not exists")
		log.Errorf("Task: New: container inspect err: %s", err)
		t.error = errorBuildFailed
		t.sendEvent(taskEvent{step: types.BuildStepBuild})
		return err
	}
	if err == nil && inspect.ExitCode != 0 {
		err := fmt.Errorf("container exited with %d code", inspect.ExitCode)
		log.Errorf("Task: Build: container exit with err %s", err)
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
		registry  = t.job.Image.Host
		image     = fmt.Sprintf("%s/%s/%s", registry, t.job.Image.Owner, t.job.Image.Name)
		namespace = fmt.Sprintf("%s:%s", image, t.job.Image.Tag)
		auth      = t.job.Image.Token
	)

	log.Infof("Task: Push: running push image %s process for job %s to registry %s", image, t.job.ID, registry)

	cli, err := docker.NewWithHost(t.endpoint)
	switch err {
	case nil:
	case context.Canceled:
		log.Debugf("Task: Build: process canceled")
		t.canceled = true
		t.sendEvent(taskEvent{step: types.BuildStepUpload})
		return nil
	default:
		log.Errorf("Task: Push: create docker client %s", err)
		t.error = errorUploadFailed
		t.sendEvent(taskEvent{step: types.BuildStepUpload})
		return err
	}

	req, err := cli.ImagePush(t.ctx, &lbt.SpecTemplateContainerImage{Name: namespace, Auth: auth})
	switch err {
	case nil:
	case context.Canceled:
		log.Debugf("Task: Build: process canceled")
		t.canceled = true
		t.sendEvent(taskEvent{step: types.BuildStepUpload})
		return nil
	default:
		log.Errorf("Task: Push: running push process err: %s", err)
		t.error = errorUploadFailed
		t.sendEvent(taskEvent{step: types.BuildStepUpload})
		return err
	}

	defer func() {
		if req != nil {
			if err := req.Close(); err != nil {
				log.Errorf("Task: Push: close request stream err: %s", err)
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

	// Сhecking the stream for the existence of an error error
	err = func(stream io.ReadCloser, data interface{}) error {

		const DEFAULT_BUFFER_SIZE = 5e+6 //  5e+6 = 5MB

		readBytesLast := 0
		bufferLast := make([]byte, DEFAULT_BUFFER_SIZE)

		for {
			buffer := make([]byte, DEFAULT_BUFFER_SIZE)
			readBytes, err := stream.Read(buffer)
			if err != nil && err != io.EOF {
				log.Warnf("Read bytes from reader err: %s", err)
			}
			if readBytes == 0 {
				if err := json.Unmarshal(bufferLast[:readBytesLast], &data); err != nil {
					log.Errorf("Task: Push: Parse result stream err: %s", err)
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
		log.Debugf("Task: Build: process canceled")
		t.canceled = true
		t.sendEvent(taskEvent{step: types.BuildStepUpload})
		return nil
	default:
		log.Errorf("Task: Push: push image err: %s", err)
		t.error = errorUploadFailed
		t.sendEvent(taskEvent{step: types.BuildStepUpload})
		return err
	}

	info, _, err := cli.ImageInspect(t.ctx, namespace)
	switch err {
	case nil:
	case context.Canceled:
		log.Debugf("Task: Build: process canceled")
		t.canceled = true
		t.sendEvent(taskEvent{step: types.BuildStepUpload})
		return nil
	default:
		log.Errorf("Task: Push: get image info err: %s", err)
		t.error = errorUploadFailed
		t.sendEvent(taskEvent{step: types.BuildStepUpload})
		return err
	}

	t.sendInfo(info)

	return nil
}

// Handler after task completion
func (t *task) finish() error {

	log.Infof("Task: Finish: handler after task %s completion", t.job.ID)

	defer func() {
		if err := t.cri.ContainerRemove(context.Background(), t.dcid, true, true); err != nil {
			log.Errorf("Task: Finish: cleanup err: %s", err)
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
			log.Errorf("Task: Finish: can not open log file for reading: %s", err)
			return err
		}

		if bs.AccountName != "" || bs.AccountKey != "" {
			cli := blob.NewClient(bs.AccountName, bs.AccountKey)
			if err := cli.Write(blob.CONTAINER_LOGS_NAME, t.id, reader); err != nil {
				log.Errorf("Task: Finish: write blob err: %s", err)
			} else {
				err := os.Remove(filePath)
				if err != nil {
					log.Errorf("Task: Finish: remove file [%s] err: %s", filePath, err)
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

	log.Infof("Task: WriteLogFile: handler logger to file for task %s", t.job.ID)

	logdir := t.ctx.Value("logdir").(string)

	if !strings.HasSuffix(logdir, string(os.PathSeparator)) {
		logdir += "/"
	}

	filePath := fmt.Sprintf("%s%s", logdir, t.id)

	if err := os.MkdirAll(logdir, os.ModePerm); err != nil {
		log.Errorf("Task: WriteLogFile: create directories [%s] err: %s", logdir, err)
		return err
	}

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		log.Errorf("Task: WriteLogFile: open file [%s] err: %s", filePath, err)
		return err
	}

	// close fi on exit and check for its returned error
	defer func() {
		if err := file.Close(); err != nil {
			log.Errorf("Task: WriteLogFile: close file [%s] err: %s", filePath, err)
		}
	}()

	if err := t.logger.Pipe(file); err != nil {
		log.Errorf("Task: WriteLogFile: write to file [%s] err: %s", filePath, err)
		return err
	}

	return nil
}

// Send status build event to controller
func (t *task) sendEvent(event taskEvent) {

	log.Debugf("Task: SendEvent: send task status event %s", t.job.ID)

	e := types.BuildStateBuilderEventPayload{}
	e.Build = t.job.ID
	e.Builder = t.builder
	e.Task = t.id
	e.State.Step = event.step
	e.State.Message = event.message
	e.State.Error = event.error
	e.State.Canceled = t.ctx.Err() == context.Canceled

	events.BuildStateEventRequest(envs.Get().GetRPC(), e)
}

// Send status build event to controller
func (t *task) sendInfo(info *lbt.ImageInfo) {
	log.Debugf("Task: SendEvent: send task status event %s", t.job.ID)
	events.BuildImageInfoRequest(envs.Get().GetRPC(), t.job.ID, info)
}
