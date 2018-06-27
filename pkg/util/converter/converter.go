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

package converter

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"github.com/lastbackend/registry/pkg/distribution/types"
	lbviews "github.com/lastbackend/lastbackend/pkg/api/types/v1/views"
	lbtypes "github.com/lastbackend/lastbackend/pkg/distribution/types"
)

type sources struct {
	Resource string
	Hub      string
	Name     string
	Owner    string
	Vendor   string
	Branch   string
}

func StringToInt64(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

func StringToInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func IntToString(i int) string {
	return strconv.Itoa(i)
}

func StringToBool(s string) bool {
	s = strings.ToLower(s)
	if s == "true" || s == "1" || s == "t" {
		return true
	}
	return false
}

func Int64ToInt(i int64) int {
	return StringToInt(strconv.FormatInt(i, 10))
}

func DecodeBase64(s string) string {
	buf, _ := base64.StdEncoding.DecodeString(s)
	return string(buf)
}

// Parse incoming string git url in sources type
// Ex:
// 	* https://github.com/lastbackend/enterprise.git
// 	* git@github.com:lastbackend/enterprise.git
func GitUrlParse(url string) (*sources, error) {

	var match []string = regexp.MustCompile(`^(?:ssh|git|http(?:s)?)(?:@|:\/\/(?:.+@)?)((\w+)\.\w+)(?:\/|:)(.+)(?:\/)(.+)(?:\..+)$`).FindStringSubmatch(url)

	if len(match) < 5 {
		return nil, errors.New("can't parse url")
	}

	return &sources{
		Resource: match[0],
		Hub:      match[1],
		Vendor:   match[2],
		Owner:    match[3],
		Name:     match[4],
		Branch:   "master",
	}, nil

}

type image struct {
	Hub   string
	Owner string
	Name  string
	Tag   string
}

func DockerNamespaceMake(hub, owner, repo, tag string) string {
	var ns = repo
	if tag != "" {
		ns = strings.Join([]string{ns, tag}, ":")
	}
	if owner != "" {
		ns = strings.Join([]string{owner, ns}, "/")
	}
	if hub != "" {
		ns = strings.Join([]string{hub, ns}, "/")
	}
	return ns
}

func DockerNamespaceParse(namespace string) *image {

	var (
		i     *image
		match = strings.Split(namespace, ":")
	)

	if len(match) != 0 {
		i = new(image)
	}

	if len(match) == 2 {
		i.Tag = match[1]
	} else {
		i.Tag = "latest"
	}

	match = strings.Split(match[0], "/")
	switch len(match) {
	case 1:
		i.Name = match[0]
	case 2:
		i.Owner = match[0]
		i.Name = match[1]
	case 3:
		i.Hub = match[0]
		i.Owner = match[1]
		i.Name = match[2]
	}
	return i
}

func EnforcePtr(obj interface{}) (reflect.Value, error) {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr {
		if v.Kind() == reflect.Invalid {
			return reflect.Value{}, fmt.Errorf("expected pointer, but got invalid kind")
		}
		return reflect.Value{}, fmt.Errorf("expected pointer, but got %v type", v.Type())
	}
	if v.IsNil() {
		return reflect.Value{}, fmt.Errorf("expected pointer, but got nil")
	}
	return v.Elem(), nil
}

func UnmarshalStringJSON(b []byte) ([]string, error) {
	if len(b) == 0 {
		return make([]string, 0), nil
	}

	p := make([]string, 0, 1)
	if err := json.Unmarshal(b, &p); err != nil {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return nil, err
		}
		p = append(p, s)
	}

	return p, nil
}

func ConvertLbNamespaceViewToCreateNamespaceOpts(cluster *types.Cluster, namespace *lbviews.Namespace) *types.NamespaceCreateOptions {

	opts := new(types.NamespaceCreateOptions)
	opts.Description = namespace.Meta.Description
	opts.Cluster = &cluster.Meta.ID
	opts.Name = strings.ToLower(namespace.Meta.Name)
	opts.SelfLink = &namespace.Meta.SelfLink

	opts.Quotas = new(lbtypes.NamespaceQuotasOptions)
	opts.Quotas.Routes = namespace.Spec.Quotas.Routes
	opts.Quotas.RAM = namespace.Spec.Quotas.RAM
	opts.Quotas.Disabled = namespace.Spec.Quotas.Disabled

	return opts
}

func ConvertLbNamespaceViewToUpdateNamespaceOpts(namespace *lbviews.Namespace) *types.NamespaceUpdateOptions {

	opts := new(types.NamespaceUpdateOptions)
	opts.Description = &namespace.Meta.Description

	opts.Quotas = new(lbtypes.NamespaceQuotasOptions)
	opts.Quotas.Routes = namespace.Spec.Quotas.Routes
	opts.Quotas.RAM = namespace.Spec.Quotas.RAM
	opts.Quotas.Disabled = namespace.Spec.Quotas.Disabled

	return opts
}