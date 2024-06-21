#!/bin/sh

docker compose up -d 

# Install frontend dependencies
cd client && yarn install

# Setup the database
cd ../server && go run ./db/migration db init && go run ./db/migration db migrate

docker compose down
