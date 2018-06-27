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

package url

import (
	"log"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestDecode(t *testing.T) {

	viper.AddConfigPath("../../../contrib")

	// Find and read the config file; Handle errors reading the config file
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read config file: %s\n", err)
	}

	var assets = make(map[string]Sources)

	assets["hub.lstbknd.net/owner/name:tag"] = Sources{
		Type:   "registry",
		Hub:    "hub.lstbknd.net",
		Owner:  "owner",
		Name:   "name",
		Branch: "tag",
	}

	assets["index.docker.io/library/redis:latest"] = Sources{
		Type:   "image",
		Hub:    "index.docker.io",
		Owner:  "library",
		Name:   "redis",
		Branch: "latest",
	}

	assets["index.docker.io/library/redis-demo:alpine3-5"] = Sources{
		Type:   "image",
		Hub:    "index.docker.io",
		Owner:  "library",
		Name:   "redis-demo",
		Branch: "alpine3-5",
	}

	assets["index.docker.io/library/redis_demo:alpine3_5"] = Sources{
		Type:   "image",
		Hub:    "index.docker.io",
		Owner:  "library",
		Name:   "redis_demo",
		Branch: "alpine3_5",
	}

	assets["index.docker.io/library/3redis_demo:3alpine3_5"] = Sources{
		Type:   "image",
		Hub:    "index.docker.io",
		Owner:  "library",
		Name:   "3redis_demo",
		Branch: "3alpine3_5",
	}

	assets["https://github.com:lastbackend/genesis.git#develop"] = Sources{
		Type:   "git",
		Hub:    "github.com",
		Owner:  "lastbackend",
		Name:   "genesis",
		Branch: "develop",
	}

	assets["https://github.com:lastbackend/genesis.git"] = Sources{
		Type:  "git",
		Hub:   "github.com",
		Owner: "lastbackend",
		Name:  "genesis",
	}

	assets["git@github.com:lastbackend/genesis.git#develop"] = Sources{
		Type:   "git",
		User:   "git",
		Hub:    "github.com",
		Owner:  "lastbackend",
		Name:   "genesis",
		Branch: "develop",
	}

	assets["git:demo@github.com:lastbackend/genesis.git#develop"] = Sources{
		Type:   "git",
		User:   "git",
		Pass:   "demo",
		Hub:    "github.com",
		Owner:  "lastbackend",
		Name:   "genesis",
		Branch: "develop",
	}

	assets["hub.lstbknd.net/:tag"] = Sources{}

	for url, valid := range assets {
		i := Decode(url)
		assert.Equal(t, valid.Type, i.Type, "type")
		assert.Equal(t, valid.User, i.User, "user")
		assert.Equal(t, valid.Pass, i.Pass, "pass")
		assert.Equal(t, valid.Hub, i.Hub, "hub")
		assert.Equal(t, valid.Owner, i.Owner, "owner")
		assert.Equal(t, valid.Name, i.Name, "name")
		assert.Equal(t, valid.Branch, i.Branch, "name")
	}

}
