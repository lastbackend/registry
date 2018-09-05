--------------------------------------------------------------------------------------------------
------------------------------------------- LB Migration ------------------------------------------
---------------------------------------------------------------------------------------------------

REVOKE usage ON schema public FROM public;

---------------------------------------------------------------------------------------------------
---------------------------------------- Drops all triggers ---------------------------------------
---------------------------------------------------------------------------------------------------

CREATE OR REPLACE FUNCTION strip_all_triggers()
  RETURNS TEXT AS $$ DECLARE
  triggNameRecord  RECORD;
  triggTableRecord RECORD;
BEGIN
  FOR triggNameRecord IN SELECT DISTINCT (trigger_name)
                         FROM information_schema.triggers
                         WHERE trigger_schema = 'public' LOOP
    FOR triggTableRecord IN SELECT DISTINCT (event_object_table)
                            FROM information_schema.triggers
                            WHERE trigger_name = triggNameRecord.trigger_name LOOP
      RAISE NOTICE 'Dropping trigger: % on table: %', triggNameRecord.trigger_name, triggTableRecord.event_object_table;
      EXECUTE 'DROP TRIGGER ' || triggNameRecord.trigger_name || ' ON ' || triggTableRecord.event_object_table || ';';
    END LOOP;
  END LOOP;

  RETURN 'done';
END;
$$
LANGUAGE plpgsql
SECURITY DEFINER;

SELECT strip_all_triggers();

---------------------------------------------------------------------------------------------------
---------------------------------------- Drops all tables -----------------------------------------
---------------------------------------------------------------------------------------------------

DROP SCHEMA public CASCADE;

---------------------------------------------------------------------------------------------------
---------------------------------------- Drops extensions -----------------------------------------
---------------------------------------------------------------------------------------------------

DROP EXTENSION IF EXISTS "uuid-ossp" CASCADE;

---------------------------------------------------------------------------------------------------
--------------------------------------- Creates schema ----------------------------------------
---------------------------------------------------------------------------------------------------

CREATE SCHEMA public;

---------------------------------------------------------------------------------------------------
--------------------------------------- Creates extensions ----------------------------------------
---------------------------------------------------------------------------------------------------

CREATE EXTENSION "uuid-ossp" WITH SCHEMA public;

---------------------------------------------------------------------------------------------------
----------------------------------------- Creates infrastructure tables ---------------------------
---------------------------------------------------------------------------------------------------

CREATE TABLE builders (
  id       UUID PRIMARY KEY             NOT NULL    DEFAULT uuid_generate_v4(),
  hostname VARCHAR(512)                 NOT NULL,
  ip       VARCHAR(512)                 NOT NULL,
  port     INTEGER                      NOT NULL,
  online   BOOLEAN                                  DEFAULT FALSE,
  tls      BOOLEAN                                  DEFAULT FALSE,
  ssl      JSONB                                    DEFAULT NULL,
  created  TIMESTAMPTZ                              DEFAULT (now() AT TIME ZONE 'utc'),
  updated  TIMESTAMPTZ                              DEFAULT (now() AT TIME ZONE 'utc')
);

CREATE TABLE images (
  id          UUID PRIMARY KEY             NOT NULL    DEFAULT uuid_generate_v4(),
  owner       VARCHAR(256)                 NOT NULL,
  name        VARCHAR(256)                 NOT NULL,
  private     BOOLEAN                                  DEFAULT FALSE,
  description TEXT                                     DEFAULT '',
  stats       JSONB                                    DEFAULT '{}',
  created     TIMESTAMPTZ                              DEFAULT (now() AT TIME ZONE 'utc'),
  updated     TIMESTAMPTZ                              DEFAULT (now() AT TIME ZONE 'utc')
);

CREATE TABLE images_tags (
  id       UUID PRIMARY KEY             NOT NULL    DEFAULT uuid_generate_v4(),
  image_id UUID                         NOT NULL,
  name     VARCHAR(256)                             DEFAULT '',
  created  TIMESTAMPTZ                              DEFAULT (now() AT TIME ZONE 'utc'),
  updated  TIMESTAMPTZ                              DEFAULT (now() AT TIME ZONE 'utc')
);

