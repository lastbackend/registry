---------------------------------------------------------------------------------------------------
------------------------------------------- LB Migration ------------------------------------------
---------------------------------------------------------------------------------------------------

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

CREATE EXTENSION "uuid-ossp";

---------------------------------------------------------------------------------------------------
----------------------------------------- Creates infrastructure tables ---------------------------
---------------------------------------------------------------------------------------------------

CREATE TABLE repositories (
  id          UUID PRIMARY KEY             NOT NULL    DEFAULT uuid_generate_v4(),
  owner       VARCHAR(256)                 NOT NULL,
  name        VARCHAR(256)                 NOT NULL,
  technology  VARCHAR(256)                             DEFAULT '',
  description TEXT                                     DEFAULT '',
  readme      TEXT                                     DEFAULT '',
  self_link   TEXT                                     DEFAULT '',
  sources     JSONB                        NOT NULL,
  labels      JSONB                                    DEFAULT '{}',
  remote      BOOLEAN                                  DEFAULT FALSE,
  deleted     BOOLEAN                                  DEFAULT FALSE,
  created     TIMESTAMPTZ                              DEFAULT (now() AT TIME ZONE 'utc'),
  updated     TIMESTAMPTZ                              DEFAULT (now() AT TIME ZONE 'utc')
);

CREATE TABLE repositories_tags (
  id             UUID PRIMARY KEY     NOT NULL  DEFAULT uuid_generate_v4(),
  repo_id        UUID                 NOT NULL,
  name           VARCHAR(256)                   DEFAULT '',
  spec           JSONB                          DEFAULT '{}',
  build_total    INTEGER                        DEFAULT 0,
  build_size     INTEGER                        DEFAULT 0,
  build_id_0     UUID,
  build_status_0 VARCHAR(256)                   DEFAULT '',
  build_number_0 INTEGER                        DEFAULT 0,
  build_id_1     UUID,
  build_status_1 VARCHAR(256)                   DEFAULT '',
  build_number_1 INTEGER                        DEFAULT 0,
  build_id_2     UUID,
  build_status_2 VARCHAR(256)                   DEFAULT '',
  build_number_2 INTEGER                        DEFAULT 0,
  build_id_3     UUID,
  build_status_3 VARCHAR(256)                   DEFAULT '',
  build_number_3 INTEGER                        DEFAULT 0,
  build_id_4     UUID,
  build_status_4 VARCHAR(256)                   DEFAULT '',
  build_number_4 INTEGER                        DEFAULT 0,
  disabled       BOOLEAN                        DEFAULT FALSE,
  created        TIMESTAMPTZ                    DEFAULT (now() AT TIME ZONE 'utc'),
  updated        TIMESTAMPTZ                    DEFAULT (now() AT TIME ZONE 'utc')
);

CREATE TABLE repositories_builds
(
  id               UUID PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
  repo_id          UUID             NOT NULL,
  tag_id           UUID             NOT NULL,
  builder_id       UUID                      DEFAULT NULL,
  task_id          UUID                      DEFAULT NULL,
  self_link        TEXT                      DEFAULT '',
  number           INTEGER                   DEFAULT 0,
  size             INTEGER                   DEFAULT 0,
  sources          JSONB                     DEFAULT '{}',
  config           JSONB                     DEFAULT '{}',
  image            JSONB                     DEFAULT '{}',
  state_step       VARCHAR(32)               DEFAULT '',
  state_status     VARCHAR(32)               DEFAULT '',
  state_message    TEXT                      DEFAULT '',
  state_processing BOOLEAN                   DEFAULT FALSE,
  state_done       BOOLEAN                   DEFAULT FALSE,
  state_error      BOOLEAN                   DEFAULT FALSE,
  state_canceled   BOOLEAN                   DEFAULT FALSE,
  state_started    TIMESTAMPTZ               DEFAULT NULL,
  state_finished   TIMESTAMPTZ               DEFAULT NULL,
  created          TIMESTAMPTZ               DEFAULT (now() AT TIME ZONE 'utc'),
  updated          TIMESTAMPTZ               DEFAULT (now() AT TIME ZONE 'utc')
);

---------------------------------------------------------------------------------------------------
--------------------------------------- Creates procedures ----------------------------------------
---------------------------------------------------------------------------------------------------

-- Repositories triggers --

CREATE OR REPLACE FUNCTION lb_before_builds_function()
  RETURNS TRIGGER AS
$$
BEGIN

  IF NEW.task_id :: VARCHAR = ''
  THEN
    NEW.task_id := NULL;
  END IF;

  RETURN NEW;
END;
$$
LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION lb_after_builds_function()
  RETURNS TRIGGER AS
$$
BEGIN
  IF TG_OP = 'INSERT'
  THEN


    UPDATE repositories_tags
    SET
      build_total = build_total + 1
    WHERE id = NEW.tag_id;

    RETURN NEW;

  ELSIF TG_OP = 'UPDATE'
    THEN
      RETURN NEW;
  ELSIF TG_OP = 'DELETE'
    THEN

      UPDATE repositories_tags
      SET
        build_total = build_total - 1
      WHERE id = OLD.tag_id;

      RETURN OLD;
  ELSE
    RETURN NEW;
  END IF;
END;
$$
LANGUAGE plpgsql;

---------------------------------------------------------------------------------------------------
---------------------------------------- Creates triggers -----------------------------------------
---------------------------------------------------------------------------------------------------

-- Run trigger after_service_specs_change after transaction, not on each row
-- The trigger has to be a constraint and "after" trigger. Then you can use DEFERRABLE

-- repositories triggers
CREATE TRIGGER lb_before_builds_change
  BEFORE INSERT OR UPDATE
  ON repositories_builds
  FOR EACH ROW EXECUTE PROCEDURE lb_before_builds_function();

CREATE CONSTRAINT TRIGGER lb_after_builds_change
  AFTER INSERT OR UPDATE OR DELETE
  ON repositories_builds
  DEFERRABLE
  FOR EACH ROW EXECUTE PROCEDURE lb_after_builds_function();