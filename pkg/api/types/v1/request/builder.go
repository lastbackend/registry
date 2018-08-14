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
	IP   *string `json:"ip"`
	Port *uint16 `json:"port"`
}

type BuilderConnectOptions struct {
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

type BuilderStatusUpdateOptions struct {
}

type BuilderCreateManifestOptions struct {
	PID string `json:"pid"`
}
