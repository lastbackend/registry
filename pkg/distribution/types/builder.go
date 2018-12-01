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
	BuilderInfo
}

type BuilderInfo struct {
	Hostname     string `json:"hostname"`
	Version      string `json:"version"`
	Architecture string `json:"architecture"`
	OSName       string `json:"os_name"`
	OSType       string `json:"os_type"`
}

type BuilderStatus struct {
	Insecure bool `json:"insecure"`
	Online   bool `json:"online"`
	TLS      bool `json:"tls"`
	// Builder Capacity
	Capacity BuilderResources `json:"capacity"`
	// Builder Allocated
	Allocated BuilderResources `json:"allocated"`
}

type BuilderSpec struct {
	Network BuilderSpecNetwork `json:"network"`
	Limits  BuilderSpecLimits  `json:"limits"`
}

type BuilderSpecNetwork struct {
	IP   string `json:"ip"`
	Port uint16 `json:"port"`
	TLS  bool   `json:"tls"`
	SSL  *SSL   `json:"ssl"`
}

type BuilderSpecLimits struct {
	WorkerLimit  bool  `json:"worker_limit"`
	Workers      int   `json:"workers"`
	WorkerMemory int64 `json:"worker_memory"`
}

type SSL struct {
	CA   []byte `json:"ca"`
	Cert []byte `json:"cert"`
	Key  []byte `json:"key"`
}

type BuilderResources struct {
	// Builder total containers
	Workers uint `json:"workers"`
	// Builder total memory
	Memory uint64 `json:"memory"`
	// Builder total cpu
	Cpu uint64 `json:"cpu"`
	// Builder storage
	Storage uint64 `json:"storage"`
}

// *********************************************
// Builder distribution options
// *********************************************

type BuilderCreateOptions struct {
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
	Port     uint16 `json:"port"`
	Online   bool   `json:"online"`
	TLS      bool   `json:"tls"`
	SSL      *SSL   `json:"ssl"`
}

type BuilderUpdateOptions struct {
	Hostname  *string           `json:"hostname"`
	IP        *string           `json:"ip"`
	Port      *uint16           `json:"port"`
	Online    *bool             `json:"online"`
	TLS       *bool             `json:"tls"`
	SSL       *SSL              `json:"ssl"`
	Limits    *BuilderLimits    `json:"limits"`
	Allocated *BuilderResources `json:"allocated"`
	Capacity  *BuilderResources `json:"capacity"`
}

type BuilderLimits struct {
	WorkerLimit  bool  `json:"worker_limit"`
	Workers      int   `json:"workers"`
	WorkerMemory int64 `json:"worker_memory"`
}
