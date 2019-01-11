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
// and its suppliers and may be covered by Russian Federation and Foreign logSystemPrefixatents,
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

const (
	logSystemPrefix = "storage:pgsql:system"
)

type SystemStorage struct {
	storage.System
}

func (s *SystemStorage) Get(ctx context.Context) (*types.System, error) {

	log.V(logLevel).Debugf("%s:get:> get system info", logSystemPrefix)

	const query = `
		SELECT 
			access_token, 
      auth_server, 
      COALESCE(ctrl_master, ''), 
      ctrl_updated, 
      ctrl_last_event, 
      updated, 
      created
		FROM systems`

	var system = new(types.System)

	err := getClient(ctx).QueryRow(query).
		Scan(
			&system.AccessToken,
			&system.AuthServer,
			&system.CtrlMaster,
			&system.CtrlUpdated,
			&system.CtrlLastEvent,
			&system.Updated,
			&system.Created,
		)
	switch err {
	case nil:
	case sql.ErrNoRows:
		return nil, nil
	default:
		log.V(logLevel).Errorf("%s:get:> get system err: %v", logSystemPrefix, err)
		return nil, err
	}

	return system, nil
}

func (s *SystemStorage) Update(ctx context.Context, system *types.System) error {
	log.V(logLevel).Debugf("%s:update:> update system %#v", logSystemPrefix, system)

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
		log.V(logLevel).Errorf("%s:update:> exec query err: %v", logSystemPrefix, err)
		return err
	}

	return nil
}

func (s *SystemStorage) UpdateControllerMaster(ctx context.Context, system *types.System) error {

	log.V(logLevel).Debugf("%s:update_ctrl_master:> update controller %s", logSystemPrefix, system.CtrlMaster)

	const query = `
		UPDATE systems
		SET
			ctrl_master = $1,
			ctrl_updated = now() at time zone 'utc',
      updated = now() at time zone 'utc'
    WHERE COALESCE(ctrl_master, '') = '' 
       OR ctrl_updated IS NULL
			 OR ctrl_updated < (NOW() :: timestamp - '1 minutes' :: interval) 
       OR ctrl_master = $1;`

	_, err := getClient(ctx).ExecContext(ctx, query, system.CtrlMaster)
	if err != nil {
		log.V(logLevel).Errorf("%s:update_ctrl_master:> exec query err: %v", logSystemPrefix, err)
		return err
	}

	item, err := s.Get(ctx)
	if err != nil {
		log.V(logLevel).Errorf("%s:update_ctrl_master:> get system row err: %v", logSystemPrefix, err)
		return err
	}

	system.CtrlMaster = item.CtrlMaster
	system.AccessToken = item.AccessToken
	system.AuthServer = item.AuthServer
	system.CtrlMaster = item.CtrlMaster
	system.CtrlUpdated = item.CtrlUpdated
	system.CtrlLastEvent = item.CtrlLastEvent
	system.Created = item.Created
	system.Updated = item.Updated

	return err
}

func (s *SystemStorage) UpdateControllerLastEvent(ctx context.Context, system *types.System) error {

	log.V(logLevel).Debugf("%s:update_ctrl_last_event:> update controller last event %s", logSystemPrefix, system.CtrlMaster)

	const query = `
		UPDATE systems
		SET
			ctrl_last_event = $2,
      updated = now() at time zone 'utc'
    WHERE COALESCE(ctrl_master, '') <> '' AND ctrl_master = $1
		RETURNING ctrl_last_event, updated;`

	err := getClient(ctx).QueryRowContext(ctx, query,
		system.CtrlMaster,
		system.CtrlLastEvent,
	).Scan(&system.CtrlLastEvent, &system.Updated)
	if err != nil {
		log.V(logLevel).Errorf("%s:update_ctrl_last_event:> exec query err: %v", logSystemPrefix, err)
		return err
	}

	return err
}

func newSystemStorage() *SystemStorage {
	s := new(SystemStorage)
	return s
}
