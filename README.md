An implementation of authentication with passkeys using the [go-webauthn](https://github.com/go-webauthn/webauthn) library

## Local Setup

Start Postgres and Redis with Docker Compose:

```
docker compose up
```

Setup the database:

```
go run ./db/migration db init
go run ./db/migration db migrate
```

Start the server:

```
go run .
```
