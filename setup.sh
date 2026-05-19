#! /bin/bash


# This bash script will setup the Database Environment 
# Variables as well as setting up the .env for the 
# Database service client to connect to the Database
# engine via Unix Socket by default

echo -e "Welcome to zensearch setup.\n"

echo -e "This Setup will need to access your database engine via root for creating
your zensearch database and user.\n"

read -p "Zensearch Database Name: " DB_NAME

read -p "Zensearch Database User: " DB_USER

read -p "Zensearch Database Host: " DB_HOST

read -sp "Zensearch Database Pass: " DB_PASS


echo -e "\nCreating $DB_NAME Database and $DB_USER User."

sudo mariadb << EOF
CREATE DATABASE IF NOT EXISTS $DB_NAME;
CREATE USER IF NOT EXISTS '$DB_USER'@'$DB_HOST' IDENTIFIED BY '$DB_PASS';
GRANT ALL PRIVILEGES ON $DB_NAME.* TO '$DB_USER'@'$DB_HOST';
FLUSH PRIVILEGES;
EOF

echo -e "\nSetting up Database Environment Variables."


cat > .env << EOF
DB_NAME=$DB_NAME
DB_HOST=$DB_HOST
DB_USER=$DB_USER
DB_PASS=$DB_PASS
EOF

echo -e "\n.env generated in the root directory.\n"

echo -e "Done, Database and user created."



