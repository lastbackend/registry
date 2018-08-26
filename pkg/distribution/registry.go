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

package distribution

import (
	"context"
	"github.com/spf13/viper"
	"strings"
	"sort"
	"errors"

	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/storage"
)

const (
	logRegistryPrefix = "distribution:registry"
)

// Registry - distribution model
type Registry struct {
	context context.Context
	storage storage.IStorage
}

// Info - get registry info
func (c *Registry) Get() (*types.Registry, error) {

	log.V(logLevel).Debugf("%s:get:> get info", logRegistryPrefix)

	registry := new(types.Registry)

	if viper.IsSet("api.tls") {
		registry.Status.TLS = !viper.GetBool("api.tls.insecure")
	}

	return registry, nil
}

func (r *Registry) ParseScope(str string) (*types.Scope, error) {
	parts := strings.Split(str, ":")

	if len(parts) != 3 {
		err := errors.New("incorrect scope")
		log.V(logLevel).Errorf("%s:parse_scope:> parse scope err: %v", logRegistryPrefix, err)
		return nil, err
	}

	scope := new(types.Scope)
	scope.Type = parts[0]
	scope.Name = parts[1]
	scope.Actions = strings.Split(parts[2], ",")

	if strings.Contains(scope.Name, "/") == false {
		err := errors.New("incorrect name")
		log.V(logLevel).Errorf("%s:parse_scope:> parse name err: %v", logRegistryPrefix, err)
		return nil, err
	}

	sort.Strings(scope.Actions)

	return scope, nil
}

func (r *Registry) CreateSignature(account *types.RegistryUser, scopes *types.Scopes) (string, error) {

	log.V(logLevel).Debugf("%s:create_signature:> Creating signature from account and scopes", logRegistryPrefix)

	// Creating new JWT
	token, err := types.NewJwtToken(account.Username, scopes, viper.GetString("service"), viper.GetString("issuer"), viper.GetString("key"))
	if err != nil {
		log.V(logLevel).Errorf("%:create_signature:> create jwt token err: %v", logRegistryPrefix, err)
		return "", err
	}

	// Creating claims for JWT
	claim, err := token.Claim(token.Account, *token.Scope)
	if err != nil {
		log.V(logLevel).Errorf("%:create_signature:> creating clams err: %v", logRegistryPrefix, err)
		return "", err
	}

	// Signing JWT
	signed, err := token.SignedString(claim, token.PrivateKey)
	if err != nil {
		log.V(logLevel).Errorf("%:create_signature:> signed jwt err: %v", logRegistryPrefix, err)
		return "", err
	}

	return signed, err
}

// NewRegistryModel - return new registry model
func NewRegistryModel(ctx context.Context, stg storage.IStorage) *Registry {
	return &Registry{ctx, stg}
}
