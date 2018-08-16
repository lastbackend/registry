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
	"errors"

	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/storage/storage"
	"github.com/lastbackend/registry/pkg/storage/types/filter"
)

type BuilderStorage struct {
	storage.Builder
}

func (s *BuilderStorage) Get(ctx context.Context, builder string) (*types.Builder, error) {
	log.V(logLevel).Debugf("%s:builder:get:> get builder `%s`", logPrefix, builder)

	if len(builder) == 0 {
		err := errors.New("hostname can not be empty")
		log.V(logLevel).Errorf("%s:builder:get:> get image err: %v", logPrefix, err)
		return nil, err
	}

	const query = `
		SELECT to_jsonb(
		  jsonb_build_object(
		    'meta', jsonb_build_object(
		      'id', b.id,
		      'hostname', b.hostname,
		      'created', b.created,
		      'updated', b.updated
		    ),
		    'status', jsonb_build_object(
          'online', b.online,
			    'insecure', b.tls
        ),
				'spec', jsonb_build_object(
					'network', jsonb_build_object(
						'tls', b.tls,
						'ssl', b.ssl,
						'ip', b.ip,
						'port', b.port
					)
				)
		  )
		)
		FROM builders AS b
		WHERE b.hostname::text = $1 OR b.id::text = $1;`

	var buf string

	err := getClient(ctx).QueryRow(query, builder).Scan(&buf)
	switch err {
	case nil:
	case sql.ErrNoRows:
		return nil, nil
	default:
		log.V(logLevel).Errorf("%s:builder:get:> get builder err: %v", logPrefix, err)
		return nil, err
	}

	b := new(types.Builder)

	if err := json.Unmarshal([]byte(buf), &b); err != nil {
		return nil, err
	}

	return b, nil
}

func (s *BuilderStorage) List(ctx context.Context, f *filter.BuilderFilter) ([]*types.Builder, error) {

	log.V(logLevel).Debug("%s:builder:list:> get builders list", logPrefix)

	where := types.EmptyString

	if f != nil {

		if f.Online != nil {
			if *f.Online {
				where += "online IS TRUE"
			} else {
				where += "online IS FALSE"
			}
		}

		if where != types.EmptyString {
			where = fmt.Sprintf("WHERE %s", where)
		}
	}

	var query = fmt.Sprintf(`
		SELECT COALESCE(jsonb_build_array(builders), '[]')
		FROM (SELECT 
		json_build_object(
			'id', id,
			'hostname', hostname,
			'created', created,
			'updated', updated
		) AS meta,
		json_build_object(
			'online', online,
			'insecure', tls
		) AS status,
		json_build_object(
			'network', json_build_object(
				'ip', ip,
				'port', port,
				'tls', tls,
				'ssl', ssl
			)
		) AS spec
		FROM builders
		%s
		ORDER BY created DESC) AS builders;`, where)

	var buf string

	err := getClient(ctx).QueryRow(query).Scan(&buf)
	switch err {
	case nil:
	case sql.ErrNoRows:
		return nil, nil
	default:
		log.V(logLevel).Errorf("%s:builder:list:> get builders list err: %v", logPrefix, err)
		return nil, err
	}

	b := make([]*types.Builder, 0)

	if err := json.Unmarshal([]byte(buf), &b); err != nil {
		return nil, err
	}

	return b, nil
}

func (s *BuilderStorage) Insert(ctx context.Context, builder *types.Builder) error {

	log.V(logLevel).Debugf("%s:builder:insert:> insert builder: %#v", logPrefix, builder)

	if builder == nil {
		err := errors.New("builder can not be empty")
		log.V(logLevel).Errorf("%s:builder:insert:> insert builder err: %v", logPrefix, err)
		return err
	}

	const query = `
    INSERT INTO builders(hostname, online, tls, ip, port, ssl)
		VALUES ($1, $2, $3, $4, $5, $6)
   	RETURNING id, created, updated;`

	ssl, err := json.Marshal(builder.Spec.Network.SSL)
	if err != nil {
		log.Errorf("%s:insert:> prepare ssl struct to database write: %s", logPrefix, err)
		ssl = []byte("{}")
	}

	err = getClient(ctx).QueryRowContext(ctx, query,
		builder.Meta.Hostname,
		builder.Status.Online,
		builder.Spec.Network.TLS,
		builder.Spec.Network.IP,
		builder.Spec.Network.Port,
		string(ssl),
	).
		Scan(&builder.Meta.ID, &builder.Meta.Created, &builder.Meta.Updated)
	if err != nil {
		log.V(logLevel).Errorf("%s:builder:insert:> insert builder err: %v", logPrefix, err)
		return err
	}

	return nil
}

func (s *BuilderStorage) Update(ctx context.Context, builder *types.Builder) error {
	log.V(logLevel).Debugf("%s:builder:update:> update builder %#v", logPrefix, builder)

	const query = `
		UPDATE builders
		SET
			online = $2,
			ip = $3,
			port = $4,
			tls = $5,
			ssl = $6,
			updated = now() at time zone 'utc'
		WHERE id = $1
		RETURNING updated;`

	ssl, err := json.Marshal(builder.Spec.Network.SSL)
	if err != nil {
		log.Errorf("%s:insert:> prepare ssl struct to database write: %s", logPrefix, err)
		ssl = []byte("{}")
	}

	err = getClient(ctx).QueryRowContext(ctx, query,
		builder.Meta.ID,
		builder.Status.Online,
		builder.Spec.Network.IP,
		builder.Spec.Network.Port,
		builder.Spec.Network.TLS,
		string(ssl),
	).
		Scan(&builder.Meta.Updated)
	if err != nil {
		log.V(logLevel).Errorf("%s:builder:update:> exec query err: %v", logPrefix, err)
		return err
	}

	return nil
}

func (s *BuilderStorage) MarkOffline(ctx context.Context) error {

	log.V(logLevel).Debugf("%s:builder:mark_offline:> mark offline builders", logPrefix)

	const query = `
		UPDATE builders
		SET
		  online = FALSE
		WHERE updated <= (NOW() :: timestamp - '5 minutes' :: interval);`

	result, err := getClient(ctx).ExecContext(ctx, query)
	if err != nil {
		log.V(logLevel).Errorf("%s:builder:mark_offline:> makr offline builders query err: %v", logPrefix, err)
		return err
	}

	if _, err := result.RowsAffected(); err != nil {
		log.V(logLevel).Errorf("%s:builder:mark_offline:> check query affected err: %v", logPrefix, err)
		return err
	}

	return nil
}

func newBuilderStorage() *BuilderStorage {
	s := new(BuilderStorage)
	return s
}
