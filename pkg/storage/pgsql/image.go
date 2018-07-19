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

	"fmt"
	"github.com/lastbackend/registry/pkg/storage/types/filter"
)

type ImageStorage struct {
	storage.Image
}

func (s *ImageStorage) Get(ctx context.Context, owner, name string) (*types.Image, error) {

	log.V(logLevel).Debugf("%s:image:get:> get image `%s/%s`", logPrefix, owner, name)

	if len(owner) == 0 {
		err := errors.New("owner can not be empty")
		log.V(logLevel).Errorf("%s:image:get:> get image err: %v", logPrefix, err)
		return nil, err
	}

	if len(name) == 0 {
		err := errors.New("name can not be empty")
		log.V(logLevel).Errorf("%s:image:get:> get image err: %v", logPrefix, err)
		return nil, err
	}

	const query = `
		SELECT to_jsonb(
		  jsonb_build_object(
		    'meta', jsonb_build_object(
		      'id', i.id,
		      'owner', i.owner,
		      'name', i.name,
		      'description', i.description,
		      'created', i.created,
		      'updated', i.updated
		    ),
		    'status', jsonb_build_object(
          'last_build', build,
          'stats', i.stats
        ),
				'spec', jsonb_build_object(
				  'private', i.private,
		      'tags', tags
				)
		  )
		)
		FROM (
       SELECT
         *,
         (SELECT jsonb_build_object(
           'id', id,
           'number', number,
           'status', state_status,
           'updated', updated
         )
          FROM images_builds
          WHERE image_id = images.id
          ORDER BY created DESC
          LIMIT 1) AS build,
         (SELECT COALESCE(json_object_agg(name, tags), '{}')
          FROM (
            SELECT
              image_id AS image,
              name,
              updated,
              created
            FROM images_tags AS it
            WHERE it.image_id = images.id) tags) AS tags
       FROM images
    ) AS i
		WHERE owner = $1 AND name = $2;`

	var buf string

	err := getClient(ctx).QueryRow(query, owner, name).Scan(&buf)
	switch err {
	case nil:
	case sql.ErrNoRows:
		return nil, nil
	default:
		log.V(logLevel).Errorf("%s:image:get:> get image err: %v", logPrefix, err)
		return nil, err
	}

	i := new(types.Image)

	if err := json.Unmarshal([]byte(buf), &i); err != nil {
		return nil, err
	}

	return i, nil
}

func (s *ImageStorage) List(ctx context.Context, f *filter.ImageFilter) ([]*types.Image, error) {

	log.V(logLevel).Debug("%s:image:list:> get images list", logPrefix)

	where := ""
	if f != nil {
		if len(f.Owner) != 0 {
			where = fmt.Sprintf(`WHERE owner='%s'`, f.Owner)
		}
	}

	var query = fmt.Sprintf(`
		SELECT owner, name, description, created, updated
		FROM images
		%s
		ORDER BY created DESC;`, where)

	rows, err := getClient(ctx).QueryContext(ctx, query)
	if err != nil {
		log.V(logLevel).Errorf("%s:image:list:> get images list err: %v", logPrefix, err)
		return nil, err
	}
	defer rows.Close()

	i := make([]*types.Image, 0)

	for rows.Next() {
		item := new(types.Image)
		err := rows.Scan(
			&item.Meta.Owner,
			&item.Meta.Name,
			&item.Meta.Description,
			&item.Meta.Updated,
			&item.Meta.Created,
		)
		if err != nil {
			log.V(logLevel).Errorf("%s:image:list:> get games err: %v", logPrefix, err)
			return nil, err
		}

		i = append(i, item)
	}

	return i, nil
}

func (s *ImageStorage) Insert(ctx context.Context, image *types.Image) error {

	log.V(logLevel).Debugf("%s:image:insert:> insert image: %#v", logPrefix, image)

	if image == nil {
		err := errors.New("image can not be empty")
		log.V(logLevel).Errorf("%s:image:insert:> insert image err: %v", logPrefix, err)
		return err
	}

	const query = `
    INSERT INTO images(owner, name, private, description)
		VALUES ($1, $2, $3, $4)
   	RETURNING id, created, updated;`

	err := getClient(ctx).QueryRowContext(ctx, query,
		image.Meta.Owner,
		image.Meta.Name,
		image.Spec.Private,
		image.Meta.Description,
	).
		Scan(&image.Meta.ID, &image.Meta.Created, &image.Meta.Updated)
	if err != nil {
		log.V(logLevel).Errorf("%s:image:insert:> insert image err: %v", logPrefix, err)
		return err
	}

	return nil
}

func (s *ImageStorage) Update(ctx context.Context, image *types.Image) error {

	log.V(logLevel).Debugf("%s:image:update:> update image %#v", logPrefix, image)

	const query = `
		UPDATE images
		SET
			description = $2,
			private = $3,
			updated = now() at time zone 'utc'
		WHERE id = $1
		RETURNING updated;`

	err := getClient(ctx).QueryRowContext(ctx, query,
		image.Meta.ID,
		image.Meta.Description,
		image.Spec.Private,
	).
		Scan(&image.Meta.Updated)
	if err != nil {
		log.V(logLevel).Errorf("%s:image:update:> update image query err: %v", logPrefix, err)
		return err
	}

	return nil
}

func (s *ImageStorage) Remove(ctx context.Context, image *types.Image) error {

	log.V(logLevel).Debugf("%s:image:remove:> remove image %s/%s", logPrefix, image.Meta.Owner, image.Meta.Name)

	const query = `
		DELETE FROM images 
		WHERE id = $1;`

	result, err := getClient(ctx).ExecContext(ctx, query, image.Meta.ID)
	if err != nil {
		log.V(logLevel).Errorf("%s:image:remove:> remove image query err: %v", logPrefix, err)
		return err
	}

	if _, err := result.RowsAffected(); err != nil {
		log.V(logLevel).Errorf("%s:image:remove:> check query affected err: %v", logPrefix, err)
		return err
	}

	return nil
}

func newImageStorage() *ImageStorage {
	s := new(ImageStorage)
	return s
}
