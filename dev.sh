#!/bin/sh
set -m 

# Spin up postgres and redis using docker
docker compose up -d
# Generate frontend
cd client && yarn dev & 
# Start the server and watch for changes
cd server && air && fg
