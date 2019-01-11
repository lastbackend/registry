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

package url

import (
	"errors"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"net/url"
	"regexp"
	"strings"
)

var (
	// RFC 1035.
	domainRegexp = regexp.MustCompile(`^([a-zA-Z0-9-]{1,63}\.)+[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9]$`)
	ipv4Regexp   = regexp.MustCompile(`^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$`)
	ipv6Regexp   = regexp.MustCompile(`^\[[a-fA-F0-9:]+\]$`)
	urlRegexp    = regexp.MustCompile(`([\w.]+)[\:\/](\w+)\/([\w_-]+)[.git]*[\#\:]?([\w_-]+)?`)
)

func Parse(rawURL string) (*url.URL, error) {
	if strings.Index(rawURL, "//") == 0 {
		rawURL = "http:" + rawURL
	}
	if strings.Index(rawURL, "://") == -1 {
		rawURL = "http://" + rawURL
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	if err := check(strings.Split(u.Host, ":")[0]); err != nil {
		return nil, err
	}

	u.Host = strings.ToLower(u.Host)
	u.Scheme = strings.ToLower(u.Scheme)

	return u, nil
}

func Decode(url string) *types.Source {
	var src = new(types.Source)
	var namespace string

	s := strings.Split(url, "@")

	if len(s) == 0 {
		return src
	}

	if len(s) == 1 {
		namespace = s[0]
	}

	if len(s) == 2 {
		namespace = s[1]
	}

	parts := urlRegexp.FindStringSubmatch(namespace)

	if len(parts) <= 1 {
		return src
	}

	src.Hub = parts[1]
	src.Owner = parts[2]
	src.Name = parts[3]
	src.Branch = parts[4]

	return src
}

func check(host string) error {
	if host == "" {
		return errors.New("host is empty")
	}

	host = strings.ToLower(host)
	if domainRegexp.MatchString(host) || host == "localhost" {
		return nil
	}
	if ipv4Regexp.MatchString(host) || ipv6Regexp.MatchString(host) {
		return nil
	}

	return errors.New("invalid host")
}
