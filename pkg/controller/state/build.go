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

package state

import (
	"github.com/lastbackend/registry/pkg/distribution/types"
	"sync"
)

type BuildState struct {
	lock   sync.RWMutex
	builds map[string]types.Build
}

func (s *BuildState) Set(key string, build *types.Build) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.builds[key] = *build
}

func (s BuildState) List() map[string]types.Build {
	return s.builds
}

func (s BuildState) Get(key string) *types.Build {
	if _, ok := s.builds[key]; ok {
		t := s.builds[key]
		return &t
	}
	return nil
}

func (s *BuildState) Del(key string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.builds, key)
}

func NewBuildState() *BuildState {
	bs := new(BuildState)
	bs.builds = make(map[string]types.Build, 0)
	return bs
}
