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
	"github.com/lastbackend/lastbackend/pkg/log"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/storage/storage"
	"github.com/lastbackend/registry/pkg/storage/types/filter"
)

type BuildStorage struct {
	storage.Build
}

func (s *BuildStorage) Get(ctx context.Context, id string) (*types.Build, error) {
	log.V(logLevel).Debugf("%s:build:get:> get build `%s`", logPrefix, id)

	const query = `
		SELECT to_json(
		 json_build_object(
		   'meta', json_build_object(
		     'id', ib.id,
		     'builder', ib.builder_id,
		     'number', ib.number,
		     'created', ib.created,
		     'updated', ib.updated
		   ),
		   'status', json_build_object(
		     'size', ib.size,
		     'step', ib.state_step,
		     'status', ib.state_status,
		     'message', ib.state_message,
		     'processing', ib.state_processing,
		     'done', ib.state_done,
		     'error', ib.state_error,
		     'canceled', ib.state_canceled,
		     'started', ib.state_started,
		     'finished', ib.state_finished
		   ),
		   'spec', json_build_object(
		     'image', ib.image,
		     'source', ib.source,
		     'config', ib.config
		   )
		  )
		 )
		FROM images_builds AS ib
		WHERE ib.id::text = $1;`

	var buf string

	err := getClient(ctx).QueryRow(query, id).Scan(&buf)
	switch err {
	case nil:
	case sql.ErrNoRows:
		return nil, nil
	default:
		log.V(logLevel).Errorf("%s:build:get:> get build %s query err: %v", logPrefix, id, err)
		return nil, err
	}

	b := new(types.Build)

	if err := json.Unmarshal([]byte(buf), &b); err != nil {
		return nil, err
	}

	return b, nil

}

func (s *BuildStorage) List(ctx context.Context, image string, f *filter.BuildFilter) ([]*types.Build, error) {
	log.V(logLevel).Debugf("%s:build:list:> get builds list by image", logPrefix)

	where := "WHERE ib.image_id = $1"

	if f != nil {

		if f.Active != nil {
			if *f.Active {
				where += "AND ib.state_processing IS TRUE"
			} else {
				where += "AND ib.state_processing IS FALSE"
			}
		}

		if where != types.EmptyString {
			where = fmt.Sprintf(" %s", where)
		}
	}

	var query = fmt.Sprintf(`
		SELECT COALESCE(to_json(json_agg(builds)), '[]')
		FROM (
		 SELECT
		   ib.id,
		   json_build_object(
		       'id', ib.id,
		       'builder', ib.builder_id,
		       'number', ib.number,
		       'created', ib.created,
		       'updated', ib.updated
		   ) AS meta,
		   json_build_object(
		       'size', ib.size,
		       'step', ib.state_step,
		       'status', ib.state_status,
		       'message', ib.state_message,
		       'processing', ib.state_processing,
		       'done', ib.state_done,
		       'error', ib.state_error,
		       'canceled', ib.state_canceled,
		       'started', ib.state_started,
		       'finished', ib.state_finished
		   ) AS status,
		   json_build_object(
		       'image', ib.image,
		       'source', ib.source,
		       'config', ib.config
		   ) AS spec
		 FROM images_builds AS ib
		 %s
		 ORDER BY ib.created DESC) builds;`, where)

	var buf string

	err := getClient(ctx).QueryRow(query, image).Scan(&buf)
	switch err {
	case nil:
	case sql.ErrNoRows:
		return nil, nil
	default:
		log.V(logLevel).Errorf("%s:build:list:> get builds list query err: %v", logPrefix, err)
		return nil, err
	}

	b := make([]*types.Build, 0)

	if err := json.Unmarshal([]byte(buf), &b); err != nil {
		return nil, err
	}

	return b, nil

}

func (s *BuildStorage) Insert(ctx context.Context, build *types.Build) error {

	log.V(logLevel).Debugf("%s:build:insert:> insert build: %#v", logPrefix, build)

	if build == nil {
		err := errors.New("build can not be empty")
		log.V(logLevel).Errorf("%s:build:insert:> insert build err: %v", logPrefix, err)
		return err
	}

	source, err := json.Marshal(build.Spec.Source)
	if err != nil {
		log.Errorf("%s:build:insert:> prepare source struct to database write: %v", logPrefix, err)
		source = []byte("{}")
	}

	config, err := json.Marshal(build.Spec.Config)
	if err != nil {
		log.Errorf("%s:build:insert:> prepare config struct to database write: %v", logPrefix, err)
		config = []byte("{}")
	}

	image, err := json.Marshal(build.Spec.Image)
	if err != nil {
		log.Errorf("%s:build:insert:> prepare image struct to database write: %v", logPrefix, err)
		image = []byte("{}")
	}

	const query = `
		WITH info AS (
		  SELECT
		    (SELECT COUNT(*) + 1
		     FROM images_builds AS ib
		     WHERE ib.image_id = i.id AND ib.image ->> 'tag' = $2) AS number
		  FROM images AS i
		  WHERE i.id = $1
		)
		INSERT INTO images_builds(image_id, number, state_status, source, config, image)
		  SELECT $1, info.number, $3, $4, $5, $6
		  FROM info
		RETURNING id, number, created, updated, created, updated;`

	err = getClient(ctx).QueryRowContext(ctx, query,
		build.Spec.Image.ID,
		build.Spec.Image.Tag,
		build.Status.Status,
		string(source),
		string(config),
		string(image),
	).
		Scan(&build.Meta.ID, &build.Meta.Number, &build.Meta.Created, &build.Meta.Updated,
			&build.Meta.Created, &build.Meta.Updated)
	if err != nil {
		log.V(logLevel).Errorf("%s:build:insert:> insert build err: %v", logPrefix, err)
		return err
	}
	return nil
}

