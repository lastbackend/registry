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

import "time"

const (
	ImageDefaultTag             = "latest"
	ImageDefaultDockerfilePath  = "./DockerFile"
	ImageDefaultContextLocation = "/"
)

type Image struct {
	Meta    ImageMeta            `json:"meta"`
	Status  ImageStatus          `json:"status"`
	TagList map[string]*ImageTag `json:"tags"`
}

type ImageMeta struct {
	Meta
	Owner string `json:"owner"`
}

type ImageStatus struct {
	Stats   ImageStats `json:"stats"`
	Private bool       `json:"private"`
}

type ImageSource struct {
	Hub   string `json:"hub"`
	Owner string `json:"owner"`
	Name  string `json:"name"`
}

type ImageStats struct {
	// Image stats pulls quantity
	PullsQuantity int64 `json:"pulls_quantity"`
	// Image stats builds quantity
	BuildsQuantity int64 `json:"builds_quantity"`
	// Image stats stars quantity
	StarsQuantity int64 `json:"stars_quantity"`
	// Image stats views quantity
	ViewsQuantity int64 `json:"views_quantity"`
}

type ImageTag struct {
	ID       string       `json:"id"`
	ImageID  string       `json:"image"`
	Name     string       `json:"name"`
	Tag      string       `json:"tag"`
	Disabled bool         `json:"disable"`
	Private  bool         `json:"private"`
	Spec     ImageTagSpec `json:"spec"`
	Created  time.Time    `json:"created"`
	Updated  time.Time    `json:"updated"`
}

type ImageTagSpec struct {
	DockerFile string   `json:"dockerfile"`
	Context    string   `json:"context"`
	Command    string   `json:"command"`
	Workdir    string   `json:"workdir"`
	EnvVars    []string `json:"environments"`
}

// *********************************************
// Image distribution options
// *********************************************

type ImageCreateOptions struct {
	Name        string `json:"name"`
	Owner       string `json:"owner"`
	Description string `json:"description"`
	Private     bool   `json:"tag"`
}

type ImageUpdateOptions struct {
	Description *string `json:"description"`
	Private     *bool   `json:"private"`
}

type ImageRemoveOptions struct {
	Force *bool `json:"force"`
}

type ImageTagCreateOptions struct {
	Name string `json:"name"`
}

type ImageListOptions struct {
	Owner string `json:"owner"`
}
