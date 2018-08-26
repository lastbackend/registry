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

package types

import (
	"crypto/rsa"
	"github.com/dgrijalva/jwt-go"
	"time"
	"io/ioutil"
	"crypto/x509"
	"crypto"
	"strings"
	"encoding/base32"
	"bytes"
	"github.com/lastbackend/registry/pkg/util/generator"
)

type RegistryList []*Registry
type RegistryMap map[string]*Registry

type Registry struct {
	Meta   RegistryMeta   `json:"meta"`
	Status RegistryStatus `json:"status"`
}

type RegistryMeta struct {
	Meta
}

type RegistryStatus struct {
	TLS bool `json:"tls"`
}

type RegistryUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Scope struct {
	Type      string   `json:"type"`
	Name      string   `json:"name"`
	Namespace string   `json:"namespace"`
	Actions   []string `json:"actions"`
}

type Scopes []*Scope

type AccessItem struct {
	Type    string   `json:"type"`
	Name    string   `json:"name"`
	Actions []string `json:"actions"`
}

type AccessItems []AccessItem

type jwtToken struct {
	Account    string
	Issuer     string
	KeyPath    string
	Service    string
	Scope      *Scopes
	PrivateKey *rsa.PrivateKey
}

// SERVICE - The name of the token issuer.
// The issuer inserts this into the token so it must match the value configured for the issuer.
// ISSUER - The absolute path to the root certificate bundle.
// This bundle contains the public part of the certificates used to sign authentication tokens.
// KEYPATH - The absolute path to the root certificate bundle.
// This bundle contains the public part of the certificates used to sign authentication tokens.
func NewJwtToken(account string, scope *Scopes, service, issuer, keypath string) (*jwtToken, error) {

	token := new(jwtToken)
	token.Account = account
	token.Service = service
	token.Issuer = issuer
	token.KeyPath = keypath
	token.Scope = scope

	key, err := ioutil.ReadFile(keypath)
	if err != nil {
		return token, err
	}

	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(key)
	if err != nil {
		return token, err
	}

	token.PrivateKey = privKey

	return token, err
}

func (t *jwtToken) Claim(account string, scopes Scopes) (jwt.MapClaims, error) {

	claims := make(jwt.MapClaims)

	claims["iss"] = t.Issuer
	claims["sub"] = account
	claims["aud"] = t.Service

	now := time.Now()

	claims["exp"] = now.Add(time.Minute * 240).Unix()
	claims["nbf"] = now.Add(time.Minute * -240).Unix()
	claims["iat"] = now.Unix()

	claims["jti"] = generator.GetUUIDV4()

	accessItems := AccessItems{}

	for _, s := range scopes {
		action := AccessItem{
			Type:    s.Type,
			Name:    s.Name,
			Actions: s.Actions,
		}
		accessItems = append(accessItems, action)
	}

	claims["access"] = accessItems

	return claims, nil
}

func (t *jwtToken) SignedString(claim jwt.MapClaims, privateKey *rsa.PrivateKey) (string, error) {

	var err error

	token := jwt.New(jwt.SigningMethodRS256)
	token.Claims = claim

	derBytes, err := x509.MarshalPKIXPublicKey(privateKey.Public())
	if err != nil {
		return "", err
	}

	hasher := crypto.SHA256.New()
	_, err = hasher.Write(derBytes)
	if err != nil {
		return "", err
	}

	s := strings.TrimRight(base32.StdEncoding.EncodeToString(hasher.Sum(nil)[:30]), "=")
	var buf bytes.Buffer
	var i int
	for i = 0; i < len(s)/4-1; i++ {
		start := i * 4
		end := start + 4
		_, err = buf.WriteString(s[start:end] + ":")
		if err != nil {
			return "", err
		}
	}

	_, err = buf.WriteString(s[i*4:])
	if err != nil {
		return "", err
	}

	token.Header["kid"] = buf.String()

	signed, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}

	return signed, nil
}
