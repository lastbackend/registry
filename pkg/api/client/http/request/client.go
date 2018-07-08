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

package request

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/lastbackend/registry/pkg/api/client/config"
	"github.com/lastbackend/registry/pkg/util/serializer"
	"github.com/pkg/errors"
)

type RESTClient struct {
	base        *url.URL
	serializers serializer.Codec
	Client      *http.Client
	BearerToken string
}

func NewRESTClient(uri string, cfg *config.Config) (*RESTClient, error) {

	if cfg == nil {
		return DefaultRESTClient(uri), nil
	}

	c := new(RESTClient)
	c.base = parseURL(uri, cfg.TLS.Insecure)
	c.Client = new(http.Client)

	c.Client.Timeout = time.Second * cfg.Timeout
	c.Client.Transport = new(http.Transport)

	if err := withTLSClientConfig(cfg)(c.Client); err != nil {
		return nil, err
	}

	return c, nil
}

func DefaultRESTClient(uri string) *RESTClient {
	return &RESTClient{
		base: parseURL(uri, true),
		Client: &http.Client{
			Timeout:   time.Second * 10,
			Transport: new(http.Transport),
		},
	}
}

// WithTLSClientConfig applies a tls config to the client transport.
func withTLSClientConfig(cfg *config.Config) func(*http.Client) error {
	return func(c *http.Client) error {

		tc, err := config.NewTLSConfig(cfg)
		if err != nil {
			return errors.Wrap(err, "failed to create tls config")
		}

		if transport, ok := c.Transport.(*http.Transport); ok {
			transport.TLSClientConfig = tc
			return nil
		}

		return errors.Errorf("cannot apply tls config to transport: %T", c.Transport)
	}
}

func (c *RESTClient) Do(verb string, path string) *Request {
	c.base.Path = path
	req := New(c.Client, verb, c.base)
	if len(c.BearerToken) != 0 {
		req.AddHeader("Authorization", fmt.Sprintf("Bearer %s", string(c.BearerToken)))
	}
	return req
}

func (c *RESTClient) Post(path string) *Request {
	return c.Do(http.MethodPost, path)
}

func (c *RESTClient) Put(path string) *Request {
	return c.Do(http.MethodPut, path)
}

func (c *RESTClient) Get(path string) *Request {
	return c.Do(http.MethodGet, path)
}

func (c *RESTClient) Delete(path string) *Request {
	return c.Do(http.MethodDelete, path)
}

func parseURL(u string, insecure bool) *url.URL {

	uri, err := url.Parse(u)
	if err != nil || uri.Scheme == "" || uri.Host == "" {
		scheme := "http://"
		if !insecure {
			scheme = "https://"
		}
		uri, err = url.Parse(scheme + u)
		if err != nil {
			return nil
		}
	}

	if !strings.HasSuffix(uri.Path, "/") {
		uri.Path += "/"
	}

	return uri
}
