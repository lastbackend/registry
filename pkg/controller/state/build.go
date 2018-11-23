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
	"sort"
	"sync"
)

type BuildState struct {
	lock   sync.RWMutex
	items  dataSlice
	builds map[string]*types.Build
}

type dataSlice []*types.Build

// Len is part of sort.Interface.
func (d dataSlice) Len() int {
	return len(d)
}

// Swap is part of sort.Interface.
func (d dataSlice) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

// Less is part of sort.Interface. We use count as the value to sort by
func (d dataSlice) Less(i, j int) bool {
	return d[i].Meta.Updated.Before(d[j].Meta.Updated)
}

func (s *BuildState) Set(key string, build *types.Build) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.builds[key] = build
	s.items = append(s.items, build)
}

func (s BuildState) List() []*types.Build {
	sort.Sort(s.items)
	return s.items
}

func (s BuildState) Get(key string) *types.Build {
	if _, ok := s.builds[key]; ok {
		t := s.builds[key]
		return t
	}
	return nil
}

func (s *BuildState) Del(key string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for j, item := range s.items {
		if item.Meta.ID == key {
			s.items = append(s.items[:j], s.items[j+1:]...)
			break
		}
	}

	delete(s.builds, key)
}

func NewBuildState() *BuildState {
	bs := new(BuildState)
	bs.builds = make(map[string]*types.Build, 0)
	bs.items = make(dataSlice, 0)
	return bs
}
