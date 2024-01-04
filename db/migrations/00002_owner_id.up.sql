BEGIN TRANSACTION;

DO
$$
    BEGIN
        IF EXISTS(SELECT *
                  FROM information_schema.columns
                  WHERE table_name = 'links_couples'
                    and column_name = '__owner_id')
        THEN
            ALTER TABLE links_couples
                RENAME COLUMN "__owner_id" TO "owner_id";
        ELSE
            ALTER TABLE links_couples
                ADD COLUMN owner_id INT;
        END IF;
    END
$$;

COMMIT;