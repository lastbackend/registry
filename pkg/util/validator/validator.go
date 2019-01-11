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

package validator

import (
	"github.com/asaskevich/govalidator"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

func IsNil(a interface{}) bool {
	defer func() { recover() }()
	return a == nil || reflect.ValueOf(a).IsNil()
}

func IsBool(s string) bool {
	s = strings.ToLower(s)
	if s == "true" || s == "1" || s == "t" || s == "false" || s == "0" || s == "f" {
		return true
	}
	return false
}

func IsIP(ip string) bool {
	return govalidator.IsIP(ip)
}

func IsUUID(uuid string) bool {
	return govalidator.IsUUIDv4(uuid)
}

func IsPort(port int) bool {
	return govalidator.IsPort(strconv.Itoa(port))
}

// Check incoming string on git valid utl
// Ex:
// 	* https://github.com/lastbackend/registry.git
// 	* git@github.com:lastbackend/enterprise.git
func IsGitUrl(url string) bool {
	res, err := regexp.MatchString(`^(?:ssh|git|http(?:s)?)(?:@|:\/\/(?:.+@)?)((\w+)\.\w+)(?:\/|:)(.+)(?:\/)(.+)(?:\..+)$`, url)
	if err != nil {
		return false
	}

	return res
}

func IsImageName(s string) bool {
	reg, _ := regexp.Compile("[a-z0-9]+(?:[._-][a-z0-9]+)*")
	str := reg.FindStringSubmatch(s)
	if len(str) == 1 && str[0] == s && len(s) > 0 {
		return true
	}
	return false
}

func IsValueInList(value string, list []string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}
