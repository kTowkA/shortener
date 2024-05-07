CREATE TABLE IF NOT EXISTS url_list (
    uuid uuid,
    short_url text,
    original_url text,
    PRIMARY KEY(uuid),
    UNIQUE(short_url),
    UNIQUE(original_url)
);