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

type Registry struct {
	Meta   RegistryMeta   `json:"meta"`
	Status RegistryStatus `json:"status"`
}

type RegistryMeta struct {
	ImageHub string `json:"image_hub"`
}

type RegistryStatus struct {
	// Registry state secure
	TLS bool `json:"tls"`
	// Registry state capacity
	Capacity RegistryResources `json:"capacity"`
	// Registry state allocated
	Allocated RegistryResources `json:"allocated"`
}

type RegistryList []*Registry

type RegistryToken struct {
	Token string `json:"token"`
}

type RegistryResources struct {
	// Registry total builders
	Builders int `json:"builders"`
	// Registry total memory
	Memory int64 `json:"memory"`
	// Registry total cpu
	Cpu int `json:"cpu"`
	// Registry storage
	Storage int `json:"storage"`
}
