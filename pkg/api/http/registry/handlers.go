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

package registry

import (
	"net/http"

	"github.com/lastbackend/registry/pkg/api/envs"
	"github.com/lastbackend/registry/pkg/api/types/v1"
	"github.com/lastbackend/registry/pkg/distribution"
	"github.com/lastbackend/registry/pkg/distribution/errors"
	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/util/url"
	"github.com/spf13/viper"
	"io"
)

const (
	logLevel  = 2
	logPrefix = "api:handler:registry"
)

func RegistryInfoH(w http.ResponseWriter, r *http.Request) {

	log.V(logLevel).Debugf("%s:info:> get registry", logPrefix)

	var rm = distribution.NewRegistryModel(r.Context(), envs.Get().GetStorage())

	ri, err := rm.Get()
	if err != nil {
		log.V(logLevel).Errorf("%s:info:> get registry err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	response, err := v1.View().Registry().New(ri).ToJson()
	if err != nil {
		log.V(logLevel).Errorf("%s:info:> convert struct to json err: %v", logPrefix)
		errors.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(response); err != nil {
		log.V(logLevel).Errorf("%s:info:> write response err: %v", logPrefix)
		return
	}
}

func RegistryAuthH(w http.ResponseWriter, r *http.Request) {

	log.V(logLevel).Debugf("%s:auth:> authentication registry", logPrefix)

	u, err := url.Parse(viper.GetString("auth_server"))
	if err != nil {
		log.V(logLevel).Errorf("%s:info:> auth server url incorrect: %v", logPrefix)
		errors.HTTP.InternalServerError(w)
		return
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), r.Body)
	if err != nil {
		log.V(logLevel).Errorf("%s:auth:> create http request err: %v", logPrefix, err)
		return
	}

	q := req.URL.Query()
	for name, value := range r.URL.Query() {
		q.Add(name, value[0])
	}
	req.URL.RawQuery = q.Encode()

	// Copy incoming request headers
	for name, value := range r.Header {
		req.Header.Set(name, value[0])
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.V(logLevel).Errorf("%s:auth:> calling http query err: %v", logPrefix, err)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		log.V(logLevel).Errorf("%s:auth:> write response err: %v", logPrefix, err)
		return
	}
}
