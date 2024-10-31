docker images | grep "zensearch" | awk '{print $1}' | xargs docker rmi -f \
&& docker compose up
