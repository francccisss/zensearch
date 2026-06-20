echo -e "Removing User and Database\n"

USER_ENV=$(cat .env)
DB_USER=""
DB_NAME=""
DB_HOST=""
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
done


read -p "Are you sure? [y/n]: " user_confirmation
read -p "Do you also want to remove the generated .env? [y/n]: " clear_env

if [ "$user_confirmation" = "n" ]; then
	echo -e "Cancelled"
	exit 0
fi

echo -e "Running sudo to delete $DB_NAME and $DB_USER in zensearch\nEnter root password:"
echo "--------------------------"
sudo mariadb << EOF
DROP USER IF EXISTS '$DB_USER'@'$DB_HOST';
DROP DATABASE IF EXISTS $DB_NAME;
SHOW DATABASES;
EOF

if [ "$clear_env" = y ];then
	rm -rf .env
	echo "Removed .env file"
fi


echo "--------------------"
echo -e "Removed zensearch_db and user for zensearch_db database.\nrun setup.sh again if you want to create zensearch database and user."
