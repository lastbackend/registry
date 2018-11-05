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

package views

import "time"

type Build struct {
	Meta   *BuildMeta   `json:"meta"`
	Status *BuildStatus `json:"status"`
	Spec   *BuildSpec   `json:"spec"`
}

type BuildList struct {
	Total int64    `json:"total"`
	Page  int64    `json:"page"`
	Limit int64    `json:"limit"`
	Items []*Build `json:"items"`
}

type BuildMeta struct {
	ID      string            `json:"id"`
	Labels  map[string]string `json:"labels"`
	Number  int64             `json:"number"`
	Updated time.Time         `json:"updated"`
	Created time.Time         `json:"created"`
}

type BuildStatus struct {
	Size       int64      `json:"size"`
	Step       string     `json:"step"`
	Status     string     `json:"status"`
	Message    string     `json:"message"`
	Processing bool       `json:"processing"`
	Done       bool       `json:"done"`
	Error      bool       `json:"error"`
	Canceled   bool       `json:"canceled"`
	Finished   *time.Time `json:"finished"`
	Started    *time.Time `json:"started"`
}

type BuildSpec struct {
	Image  BuildImage  `json:"image"`
	Source BuildSource `json:"source"`
	Config BuildConfig `json:"config"`
}

type BuildSource struct {
	Hub    string       `json:"hub"`
	Owner  string       `json:"owner"`
	Name   string       `json:"name"`
	Branch string       `json:"branch"`
	Commit *BuildCommit `json:"commit,omitempty"`
}

type BuildImage struct {
	Name string `json:"name"`
	Tag  string `json:"tag"`
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