func (s *BuildStorage) Update(ctx context.Context, build *types.Build) error {
	log.V(logLevel).Debugf("%s:build:update:> update build %#v", logPrefix, build)

	const query = `
		UPDATE images_builds
		SET
			state_step = $2,
			state_status = $3,
			state_message = $4,
			state_processing = $5,
			state_done = $6,
			state_error = $7,
			state_canceled = $8,
  		-- set state_started if state_step == '' and processing = true
  		state_started = CASE WHEN (state_step = '' AND $5 = TRUE) THEN now() at time zone 'utc' ELSE state_started END,
  		-- set state_finished if state_done = true or state_error = true or state_canceled = true
  		state_finished = CASE WHEN ($6 = TRUE OR $7 = TRUE OR $8 = TRUE) THEN now() at time zone 'utc' ELSE state_finished END,
			size = $9,
			image = json_build_object(
				'name', image ->> 'name',
				'owner', image ->> 'owner',
				'tag', image ->> 'tag',
				'hash', $10 :: TEXT
			),
			builder_id = CASE WHEN $11 <> '' THEN $11 :: UUID ELSE NULL END,
			updated = now() at time zone 'utc'
		WHERE id = $1
		RETURNING updated;`

	err := getClient(ctx).QueryRowContext(ctx, query, build.Meta.ID,
		build.Status.Step, build.Status.Status, build.Status.Message, build.Status.Processing,
		build.Status.Done, build.Status.Error, build.Status.Canceled, build.Status.Size, build.Spec.Image.Hash,
		build.Meta.Builder).Scan(&build.Meta.Updated)
	if err != nil {
		log.V(logLevel).Errorf("%s:build:update:> exec query err: %v", logPrefix, err)
		return err
	}

	return nil
}

func (s *BuildStorage) Remove(ctx context.Context, build *types.Build) error {

	log.V(logLevel).Debugf("%s:build:remove:> remove build %s", logPrefix, build.Meta.ID)

	const query = `
		DELETE FROM images_builds 
		WHERE id = $1;`

	result, err := getClient(ctx).ExecContext(ctx, query, build.Meta.ID)
	if err != nil {
		log.V(logLevel).Errorf("%s:build:remove:> remove build query err: %v", logPrefix, err)
		return err
	}

	if _, err := result.RowsAffected(); err != nil {
		log.V(logLevel).Errorf("%s:build:remove:> check query affected err: %v", logPrefix, err)
		return err
	}

	return nil
}

func (s *BuildStorage) Attach(ctx context.Context, builder *types.Builder) (*types.Build, error) {

	log.V(logLevel).Debugf("%s:build:attach:> attach build", logPrefix)

	const query = `
		UPDATE images_builds
		SET builder_id       = $1,
		    state_status     = 'queued',
		    state_processing = TRUE
		WHERE images_builds.id = (
		 SELECT ib1.id
		  FROM images_builds AS ib1
		  WHERE 
		  (
		      -- check if there are builds that are not active or finished
		      NOT (ib1.state_done OR ib1.state_canceled OR ib1.state_error OR
		           ib1.state_processing)
		      -- check if exists builds that are set as processing for one tag
		      AND NOT EXISTS(
		        SELECT TRUE
		        FROM images_builds AS ib2
		        WHERE ib2.state_processing
		          AND ib2.image ->> 'name' = ib1.image ->> 'name'
		          AND ib2.image ->> 'owner' = ib1.image ->> 'owner'
		          AND ib2.image ->> 'tag' = ib1.image ->> 'tag')
		  )
		  ORDER BY ib1.created
		  LIMIT 1)
		RETURNING images_builds.id;`

	var id sql.NullString

	if _, err := getClient(ctx).Exec(`SET TRANSACTION ISOLATION LEVEL SERIALIZABLE;`); err != nil {
		return nil, nil
	}

	err := getClient(ctx).QueryRowContext(ctx, query, builder.Meta.ID).Scan(&id)
	switch err {
	case nil:
	case sql.ErrNoRows:
		return nil, nil
	default:
		log.V(logLevel).Errorf("%s:build:attach:> attach build query err: %v", logPrefix, err)
		return nil, err
	}

	if !id.Valid {
		return nil, nil
	}

	return s.Get(ctx, id.String)
}

func (s *BuildStorage) Unfreeze(ctx context.Context) error {

	log.V(logLevel).Debugf("%s:build:unfreeze:> unfreeze builds", logPrefix)

	const query = `
   UPDATE images_builds
   SET builder_id     = NULL,
     state_step       = '',
     state_status     = 'queued',
     state_processing = FALSE,
     state_started    = NULL,
     updated          = now() at time zone 'utc'
   WHERE (updated <= (CURRENT_DATE :: timestamp - '1 day' :: interval)
      OR (
        (EXISTS(SELECT TRUE
                FROM builders
                WHERE online = FALSE
                  AND id = images_builds.builder_id))
        AND state_processing IS TRUE));`

	result, err := getClient(ctx).ExecContext(ctx, query)
	if err != nil {
		log.V(logLevel).Errorf("%s:build:unfreeze:> unfreeze builds query err: %v", logPrefix, err)
		return err
	}

	if _, err := result.RowsAffected(); err != nil {
		log.V(logLevel).Errorf("%s:build:unfreeze:> check query affected err: %v", logPrefix, err)
		return err
	}

	return nil
}

func newBuildStorage() *BuildStorage {
	s := new(BuildStorage)
	return s
}
