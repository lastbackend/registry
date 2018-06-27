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

package types

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/lastbackend/pkg/distribution/errors"
)

// Statuses
const (
	BuildStatusQueued    = "queued"
	BuildStatusFetching  = "fetching"
	BuildStatusBuilding  = "building"
	BuildStatusUploading = "uploading"
	BuildStatusSuccess   = "success"
	BuildStatusFailed    = "failed"
	BuildStatusCanceled  = "canceled"
)

// STEPS
const (
	BuildStepFetch  = "fetch"
	BuildStepBuild  = "build"
	BuildStepUpload = "upload"
	BuildStepDone   = "done"
)

type BuildList []*Build

type Build struct {
	Meta    BuildMeta    `json:"meta"`
	Repo    BuildRepo    `json:"repo"`
	State   BuildState   `json:"state"`
	Stats   BuildInfo    `json:"info"`
	Image   BuildImage   `json:"image"`
	Sources BuildSources `json:"sources"`
	Config  BuildConfig  `json:"config"`
}

type BuildMeta struct {
	Meta
	Number   int64  `json:"number"`
	SelfLink string `json:"self_link"`
	Builder  string `json:"builder"`
	Task     string `json:"task"`
}

type BuildRepo struct {
	ID       string `json:"id"`
	Owner    string `json:"owner"`
	Name     string `json:"name"`
	SelfLink string `json:"self_link"`
}

type BuildState struct {
	Step       string    `json:"step"`
	Status     string    `json:"status"`
	Message    string    `json:"message"`
	Processing bool      `json:"processing"`
	Done       bool      `json:"done"`
	Error      bool      `json:"error"`
	Canceled   bool      `json:"canceled"`
	Created    time.Time `json:"created"`
	Updated    time.Time `json:"updated"`
	Finished   time.Time `json:"finished"`
	Started    time.Time `json:"started"`
}

type BuildInfo struct {
	ImageHash string `json:"image_hash"`
	Size      int64  `json:"size"`
}

type BuildImage struct {
	Hub   string `json:"hub"`
	Name  string `json:"name"`
	Owner string `json:"owner"`
	Tag   string `json:"tag"`
	Hash  string `json:"hash"`
}

type BuildSources struct {
	Token  string      `json:"token,omitempty"`
	Hub    string      `json:"hub,omitempty"`
	Owner  string      `json:"owner,omitempty"`
	Name   string      `json:"name,omitempty"`
	Branch string      `json:"branch,omitempty"`
	Commit BuildCommit `json:"commit,omitempty"`
}

type BuildCommit struct {
	Hash     string    `json:"hash"`
	Username string    `json:"username"`
	Message  string    `json:"message"`
	Email    string    `json:"email"`
	Date     time.Time `json:"date"`
}

type BuildConfig struct {
	Dockerfile string   `json:"dockerfile"`
	Workdir    string   `json:"workdir"`
	EnvVars    []string `json:"env"`
}

type BuildStep string

type BuildJob struct {
	ID     string         `json:"id,omitempty"`
	Meta   BuildJobMeta   `json:"meta"`
	Image  BuildJobImage  `json:"image"`
	Config BuildJobConfig `json:"config"`
	Repo   string         `json:"repo"`
	Branch string         `json:"branch"`
	LogUri string         `json:"log_uri"`
}

type BuildJobMeta struct {
	ID      string `json:"id"`
	LogsUri string `json:"logs_uri"`
}

type BuildJobImage struct {
	Host  string `json:"host"`
	Name  string `json:"name"`
	Owner string `json:"owner"`
	Tag   string `json:"tag"`
	Token string `json:"token"`
}

type BuildJobConfig struct {
	BuildConfig
}

type SourceJobConfig struct {
	Hub    string `json:"hub"`
	Owner  string `json:"owner"`
	Name   string `json:"name"`
	Branch string `json:"branch"`
	// Link to source code in tar
	TarUri string `json:"tar_uri"`
	// Directory where sources are unpacked (Ex: github => ./owner/repo/ref)
	WorkDir string `json:"dir"`
}

