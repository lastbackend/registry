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
	"fmt"
	"reflect"
	"strings"

	"github.com/lastbackend/lastbackend/pkg/log"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/storage/storage"
	"github.com/lastbackend/registry/pkg/storage/types/filter"
)

const (
	logImagePrefix = "storage:pgsql:image"
)

type ImageStorage struct {
	storage.Image
}

func (s *ImageStorage) Get(ctx context.Context, owner, name string) (*types.Image, error) {

	log.V(logLevel).Debugf("%s:get:> get image `%s/%s`", logImagePrefix, owner, name)

	if len(owner) == 0 {
		err := errors.New("owner can not be empty")
		log.V(logLevel).Errorf("%s:get:> get image err: %v", logImagePrefix, err)
		return nil, err
	}

	if len(name) == 0 {
		err := errors.New("name can not be empty")
		log.V(logLevel).Errorf("%s:get:> get image err: %v", logImagePrefix, err)
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
          'stats', i.stats,
          'private', i.private
        ),
				'tags', tags
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
		log.V(logLevel).Errorf("%s:get:> get image err: %v", logImagePrefix, err)
		return nil, err
	}

	i := new(types.Image)

	if err := json.Unmarshal([]byte(buf), &i); err != nil {
		return nil, err
	}

	return i, nil
}

func (s *ImageStorage) List(ctx context.Context, f *filter.ImageFilter) ([]*types.Image, error) {

	log.V(logLevel).Debugf("%s:list:> get images list", logImagePrefix)

	var values = make([]interface{}, 0)
	var where = make([]string, 0)
	var whereCondition = types.EmptyString

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
				where = append(where, fmt.Sprintf("%s=$%d", name, i+1))
			case reflect.Bool:
				where = append(where, fmt.Sprintf("%s=$%d", name, i+1))
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

	var query = fmt.Sprintf(`
    SELECT COALESCE(json_agg(
      json_build_object(
          'meta', jsonb_build_object(
          'id', tmp.id,
          'owner', tmp.owner,
          'name', tmp.name,
          'description', tmp.description,
          'created', tmp.created,
          'updated', tmp.updated
        ),
          'status', jsonb_build_object(
              'last_build', tmp.build,
              'stats', tmp.stats
            ),
          'spec', jsonb_build_object(
              'private', tmp.private,
              'tags', tmp.tags
            )
        )), '[]')
      FROM (SELECT
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
       LIMIT 1)  AS build,
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
    %s) AS tmp;`, whereCondition)

	var buf string

	err := getClient(ctx).QueryRow(query).Scan(&buf)
	switch err {
	case nil:
	case sql.ErrNoRows:
		return nil, nil
	default:
		log.V(logLevel).Errorf("%s:list:> get images list err: %v", logBuilderPrefix, err)
		return nil, err
	}

	i := make([]*types.Image, 0)

	if err := json.Unmarshal([]byte(buf), &i); err != nil {
		return nil, err
	}
	return i, nil
}

func (s *ImageStorage) Insert(ctx context.Context, image *types.Image) error {

	log.V(logLevel).Debugf("%s:insert:> insert image: %#v", logImagePrefix, image)

	if image == nil {
		err := errors.New("image can not be empty")
		log.V(logLevel).Errorf("%s:insert:> insert image err: %v", logImagePrefix, err)
		return err
	}

	const query = `
    INSERT INTO images(owner, name, private, description)
		VALUES ($1, $2, $3, $4)
   	RETURNING id, created, updated;`

	err := getClient(ctx).QueryRowContext(ctx, query,
		image.Meta.Owner,
		image.Meta.Name,
		image.Status.Private,
		image.Meta.Description,
	).
		Scan(&image.Meta.ID, &image.Meta.Created, &image.Meta.Updated)
	if err != nil {
		log.V(logLevel).Errorf("%s:insert:> insert image err: %v", logImagePrefix, err)
		return err
	}

	return nil
}

func (s *ImageStorage) Update(ctx context.Context, image *types.Image) error {

	log.V(logLevel).Debugf("%s:update:> update image %#v", logImagePrefix, image)

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
		image.Status.Private,
	).
		Scan(&image.Meta.Updated)
	if err != nil {
		log.V(logLevel).Errorf("%s:update:> update image query err: %v", logImagePrefix, err)
		return err
	}

	return nil
}

func (s *ImageStorage) Remove(ctx context.Context, image *types.Image) error {

	log.V(logLevel).Debugf("%s:remove:> remove image %s/%s", logImagePrefix, image.Meta.Owner, image.Meta.Name)

	const query = `
		DELETE FROM images 
		WHERE id = $1;`

	result, err := getClient(ctx).ExecContext(ctx, query, image.Meta.ID)
	if err != nil {
		log.V(logLevel).Errorf("%s:remove:> remove image query err: %v", logImagePrefix, err)
		return err
	}

	if _, err := result.RowsAffected(); err != nil {
		log.V(logLevel).Errorf("%s:remove:> check query affected err: %v", logImagePrefix, err)
		return err
	}

	return nil
}

func newImageStorage() *ImageStorage {
	s := new(ImageStorage)
	return s
}
