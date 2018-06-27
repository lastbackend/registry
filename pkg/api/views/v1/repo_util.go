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

package v1

import (
	"encoding/json"
	"github.com/lastbackend/registry/pkg/distribution/types"
)

type RepoView struct{}

func (rv *RepoView) New(obj *types.Repo) *Repo {
	i := new(Repo)
	i.Meta = rv.ToRepoMeta(&obj.Meta)
	i.State = rv.ToRepoState(&obj.State)
	i.Stats = rv.ToRepoStats(&obj.Stats)
	i.Sources = rv.ToRepoSource(&obj.Sources)
	i.TagList = rv.ToRepoTagList(obj.Tags)
	i.RuleList = rv.ToRepoRuleList(obj.Tags)
	i.LastBuild = rv.ToRepoLastBuild(obj.LastBuild)
	i.Readme = obj.Readme
	i.Remote = obj.Remote
	i.Private = obj.Private
	return i
}

func (obj *Repo) ToJson() ([]byte, error) {
	return json.Marshal(obj)
}

func (rv *RepoView) NewList(obj map[string]*types.Repo) *RepoList {
	if obj == nil {
		return nil
	}

	r := make(RepoList, 0)
	for _, v := range obj {
		r = append(r, rv.New(v))
	}
	return &r
}

func (obj *RepoList) ToJson() ([]byte, error) {
	if obj == nil {
		obj = &RepoList{}
	}
	return json.Marshal(obj)
}

func (rv *RepoView) ToRepoMeta(obj *types.RepoMeta) RepoMeta {
	return RepoMeta{
		Name:        obj.Name,
		Owner:       obj.Owner,
		Technology:  obj.Technology,
		Technical:   obj.Technical,
		Description: obj.Description,
		SelfLink:    obj.SelfLink,
		Created:     obj.Created,
		Updated:     obj.Updated,
	}
}

func (rv *RepoView) ToRepoState(obj *types.RepoState) RepoState {
	return RepoState{
		State:   obj.State,
		Status:  obj.Status,
		Deleted: obj.Deleted,
		Liked:   obj.Liked,
	}
}

func (rv *RepoView) ToRepoStats(obj *types.RepoStats) RepoStats {
	return RepoStats{
		ViewsQuantity:    obj.ViewsQuantity,
		StarsQuantity:    obj.StarsQuantity,
		PullsQuantity:    obj.PullsQuantity,
		BuildsQuantity:   obj.BuildsQuantity,
		ServicesQuantity: obj.ServicesQuantity,
	}
}

func (rv *RepoView) ToRepoSource(obj *types.RepoSources) RepoSources {
	return RepoSources{
		Hub:    obj.Hub,
		Owner:  obj.Owner,
		Name:   obj.Name,
		Branch: obj.Branch,
	}
}

func (rv *RepoView) ToRepoTagList(obj map[string]*types.RepoTag) []*RepoTag {
	tags := make([]*RepoTag, 0)

	for _, item := range obj {
		tag := new(RepoTag)
		tag.Name = item.Name
		tag.Builds.Size = item.Builds.Size
		tag.Builds.Total = item.Builds.Total
		tag.Layers.Count = item.Layers.Count
		tag.Layers.Size.Average = item.Layers.Size.Average
		tag.Layers.Size.Max = item.Layers.Size.Max
		tag.Build0.ID = item.Build0.ID
		tag.Build0.Number = item.Build0.Number
		tag.Build0.Status = item.Build0.Status
		tag.Build1.ID = item.Build1.ID
		tag.Build1.Number = item.Build1.Number
		tag.Build1.Status = item.Build1.Status
		tag.Build2.ID = item.Build2.ID
		tag.Build2.Number = item.Build2.Number
		tag.Build2.Status = item.Build2.Status
		tag.Build3.ID = item.Build3.ID
		tag.Build3.Number = item.Build3.Number
		tag.Build3.Status = item.Build3.Status
		tag.Build4.ID = item.Build4.ID
		tag.Build4.Number = item.Build4.Number
		tag.Build4.Status = item.Build4.Status
		tag.Disabled = item.Disabled
		tag.AutoBuild = item.AutoBuild
		tag.Updated = item.Updated
		tag.Created = item.Created

		tag.Spec.EnvVars = item.Spec.EnvVars
		tag.Spec.Registry = item.Spec.Registry
		tag.Spec.Branch = item.Spec.Branch
		tag.Spec.FilePath = item.Spec.FilePath

		tags = append(tags, tag)
	}
	return tags
}

func (rv *RepoView) ToRepoRuleList(obj map[string]*types.RepoTag) []*RepoBuildRule {
	rules := make([]*RepoBuildRule, 0)

	for _, item := range obj {
		if item.Disabled {
			continue
		}

		rule := new(RepoBuildRule)
		rule.ID = item.ID
		rule.Branch = item.Spec.Branch
		rule.FilePath = item.Spec.FilePath
		rule.Tag = item.Name
		rule.Registry = item.Spec.Registry
		rule.AutoBuild = item.AutoBuild
		rule.Config.EnvVars = item.Spec.EnvVars

		rules = append(rules, rule)
	}
	return rules
}

func (rv *RepoView) ToRepoLastBuild(obj types.RepoLastBuild) RepoLastBuild {
	return RepoLastBuild{
		ID:      obj.ID,
		Tag:     obj.Tag,
		Number:  obj.Number,
		Status:  obj.Status,
		Updated: obj.Updated,
	}
}
