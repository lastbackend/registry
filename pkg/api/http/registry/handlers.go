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
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/lastbackend/registry/pkg/api/envs"
	"github.com/lastbackend/registry/pkg/api/types/v1"
	"github.com/lastbackend/registry/pkg/distribution"
	"github.com/lastbackend/registry/pkg/distribution/errors"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/util/url"
	"github.com/spf13/viper"
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

func RegistryUpdateH(w http.ResponseWriter, r *http.Request) {

	log.V(logLevel).Debugf("%s:update:> update registry info", logPrefix)

	var (
		rm = distribution.NewRegistryModel(r.Context(), envs.Get().GetStorage())
		sm = distribution.NewSystemModel(r.Context(), envs.Get().GetStorage())
	)

	// request body struct
	rq := v1.Request().Registry().UpdateOptions()
	if e := rq.DecodeAndValidate(r.Body); e != nil {
		log.V(logLevel).Errorf("%s:update:> validation incoming data err: %v", logPrefix, e)
		errors.New("Invalid incoming data").Unknown().Http(w)
		return
	}

	system, err := sm.Get()
	if err != nil {
		log.V(logLevel).Errorf("%s:update:> update registry err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	opts := new(types.SystemUpdateOptions)
	opts.AccessToken = rq.AccessToken
	opts.AuthServer = rq.AuthServer

	err = sm.Update(system, opts)
	if err != nil {
		log.V(logLevel).Errorf("%s:update:> update registry err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	ri, err := rm.Get()
	if err != nil {
		log.V(logLevel).Errorf("%s:update:> get registry err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	response, err := v1.View().Registry().New(ri).ToJson()
	if err != nil {
		log.V(logLevel).Errorf("%s:update:> convert struct to json err: %v", logPrefix)
		errors.HTTP.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(response); err != nil {
		log.V(logLevel).Errorf("%s:update:> write response err: %v", logPrefix, err)
		return
	}
}

func RegistryAuthH(w http.ResponseWriter, r *http.Request) {

	log.V(logLevel).Debugf("%s:auth:> authentication registry", logPrefix)

	var (
		err   error
		token string
		// The name of the service which hosts the resource.
		service = r.URL.Query().Get("service")
		// Scope parameter as: scope=repo:lastbackend/my-app:push,pull.
		// The scope field may be empty to request a refresh token without providing any resource permissions
		// to the returned bearer token.
		scope   = r.URL.Query()["scope"]
		account = new(types.RegistryUser)
		scopes  = new(types.Scopes)
		rgm     = distribution.NewRegistryModel(r.Context(), envs.Get().GetStorage())
		sm = distribution.NewSystemModel(r.Context(), envs.Get().GetStorage())
	)

	// Checking for service being authenticated.
	if service == "" || service != viper.GetString("service") {
		log.V(logLevel).Errorf("%s:auth:> error checking for service variable existence", logPrefix)
		errors.New("registry").Unauthorized().Http(w)
		return
	}

	if len(r.Header.Get("Authorization")) != 0 {
		match := strings.Split(r.Header.Get("Authorization"), " ")

		if len(match) != 2 {
			err := errors.New("token incorrect")
			log.V(logLevel).Errorf("%s:auth:> parse token err: %v", logPrefix, err)
			errors.New("registry").Unauthorized().Http(w)
			return
		}

		token = match[1]
	}

	if len(scope) == 0 {
		log.V(logLevel).Errorf("%s:auth:> check scope", logPrefix)
		errors.New("registry").Unauthorized().Http(w)
		return
	}

	log.V(logLevel).Debugf("%s:auth:> check scope", logPrefix)

	system, err := sm.Get()
	if err != nil {
		log.V(logLevel).Errorf("%s:update:> update registry err: %v", logPrefix, err)
		errors.HTTP.InternalServerError(w)
		return
	}

	for _, scopeStr := range scope {

		s, err := rgm.ParseScope(scopeStr)
		if err != nil {
			log.V(logLevel).Errorf("%s:auth:> error checking for service variable existence err: %s", logPrefix, err)
			errors.New("registry").Unauthorized().Http(w)
			return
		}

		u, err := url.Parse(system.AuthServer)
		if err != nil {
			log.V(logLevel).Errorf("%s:auth:> auth server url incorrect: %v", logPrefix)
			errors.HTTP.InternalServerError(w)
			return
		}

		req, err := http.NewRequest(http.MethodGet, u.String(), nil)
		if err != nil {
			log.V(logLevel).Errorf("%s:auth:> create http request err: %v", logPrefix, err)
			return
		}

		req.Header.Set("Authorization", r.Header.Get("Authorization"))
		req.Header.Set("X-Registry-Auth", system.AccessToken)

		q := req.URL.Query()

		q.Add("type", "repository")
		q.Add("name", s.Name)
		req.URL.RawQuery = q.Encode()

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.V(logLevel).Errorf("%s:auth:> calling http query err: %v", logPrefix, err)
			errors.HTTP.InternalServerError(w)
			return
		}

		if resp.StatusCode == http.StatusBadGateway {
			log.V(logLevel).Errorf("%s:auth:> bad gateway", logPrefix)
			errors.HTTP.BadGateway(w)
			return
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.V(logLevel).Errorf("%s:auth:> read response body: %v", logPrefix, err)
			errors.HTTP.InternalServerError(w)
			return
		}

		res := make([]string, 0)

		err = json.Unmarshal(body, &res)
		if err != nil {
			log.V(logLevel).Errorf("%s:auth:> parse json: %v", logPrefix)
			errors.New("registry").IncorrectJSON(err)
			return
		}

		resp.Body.Close()

		s.Actions = res
		*scopes = append(*scopes, s)
	}

	if scopes == nil || len(*scopes) == 0 {
		errors.HTTP.Unauthorized(w)
		return
	}

	if len(r.Header.Get("Authorization")) != 0 {
		// Decoding second part (login:password) of auth
		auth, err := base64.StdEncoding.DecodeString(token)
		if err != nil {
			log.V(logLevel).Errorf("%s:auth:> decode token err: %v", logPrefix, err)
			errors.HTTP.InternalServerError(w)
			return
		}

		// Splitting into two part: login and password
		login := strings.Split(string(auth), ":")

		account.Username = login[0]
		account.Password = login[1]
	}

	sign, err := rgm.CreateSignature(account, scopes)
	if err != nil {
		log.V(logLevel).Errorf("%s:auth:> create signature err: %v", logPrefix, err)
		errors.HTTP.Unauthorized(w)
		return
	}

	response, err := v1.View().Registry().NewToken(sign).ToJson()
	if err != nil {
		log.V(logLevel).Errorf("%s:auth:> convert struct to json err: %v", logPrefix, err)
		errors.HTTP.Unauthorized(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(response); err != nil {
		log.V(logLevel).Errorf("%s:auth:> write message response err: %v", logPrefix, err)
		return
	}

}
