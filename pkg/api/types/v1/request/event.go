//
// Last.Backend LLC CONFIDENTIAL
// __________________
//
// [2014] - [2019] Last.Backend LLC
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

package request

import (
	"time"
)

type EventOptions struct {
	Builds map[string]BuildEvent `json:"builds"`
}

type BuildEvent struct {
	ID         string      `json:"id"`
	Number     int64       `json:"number"`
	Commit     BuildCommit `json:"commit"`
	Branch     string      `json:"branch"`
	Image      string      `json:"image"`
	ImageSha   string      `json:"image_sha"`
	Source     string      `json:"source"`
	Size       int64       `json:"size"`
	Step       string      `json:"step"`
	Status     string      `json:"status"`
	Message    string      `json:"message"`
	Processing bool        `json:"processing"`
	Done       bool        `json:"done"`
	Error      bool        `json:"error"`
	Canceled   bool        `json:"canceled"`
	Finished   *time.Time  `json:"finished"`
	Started    *time.Time  `json:"started"`
}

type BuildCommit struct {
	Hash     string    `json:"hash"`
	Username string    `json:"username"`
	Message  string    `json:"message"`
	Email    string    `json:"email"`
	Date     time.Time `json:"date"`
}
