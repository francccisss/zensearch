CREATE TABLE queues (
  id TEXT PRIMARY KEY,
  domain TEXT NOT NULL
);

CREATE TABLE nodes (
  id INTEGER PRIMARY KEY,
  url TEXT NOT NULL,
  status TEXT DEFAULT 'pending',
  queue_id TEXT REFERENCES queues(id)
);

CREATE TABLE visited_nodes (
  id TEXT PRIMARY KEY,
  url TEXT NOT NULL,
  queue_id TEXT REFERENCES queues(id)
);

