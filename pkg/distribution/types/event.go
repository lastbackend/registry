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

import (
	"encoding/json"
)

const (
	StorageInsertAction = "insert"
	StorageUpdateAction = "update"
	StorageDeleteAction = "delete"
)

type StorageEvent struct {
	Channel   string `json:"channel"`
	Operation string `json:"operation"`
	Entity    string `json:"entity"`
}

type Event struct {
	Name    string          `json:"name"`
	Payload json.RawMessage `json:"payload"`
}
