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
	"encoding/json"
	"errors"

	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/storage/storage"
)

// Service Build type for interface in interfaces folder
type BuildStorage struct {
	storage.Build
}

func (s *BuildStorage) Insert(ctx context.Context, build *types.Build) error {

	log.V(logLevel).Debugf("%s:build:insert:> insert build: %#v", logPrefix, build)

	if build == nil {
		err := errors.New("build can not be empty")
		log.V(logLevel).Errorf("%s:build:insert:> insert build err: %v", logPrefix, err)
		return err
	}

	sources, err := json.Marshal(build.Sources)
	if err != nil {
		log.Errorf("Storage: Account: prepare sources struct to database write: %v", logPrefix, err)
		sources = []byte("{}")
	}

	config, err := json.Marshal(build.Config)
	if err != nil {
		log.Errorf("Storage: Account: prepare config struct to database write: %v", logPrefix, err)
		config = []byte("{}")
	}

	image, err := json.Marshal(build.Image)
	if err != nil {
		log.Errorf("Storage: Account: prepare image struct to database write: %v", logPrefix, err)
		image = []byte("{}")
	}

	const query = `
		WITH info AS (
		    SELECT
		      (SELECT COUNT(*) + 1
		       FROM repositories_builds
		       WHERE repo_id = r.id AND image ->> 'tag' = $2) AS number,
		      rt.id                  AS tag_id,
		      r.owner,
		      r.name
		    FROM repositories AS r
		      LEFT JOIN repositories_tags AS rt ON rt.repo_id = r.id AND rt.name = $2
		    WHERE r.id = $1
		)
		INSERT INTO repositories_builds (repo_id, tag_id, self_link, number, state_status, sources, config, image)
	  	SELECT $1, info.tag_id, (info.owner || '.' || info.name || '.' || $2 || '.' || info.number), info.number, $3, $4, $5, $6
  		FROM info
		RETURNING id, self_link, number, created, updated;`

	err = getClient(ctx).QueryRowContext(ctx, query, build.Repo.ID, build.Image.Tag, build.State.Status,
		string(sources), string(config), string(image)).
		Scan(&build.Meta.ID, &build.Meta.SelfLink, &build.Meta.Number, &build.Meta.Updated, &build.Meta.Created)
	if err != nil {
		log.V(logLevel).Errorf("%s:build:insert:> insert build meta err: %v", logPrefix, err)
		return err
	}

	return nil
}

func (s *BuildStorage) Get(ctx context.Context, id string) (*types.Build, error) {

	log.V(logLevel).Debugf("%s:build:get:> get build: %v", logPrefix, id)

	if len(id) == 0 {
		err := errors.New("id can not be empty")
		log.V(logLevel).Errorf("%s:build:get:> get build err: %v", logPrefix, err)
		return nil, err
	}

	const query = `
    SELECT to_json(
        json_build_object(
            'meta', json_build_object(
                'id', rb.id,
                'self_link', rb.self_link,
                'tag', rb.tag_id,
                'builder', rb.builder_id,
                'task', rb.task_id,
                'number', rb.number,
                'created', rb.created,
                'updated', rb.updated
            ),
            'repo', json_build_object(
                'id', r.id,
                'owner', r.owner,
                'name', r.name,
                'self_link', r.self_link
            ),
            'state', json_build_object(
                'step', rb.state_step,
                'status', rb.state_status,
                'message', rb.state_message,
                'processing', rb.state_processing,
                'done', rb.state_done,
                'error', rb.state_error,
                'canceled', rb.state_canceled,
                'started', rb.state_started,
                'finished', rb.state_finished
            ),
            'stats', json_build_object(
                'size', rb.size
            ),
            'image', rb.image,
            'sources', json_build_object(
                'hub', rb.sources -> 'hub',
                'name', rb.sources -> 'name',
                'owner', rb.sources -> 'owner',
                'branch', rb.sources -> 'branch',
                'commit', rb.sources -> 'commit'
            ),
            'config', rb.config
        )
    )
    FROM repositories_builds AS rb
      LEFT JOIN repositories AS r ON r.id = rb.repo_id
    WHERE rb.id ::TEXT = $1 OR rb.task_id :: TEXT = $1 OR rb.self_link = $1;`

	var buf string

	err := getClient(ctx).QueryRowContext(ctx, query, id).Scan(&buf)
	switch err {
	case nil:
	case sql.ErrNoRows:
		return nil, nil
	default:
		log.V(logLevel).Errorf("%s:build:get:> get build err: %v", logPrefix, err)
		return nil, err
	}

	b := new(types.Build)

	if err := json.Unmarshal([]byte(buf), &b); err != nil {
		return nil, err
	}

	return b, nil
}

