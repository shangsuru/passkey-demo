#!/bin/sh

# Install frontend dependencies
cd client && yarn install

# Setup the database
docker compose up -d 
cd ../server && go run ./db/migration db init && go run ./db/migration db migrate
docker compose down
