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

package validator

import (
	"github.com/asaskevich/govalidator"
	"regexp"
)

func IsUsername(s string) bool {
	reg, _ := regexp.Compile("[A-Za-z0-9]+(?:[_-][A-Za-z0-9]+)*")
	str := reg.FindStringSubmatch(s)
	if len(str) == 1 && str[0] == s && len(s) >= 4 && len(s) <= 64 {
		return true
	}
	return false
}

func IsPassword(s string) bool {
	return len(s) > 6
}

func IsEmail(s string) bool {
	return govalidator.IsEmail(s)
}

func IsServiceName(s string) bool {
	reg, _ := regexp.Compile("[A-Za-z0-9]+(?:[_-][A-Za-z0-9]+)*")
	str := reg.FindStringSubmatch(s)
	if len(str) == 1 && str[0] == s && len(s) >= 4 && len(s) <= 64 {
		return true
	}
	return false
}

func IsNamespaceName(s string) bool {
	reg, _ := regexp.Compile("^[A-Za-z0-9][A-Za-z0-9\\.]+(?:[_-][A-Za-z0-9\\.]+)*")
	str := reg.FindStringSubmatch(s)
	if len(str) == 1 && str[0] == s && len(s) >= 4 && len(s) <= 64 {
		return true
	}
	return false
}

func IsClusterName(s string) bool {
	reg, _ := regexp.Compile("^([A-Za-z0-9]+(?:[_-][A-Za-z0-9]+)*)\\.([A-Za-z0-9]+(?:[_-][A-Za-z0-9]+)*)$")
	str := reg.FindStringSubmatch(s)
	return len(str) == 3
}

func IsRepoName(s string) bool {
	reg, _ := regexp.Compile("[a-z0-9]+(?:[._-][a-z0-9]+)*")
	str := reg.FindStringSubmatch(s)
	if len(str) == 1 && str[0] == s && len(s) > 0 {
		return true
	}
	return false
}

// Check incoming string on git valid utl
// Ex:
// 	* https://github.com/lastbackend/enterprise.git
// 	* git@github.com:lastbackend/enterprise.git
func IsGitUrl(url string) bool {
	res, err := regexp.MatchString(`^(?:ssh|git|http(?:s)?)(?:@|:\/\/(?:.+@)?)((\w+)\.\w+)(?:\/|:)(.+)(?:\/)(.+)(?:\..+)$`, url)
	if err != nil {
		return false
	}

	return res
}

func IsValueInList(value string, list []string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}
