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
	Insecure bool   `json:"insecure"`
	Online   bool   `json:"online"`
	TLS      bool   `json:"tls"`
	Error    string `json:"error"`
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
	Hostname *string `json:"hostname"`
	IP       *string `json:"ip"`
	Port     *uint16 `json:"port"`
	Online   *bool   `json:"online"`
	TLS      *bool   `json:"tls"`
	SSL      *SSL    `json:"ssl"`
}
