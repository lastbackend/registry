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

type BuilderConnectOptions struct {
	IP       string           `json:"ip"`
	Port     uint16           `json:"port"`
	TLS      bool             `json:"tls"`
	SSL      *SSL             `json:"ssl"`
	System   SystemInfo       `json:"system"`
	Resource BuilderResources `json:"resource"`
}

type SystemInfo struct {
	Version      string
	Architecture string
	OSName       string
	OSType       string
}

type SSL struct {
	CA   []byte `json:"ca"`
	Cert []byte `json:"cert"`
	Key  []byte `json:"key"`
}

type BuilderStatusUpdateOptions struct {
	Allocated BuilderResources `json:"allocated"`
	Capacity  BuilderResources `json:"capacity"`
	Usage     BuilderResources `json:"usage"`
}

type BuilderResources struct {
	// Builder total containers
	Workers int `json:"workers"`
	// Builder total memory
	RAM int64 `json:"memory"`
	// Builder total cpu
	CPU int64 `json:"cpu"`
	// Builder storage
	Storage int64 `json:"storage"`
}

type BuilderCreateManifestOptions struct {
	PID string `json:"pid"`
}
