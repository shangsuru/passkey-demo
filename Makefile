install:
	docker compose up -d 
# Install frontend dependencies
	cd client && yarn install
# Setup the database
	cd server && go run ./db/migration db init
	cd server && go run ./db/migration db migrate
	
	docker compose down

run:
	set -m 
# Spin up postgres and redis using docker
	docker compose up -d
# Generate frontend and watch for changes
	cd client && yarn watch &
# Start the server and watch for changes
	cd server && air

delete:
	docker compose down
	rm -rf ./server/db/data/**
	rm -rf ./server/tmp/**
