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

type BuildManifest struct {
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
	Workdir    string   `json:"workdir"`
	EnvVars    []string `json:"env"`
	Command    string   `json:"command"`
}
