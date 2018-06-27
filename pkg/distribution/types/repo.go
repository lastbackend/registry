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
	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/util/validator"
	"github.com/lastbackend/lastbackend/pkg/distribution/errors"
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
	// Repo state
	State RepoState `json:"state"`
	// Repo stats
	Stats RepoStats `json:"stats"`
	// Repo source
	Sources RepoSources `json:"sources"`
	// Repo last build
	LastBuild RepoLastBuild `json:"last_build"`
	// Repo tags
	Tags map[string]*RepoTag `json:"tags"`
	// Repo created time
	Registry string `json:"registry"`
	// Repo meta readme
	Readme string `json:"readme"`
	// Remote repo option
	Remote bool `json:"remote"`
	// Private repo option
	Private bool `json:"private"`
	// Technial repo
	Technical bool `json:"technical"`
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
	// Repo meta user
	AccountID string `json:"account"`
	// Meta technical
	Technical bool `json:"technical"`
	// Repo readme
	Technology string `json:"technology"`
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

type RepoState struct {
	// Repo state
	State string `json:"state"`
	// Repo state status
	Status string `json:"status"`
	// Meta deleted
	Deleted bool `json:"deleted"`
	// Meta liked
	Liked bool `json:"liked"`
	// Last build
	LastBuild struct {
		Status  string    `json:"status"`
		Updated time.Time `json:"updated"`
	} `json:"last_build"`
}

type RepoStats struct {
	// Repo last build
	LastBuild time.Time `json:"last_build"`
	// Repo stats pulls quantity
	PullsQuantity int64 `json:"pulls_quantity"`
	// Repo stats builds quantity
	BuildsQuantity int64 `json:"builds_quantity"`
	// Repo stats services quantity
	ServicesQuantity int64 `json:"services_quantity"`
	// Repo stats stars quantity
	StarsQuantity int64 `json:"stars_quantity"`
	// Repo stats views quantity
	ViewsQuantity int64 `json:"views_quantity"`
}

type RepoTag struct {
	ID     string      `json:"id"`
	RepoID string      `json:"repo"`
	Name   string      `json:"name"`
	Spec   RepoTagSpec `json:"spec"`
	Builds struct {
		Total int64 `json:"total"`
		Size  int64 `json:"size"`
	} `json:"builds"`
	Layers struct {
		Count int64 `json:"count"`
		Size  struct {
			Average int64 `json:"average"`
			Max     int64 `json:"max"`
		} `json:"size"`
	} `json:"layers"`
	Build0    BuildView `json:"build_0"`
	Build1    BuildView `json:"build_1"`
	Build2    BuildView `json:"build_2"`
	Build3    BuildView `json:"build_3"`
	Build4    BuildView `json:"build_4"`
	Disabled  bool      `json:"disabled"`
	AutoBuild bool      `json:"autobuild"`
	Created   time.Time `json:"created"`
	Updated   time.Time `json:"updated"`
}

type RepoTagSpec struct {
	Branch   string   `json:"branch"`
	FilePath string   `json:"filepath"`
	Registry string   `json:"registry"`
	EnvVars  []string `json:"environments"`
}

type BuildView struct {
	ID     string `json:"id"`
	Number int64  `json:"number"`
	Status string `json:"status"`
}

type RepoCreateOptions struct {
	Name       string          `json:"name"`
	Url        *string         `json:"url"`
	Technology string          `json:"technology"`
	Private    bool            `json:"private"`
	Technical  bool            `json:"technical"`
	Rules      []RepoBuildRule `json:"rules"`
}

type RepoBuildRule struct {
	Branch    string              `json:"branch"`
	FilePath  string              `json:"filepath"`
	Tag       string              `json:"tag"`
	Registry  string              `json:"registry"`
	AutoBuild bool                `json:"autobuild"`
	Config    RepoBuildRuleConfig `json:"config"`
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
	if s.Url != nil {

		if *s.Url == "" {
			log.V(logLevel).Error("Request: Repo: url parameter is required")
			return errors.New("repo").BadParameter("url")
		}

		match := strings.Split(*s.Url, "#")

		if !validator.IsGitUrl(match[0]) {
			log.V(logLevel).Error("Request: Repo: parameter url not valid")
			return errors.New("repo").BadParameter("url")
		}

		uniq := make(map[string]bool, len(s.Rules))
		for index := range s.Rules {

			setting := s.Rules[index]

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

			if s.Rules[index].Config.EnvVars == nil {
				emptySlice := make([]string, 0)
				s.Rules[index].Config.EnvVars = &emptySlice
			}

			if s.Rules[index].Config.Command == nil {
				s.Rules[index].Config.Command = new(string)
			}

		}
	}

	if s.Name == "" {
		log.V(logLevel).Error("Request: Repo: parameter name can not be empty")
		return errors.New("repo").BadParameter("name")
	}

	s.Name = strings.ToLower(s.Name)

	if !validator.IsRepoName(s.Name) {
		return errors.New("repo").BadParameter("name")
	}

	return nil
}

type RepoUpdateOptions struct {
	Technology  *string `json:"technology,omitempty"`
	Description *string `json:"description,omitempty"`
	Private     *bool   `json:"private,omitempty"`
	Rules       *[]struct {
		Branch    string `json:"branch"`
		FilePath  string `json:"filepath"`
		Tag       string `json:"tag"`
		Registry  string `json:"registry"`
		AutoBuild bool   `json:"autobuild"`
		Config    struct {
			Command *string   `json:"command"`
			EnvVars *[]string `json:"environments"`
		} `json:"config"`
	} `json:"rules,omitempty"`
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

	if s.Rules != nil {

		uniq := make(map[string]bool, len(*s.Rules))

		for index := range *s.Rules {

			setting := (*s.Rules)[index]

			if setting.Branch == "" {
				return errors.New("repo").BadParameter("branch")
			}

			if setting.Tag == "" {
				return errors.New("repo").BadParameter("tag")
			}

			if _, exists := uniq[setting.Tag]; exists {
				return errors.New("repo").NotUnique("tag")
			}

			uniq[setting.Tag] = true

			if (*s.Rules)[index].Config.EnvVars == nil {
				emptySlice := make([]string, 0)
				(*s.Rules)[index].Config.EnvVars = &emptySlice
			}

			if (*s.Rules)[index].Config.Command == nil {
				(*s.Rules)[index].Config.Command = new(string)
			}
		}
	}

	return nil
}