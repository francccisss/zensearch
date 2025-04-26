CREATE TABLE known_sites (
    id TEXT PRIMARY KEY,
    url TEXT NOT NULL,
    last_added INTEGER NOT NULL
);

CREATE TABLE indexed_sites (
    id TEXT PRIMARY KEY,
    primary_url TEXT NOT NULL UNIQUE,
    last_indexed INTEGER NOT NULL
);

CREATE TABLE webpages (
    parent TEXT REFERENCES indexed_sites(id),
    id TEXT PRIMARY KEY,
    url TEXT NOT NULL UNIQUE,
    title TEXT,
    contents TEXT
);

