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

func (s *RepoStorage) Get(ctx context.Context, owner, name string) (*types.Repo, error) {

	log.V(logLevel).Debugf("%s:repo:get:> get repo `%s/%s`", logPrefix, owner, name)

	if len(owner) == 0 {
		err := errors.New("owner can not be empty")
		log.V(logLevel).Errorf("%s:repo:get:> get repo err: %v", logPrefix, err)
		return nil, err
	}

	if len(name) == 0 {
		err := errors.New("name can not be empty")
		log.V(logLevel).Errorf("%s:repo:get:> get repo err: %v", logPrefix, err)
		return nil, err
	}

	const query = `
      SELECT to_jsonb(
          jsonb_build_object(
              'meta', jsonb_build_object(
                  'id', r.id,
                  'owner', r.owner,
                  'name', r.name,
                  'description', r.description,
                  'labels', r.labels,
                  'self_link', r.self_link,
                  'created', r.created,
                  'updated', r.updated
              ),
              'sources', r.sources,
              'remote', r.remote,
              'readme', r.readme,
              'tags', tags
          )
       )
       FROM (
       SELECT
         *,
         (SELECT jsonb_build_object(
             'id', id,
             'tag', (SELECT name
                     FROM repositories_tags
                     WHERE id = tag_id),
             'number', number
         )
          FROM repositories_builds
          WHERE repo_id = repositories.id
          ORDER BY created DESC
          LIMIT 1)                                         AS build,
         (SELECT COALESCE(json_object_agg(name, tags), '{}')
          FROM (
                 SELECT
                   repo_id       AS repo,
                   name,
                   spec,
                   disabled,
                   updated,
                   created,
                   CASE WHEN build_id_0 :: TEXT <> ''
                     THEN
                       jsonb_build_object(
                           'id', build_id_0,
                           'status', build_status_0,
                           'number', build_number_0
                       )
                   ELSE NULL END AS build_0,

                   CASE WHEN build_id_1 :: TEXT <> ''
                     THEN
                       jsonb_build_object(
                           'id', build_id_1,
                           'status', build_status_1,
                           'number', build_number_1
                       )
                   ELSE NULL END AS build_1,

                   CASE WHEN build_id_2 :: TEXT <> ''
                     THEN
                       jsonb_build_object(
                           'id', build_id_2,
                           'status', build_status_2,
                           'number', build_number_2
                       )
                   ELSE NULL END AS build_2,

                   CASE WHEN build_id_3 :: TEXT <> ''
                     THEN
                       jsonb_build_object(
                           'id', build_id_3,
                           'status', build_status_3,
                           'number', build_number_3
                       )
                   ELSE NULL END AS build_3,

                   CASE WHEN build_id_4 :: TEXT <> ''
                     THEN
                       jsonb_build_object(
                           'id', build_id_4,
                           'status', build_status_4,
                           'number', build_number_4
                       )
                   ELSE NULL END AS build_4
                 FROM repositories_tags AS rt
                 WHERE rt.repo_id = repositories.id) tags) AS tags
       FROM repositories
     ) AS r
     WHERE owner = $1 AND name = $2 AND deleted = FALSE;`

	var buf string

	err := getClient(ctx).QueryRowContext(ctx, query, owner, name).Scan(&buf)
	switch err {
	case nil:
	case sql.ErrNoRows:
		return nil, nil
	default:
		log.V(logLevel).Errorf("%s:repo:get:> get repo err: %v", logPrefix, err)
		return nil, err
	}

	r := new(types.Repo)

	if err := json.Unmarshal([]byte(buf), &r); err != nil {
		return nil, err
	}

	return r, nil
}

func (s *RepoStorage) List(ctx context.Context) (map[string]*types.Repo, error) {

	log.V(logLevel).Debug("%s:repo:list:> get repositories list", logPrefix)

	const query = `
		SELECT COALESCE(to_json(json_object_agg(id, repositories)), '{}')
		FROM (
       SELECT
         r.id,
         jsonb_build_object(
					 'id', r.id,
					 'owner', r.owner,
					 'name', r.name,
					 'labels', r.labels,
					 'description', r.description,
					 'self_link', r.self_link,
					 'created', r.created,
					 'updated', r.updated
         ) AS meta,
         r.sources,
         r.readme,
         r.remote,
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
									updated,
									created,
                   CASE WHEN build_id_0 :: TEXT <> ''
                     THEN
                       jsonb_build_object(
                           'id', build_id_0,
                           'status', build_status_0,
                           'number', build_number_0
                       )
                   ELSE NULL END AS build_0,

                   CASE WHEN build_id_1 :: TEXT <> ''
                     THEN
                       jsonb_build_object(
                           'id', build_id_1,
                           'status', build_status_1,
                           'number', build_number_1
                       )
                   ELSE NULL END AS build_1,

                   CASE WHEN build_id_2 :: TEXT <> ''
                     THEN
                       jsonb_build_object(
                           'id', build_id_2,
                           'status', build_status_2,
                           'number', build_number_2
                       )
                   ELSE NULL END AS build_2,

                   CASE WHEN build_id_3 :: TEXT <> ''
                     THEN
                       jsonb_build_object(
                           'id', build_id_3,
                           'status', build_status_3,
                           'number', build_number_3
                       )
                   ELSE NULL END AS build_3,

                   CASE WHEN build_id_4 :: TEXT <> ''
                     THEN
                       jsonb_build_object(
                           'id', build_id_4,
                           'status', build_status_4,
                           'number', build_number_4
                       )
                   ELSE NULL END AS build_4
								FROM repositories_tags AS rt
								WHERE rt.repo_id = repositories.id) tags) AS tags
					FROM repositories
				) AS r
       WHERE deleted = FALSE) repositories`

	var buf string

	err := getClient(ctx).QueryRowContext(ctx, query).Scan(&buf)
	switch err {
	case nil:
	case sql.ErrNoRows:
		return nil, nil
	default:
		log.V(logLevel).Errorf("%s:repo:list:> get repo err: %v", logPrefix, err)
		return nil, err
	}

	r := make(map[string]*types.Repo)

	if err := json.Unmarshal([]byte(buf), &r); err != nil {
		return nil, err
	}

	return r, nil
}

