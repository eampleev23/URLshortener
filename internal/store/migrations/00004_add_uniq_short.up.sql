BEGIN TRANSACTION;
CREATE UNIQUE INDEX IF NOT EXISTS links_couples_index_by_short_url_unique
    ON links_couples
        USING btree (short_url);

COMMIT;