CREATE TABLE images_builds (
  id               UUID PRIMARY KEY             NOT NULL    DEFAULT uuid_generate_v4(),
  image_id         UUID                         NOT NULL,
  builder_id       UUID                                     DEFAULT NULL,
  number           INTEGER                                  DEFAULT 0,
  size             INTEGER                                  DEFAULT 0,
  source           JSONB                                    DEFAULT '{}',
  image            JSONB                                    DEFAULT '{}',
  config           JSONB                                    DEFAULT '{}',
  state_step       VARCHAR(32)                              DEFAULT '',
  state_status     VARCHAR(32)                              DEFAULT '',
  state_message    TEXT                                     DEFAULT '',
  state_processing BOOLEAN                                  DEFAULT FALSE,
  state_done       BOOLEAN                                  DEFAULT FALSE,
  state_error      BOOLEAN                                  DEFAULT FALSE,
  state_canceled   BOOLEAN                                  DEFAULT FALSE,
  state_started    TIMESTAMPTZ                              DEFAULT NULL,
  state_finished   TIMESTAMPTZ                              DEFAULT NULL,
  created          TIMESTAMPTZ                              DEFAULT (now() AT TIME ZONE 'utc'),
  updated          TIMESTAMPTZ                              DEFAULT (now() AT TIME ZONE 'utc')
);

CREATE TABLE systems (
  id           boolean PRIMARY KEY             NOT NULL    DEFAULT TRUE,
  access_token VARCHAR(512)                                DEFAULT '',
  auth_server  TEXT                                        DEFAULT '',
  created      TIMESTAMPTZ                                 DEFAULT (now() AT TIME ZONE 'utc'),
  updated      TIMESTAMPTZ                                 DEFAULT (now() AT TIME ZONE 'utc'),
  CONSTRAINT only_one_row CHECK (id = TRUE)
);

---------------------------------------------------------------------------------------------------
-------------------------------------- Creates foreign keys ---------------------------------------
---------------------------------------------------------------------------------------------------

ALTER TABLE images_builds
  ADD FOREIGN KEY (image_id) REFERENCES images (id) ON DELETE CASCADE;
ALTER TABLE images_tags
  ADD FOREIGN KEY (image_id) REFERENCES images (id) ON DELETE CASCADE;

---------------------------------------------------------------------------------------------------
--------------------------------------- Creates procedures ----------------------------------------
---------------------------------------------------------------------------------------------------

----------------------------------------- STORE PROCEDURE -----------------------------------------

----------------------------------------- TRIGGER PROCEDURE ----------------------------------------


CREATE OR REPLACE FUNCTION lb_after_images_builds_function()
  RETURNS TRIGGER AS
$$
BEGIN
  IF TG_OP = 'INSERT'
  THEN

    IF NOT EXISTS(SELECT FALSE FROM images_tags AS it WHERE it.image_id = NEW.image_id
                                                        AND it.name = NEW.image ->> 'tag')
    THEN
      INSERT INTO images_tags (image_id, name) VALUES (NEW.image_id, NEW.image ->> 'tag');
    END IF;

    RETURN NEW;
  ELSIF TG_OP = 'UPDATE'
    THEN
      RETURN NEW;
  ELSE
    RETURN NEW;
  END IF;
END;
$$
LANGUAGE plpgsql;

---------------------------------------------------------------------------------------------------
---------------------------------------- Creates triggers -----------------------------------------
---------------------------------------------------------------------------------------------------

CREATE CONSTRAINT TRIGGER lb_after_images_builds_change
  AFTER INSERT OR UPDATE
  ON images_builds
  DEFERRABLE
  FOR EACH ROW EXECUTE PROCEDURE lb_after_images_builds_function();

---------------------------------------------------------------------------------------------------
---------------------------------------- Creates default records  ---------------------------------
---------------------------------------------------------------------------------------------------

INSERT INTO systems (access_token, auth_server)
VALUES ('', '');

---------------------------------------------------------------------------------------------------
---------------------------------------- Creates rules  -------------------------------------------
---------------------------------------------------------------------------------------------------

CREATE RULE systems_rule AS ON DELETE TO systems DO INSTEAD NOTHING
