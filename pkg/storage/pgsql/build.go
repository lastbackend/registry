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
	"reflect"
	"strings"

	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/lastbackend/lastbackend/pkg/log"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/storage/storage"
	"github.com/lastbackend/registry/pkg/storage/types/filter"
)

const (
	logBuildPrefix = "storage:pgsql:build"
)

type BuildStorage struct {
	storage.Build
}

func (s *BuildStorage) Get(ctx context.Context, id string) (*types.Build, error) {
	log.V(logLevel).Debugf("%s:get:> get build `%s`", logBuildPrefix, id)

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
		log.V(logLevel).Errorf("%s:get:> get build %s query err: %v", logBuildPrefix, id, err)
		return nil, err
	}

	b := new(types.Build)

	if err := json.Unmarshal([]byte(buf), &b); err != nil {
		return nil, err
	}

	return b, nil

}

func (s *BuildStorage) List(ctx context.Context, image string, f *filter.BuildFilter) (*types.BuildList, error) {
	log.V(logLevel).Debugf("%s:list:> get builds list by image", logBuildPrefix)

	var values = make([]interface{}, 0)
	var where = make([]string, 0)

	var whereCondition = types.EmptyString

	values = append(values, image)
	where = append(where, "image_id=$1")
	if f != nil {

		t := reflect.TypeOf(*f)
		v := reflect.ValueOf(*f)

		for i := 0; i < t.NumField(); i++ {
			tp := v.Field(i)

			if (tp.Kind() == reflect.Ptr || tp.Kind() == reflect.Interface) && tp.IsNil() {
				continue
			}

			name := t.Field(i).Tag.Get("db")

			switch tp.Elem().Kind() {
			case reflect.String:
				values = append(values, tp.Elem().String())
				where = append(where, fmt.Sprintf("%s=$%d", name, len(values)+1))
			case reflect.Bool:
				where = append(where, fmt.Sprintf("%s=$%d", name, len(values)+1))
				if tp.Elem().Bool() == true {
					values = append(values, "TRUE")
				} else {
					values = append(values, "FALSE")
				}
			}
		}

		if len(where) > 0 {
			whereCondition = fmt.Sprintf("WHERE %s", strings.Join(where, " AND "))
		}
	}

	var filterCondition = types.EmptyString
	var page int64 = 0
	var limit int64 = 0

	if f.Limit != nil && f.Page != nil {
		page = *f.Page
		limit = *f.Limit

		if page == 0 {
			page = 1
		}
		if limit == 0 {
			limit = 10
		}

		filterCondition = fmt.Sprintf(`
			OFFSET %d
			LIMIT %d`, (page-1)*limit, limit)
	}

	query := fmt.Sprintf(`
		SELECT json_build_object(
			'total', (SELECT COUNT(*) 
                FROM images_builds 
								WHERE image_id = $1),
			'items', COALESCE(json_agg(
				json_build_object(
		    	'meta', json_build_object(
		    	  'id', tmp.id,
		    	  'builder', tmp.builder_id,
		    	  'number', tmp.number,
		    	  'created', tmp.created,
		    	  'updated', tmp.updated
		    	),
		    	'status', json_build_object(
		    	  'size', tmp.size,
		    	  'step', tmp.state_step,
		    	  'status', tmp.state_status,
		    	  'message', tmp.state_message,
		    	  'processing', tmp.state_processing,
		    	  'done', tmp.state_done,
		    	  'error', tmp.state_error,
		    	  'canceled', tmp.state_canceled,
		    	  'started', tmp.state_started,
		    	  'finished', tmp.state_finished
		    	),
		   		'spec', json_build_object(
		    	  'image', tmp.image,
		    	  'source', tmp.source,
		    	  'config', tmp.config
		   		)
				)
			), '[]')
		)
		FROM (
			SELECT *
			FROM images_builds
			%s
		  ORDER BY created DESC
		  %s
		) AS tmp;`, whereCondition, filterCondition)

	var buf string

	err := getClient(ctx).QueryRow(query, image).Scan(&buf)
	switch err {
	case nil:
	case sql.ErrNoRows:
		return nil, nil
	default:
		log.V(logLevel).Errorf("%s:list:> get builds list query err: %v", logBuildPrefix, err)
		return nil, err
	}

	b := new(types.BuildList)

	if err := json.Unmarshal([]byte(buf), &b); err != nil {
		return nil, err
	}

	b.Page = page
	b.Limit = limit

	return b, nil
}

func (s *BuildStorage) Insert(ctx context.Context, build *types.Build) error {

	log.V(logLevel).Debugf("%s:insert:> insert build: %#v", logBuildPrefix, build)

	if build == nil {
		err := errors.New("build can not be empty")
		log.V(logLevel).Errorf("%s:insert:> insert build err: %v", logBuildPrefix, err)
		return err
	}

	source, err := json.Marshal(build.Spec.Source)
	if err != nil {
		log.Errorf("%s:insert:> prepare source struct to database write: %v", logBuildPrefix, err)
		source = []byte("{}")
	}

	config, err := json.Marshal(build.Spec.Config)
	if err != nil {
		log.Errorf("%s:insert:> prepare config struct to database write: %v", logBuildPrefix, err)
		config = []byte("{}")
	}

	image, err := json.Marshal(build.Spec.Image)
	if err != nil {
		log.Errorf("%s:insert:> prepare image struct to database write: %v", logBuildPrefix, err)
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
		log.V(logLevel).Errorf("%s:insert:> insert build err: %v", logBuildPrefix, err)
		return err
	}
	return nil
}

