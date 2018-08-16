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

type BuildManifest struct {
	Meta BuildManifestMeta
	Spec BuildManifestSpec
}

type BuildManifestMeta struct {
	ID string `json:"id"`
}

type BuildManifestSpec struct {
	Source BuildManifestSource `json:"source"`
	Image  BuildManifestImage  `json:"image"`
	Config BuildManifestConfig `json:"config"`
}

type BuildManifestSource struct {
	Url    string `json:"url"`
	Branch string `json:"branch"`
}

type BuildManifestImage struct {
	Host  string `json:"host"`
	Name  string `json:"name"`
	Owner string `json:"owner"`
	Tag   string `json:"tag"`
	Auth  string `json:"auth"`
}

type BuildManifestConfig struct {
	Dockerfile string   `json:"dockerfile"`
	Context    string   `json:"context"`
	Workdir    string   `json:"workdir"`
	EnvVars    []string `json:"env"`
	Command    string   `json:"command"`
}

type Builder struct {
	Meta   BuilderMeta   `json:"meta"`
	Status BuilderStatus `json:"status"`
	Spec   BuilderSpec   `json:"spec"`
}

type BuilderMeta struct {
	ID       string    `json:"id"`
	Hostname string    `json:"hostname"`
	Created  time.Time `json:"created"`
	Updated  time.Time `json:"updated"`
}

type BuilderStatus struct {
	Insecure bool `json:"insecure"`
	Online   bool `json:"online"`
	TLS      bool `json:"tls"`
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
	CA         []byte `json:"ca"`
	ClientCert []byte `json:"cert"`
	ClientKey  []byte `json:"key"`
}

type BuilderList []*Builder
