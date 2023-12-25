BEGIN TRANSACTION;

CREATE TABLE IF NOT EXISTS links_couples
(
    uuid         INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    short_url    VARCHAR(250)  NOT NULL DEFAULT '',
    original_url VARCHAR(1000) NOT NULL DEFAULT ''
);

CREATE UNIQUE INDEX IF NOT EXISTS links_couples_index_by_original_url_unique
    ON links_couples
        USING btree (original_url);

COMMIT;