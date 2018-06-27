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

type RepoStorage struct {
	storage.Repo
}

func (s *RepoStorage) Get(ctx context.Context, id string) (*types.Repo, error) {

	log.V(logLevel).Debugf("Storage: Repo: GetByID: get repo by id: %s", id)
	if len(id) == 0 {
		err := errors.New("id can not be empty")
		log.V(logLevel).Errorf("Storage: Repo: GetByID: get repo err: %s", err)
		return nil, err
	}

	const query = `
		SELECT to_jsonb(
			jsonb_build_object(
				'meta', jsonb_build_object(
					'id', r.id,
					'account', r.account_id,
					'owner', r.owner,
					'name', r.name,
					'technology', r.technology,
					'technical', r.technical,
					'description', r.description,
					'self_link', r.self_link,
					'created', r.created,
					'updated', r.updated
				),
				'sources', r.sources,
				'readme', r.readme,
				'remote', r.remote,
				'private', r.private,
				'technical', r.technical,
				'stats', jsonb_build_object(
				'builds', r.stats_builds,
				'pulls', r.stats_pulls,
				'services', r.stats_services,
				'stars', r.stats_stars,
				'views', r.stats_views
				),
				'last_build', build,
				'tags', tags
			)
		)
	 FROM (
		SELECT
			*,
		 (SELECT jsonb_build_object(
				'id', id,
				'tag', (SELECT name FROM repositories_tags WHERE id = tag_id),
				'number', number
			)
			FROM repositories_builds
			WHERE repo_id = repositories.id
			ORDER BY created DESC
			LIMIT 1) AS build,
			(SELECT COALESCE(json_object_agg(name, tags), '{}')
			 FROM (
				SELECT
					repo_id AS repo,
					name,
					spec,
					disabled,
					autobuild,
					updated,
					created,
					jsonb_build_object(
							'id', build_id_0,
							'status', build_status_0,
							'number', build_number_0
					)       AS build_0,
					jsonb_build_object(
							'id', build_id_1,
							'status', build_status_1,
							'number', build_number_1
					)       AS build_1,
					jsonb_build_object(
							'id', build_id_2,
							'status', build_status_2,
							'number', build_number_2
					)       AS build_2,
					jsonb_build_object(
							'id', build_id_3,
							'status', build_status_3,
							'number', build_number_3
					)       AS build_3,
					jsonb_build_object(
							'id', build_id_4,
							'status', build_status_4,
							'number', build_number_4
					)       AS build_4
				FROM repositories_tags AS rt
				WHERE rt.repo_id = repositories.id) tags) AS tags
			FROM repositories
		) AS r
		WHERE id = $1 AND deleted = FALSE;`

	var buf string

	err := getClient(ctx).QueryRow(query, id).Scan(&buf)
	switch err {
	case nil:
	case sql.ErrNoRows:
		return nil, nil
	default:
		log.V(logLevel).Errorf("Storage: Repo: GetByName: get repo err: %s", err)
		return nil, err
	}

	r := new(types.Repo)

	if err := json.Unmarshal([]byte(buf), &r); err != nil {
		return nil, err
	}

	return r, nil
}

func (s *RepoStorage) List(ctx context.Context) (map[string]*types.Repo, error) {

	log.V(logLevel).Debug("Storage: Repo: List: get repositories list")

	const query = `
		SELECT COALESCE(to_json(json_object_agg(id, repositories)), '{}')
		FROM (
       SELECT
         r.id,
         jsonb_build_object(
					 'id', r.id,
					 'account', r.account_id,
					 'owner', r.owner,
					 'name', r.name,
					 'technology', r.technology,
					 'technical', r.technical,
					 'description', r.description,
					 'self_link', r.self_link,
					 'created', r.created,
					 'updated', r.updated
         ) AS meta,
         jsonb_build_object(
					 'builds', r.stats_builds,
					 'pulls', r.stats_pulls,
					 'services', r.stats_services,
					 'stars', r.stats_stars,
					 'views', r.stats_views
         ) AS stats,
         r.sources,
         r.readme,
         r.remote,
         r.private,
         r.technical,
         tags
       FROM (
					SELECT
						*,
						(SELECT COALESCE(json_object_agg(name, tags), '{}')
						 FROM (
								SELECT
									repo_id AS repo,
									name,
									spec,
									disabled,
									autobuild,
									updated,
									created,
									jsonb_build_object(
										'id', build_id_0,
										'status', build_status_0,
										'number', build_number_0
									) AS build_0,
									jsonb_build_object(
										'id', build_id_1,
										'status', build_status_1,
										'number', build_number_1
									) AS build_1,
									jsonb_build_object(
										'id', build_id_2,
										'status', build_status_2,
										'number', build_number_2
									) AS build_2,
									jsonb_build_object(
										'id', build_id_3,
										'status', build_status_3,
										'number', build_number_3
									) AS build_3,
									jsonb_build_object(
										'id', build_id_4,
										'status', build_status_4,
										'number', build_number_4
									) AS build_4
								FROM repositories_tags AS rt
								WHERE rt.repo_id = repositories.id) tags) AS tags
					FROM repositories
				) AS r
         INNER JOIN repositories_acl AS al ON al.entity_id = r.id
       WHERE r.account_id = $2 AND (al.member_id = $1 OR r.private IS NOT TRUE) AND deleted = FALSE) repositories`

	var buf string

	err := getClient(ctx).QueryRow(query).Scan(&buf)
	switch err {
	case nil:
	case sql.ErrNoRows:
		return nil, nil
	default:
		log.V(logLevel).Errorf("Storage: Repo: GetByName: get repo err: %s", err)
		return nil, err
	}

	r := make(map[string]*types.Repo)

	if err := json.Unmarshal([]byte(buf), &r); err != nil {
		return nil, err
	}

	return r, nil
}

