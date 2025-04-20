CREATE TABLE queues (
  id TEXT PRIMARY KEY,
  domain TEXT NOT NULL
);

CREATE TABLE nodes (
  id INTEGER PRIMARY KEY,
  url TEXT NOT NULL UNIQUE,
  status TEXT DEFAULT 'pending',
  queue_id TEXT REFERENCES queues(id)
);

CREATE TABLE visited_nodes (
  id INTEGER PRIMARY KEY,
  node_url TEXT NOT NULL REFERENCES nodes(url),
  queue_id TEXT REFERENCES queues(id)
);