func (s *BuildStorage) List(ctx context.Context, repo string) (map[string]*types.Build, error) {

	log.V(logLevel).Debugf("%s:build:list:> get builds list", logPrefix)

	const query = `
		SELECT COALESCE(to_json(json_object_agg(id, builds)), '{}')
		FROM (
		 SELECT
			 rb.id,
			 json_build_object(
				 'id', rb.id,
         'self_link', rb.self_link,
				 'tag', rb.tag_id,
				 'builder', rb.builder_id,
				 'task', rb.task_id,
				 'number', rb.number,
				 'created', rb.created,
				 'updated', rb.updated
			 ) AS meta,
			 json_build_object(
					'id', r.id,
					'owner', r.owner,
					'name', r.name,
					'self_link', r.self_link
			 ) AS repo,
			 json_build_object(
				 'step', rb.state_step,
				 'status', rb.state_status,
				 'message', rb.state_message,
				 'processing', rb.state_processing,
				 'done', rb.state_done,
				 'error', rb.state_error,
				 'canceled', rb.state_canceled,
				 'started', rb.state_started,
				 'finished', rb.state_finished
			 ) AS state,
			 json_build_object(
				 'size', rb.size
			 ) AS stats,
			 rb.image,
			 rb.sources,
			 rb.config
		 FROM repositories_builds AS rb
			LEFT JOIN repositories AS r ON r.id = rb.repo_id
		 WHERE rb.repo_id = $1) builds;`

	var buf string

	err := getClient(ctx).QueryRowContext(ctx, query, repo).Scan(&buf)
	switch err {
	case nil:
	case sql.ErrNoRows:
		return nil, nil
	default:
		log.V(logLevel).Errorf("%s:build:list:> get builds list query err: %v", logPrefix, err)
		return nil, err
	}

	b := make(map[string]*types.Build)

	if err := json.Unmarshal([]byte(buf), &b); err != nil {
		return nil, err
	}

	return b, nil
}

func (s *BuildStorage) Update(ctx context.Context, build *types.Build) error {
	log.V(logLevel).Debugf("%s:build:update:> update build %s data", logPrefix, build.Meta.SelfLink)

	const query = `
		UPDATE repositories_builds
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
				'hub', image ->> 'hub',
				'name', image ->> 'name',
				'owner', image ->> 'owner',
				'tag', image ->> 'tag',
				'hash', $10 :: TEXT
			),

			builder_id = CASE WHEN $11 <> '' THEN $11 :: UUID ELSE NULL END,
			task_id = CASE WHEN $12 <> '' THEN $12 :: UUID ELSE NULL END,

			updated = now() at time zone 'utc'
		WHERE id = $1
		RETURNING updated;`

	err := getClient(ctx).QueryRowContext(ctx, query, build.Meta.ID,
		build.State.Step, build.State.Status, build.State.Message, build.State.Processing,
		build.State.Done, build.State.Error, build.State.Canceled, build.Stats.Size, build.Image.Hash,
		build.Meta.Builder, build.Meta.Task).Scan(&build.Meta.Updated)
	if err != nil {
		log.V(logLevel).Errorf("%s:build:update:> exec query err: %v", logPrefix, err)
		return err
	}

	return nil
}

func newBuildStorage() *BuildStorage {
	s := new(BuildStorage)
	return s
}
