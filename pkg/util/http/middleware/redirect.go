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

package middleware

import (
	"fmt"
	"net/http"
)

func RedirectToHTTPSRouter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		proto := req.Header.Get("x-forwarded-proto")
		if proto == "http" || proto == "HTTP" {
			http.Redirect(res, req, fmt.Sprintf("https://%s%s", req.Host, req.URL), http.StatusPermanentRedirect)
			return
		}

		next.ServeHTTP(res, req)

	})
}
