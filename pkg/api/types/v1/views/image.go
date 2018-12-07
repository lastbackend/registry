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

type Image struct {
	Meta    ImageMeta    `json:"meta"`
	Status  *ImageStatus `json:"status,omitempty"`
	TagList *ImageTags   `json:"tags,omitempty"`
}

type ImageList []*Image

type ImageMeta struct {
	Owner       string            `json:"owner"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Labels      map[string]string `json:"labels"`
	Created     time.Time         `json:"created"`
	Updated     time.Time         `json:"updated"`
}

type ImageStatus struct {
	Stats   ImageStats `json:"stats"`
	Private bool       `json:"private"`
}

type ImageTag struct {
	Meta   ImageTagMeta   `json:"meta"`
	Status ImageTagStatus `json:"status"`
	Spec   ImageTagSpec   `json:"spec"`
}

type ImageTags []*ImageTag

type ImageTagMeta struct {
	Name string `json:"name"`
}

type ImageTagStatus struct {
	Disabled bool `json:"disabled"`
}

type ImageTagSpec struct {
	Branch     string   `json:"branch"`
	DockerFile string   `json:"dockerfile"`
	Context    string   `json:"context"`
	Command    string   `json:"command"`
	EnvVars    []string `json:"environments"`
	Workdir    string   `json:"workdir"`
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
