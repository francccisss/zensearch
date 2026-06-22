#! /bin/bash


###################### NOTICE ##########################
#                                                      #
# This bash script will setup the Database Environment #
# Variables as well as setting up the .env for the     #
# Database service client to connect to the Database   #
# engine via Unix Socket by default                    #
#                                                      #
# ######################################################

echo -e "Welcome to zensearch setup.\n"

echo -e "This Setup will need to access your database engine via root for creating
your zensearch database and user.\n"



read -p "Zensearch Database Name: " DB_NAME

read -p "Zensearch Database User: " DB_USER

read -p "Zensearch Database Host: " DB_HOST

read -sp "Zensearch Database Pass: " DB_PASS


echo -e "\nCreating $DB_NAME Database and $DB_USER User.\n"

sudo mariadb << EOF
CREATE DATABASE IF NOT EXISTS $DB_NAME;
CREATE USER IF NOT EXISTS '$DB_USER'@'$DB_HOST' IDENTIFIED BY '$DB_PASS';
GRANT ALL PRIVILEGES ON $DB_NAME.* TO '$DB_USER'@'$DB_HOST';
FLUSH PRIVILEGES;
EOF
echo -e "--------------------------------------------"
echo -e "Checking if Database: $DB_NAME was created.\nEnter your password for $DB_USER in $DB_NAME\n"

DATABASE_LIST=$(echo "show databases;" | sudo mariadb -u root -p) 

for DB in $DATABASE_LIST 
do
	if  [ "zensearch_db"="$DB" ]; then
		echo "Database is created moving on."
		break
	fi
done


echo -e "--------------------------------------------\n"
echo -e "Now Creating Zensearch Tables,\nEnter your password for $DB_USER in $DB_NAME\n"

mariadb -u "$DB_USER" -h "$DB_HOST" -D "$DB_NAME" -p < db.init.sql

echo -e "Created Tables for Zensearch."

echo -e "--------------------------------------------\n"
echo -e "Setting up Database Environment Variables.\n"

cat > .env << EOF
DB_NAME=$DB_NAME
DB_HOST=$DB_HOST
DB_USER=$DB_USER
DB_PASS=$DB_PASS
EOF

echo -e "\n.env generated in the root directory.\n"
echo -e "Done, Database and user created."



