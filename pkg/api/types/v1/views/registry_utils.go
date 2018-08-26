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

import (
	"encoding/json"

	"github.com/lastbackend/registry/pkg/distribution/types"
)

type RegistryView struct{}

func (rv *RegistryView) New(obj *types.Registry) *Registry {
	r := Registry{}
	r.Meta = rv.ToRegistryMeta(obj.Meta)
	r.Status = rv.ToRegistryStatus(obj.Status)
	return &r
}

func (rl *Registry) ToJson() ([]byte, error) {
	return json.Marshal(rl)
}

func (rl *RegistryList) ToJson() ([]byte, error) {
	if rl == nil {
		rl = &RegistryList{}
	}
	return json.Marshal(rl)
}

func (rv *RegistryView) ToRegistryMeta(meta types.RegistryMeta) RegistryMeta {
	return RegistryMeta{}
}

func (rv *RegistryView) ToRegistryStatus(status types.RegistryStatus) RegistryStatus {
	return RegistryStatus{
		TLS: status.TLS,
	}
}

func (RegistryView) NewToken(token string) *RegistryToken {
	return &RegistryToken{token}
}

func (obj *RegistryToken) ToJson() ([]byte, error) {
	return json.Marshal(obj)
}
