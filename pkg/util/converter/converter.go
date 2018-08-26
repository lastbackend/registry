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
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type source struct {
	Resource string
	Hub      string
	Name     string
	Owner    string
	Vendor   string
	Branch   string
}

func IntToString(i int) string {
	return strconv.Itoa(i)
}

func ParseBool(str string) (bool, error) {
	switch str {
	case "":
		return false, nil
	case "1", "t", "T", "true", "TRUE", "True":
		return true, nil
	case "0", "f", "F", "false", "FALSE", "False":
		return false, nil
	}
	return false, errors.New(fmt.Sprintf("parse bool string: %s", str))
}

// Parse incoming string git url in source type
// Ex:
// 	* https://github.com/registry/registry.git
// 	* git@github.com:lastbackend/registry.git
func GitUrlParse(url string) (*source, error) {

	var match []string = regexp.MustCompile(`^(?:ssh|git|http(?:s)?)(?:@|:\/\/(?:.+@)?)((\w+)\.\w+)(?:\/|:)(.+)(?:\/)(.+)(?:\..+)$`).FindStringSubmatch(url)

	if len(match) < 5 {
		return nil, errors.New("can't parse url")
	}

	return &source{
		Resource: match[0],
		Hub:      match[1],
		Vendor:   match[2],
		Owner:    match[3],
		Name:     match[4],
		Branch:   "master",
	}, nil

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