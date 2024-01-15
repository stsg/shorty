CREATE TABLE urls (
    uuid SERIAL PRIMARY KEY,
    short_url text NOT NULL,
    original_url text NOT NULL
);

ALTER TABLE urls
    ADD CONSTRAINT unique_url
        UNIQUE (original_url);

INSERT INTO
    urls (short_url, original_url)
        VALUES ('123456', 'https://www.google.com');
