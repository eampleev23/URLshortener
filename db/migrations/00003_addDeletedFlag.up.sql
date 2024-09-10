BEGIN TRANSACTION;

DO
$$
    BEGIN
        IF EXISTS(SELECT *
                  FROM information_schema.columns
                  WHERE table_name = 'links_couples'
                    and column_name = '__is_deleted')
        THEN
            ALTER TABLE links_couples
                RENAME COLUMN "__is_deleted" TO "is_deleted";
        ELSE
            ALTER TABLE links_couples
                ADD COLUMN is_deleted BOOLEAN NOT NULL DEFAULT FALSE;
        END IF;
    END
$$;

COMMIT;