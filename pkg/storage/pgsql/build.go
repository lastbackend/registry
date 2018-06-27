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

	log.V(logLevel).Debugf("Storage: Build: insert build: %#v", build)

	if build == nil {
		err := errors.New("build can not be empty")
		log.V(logLevel).Errorf("Storage: Build: insert build err: %s", err)
		return err
	}

	sources, err := json.Marshal(build.Sources)
	if err != nil {
		log.Errorf("Storage: Account: prepare sources struct to database write: %s", err)
		sources = []byte("{}")
	}

	config, err := json.Marshal(build.Config)
	if err != nil {
		log.Errorf("Storage: Account: prepare config struct to database write: %s", err)
		config = []byte("{}")
	}

	image, err := json.Marshal(build.Image)
	if err != nil {
		log.Errorf("Storage: Account: prepare image struct to database write: %s", err)
		image = []byte("{}")
	}

	const query = `
		INSERT INTO repositories_builds(repo_id, tag_id, number, state_status, sources, config, image)
		VALUES ($1, (SELECT id FROM repositories_tags WHERE repo_id = $1 AND name = $2),
					 (SELECT COUNT(*) + 1 FROM repositories_builds WHERE repo_id = $1), $3, $4, $5, $6)
		RETURNING id, number, created, updated;`

	err = getClient(ctx).QueryRow(query, build.Repo.ID, build.Image.Tag, build.State.Status,
		string(sources), string(config), string(image)).
		Scan(&build.Meta.ID, &build.Meta.Number, &build.Meta.Updated, &build.Meta.Created)
	if err != nil {
		log.V(logLevel).Errorf("Storage: Build: Insert: insert build meta err: %s", err)
		return err
	}

	return nil
}

func (s *BuildStorage) Get(ctx context.Context, id string) (*types.Build, error) {
	log.V(logLevel).Debugf("Storage: Build: Get: get build: %s", id)

	if len(id) == 0 {
		err := errors.New("id can not be empty")
		log.V(logLevel).Errorf("Storage: Build: Get: get build err: %s", err)
		return nil, err
	}

	const query = `
		SELECT to_json(
				json_build_object(
						'meta', json_build_object(
								'id', rb.id,
								'self_link', r.owner || '.' || r.name || ':' || rb.number,
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
								'size', rb.size,
								'layers', rb.layers,
								'avg_layer_size', rb.avg_layer_size,
								'max_layer_size', rb.max_layer_size
						),
						'image', rb.image,
						'sources', json_build_object(
								'hub', rb.sources -> 'hub',
								'name', rb.sources -> 'name',
								'owner', rb.sources -> 'owner',
								'branch', rb.sources -> 'branch',
								'commit', rb.sources -> 'commit',
								'token', (SELECT vendors -> 'platform' -> split_part(rb.sources ->> 'hub', '.', 1)::TEXT -> 'token' ->> 'access_token' AS token
                        FROM accounts AS a
                          INNER JOIN repositories AS r ON r.id = rb.repo_id
                        WHERE a.id = r.account_id AND vendors -> 'platform' -> split_part(rb.sources ->> 'hub', '.', 1)::TEXT ->> 'type' = 'platform')
						),
						'config', rb.config
				)
		)
		FROM repositories_builds AS rb
			LEFT JOIN repositories AS r ON r.id = rb.repo_id
			LEFT JOIN accounts AS a ON a.id = r.account_id
		WHERE rb.id ::TEXT = $1 OR rb.task_id :: TEXT = $1 OR (r.owner || '.' || r.name || ':' || rb.number) :: TEXT = $1;`

	var buf string

	err := getClient(ctx).QueryRow(query, id).Scan(&buf)
	switch err {
	case nil:
	case sql.ErrNoRows:
		return nil, nil
	default:
		log.V(logLevel).Errorf("Storage: Build: Get: get build err: %s", err)
		return nil, err
	}

	b := new(types.Build)

	if err := json.Unmarshal([]byte(buf), &b); err != nil {
		return nil, err
	}

	return b, nil
}

func (s *BuildStorage) List(ctx context.Context, repo string) (map[string]*types.Build, error) {

	log.V(logLevel).Debugf("Storage: Build: ListByRepo: get builds list")

	const query = `
		SELECT COALESCE(to_json(json_object_agg(id, builds)), '{}')
		FROM (
		 SELECT
			 rb.id,
			 json_build_object(
				 'id', rb.id,
         'self_link', r.owner || '.' || r.name || ':' || rb.number,
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
				 'size', rb.size,
				 'layers', rb.layers,
				 'avg_layer_size', rb.avg_layer_size,
				 'max_layer_size', rb.max_layer_size
			 ) AS stats,
			 rb.image,
			 rb.sources,
			 rb.config
		 FROM repositories_builds AS rb
			LEFT JOIN repositories AS r ON r.id = rb.repo_id
		 WHERE rb.repo_id = $1) builds;`

	var buf string

	err := getClient(ctx).QueryRow(query, repo).Scan(&buf)
	switch err {
	case nil:
	case sql.ErrNoRows:
		return nil, nil
	default:
		log.V(logLevel).Errorf("Storage: Build: ListByRepo: get builds list query err: %s", err)
		return nil, err
	}

	b := make(map[string]*types.Build)

	if err := json.Unmarshal([]byte(buf), &b); err != nil {
		return nil, err
	}

	return b, nil
}

func newBuildStorage() *BuildStorage {
	s := new(BuildStorage)
	return s
}
