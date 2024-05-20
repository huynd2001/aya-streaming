# Aya Streaming Project Startup

##### Note: Every command in the following should be run at project root unless explained otherwise.

## Build from source

### Requirements

- [go](https://go.dev/doc/install) version 1.22 or above
- [node](https://nodejs.org/en/download/package-manager) version 20. Highly suggesting using a node version control
  like [nvm](https://github.com/nvm-sh/nvm)

### Build

```shell
make build-aya-streaming
```

You should see the binaries in `output/`, and the web app distribution in `aya-frontend/dist/`

### Run

Initialize the database with

```shell
./output/db-migration
```

Run the backends

```shell
./output/server-api && \
./output/server-ws
```

Run the web app

```shell
node aya-frontend/dist/analog/server/index.mjs
```

or

```shell
npm start # This will run the server in dev mode
```

## Dockerfile

### Requirements

- [Docker](https://docs.docker.com/get-docker/)

### Build

```shell
docker compose build
```

### Run

```shell
docker compose up -d
```
