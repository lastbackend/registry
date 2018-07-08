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

type Builder struct {
	Meta   BuilderMeta   `json:"meta"`
	Status BuilderStatus `json:"status"`
	Spec   BuilderSpec   `json:"spec"`
}

type BuilderMeta struct {
	Meta
	Hostname string `json:"hostname"`
}

type BuilderStatus struct {
	Online bool `json:"online"`
}

type BuilderSpec struct {
	TaskList []*struct{} `json:"tasks"`
}

// *********************************************
// Builder distribution options
// *********************************************

type BuilderCreateOptions struct {
	Hostname string `json:"hostname"`
	Online   bool   `json:"online"`
}

type BuilderUpdateOptions struct {
	Online *bool        `json:"online"`
	Tasks  *[]*struct{} `json:"tasks"`
}
