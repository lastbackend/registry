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
	"fmt"
	"github.com/spf13/viper"
	"time"
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
	Meta   BuildMeta   `json:"meta"`
	Status BuildStatus `json:"status"`
	Spec   BuildSpec   `json:"spec"`
}

type BuildMeta struct {
	Meta
	Number  int64  `json:"number"`
	Builder string `json:"builder"`
}

type BuildStatus struct {
	Size       int64     `json:"size"`
	Step       string    `json:"step"`
	Status     string    `json:"status"`
	Message    string    `json:"message"`
	Processing bool      `json:"processing"`
	Done       bool      `json:"done"`
	Error      bool      `json:"error"`
	Canceled   bool      `json:"canceled"`
	Finished   time.Time `json:"finished"`
	Started    time.Time `json:"started"`
}

type BuildSpec struct {
	Image  BuildImage  `json:"image"`
	Source BuildSource `json:"source"`
	Config BuildConfig `json:"config"`
}

type BuildInfo struct {
	ImageHash string `json:"image_hash"`
	Size      int64  `json:"size"`
}

type BuildImage struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Owner string `json:"owner"`
	Tag   string `json:"tag"`
	Auth  string `json:"auth"`
	Hash  string `json:"hash"`
}

type BuildSource struct {
	Source
	Commit BuildCommit `json:"commit"`
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
	Context    string   `json:"context"`
	Workdir    string   `json:"workdir"`
	EnvVars    []string `json:"env"`
	Command    string   `json:"command"`
}

type BuildStep string

type Task struct {
	Meta struct {
		ID string `json:"id"`
	} `json:"meta"`
	Spec struct {
		Source BuildManifestSource `json:"source"`
		Image  BuildManifestImage  `json:"image"`
		Config BuildManifestConfig `json:"config"`
	} `json:"spec"`
}

type BuildManifestSource struct {
	Url    string `json:"url"`
	Branch string `json:"branch"`
}

type BuildManifestImage struct {
	Host  string `json:"host"`
	Name  string `json:"name"`
	Owner string `json:"owner"`
	Tag   string `json:"tag"`
	Auth  string `json:"auth"`
}

type BuildManifestConfig struct {
	BuildConfig
}

type SourceJobConfig struct {
	Hub    string `json:"hub"`
	Owner  string `json:"owner"`
	Name   string `json:"name"`
	Branch string `json:"branch"`
	// Link to source code in tar
	TarUri string `json:"tar_uri"`
	// Directory where source are unpacked (Ex: github => ./owner/repo/ref)
	WorkDir string `json:"dir"`
}

func (b *Build) MarkAsFetching(step, message string) {
	b.Status.Step = step
	b.Status.Message = message
	b.Status.Status = BuildStatusFetching
	b.Status.Processing = true
	b.Status.Done = false
	b.Status.Error = false
	b.Status.Canceled = false
}

func (b *Build) MarkAsBuilding(step, message string) {
	b.Status.Step = step
	b.Status.Message = message
	b.Status.Status = BuildStatusBuilding
	b.Status.Processing = true
	b.Status.Done = false
	b.Status.Error = false
	b.Status.Canceled = false
}

func (b *Build) MarkAsUploading(step, message string) {
	b.Status.Step = step
	b.Status.Message = message
	b.Status.Status = BuildStatusUploading
	b.Status.Processing = true
	b.Status.Done = false
	b.Status.Error = false
	b.Status.Canceled = false
}

func (b *Build) MarkAsDone(step, message string) {
	b.Status.Step = step
	b.Status.Message = message
	b.Status.Status = BuildStatusSuccess
	b.Status.Processing = false
	b.Status.Done = true
	b.Status.Error = false
	b.Status.Canceled = false
}

func (b *Build) MarkAsError(step, message string) {
	b.Status.Step = step
	b.Status.Message = message
	b.Status.Status = BuildStatusFailed
	b.Status.Processing = false
	b.Status.Done = false
	b.Status.Error = true
	b.Status.Canceled = false
}

