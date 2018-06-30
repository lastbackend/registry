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

package v1

import "time"

type Repo struct {
	Meta RepoMeta `json:"meta"`
	// Repo source
	Sources RepoSources `json:"sources"`
	// Repo tags
	RuleList []*RepoBuildRule `json:"rules"`
	// Repo tags
	TagList []*RepoTag `json:"tags"`
	// Repo readme
	Readme string `json:"readme"`
	// Meta remote
	Remote bool `json:"remote"`
}

type RepoMeta struct {
	// Meta name
	Name string `json:"name"`
	// Meta description
	Description string `json:"description"`
	// Meta labels
	Labels map[string]string `json:"labels"`
	// Meta owner
	Owner string `json:"owner"`
	// Meta self link
	SelfLink string `json:"self_link"`
	// Meta created
	Created time.Time `json:"created"`
	// Meta updated
	Updated time.Time `json:"updated"`
}

type RepoState struct {
	// Repo state
	State string `json:"state"`
	// Repo status
	Status string `json:"status"`
	// Meta deleted
	Deleted bool `json:"deleted"`
	// Meta liked
	Liked bool `json:"liked"`
}

type RepoLastBuild struct {
	ID      string    `json:"id"`
	Tag     string    `json:"tag"`
	Status  string    `json:"status"`
	Number  int       `json:"number"`
	Updated time.Time `json:"updated"`
}

type RepoSources struct {
	Hub    string `json:"hub"`
	Owner  string `json:"owner"`
	Name   string `json:"name"`
	Branch string `json:"branch"`
}

type RepoBuildRule struct {
	// Repo rule branch
	Branch string `json:"branch"`
	// Repo rule filepath
	FilePath string `json:"filepath"`
	// Repo rule tag
	Tag string `json:"tag"`
	// Repo rule config
	Config struct {
		// Repo rule config environments
		EnvVars []string `json:"env"`
	} `json:"configs"`
}

type RepoTag struct {
	Name   string `json:"name"`
	Builds struct {
		Size  int64 `json:"size"`
		Total int64 `json:"total"`
	} `json:"builds"`
	Build0   *RepoBuildView `json:"build_0"`
	Build1   *RepoBuildView `json:"build_1"`
	Build2   *RepoBuildView `json:"build_2"`
	Build3   *RepoBuildView `json:"build_3"`
	Build4   *RepoBuildView `json:"build_4"`
	Spec     RepoTagSpec    `json:"spec"`
	Disabled bool           `json:"disabled"`
	Updated  time.Time      `json:"updated"`
	Created  time.Time      `json:"created"`
}

type RepoTagSpec struct {
	Branch   string   `json:"branch"`
	FilePath string   `json:"filepath"`
	EnvVars  []string `json:"environments"`
}

type RepoBuildView struct {
	ID     string `json:"id"`
	Number int64  `json:"number"`
	Status string `json:"status"`
}

type RepoList []*Repo
