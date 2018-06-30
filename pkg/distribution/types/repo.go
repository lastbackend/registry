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
	"github.com/lastbackend/registry/pkg/distribution/errors"
	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/util/validator"
	"io"
	"io/ioutil"
	"strings"
	"sync"
	"time"
)

const RepoDefaultBranch = "master"
const RepoDefaultTag = "latest"
const RepoDefaultDockerfilePath = "./Dockerfile"
const RepoDefaultTechnology = "docker"

type Repo struct {
	lock sync.RWMutex
	// Repo meta
	Meta RepoMeta `json:"meta"`
	// Repo source
	Sources RepoSources `json:"sources"`
	// Repo tags
	Tags map[string]*RepoTag `json:"tags"`
	// Repo meta readme
	Readme string `json:"readme"`
	// Remote repo option
	Remote bool `json:"remote"`
	// Repo created time
	Created time.Time `json:"created"`
	// Repo updated time
	Updated time.Time `json:"updated"`
}

type RepoList []*Repo

type RepoMeta struct {
	Meta
	// Repo meta user
	Owner string `json:"owner"`
	// Repo self_link
	SelfLink string `json:"self_link"`
}

type RepoSources struct {
	Hub    string `json:"hub"`
	Owner  string `json:"owner"`
	Name   string `json:"name"`
	Branch string `json:"branch"`
}

type RepoLastBuild struct {
	ID      string    `json:"id"`
	Status  string    `json:"status"`
	Tag     string    `json:"tag"`
	Number  int       `json:"number"`
	Updated time.Time `json:"updated"`
}

type RepoTag struct {
	ID     string      `json:"id"`
	RepoID string      `json:"repo"`
	Owner  string      `json:"owner"`
	Name   string      `json:"name"`
	Spec   RepoTagSpec `json:"spec"`
	Builds struct {
		Total int64 `json:"total"`
		Size  int64 `json:"size"`
	} `json:"builds"`
	Build0   *BuildView `json:"build_0"`
	Build1   *BuildView `json:"build_1"`
	Build2   *BuildView `json:"build_2"`
	Build3   *BuildView `json:"build_3"`
	Build4   *BuildView `json:"build_4"`
	Disabled bool       `json:"disabled"`
	Created  time.Time  `json:"created"`
	Updated  time.Time  `json:"updated"`
}

type RepoTagSpec struct {
	Branch   string   `json:"branch"`
	FilePath string   `json:"filepath"`
	EnvVars  []string `json:"environments"`
}

type BuildView struct {
	ID     string `json:"id"`
	Number int64  `json:"number"`
	Status string `json:"status"`
}

type RepoCreateOptions struct {
	Meta struct {
		Labels map[string]string `json:"labels"`
	} `json:"meta"`
	Spec struct {
		Image  RepoImageOpts   `json:"image"`
		Source *RepoSourceOpts `json:"source"`
		Rules  RepoBuildRules  `json:"rules"`
	} `json:"spec"`
}

type RepoImageOpts struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`
	Auth  string `json:"auth"`
}

type RepoSourceOpts struct {
	Url   string `json:"url"`
	Token string `json:"token"`
}

type RepoBuildRules []RepoBuildRule

type RepoBuildRule struct {
	Branch   string              `json:"branch"`
	FilePath string              `json:"filepath"`
	Tag      string              `json:"tag"`
	Config   RepoBuildRuleConfig `json:"config"`
}

type RepoBuildRuleConfig struct {
	Command *string   `json:"command"`
	EnvVars *[]string `json:"environments"`
}

func (s *RepoCreateOptions) DecodeAndValidate(reader io.Reader) *errors.Err {

	log.V(logLevel).Debug("Request: Repo: decode and validate data for creating")

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		log.V(logLevel).Errorf("Request: Repo: decode and validate data for creating err: %s", err)
		return errors.New("repo").Unknown(err)
	}

	err = json.Unmarshal(body, s)
	if err != nil {
		log.V(logLevel).Errorf("Request: Repo: convert struct from json err: %s", err)
		return errors.New("repo").IncorrectJSON(err)
	}

	// Local source does not have url
	if len(s.Spec.Source.Url) == 0 {
		log.V(logLevel).Error("Request: Repo: url parameter is required")
		return errors.New("repo").BadParameter("url")
	}

	url := s.Spec.Source.Url
	rules := s.Spec.Rules

	match := strings.Split(url, "#")

	if !validator.IsGitUrl(match[0]) {
		log.V(logLevel).Error("Request: Repo: parameter url not valid")
		return errors.New("repo").BadParameter("url")
	}

	uniq := make(map[string]bool, len(rules))
	for index := range rules {

		setting := rules[index]

		if setting.Branch == "" {
			return errors.New("repo").BadParameter("branch")
		}

		if setting.Tag == "" {
			return errors.New("repo").BadParameter("tag")
		}

		if _, exists := uniq[setting.Tag]; exists {
			return errors.New("repo").BadParameter("tag")
		}

		uniq[setting.Tag] = true

		if rules[index].Config.EnvVars == nil {
			emptySlice := make([]string, 0)
			rules[index].Config.EnvVars = &emptySlice
		}

		if rules[index].Config.Command == nil {
			rules[index].Config.Command = new(string)
		}
	}

	if s.Spec.Image.Name == "" {
		log.V(logLevel).Error("Request: Repo: parameter name can not be empty")
		return errors.New("repo").BadParameter("name")
	}

	s.Spec.Image.Name = strings.ToLower(s.Spec.Image.Name)

	if !validator.IsRepoName(s.Spec.Image.Name) {
		return errors.New("repo").BadParameter("name")
	}

	return nil
}

type RepoUpdateOptions struct {
	Meta struct {
		Labels map[string]string `json:"labels"`
	} `json:"meta"`
	Spec struct {
		Rules *RepoBuildRules `json:"rules"`
	} `json:"spec"`
}

func (s *RepoUpdateOptions) DecodeAndValidate(reader io.Reader) *errors.Err {

	log.V(logLevel).Debug("Request: Repo: decode and validate data for updating")

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		log.V(logLevel).Errorf("Request: Repo: decode and validate data for creating err: %s", err)
		return errors.New("repo").Unknown(err)
	}

	err = json.Unmarshal(body, s)
	if err != nil {
		log.V(logLevel).Errorf("Request: Repo: convert struct from json err: %s", err)
		return errors.New("repo").IncorrectJSON(err)
	}

	if s.Spec.Rules != nil {

		rules := *s.Spec.Rules

		uniq := make(map[string]bool, len(rules))
		for index := range rules {

			setting := rules[index]

			if setting.Branch == "" {
				return errors.New("repo").BadParameter("branch")
			}

			if setting.Tag == "" {
				return errors.New("repo").BadParameter("tag")
			}

			if _, exists := uniq[setting.Tag]; exists {
				return errors.New("repo").BadParameter("tag")
			}

			uniq[setting.Tag] = true

			if rules[index].Config.EnvVars == nil {
				emptySlice := make([]string, 0)
				rules[index].Config.EnvVars = &emptySlice
			}

			if rules[index].Config.Command == nil {
				rules[index].Config.Command = new(string)
			}
		}
	}

	return nil
}