func (b *Build) MarkAsCanceled(step, message string) {
	b.Status.Step = step
	b.Status.Message = message
	b.Status.Status = BuildStatusCanceled
	b.Status.Processing = false
	b.Status.Done = false
	b.Status.Error = false
	b.Status.Canceled = true
}

func (b Build) NewBuildManifest() *Task {

	manifest := new(Task)

	manifest.Meta.ID = b.Meta.ID

	manifest.Spec.Image.Host = viper.GetString("domain")
	manifest.Spec.Image.Name = b.Spec.Image.Name
	manifest.Spec.Image.Owner = b.Spec.Image.Owner
	manifest.Spec.Image.Tag = b.Spec.Image.Tag
	manifest.Spec.Image.Auth = b.Spec.Image.Auth

	manifest.Spec.Source.Branch = b.Spec.Source.Branch

	if b.Spec.Source.Hub == GithubHost {
		manifest.Spec.Source.Url = fmt.Sprintf("https://github.com/%s/%s", b.Spec.Source.Owner, b.Spec.Source.Name)

		if b.Spec.Source.Token != "" {
			manifest.Spec.Source.Url = fmt.Sprintf("https://%s@github.com/%s/%s", b.Spec.Source.Token, b.Spec.Source.Owner, b.Spec.Source.Name)
		}
	}

	if b.Spec.Source.Hub == BitbucketHost {
		manifest.Spec.Source.Url = fmt.Sprintf("https://bitbucket.org/%s/%s", b.Spec.Source.Owner, b.Spec.Source.Name)

		if b.Spec.Source.Token != "" {
			manifest.Spec.Source.Url = fmt.Sprintf("https://x-token-auth:%s@bitbucket.org/%s/%s", b.Spec.Source.Token, b.Spec.Source.Owner, b.Spec.Source.Name)
		}
	}

	if b.Spec.Source.Hub == GitlabHost {
		manifest.Spec.Source.Url = fmt.Sprintf("https://gitlab.com/%s/%s", b.Spec.Source.Owner, b.Spec.Source.Name)

		if b.Spec.Source.Token != "" {
			manifest.Spec.Source.Url = fmt.Sprintf("https://gitlab-ci-token:%s@gitlab.com/%s/%s", b.Spec.Source.Token, b.Spec.Source.Owner, b.Spec.Source.Name)
		}
	}

	manifest.Spec.Config.Dockerfile = b.Spec.Config.Dockerfile
	manifest.Spec.Config.Context = b.Spec.Config.Context
	manifest.Spec.Config.EnvVars = b.Spec.Config.EnvVars
	manifest.Spec.Config.Workdir = b.Spec.Config.Workdir
	manifest.Spec.Config.Command = b.Spec.Config.Command

	return manifest
}

// Distribution options

type BuildCreateOptions struct {
	Source Source `json:"source"`
	Image  struct {
		ID    string `json:"id"`
		Owner string `json:"owner"`
		Name  string `json:"name"`
		Tag   string `json:"tag"`
		Auth  string `json:"auth"`
	} `json:"source"`
	Spec struct {
		DockerFile string   `json:"dockerfile"`
		Context    string   `json:"context"`
		Command    string   `json:"command"`
		Workdir    string   `json:"workdir"`
		EnvVars    []string `json:"environments"`
	}
	Labels map[string]string `json:"labels"`
}

type BuildUpdateStatusOptions struct {
	Step     string `json:"step"`
	Message  string `json:"message"`
	Error    bool   `json:"error"`
	Canceled bool   `json:"canceled"`
}

type BuildUpdateInfoOptions struct {
	Size int64  `json:"size"`
	Hash string `json:"hash"`
}

type BuildListOptions struct {
	Active *bool `json:"active"`
}
