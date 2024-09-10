BEGIN TRANSACTION;

ALTER TABLE links_couples
    RENAME COLUMN "owner_id" TO "__owner_id";

COMMIT;