func (s *RepoStorage) Insert(ctx context.Context, repo *types.Repo) error {

	log.V(logLevel).Debugf("Storage: Repo: insert repo: %#v", repo)

	if repo == nil {
		err := errors.New("repo can not be empty")
		log.V(logLevel).Errorf("Storage: Repo: insert repo err: %s", err)
		return err
	}

	sources, err := json.Marshal(repo.Sources)
	if err != nil {
		log.Errorf("Storage: Repo: prepare sources struct to database write: %s", err)
		sources = []byte("{}")
	}

	const query_repo = `
    INSERT INTO repositories(account_id, owner, name, technology, sources, description, 
			readme, remote, private, technical, self_link)
		VALUES ($1, $2, lower($3), $4, $5, $6, $7, $8, $9, $10, $11)
   	RETURNING id, created, updated;`

	err = getClient(ctx).QueryRow(query_repo, repo.Meta.AccountID, repo.Meta.Owner, repo.Meta.Name,
		repo.Meta.Technology, string(sources), repo.Meta.Description, repo.Readme, repo.Remote,
		repo.Private, repo.Technical, repo.Meta.SelfLink).
		Scan(&repo.Meta.ID, &repo.Meta.Created, &repo.Meta.Updated)
	if err != nil {
		log.V(logLevel).Errorf("Storage: Repo: Insert: insert repo meta err: %s", err)
		return err
	}

	return nil
}

func (s *RepoStorage) Update(ctx context.Context, repo *types.Repo) error {

	log.V(logLevel).Debugf("Storage: Repo: Update: update repo %#v", repo)

	const query = `
		UPDATE repositories
		SET
			description = $2,
			technology = $3,
			private = $4,
			updated = now() at time zone 'utc'
		WHERE id = $1
		RETURNING updated;`

	err := getClient(ctx).QueryRow(query, repo.Meta.ID, repo.Meta.Description, repo.Meta.Technology, repo.Private).
		Scan(&repo.Meta.Updated)
	if err != nil {
		log.V(logLevel).Errorf("Storage: Repo: Update: update repo query err: %s", err)
		return err
	}

	const queryUpsertTags = `
		WITH upsert AS (
			UPDATE repositories_tags
			SET
				name      = $2,
				spec      = $3,
				disabled  = $4,
				autobuild = $5,
				updated   = now() AT TIME ZONE 'utc'
			WHERE repo_id = $1 AND name = $2
			RETURNING updated
		)
		INSERT INTO repositories_tags (repo_id, name, spec, disabled, autobuild)
		SELECT
			$1 AS repo_id,
			$2 AS name,
			$3 AS spec,
			$4 AS disabled,
			$5 AS autobuild
		WHERE NOT EXISTS(SELECT * FROM upsert)
		RETURNING updated;
`

	for _, tag := range repo.Tags {

		spec, err := json.Marshal(tag.Spec)
		if err != nil {
			log.Errorf("Storage: Repo: prepare spec struct to database write: %s", err)
			spec = []byte("{}")
		}

		_, err = getClient(ctx).Exec(queryUpsertTags, tag.RepoID, tag.Name, string(spec), tag.Disabled, tag.AutoBuild)
		if err != nil {
			log.V(logLevel).Errorf("Storage: Repo: Update: update repo query err: %s", err)
			return err
		}
	}

	return nil
}

func (s *RepoStorage) Remove(ctx context.Context, id string) error {

	log.V(logLevel).Debugf("Storage: Repo: Remove: remove repo %s", id)

	const query = `
  	UPDATE repositories
		SET
			deleted = TRUE,
			updated = now() at time zone 'utc'
		WHERE id = $1
		RETURNING updated;`

	_, err := getClient(ctx).Exec(query, id)
	if err != nil {
		log.V(logLevel).Errorf("Storage: Repo: Remove: remove repo query err: %s", err)
		return err
	}

	return nil
}

func newRepoStorage() *RepoStorage {
	s := new(RepoStorage)
	return s
}
