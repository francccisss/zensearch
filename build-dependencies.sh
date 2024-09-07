#!/bin/bash

steps=('./database-service' './gateway/' './web-crawler-service/')
for i in "${steps[@]}"; do
  cd "$i"
  echo "Processing"
  npm i
  cd ../
  echo "$i"
done
echo "Done"
