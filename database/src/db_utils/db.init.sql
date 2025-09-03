CREATE TABLE indexed_sites (
    id CHAR(60) PRIMARY KEY,
    primary_url VARCHAR(600) NOT NULL UNIQUE
);

CREATE TABLE webpages (
    id CHAR(60) PRIMARY KEY,
    url VARCHAR(600) NOT NULL UNIQUE,
    title TEXT,
    contents TEXT,
    parent CHAR(60), 
    FOREIGN KEY (parent) REFERENCES indexed_sites(id)
);


CREATE TABLE queues (
  id CHAR(60) PRIMARY KEY,
  domain VARCHAR(600) NOT NULL UNIQUE
);

CREATE TABLE nodes (
  id INTEGER AUTO_INCREMENT PRIMARY KEY,
  url VARCHAR(600) NOT NULL UNIQUE,
  status CHAR(20) DEFAULT 'pending',
  queue_id CHAR(60),
  FOREIGN KEY (queue_id) REFERENCES queues(id) ON DELETE CASCADE
);

CREATE TABLE visited_nodes (
  id INTEGER AUTO_INCREMENT PRIMARY KEY,
  node_url VARCHAR(600) NOT NULL UNIQUE,
  queue_id CHAR(60),
  FOREIGN KEY (queue_id) REFERENCES queues(id) ON DELETE CASCADE
);