func (s *RepoStorage) Insert(ctx context.Context, repo *types.Repo) error {

	log.V(logLevel).Debugf("%s:repo:insert:> insert repo: %#v", logPrefix, repo)

	if repo == nil {
		err := errors.New("repo can not be empty")
		log.V(logLevel).Errorf("%s:repo:insert:> insert repo err: %v", logPrefix, err)
		return err
	}

	sources, err := json.Marshal(repo.Sources)
	if err != nil {
		log.Errorf("%s:repo:insert:> prepare sources struct to database write: %v", logPrefix, err)
		sources = []byte("{}")
	}

	labels, err := json.Marshal(repo.Meta.Labels)
	if err != nil {
		log.Errorf("%s:repo:insert:> prepare labels struct to database write: %v", logPrefix, err)
		labels = []byte("{}")
	}

	const query = `
    INSERT INTO repositories(owner, name, description, sources, readme, remote, self_link, labels)
		VALUES ($1, $2, lower($3), $4, $5, $6, $7, $8)
   	RETURNING id, created, updated;`

	err = getClient(ctx).QueryRowContext(ctx, query, repo.Meta.Owner, repo.Meta.Name, repo.Meta.Description,
		string(sources), repo.Readme, repo.Remote, repo.Meta.SelfLink, string(labels)).
		Scan(&repo.Meta.ID, &repo.Meta.Created, &repo.Meta.Updated)
	if err != nil {
		log.V(logLevel).Errorf("%s:repo:insert:> insert repo meta err: %v", logPrefix, err)
		return err
	}

	return nil
}

func (s *RepoStorage) Update(ctx context.Context, repo *types.Repo) error {

	log.V(logLevel).Debugf("%s:repo:update:> update repo %#v", logPrefix, repo)

	const query = `
		UPDATE repositories
		SET
			description = $2,
			labels = $3,
			updated = now() at time zone 'utc'
		WHERE id = $1
		RETURNING updated;`

	labels, err := json.Marshal(repo.Meta.Labels)
	if err != nil {
		log.Errorf("%s:repo:insert:> prepare labels struct to database write: %v", logPrefix, err)
		labels = []byte("{}")
	}

	err = getClient(ctx).QueryRowContext(ctx, query, repo.Meta.ID, repo.Meta.Description, string(labels)).
		Scan(&repo.Meta.Updated)
	if err != nil {
		log.V(logLevel).Errorf("%s:repo:update:> update repo query err: %v", logPrefix, err)
		return err
	}

	const queryUpsertTags = `
		WITH upsert AS (
			UPDATE repositories_tags
			SET
				name      = $2,
				spec      = $3,
				disabled  = $4,
				updated   = now() AT TIME ZONE 'utc'
			WHERE repo_id = $1 AND name = $2
			RETURNING updated
		)
		INSERT INTO repositories_tags (repo_id, name, spec, disabled)
		SELECT
			$1 AS repo_id,
			$2 AS name,
			$3 AS spec,
			$4 AS disabled
		WHERE NOT EXISTS(SELECT * FROM upsert)
		RETURNING updated;`

	for _, tag := range repo.Tags {

		spec, err := json.Marshal(tag.Spec)
		if err != nil {
			log.Errorf("%s:repo:update:> prepare spec struct to database write: %v", logPrefix, err)
			spec = []byte("{}")
		}

		_, err = getClient(ctx).ExecContext(ctx, queryUpsertTags, tag.RepoID, tag.Name, string(spec), tag.Disabled)
		if err != nil {
			log.V(logLevel).Errorf("%s:repo:update:> update repo query err: %v", logPrefix, err)
			return err
		}
	}

	return nil
}

