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
	// Repo state
	State RepoState `json:"state"`
	// Repo stats
	Stats RepoStats `json:"stats"`
	// Repo source
	Sources RepoSources `json:"sources"`
	// Repo tags
	RuleList []*RepoBuildRule `json:"rules"`
	// Repo tags
	TagList []*RepoTag `json:"tags"`
	// Repo tags
	TemplateList map[string]*RepoTemplate `json:"templates"`
	// Repo last build
	LastBuild RepoLastBuild `json:"last_build"`
	// Repo readme
	Readme string `json:"readme"`
	// Meta remote
	Remote bool `json:"remote"`
	// Meta deleted
	Private bool `json:"private"`
}

type RepoMeta struct {
	// Meta name
	Name string `json:"name"`
	// Meta description
	Description string `json:"description"`
	// Meta owner
	Owner string `json:"owner"`
	// Meta self link
	SelfLink string `json:"self_link"`
	// Meta technology
	Technology string `json:"technology"`
	// Meta technical
	Technical bool `json:"technical"`
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

type RepoStats struct {
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

type RepoSources struct {
	Hub    string `json:"hub"`
	Owner  string `json:"owner"`
	Name   string `json:"name"`
	Branch string `json:"tag"`
}

type RepoBuildRule struct {
	// Repo rule id
	ID string `json:"id"`
	// Repo rule branch
	Branch string `json:"branch"`
	// Repo rule filepath
	FilePath string `json:"filepath"`
	// Repo rule tag
	Tag string `json:"tag"`
	// Repo rule registry
	Registry string `json:"registry"`
	// Repo rule autobuild
	AutoBuild bool `json:"autobuild"`
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
	Layers struct {
		Count int64 `json:"count"`
		Size  struct {
			Average int64 `json:"average"`
			Max     int64 `json:"max"`
		} `json:"size"`
	} `json:"layers"`
	Build0    RepoBuildView `json:"build_0"`
	Build1    RepoBuildView `json:"build_1"`
	Build2    RepoBuildView `json:"build_2"`
	Build3    RepoBuildView `json:"build_3"`
	Build4    RepoBuildView `json:"build_4"`
	Spec      RepoTagSpec   `json:"spec"`
	Disabled  bool          `json:"disabled"`
	AutoBuild bool          `json:"autobuild"`
	Updated   time.Time     `json:"updated"`
	Created   time.Time     `json:"created"`
}

type RepoTagSpec struct {
	Branch   string   `json:"branch"`
	FilePath string   `json:"filepath"`
	Registry string   `json:"registry"`
	EnvVars  []string `json:"environments"`
}

type RepoTemplate struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Main        bool     `json:"main"`
	Shared      bool     `json:"shared"`
	Memory      int64    `json:"memory"`
	Command     string   `json:"command"`
	Entrypoint  string   `json:"entrypoint"`
	Image       string   `json:"image"`
	EnvVars     []string `json:"env"`
	Ports       []Port   `json:"ports"`
}

type RepoTemplateList []*RepoTemplate

type Port struct {
	Protocol      string `json:"protocol"`
	HostPort      int    `json:"external"`
	ContainerPort int    `json:"internal"`
	Published     bool   `json:"published"`
}

type RepoTagList struct {
	Active   []*RepoTag `json:"active"`
	Inactive []*RepoTag `json:"inactive"`
}

type RepoBuildView struct {
	ID     string `json:"id"`
	Number int64  `json:"number"`
	Status string `json:"status"`
}

type RepoList []*Repo
