CREATE TABLE known_sites (
    id INTEGER PRIMARY KEY,
    url TEXT NOT NULL,
    last_added INTEGER NOT NULL
);

CREATE TABLE indexed_sites (
    id INTEGER PRIMARY KEY,
    primary_url TEXT NOT NULL UNIQUE,
    last_indexed INTEGER NOT NULL
);

CREATE TABLE webpages (
    parent INTEGER REFERENCES indexed_sites(id),
    id INTEGER PRIMARY KEY,
    url TEXT NOT NULL ,
    title TEXT,
    contents TEXT
);
