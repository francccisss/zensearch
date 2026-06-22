#! /bin/bash


USER_ENV=$(cat .env)
DB_USER=""
DB_NAME=""
DB_HOST=""
DB_PASS=""
for secret in $USER_ENV 
do
	if [ ${secret%=*} = "DB_USER" ];then
		DB_USER=${secret#*=}
	fi
	if [ ${secret%=*} = "DB_NAME" ];then
		DB_NAME=${secret#*=}

	fi
	if [ ${secret%=*} = "DB_HOST" ];then
		DB_HOST=${secret#*=}
	fi

	if [ ${secret%=*} = "DB_PASS" ];then
		DB_PASS=${secret#*=}
	fi
done




sudo mariadb << EOF

DROP USER IF EXISTS '$DB_USER'@'$DB_HOST';
DROP DATABASE IF EXISTS $DB_NAME;


CREATE DATABASE IF NOT EXISTS $DB_NAME;
CREATE USER IF NOT EXISTS '$DB_USER'@'$DB_HOST' IDENTIFIED BY '$DB_PASS';
GRANT ALL PRIVILEGES ON $DB_NAME.* TO '$DB_USER'@'$DB_HOST';
FLUSH PRIVILEGES;


USE $DB_NAME;


CREATE TABLE indexed_sites (
    id CHAR(60) PRIMARY KEY,
    hostname VARCHAR(600) NOT NULL UNIQUE
);



CREATE TABLE webpages (
    id CHAR(60) PRIMARY KEY,
    url VARCHAR(600) NOT NULL UNIQUE,
    title TEXT,
    contents LONGTEXT,
    parent CHAR(60), 
    FOREIGN KEY (parent) REFERENCES indexed_sites(id)
);


CREATE TABLE queues (
  id CHAR(60) PRIMARY KEY,
  root VARCHAR(600) NOT NULL UNIQUE
);

CREATE TABLE nodes (
  id INTEGER AUTO_INCREMENT PRIMARY KEY,
  url VARCHAR(600) NOT NULL UNIQUE,
  status CHAR(20) DEFAULT 'pending',
  queue_id CHAR(60),
  date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (queue_id) REFERENCES queues(id) ON DELETE CASCADE
);

CREATE TABLE visited_nodes (
  id INTEGER AUTO_INCREMENT PRIMARY KEY,
  node_url VARCHAR(600) NOT NULL UNIQUE,
  queue_id CHAR(60),
  FOREIGN KEY (queue_id) REFERENCES queues(id) ON DELETE CASCADE
);

EOF

echo "Database Restarted"
