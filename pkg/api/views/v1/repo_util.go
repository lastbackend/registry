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
	i.Sources = rv.ToRepoSource(&obj.Sources)
	i.TagList = rv.ToRepoTagList(obj.Tags)
	i.RuleList = rv.ToRepoRuleList(obj.Tags)
	i.Readme = obj.Readme
	i.Remote = obj.Remote
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
		Labels:      obj.Labels,
		Description: obj.Description,
		SelfLink:    obj.SelfLink,
		Created:     obj.Created,
		Updated:     obj.Updated,
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

		if item.Build0 != nil {
			tag.Build0 = new(RepoBuildView)
			tag.Build0.ID = item.Build0.ID
			tag.Build0.Number = item.Build0.Number
			tag.Build0.Status = item.Build0.Status
		}

		if item.Build1 != nil {
			tag.Build1 = new(RepoBuildView)
			tag.Build1.ID = item.Build1.ID
			tag.Build1.Number = item.Build1.Number
			tag.Build1.Status = item.Build1.Status
		}

		if item.Build2 != nil {
			tag.Build2 = new(RepoBuildView)
			tag.Build2.ID = item.Build2.ID
			tag.Build2.Number = item.Build2.Number
			tag.Build2.Status = item.Build2.Status
		}

		if item.Build3 != nil {
			tag.Build3 = new(RepoBuildView)
			tag.Build3.ID = item.Build3.ID
			tag.Build3.Number = item.Build3.Number
			tag.Build3.Status = item.Build3.Status
		}

		if item.Build4 != nil {
			tag.Build4 = new(RepoBuildView)
			tag.Build4.ID = item.Build4.ID
			tag.Build4.Number = item.Build4.Number
			tag.Build4.Status = item.Build4.Status
		}

		tag.Disabled = item.Disabled
		tag.Updated = item.Updated
		tag.Created = item.Created

		tag.Spec.EnvVars = item.Spec.EnvVars
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
		rule.Branch = item.Spec.Branch
		rule.FilePath = item.Spec.FilePath
		rule.Tag = item.Name
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
