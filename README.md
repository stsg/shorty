### URL shortener

#### DB structure

##### table url
```postgresql
CREATE TABLE urls (
    uuid            SERIAL PRIMARY KEY,
    short_url       TEXT not null,
    original_url    TEXT not null
)
```

NOTE: todo