func (s *RepoStorage) Remove(ctx context.Context, id string) error {

	log.V(logLevel).Debugf("%s:repo:remove:> remove repo %s", id)

	const query = `
  	UPDATE repositories
		SET
			deleted = TRUE,
			updated = now() at time zone 'utc'
		WHERE id = $1
		RETURNING updated;`

	_, err := getClient(ctx).ExecContext(ctx, query, id)
	if err != nil {
		log.V(logLevel).Errorf("%s:repo:remove:> remove repo query err: %v", logPrefix, err)
		return err
	}

	return nil
}

func (s *RepoStorage) InsertTag(ctx context.Context, tag *types.RepoTag) error {

	log.V(logLevel).Debugf("%s:repo:insert_tag:> update build rules for repo %#v", logPrefix, tag.RepoID)

	const query = `
    INSERT INTO repositories_tags(repo_id, name, spec)
    	VALUES ($1, $2, $3);`

	spec, err := json.Marshal(tag.Spec)
	if err != nil {
		log.Errorf("%s:repo:insert_tag:> prepare spec struct to database write: %v", logPrefix, err)
		spec = []byte("{}")
	}
	_, err = getClient(ctx).ExecContext(ctx, query, tag.RepoID, tag.Name, string(spec))
	if err != nil {
		log.V(logLevel).Errorf("%s:repo:insert_tag:> insert repo tag err: %v", logPrefix, err)
		return err
	}

	return nil
}

func (s *RepoStorage) UpdateTag(ctx context.Context, repo, tag string) error {

	log.V(logLevel).Debugf("%s:repo:update_tag:> update build rules for repo %#v", logPrefix, repo)

	if repo == "" {
		err := errors.New("repo can not be empty")
		log.V(logLevel).Errorf("%s:repo:> update tags for repo err: %v", logPrefix, err)
		return err
	}

	if tag == "" {
		err := errors.New("tag can not be empty")
		log.V(logLevel).Errorf("%s:repo:update_tag:> update tags for repo err: %v", logPrefix, err)
		return err
	}

	const queryListTop5Builds = `
		SELECT rb.id, rb.number, rb.state_status, rb.size
		FROM repositories_builds AS rb
			INNER JOIN repositories_tags AS rt ON rt.id = rb.tag_id
		WHERE rb.repo_id = $1 AND rt.name = $2
		ORDER BY rb.created DESC
		LIMIT 5;`

	rows, err := getClient(ctx).QueryContext(ctx, queryListTop5Builds, repo, tag)
	switch err {
	case nil:
	case sql.ErrNoRows:
		return nil
	default:
		log.V(logLevel).Errorf("%s:repo:update_tag:> get repositories list err: %v", logPrefix, err)
		return err
	}

	builds := [5]struct {
		ID          string
		BuildNumber int64
		Status      string
		Size        int64
	}{}

	var (
		i    = int(0)
		size = int(0)
	)

	for rows.Next() {
		err := rows.Scan(&builds[i].ID, &builds[i].BuildNumber, &builds[i].Status, &builds[i].Size)
		if err != nil {
			log.V(logLevel).Errorf("%s:repo:update_tag:> get last 5 builds for repo err: %v", logPrefix, err)
			return err
		}

		if size == 0 {
			size = int(builds[i].Size)
		}

		i++
	}
	if err := rows.Close(); err != nil {
		log.V(logLevel).Errorf("%s:repo:update_tag:> close rows err: %v", logPrefix, err)
		return err
	}

	if len(builds) == 0 {
		return nil
	}

	const queryUpdateTag = `
		UPDATE repositories_tags
		SET
			build_size = $3,
			build_id_0 = CASE WHEN $4 <> '' THEN $4::UUID ELSE NULL END,
			build_status_0 = $5,
			build_number_0 = $6,
			build_id_1 = CASE WHEN $7 <> '' THEN $7::UUID ELSE NULL END,
			build_status_1 = $8,
			build_number_1 = $9,
			build_id_2 = CASE WHEN $10 <> '' THEN $10::UUID ELSE NULL END,
			build_status_2 = $11,
			build_number_2 = $12,
			build_id_3 = CASE WHEN $13 <> '' THEN $13::UUID ELSE NULL END,
			build_status_3 = $14,
			build_number_3 = $15,
			build_id_4 = CASE WHEN $16 <> '' THEN $16::UUID ELSE NULL END,
			build_status_4 = $17,
			build_number_4 = $18,
			updated = now() at time zone 'utc'
		WHERE repo_id = $1 AND name = $2;`

	_, err = getClient(ctx).ExecContext(ctx, queryUpdateTag, repo, tag, size,
		builds[0].ID, builds[0].Status, builds[0].BuildNumber,
		builds[1].ID, builds[1].Status, builds[1].BuildNumber,
		builds[2].ID, builds[2].Status, builds[2].BuildNumber,
		builds[3].ID, builds[3].Status, builds[3].BuildNumber,
		builds[4].ID, builds[4].Status, builds[4].BuildNumber)

	if err != nil {
		log.V(logLevel).Errorf("%s:repo:update_tag:> update tag err: %v", logPrefix, err)
		return err
	}

	return nil
}

func newRepoStorage() *RepoStorage {
	s := new(RepoStorage)
	return s
}
