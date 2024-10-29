#!/bin/bash

steps=('./database-service' './gateway/' )
for i in "${steps[@]}"; do
  cd "$i"
  echo "Processing"
  echo "$i"
  npm i
  npm run build
  cd ../
  echo "Done."
done


steps=('./web-crawler-service/' './search-engine-service/' )
for i in "${steps[@]}"; do
  cd "$i"
  echo "Processing"
  echo "$i"
  go mod tidy
  go mod build
  cd ../
  echo "Done."
done
