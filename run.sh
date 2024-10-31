docker compose down && \
docker ps -a | grep "zensearch" | awk '{print $1}' | xargs docker stop \
&& docker ps -a | grep "zensearch" | awk '{print $1}' | xargs docker rm \
&& docker images | grep "zensearch" | awk '{print $1}' | xargs docker rmi \
&& docker compose up
