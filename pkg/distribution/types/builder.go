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

package types

const (
	DEFAULT_MIN_WORKERS        = 1
	DEFAULT_MIN_WORKER_CPU     = 50000000 // 0.5 vCPU
	DEFAULT_MIN_WORKER_Storage = 5
	DEFAULT_MIN_WORKER_MEMORY  = 256 * 1024 * 1024
)

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
	// Builder Usage
	Usage BuilderResources `json:"usage"`
}

type BuilderSpec struct {
	Network BuilderSpecNetwork `json:"network"`
}

type BuilderSpecNetwork struct {
	IP   string `json:"ip"`
	Port uint16 `json:"port"`
	TLS  bool   `json:"tls"`
	SSL  *SSL   `json:"ssl"`
}

type SSL struct {
	CA   []byte `json:"ca"`
	Cert []byte `json:"cert"`
	Key  []byte `json:"key"`
}

type BuilderResources struct {
	// Builder total containers
	Workers int `json:"workers"`
	// Builder total memory
	RAM int64 `json:"ram"`
	// Builder total cpu
	CPU int64 `json:"cpu"`
	// Builder storage
	Storage int64 `json:"storage"`
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
	Allocated *BuilderResources `json:"allocated"`
	Capacity  *BuilderResources `json:"capacity"`
	Usage     *BuilderResources `json:"usage"`
}