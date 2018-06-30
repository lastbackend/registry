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
	"regexp"
	"strconv"
)

type sources struct {
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