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

type BuildTaskExecuteOptions struct {
	ID     string         `json:"id,omitempty"`
	Meta   BuildJobMeta   `json:"meta"`
	Image  BuildJobImage  `json:"image"`
	Config BuildJobConfig `json:"config"`
	Repo   string         `json:"repo"`
	Branch string         `json:"branch"`
	LogUri string         `json:"log_uri"`
}

type BuildJobMeta struct {
	ID      string `json:"id"`
	LogsUri string `json:"logs_uri"`
}

type BuildJobImage struct {
	Host  string `json:"host"`
	Name  string `json:"name"`
	Owner string `json:"owner"`
	Tag   string `json:"tag"`
	Token string `json:"token"`
}

type BuildJobConfig struct {
	Dockerfile string   `json:"dockerfile"`
	Context    string   `json:"context"`
	Workdir    string   `json:"workdir"`
	EnvVars    []string `json:"env"`
}
