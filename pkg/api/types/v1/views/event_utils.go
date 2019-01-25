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

package views

import (
	"encoding/json"
	"unsafe"

	"github.com/lastbackend/registry/pkg/distribution/types"
)

type EventView struct{}

func (ev *EventView) New(obj *types.Event) *Event {
	if obj == nil {
		return nil
	}
	e := new(Event)
	return e
}

func (ev *Event) ToJson() ([]byte, error) {
	if unsafe.Sizeof(ev) == 0 {
		return []byte{}, nil
	}
	return json.Marshal(ev)
}