func (s *BuildStorage) Update(ctx context.Context, build *types.Build) error {
	log.V(logLevel).Debugf("%s:update:> update build %#v", logBuildPrefix, build)

	const query = `
		UPDATE images_builds
		SET
      builder_id       = CASE WHEN ($6 = TRUE OR $7 = TRUE OR $8 = TRUE) THEN NULL ELSE builder_id END,
			state_step       = $2,
			state_status     = $3,
			state_message    = $4,
			state_processing = $5,
			state_done       = $6,
			state_error      = $7,
			state_canceled   = $8,
  		-- set state_started if state_step == '' and processing = true
  		state_started    = CASE WHEN (state_started = NULL' AND $5 = TRUE) THEN now() at time zone 'utc' ELSE state_started END,
  		-- set state_finished if state_done = true or state_error = true or state_canceled = true
  		state_finished   = CASE WHEN state_started <> NULL AND ($6 = TRUE OR $7 = TRUE OR $8 = TRUE) THEN now() at time zone 'utc' ELSE state_finished END,
			size = $9,
			image = json_build_object(
				'name', image ->> 'name',
				'owner', image ->> 'owner',
				'tag', image ->> 'tag',
				'auth', image ->> 'auth',
				'hash', $10 :: TEXT
			),
			updated = now() at time zone 'utc'
		WHERE id = $1
    RETURNING updated;`

	err := getClient(ctx).QueryRowContext(ctx, query,
		build.Meta.ID,
		build.Status.Step,
		build.Status.Status,
		build.Status.Message,
		build.Status.Processing,
		build.Status.Done,
		build.Status.Error,
		build.Status.Canceled,
		build.Status.Size,
		build.Spec.Image.Hash,
	).Scan(&build.Meta.Updated)
	if err != nil {
		log.V(logLevel).Errorf("%s:update:> exec query err: %v", logBuildPrefix, err)
		return err
	}

	return nil
}

func (s *BuildStorage) Remove(ctx context.Context, build *types.Build) error {

	log.V(logLevel).Debugf("%s:remove:> remove build %s", logBuildPrefix, build.Meta.ID)

	const query = `
		DELETE FROM images_builds 
		WHERE id = $1;`

	result, err := getClient(ctx).ExecContext(ctx, query, build.Meta.ID)
	if err != nil {
		log.V(logLevel).Errorf("%s:remove:> remove build query err: %v", logBuildPrefix, err)
		return err
	}

	if _, err := result.RowsAffected(); err != nil {
		log.V(logLevel).Errorf("%s:remove:> check query affected err: %v", logBuildPrefix, err)
		return err
	}

	return nil
}

func (s *BuildStorage) Attach(ctx context.Context, builder *types.Builder) (*types.Build, error) {

	log.V(logLevel).Debugf("%s:attach:> attach build", logBuildPrefix)

	const query = `
		UPDATE images_builds
		SET 
      builder_id       = $1,
		  state_status     = $2,
		  state_processing = TRUE
		WHERE images_builds.id = (
		 SELECT ib1.id
		  FROM images_builds AS ib1
		  WHERE 
		  (
		      -- check if there are builds that are not active or finished
		      NOT (ib1.state_done OR ib1.state_canceled OR ib1.state_error OR ib1.state_processing)
		      
          -- check if exists builds that are set as processing for one tag
		      AND NOT EXISTS(
		        SELECT TRUE
		        FROM images_builds AS ib2
		        WHERE ib2.state_processing
		          AND ib2.image ->> 'name' = ib1.image ->> 'name'
		          AND ib2.image ->> 'owner' = ib1.image ->> 'owner'
		          AND ib2.image ->> 'tag' = ib1.image ->> 'tag'
				)
		  )
		  ORDER BY ib1.created
		  LIMIT 1)
		RETURNING images_builds.id;`

	var id sql.NullString

	if _, err := getClient(ctx).Exec(`SET TRANSACTION ISOLATION LEVEL SERIALIZABLE;`); err != nil {
		return nil, nil
	}

	err := getClient(ctx).QueryRowContext(ctx, query, builder.Meta.ID, types.BuildStatusQueued).Scan(&id)
	switch err {
	case nil:
	case sql.ErrNoRows:
		return nil, nil
	default:
		log.V(logLevel).Errorf("%s:attach:> attach build query err: %v", logBuildPrefix, err)
		return nil, err
	}

	if !id.Valid {
		return nil, nil
	}

	return s.Get(ctx, id.String)
}

func (s *BuildStorage) Unfreeze(ctx context.Context) error {

	log.V(logLevel).Debugf("%s:unfreeze:> unfreeze builds", logBuildPrefix)

	const query = `
   UPDATE images_builds
   SET 
     builder_id       = NULL,
     state_step       = '',
     state_status     = $1,
     state_processing = FALSE,
     state_started    = NULL,
     updated          = now() at time zone 'utc'
   WHERE (updated <= (CURRENT_DATE :: timestamp - '1 day' :: interval)
     OR (
      (EXISTS(SELECT TRUE
              FROM builders
              WHERE online = FALSE AND id = images_builds.builder_id))
      AND state_processing IS TRUE));`

	result, err := getClient(ctx).ExecContext(ctx, query, types.BuildStatusQueued)
	if err != nil {
		log.V(logLevel).Errorf("%s:unfreeze:> unfreeze builds query err: %v", logBuildPrefix, err)
		return err
	}

	if _, err := result.RowsAffected(); err != nil {
		log.V(logLevel).Errorf("%s:unfreeze:> check query affected err: %v", logBuildPrefix, err)
		return err
	}

	return nil
}

func newBuildStorage() *BuildStorage {
	s := new(BuildStorage)
	return s
}
