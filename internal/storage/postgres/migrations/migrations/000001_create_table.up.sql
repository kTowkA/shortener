BEGIN;
CREATE TABLE IF NOT EXISTS url_list (
    uuid uuid,
    user_id uuid,
    short_url text,
    original_url text,
    is_deleted boolean,
    PRIMARY KEY(uuid),
    UNIQUE(short_url),
    UNIQUE(original_url)
);
COMMIT;