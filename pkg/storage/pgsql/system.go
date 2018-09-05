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

package pgsql

import (
	"context"
	"database/sql"

	"github.com/lastbackend/lastbackend/pkg/log"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/storage/storage"
)

type SystemStorage struct {
	storage.Build
}

func (s *SystemStorage) Get(ctx context.Context) (*types.System, error) {

	log.V(logLevel).Debugf("%s:system:get:> get system info", logPrefix)

	const query = `
		SELECT access_token, auth_server, updated, created
		FROM systems`

	var system = new(types.System)

	err := getClient(ctx).QueryRow(query).
		Scan(
			&system.AccessToken,
			&system.AuthServer,
			&system.Updated,
			&system.Created,
		)
	switch err {
	case nil:
	case sql.ErrNoRows:
		return nil, nil
	default:
		log.V(logLevel).Errorf("%s:system:get:> get system err: %v", logPrefix, err)
		return nil, err
	}

	return system, nil
}

func (s *SystemStorage) Update(ctx context.Context, system *types.System) error {
	log.V(logLevel).Debugf("%s:system:update:> update system %#v", logPrefix, system)

	const query = `
		UPDATE systems
		SET
			access_token = $1,
			auth_server = $2,
			updated = now() at time zone 'utc'
		RETURNING updated;`

	err := getClient(ctx).QueryRowContext(ctx, query,
		system.AccessToken,
		system.AuthServer,
	).Scan(&system.Updated)
	if err != nil {
		log.V(logLevel).Errorf("%s:system:update:> exec query err: %v", logPrefix, err)
		return err
	}

	return nil
}

func newSystemStorage() *SystemStorage {
	s := new(SystemStorage)
	return s
}
