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

package request

type BuilderUpdateOptions struct {
	Workers      uint   `json:"workers"`
	WorkerMemory uint64 `json:"worker_memory"`
	WorkerLimit  bool   `json:"worker_limit"`
}

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
}

type BuilderResources struct {
	// Builder total containers
	Workers uint `json:"workers"`
	// Builder total memory
	Memory uint64 `json:"memory"`
	// Builder total cpu
	Cpu uint `json:"cpu"`
	// Builder storage
	Storage uint64 `json:"storage"`
}

type BuilderCreateManifestOptions struct {
	PID string `json:"pid"`
}
