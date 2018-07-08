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

type ImageCreateOptions struct {
	Name        string `json:"name"`
	Owner       string `json:"owner"`
	Description string `json:"description"`
	Private     bool   `json:"private"`
}

type ImageUpdateOptions struct {
	Description *string `json:"description"`
	Private     *bool   `json:"private"`
}

type ImageRemoveOptions struct {
	Force *bool `json:"force"`
}

type ImageSource struct {
	Hub    string `json:"hub"`
	Owner  string `json:"owner"`
	Name   string `json:"name"`
	Branch string `json:"branch"`
	Token  string `json:"token"`
}
