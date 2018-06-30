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

type Build struct {
	Meta    BuildMeta   `json:"meta"`
	Repo    BuildRepo   `json:"repo"`
	State   BuildState  `json:"state"`
	Stats   BuildStats  `json:"stats"`
	Sources BuildSource `json:"sources"`
	Image   BuildImage  `json:"image"`
}

type BuildMeta struct {
	Number   int64     `json:"number"`
	SelfLink string    `json:"self_link"`
	Created  time.Time `json:"created"`
	Updated  time.Time `json:"updated"`
}

type BuildRepo string

type BuildState struct {
	Step       string     `json:"step"`
	Message    string     `json:"message"`
	Status     string     `json:"status"`
	Done       bool       `json:"done"`
	Processing bool       `json:"processing"`
	Canceled   bool       `json:"canceled"`
	Error      bool       `json:"error"`
	Finished   *time.Time `json:"finished"`
	Started    *time.Time `json:"started"`
}

type BuildStats struct {
	Size int64 `json:"size"`
}

type BuildSource struct {
	Hub    string `json:"hub"`
	Owner  string `json:"owner"`
	Name   string `json:"name"`
	Branch string `json:"branch"`
	Commit struct {
		Hash     string    `json:"hash"`
		Username string    `json:"username"`
		Message  string    `json:"message"`
		Email    string    `json:"email"`
		Date     time.Time `json:"date"`
	} `json:"commit"`
}

type BuildImage struct {
	Hub   string `json:"hub"`
	Name  string `json:"name"`
	Owner string `json:"owner"`
	Tag   string `json:"tag"`
	Hash  string `json:"hash"`
}

type BuildList []*Build
