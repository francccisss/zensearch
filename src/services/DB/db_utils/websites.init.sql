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
    parent INTEGER ,
    webpage_id INTEGER PRIMARY KEY,
    webpage_url TEXT NOT NULL ,
    title TEXT,
    contents TEXT,
    FOREIGN KEY (parent) REFERENCES indexed_sites(id) ON DELETE CASCADE ON UPDATE CASCADE
);
