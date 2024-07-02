install:
	docker compose up -d 
# Install frontend dependencies
	cd client && yarn install
# Setup the database
	cd server && go run ./db/migration db init && go run ./db/migration db migrate
	docker compose down

run:
	set -m 
# Spin up postgres and redis using docker
	docker compose up -d
# Generate frontend
	cd client && yarn dev & 
# Start the server and watch for changes
	cd server && air
