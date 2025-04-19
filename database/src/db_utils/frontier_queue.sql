CREATE TABLE queue (
  id TEXT PRIMARY KEY,
  domain TEXT NOT NULL
);

CREATE TABLE node (
  iD TEXT PRIMARY KEY,
  url TEXT NOT NULL,
  queue_id TEXT REFERENCES queue(id)
);

CREATE TABLE visited_node (
  id TEXT PRIMARY KEY,
  url TEXT NOT NULL,
  node_id TEXT REFERENCES node(id)
);
