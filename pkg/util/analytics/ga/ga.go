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

package ga

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
)

const endpoint = "https://www.google-analytics.com/collect"

type GA struct {
	trackingID string
}

func New(trackingID string) *GA {
	return &GA{trackingID}
}

func (ga *GA) TrackEvent(r *http.Request, clientID, category, action, label string, value *uint) error {
	if category == "" || action == "" {
		return errors.New("analytics: category and action are required")
	}

	v := url.Values{
		"v":   {"1"},
		"tid": {ga.trackingID},
		// Anonymously identifies a particular user. See the parameter guide for
		// details:
		// https://developers.google.com/analytics/devguides/collection/protocol/v1/parameters#cid
		//
		// Depending on your application, this might want to be associated with the
		// user in a cookie.
		"cid": {clientID},
		"t":   {"event"},
		"ec":  {"api:" + category},
		"ea":  {action},
	}

	if label != "" {
		v.Set("el", strings.Title(label))
	}

	if value != nil {
		v.Set("ev", fmt.Sprintf("%d", *value))
	}

	if r != nil {
		if label != "" {
			v.Set("ua", r.UserAgent())
		}

		if remoteIP, _, err := net.SplitHostPort(r.RemoteAddr); err != nil {
			v.Set("uip", remoteIP)
		}
	}

	// NOTE: Google Analytics returns a 200, even if the request is malformed.
	_, err := http.PostForm(endpoint, v)

	return err
}
