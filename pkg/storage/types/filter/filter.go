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

package filter

func NewFilter() *Filter {
	return new(Filter)
}

type Filter struct{}

func (Filter) Image() *ImageFilter {
	return new(ImageFilter)
}

type ImageFilter struct {
	Owner   *string `db:"owner"`
	Private *bool   `db:"private"`
}

func (Filter) Builder() *BuilderFilter {
	return new(BuilderFilter)
}

type BuilderFilter struct {
	Online *bool `db:"online"`
}

func (Filter) Controller() *ControllerFilter {
	return new(ControllerFilter)
}

type ControllerFilter struct {
	Online *bool `db:"online"`
	Master *bool `db:"master"`
}

func (Filter) Build() *BuildFilter {
	return new(BuildFilter)
}

type BuildFilter struct {
	Active *bool  `db:"state_processing"`
	Page   *int64 `db:"page"`
	Limit  *int64 `db:"limit"`
}