func (b *Build) NewBuildJob() *BuildJob {

	job := new(BuildJob)
	job.ID = b.Meta.ID
	job.Image.Name = b.Image.Name
	job.Image.Owner = b.Image.Owner
	job.Image.Tag = b.Image.Tag
	job.Image.Host = b.Image.Hub

	url := ""

	if b.Sources.Hub == GithubHost {
		url = fmt.Sprintf("https://github.com/%s/%s", b.Sources.Owner, b.Sources.Name)

		if b.Sources.Token != "" {
			url = fmt.Sprintf("https://%s@github.com/%s/%s", b.Sources.Token, b.Sources.Owner, b.Sources.Name)
		}
	}

	if b.Sources.Hub == BitbucketHost {
		url = fmt.Sprintf("https://bitbucket.org/%s/%s", b.Sources.Owner, b.Sources.Name)

		if b.Sources.Token != "" {
			url = fmt.Sprintf("https://x-token-auth:%s@bitbucket.org/%s/%s", b.Sources.Token, b.Sources.Owner, b.Sources.Name)
		}
	}

	if b.Sources.Hub == GitlabHost {
		url = fmt.Sprintf("https://gitlab.com/%s/%s", b.Sources.Owner, b.Sources.Name)

		if b.Sources.Token != "" {
			url = fmt.Sprintf("https://gitlab-ci-token:%s@gitlab.com/%s/%s", b.Sources.Token, b.Sources.Owner, b.Sources.Name)
		}
	}

	job.Repo = url
	job.Branch = b.Sources.Branch

	return job
}

type BuildCreateOptions struct {
	Tag    *string `json:"tag"`
	TarUri *string `json:"tar"`
}

func (s *BuildCreateOptions) DecodeAndValidate(reader io.Reader) *errors.Err {

	log.V(logLevel).Debug("Request: Build: decode and validate data for creating")

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		log.V(logLevel).Errorf("Request: Build: decode and validate data for creating err: %s", err)
		return errors.New("build").Unknown(err)
	}

	err = json.Unmarshal(body, s)
	if err != nil {
		log.V(logLevel).Errorf("Request: Build: convert struct from json err: %s", err)
		return errors.New("build").IncorrectJSON(err)
	}

	switch true {
	case s.Tag == nil:
		log.V(logLevel).Errorf("Request: Repo: parameter tag can not be empty")
		return errors.New("build").BadParameter("tag")
	}

	return nil
}

func (b *Build) MarkAsFetching(step, message string) {
	b.State.Step = step
	b.State.Message = message
	b.State.Status = BuildStatusFetching
	b.State.Processing = true
	b.State.Done = false
	b.State.Error = false
	b.State.Canceled = false
}

func (b *Build) MarkAsBuilding(step, message string) {
	b.State.Step = step
	b.State.Message = message
	b.State.Status = BuildStatusBuilding
	b.State.Processing = true
	b.State.Done = false
	b.State.Error = false
	b.State.Canceled = false
}

func (b *Build) MarkAsUploading(step, message string) {
	b.State.Step = step
	b.State.Message = message
	b.State.Status = BuildStatusUploading
	b.State.Processing = true
	b.State.Done = false
	b.State.Error = false
	b.State.Canceled = false
}

func (b *Build) MarkAsDone(step, message string) {
	b.State.Step = step
	b.State.Message = message
	b.State.Status = BuildStatusSuccess
	b.State.Processing = false
	b.State.Done = true
	b.State.Error = false
	b.State.Canceled = false
}

func (b *Build) MarkAsError(step, message string) {
	b.State.Step = step
	b.State.Message = message
	b.State.Status = BuildStatusFailed
	b.State.Processing = false
	b.State.Done = false
	b.State.Error = true
	b.State.Canceled = false
}

func (b *Build) MarkAsCanceled(step, message string) {
	b.State.Step = step
	b.State.Message = message
	b.State.Status = BuildStatusCanceled
	b.State.Processing = false
	b.State.Done = false
	b.State.Error = false
	b.State.Canceled = true
}
