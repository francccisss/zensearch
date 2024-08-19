
-- type data_t = {
--   webpage_contents: Array<{
--     header: { title: string; page_url: string };
--     contents: string;
--   }>;
--   header: {
--     title: string;
--   };
-- };

CREATE TABLE known_sites (
    id INTEGER PRIMARY KEY,
    url TEXT NOT NULL, 
    last_added INTEGER NOT NULL
);

CREATE TABLE indexed_sites (
    id INTEGER PRIMARY KEY,
    url TEXT NOT NULL, 
    last_indexed INTEGER NOT NULL
);

CREATE TABLE webpages (
    parent INTEGER REFERENCES indexed_sites(id),
    webpage_id INTEGER PRIMARY KEY,
    webpage_url TEXT NOT NULL,
    title TEXT,
    contents TEXT
);

SELECT * FROM indexed_sites;
