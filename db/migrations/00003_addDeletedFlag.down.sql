BEGIN TRANSACTION;

ALTER TABLE links_couples
    RENAME COLUMN "is_deleted" TO "__is_deleted";

COMMIT;