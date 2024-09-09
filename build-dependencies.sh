#!/bin/bash

steps=('./database-service' './gateway/' './web-crawler-service/')
for i in "${steps[@]}"; do
  cd "$i"
  echo "Processing"
  echo "$i"
  npm i
  cd ../
  echo "Done."
done
echo "Done"
