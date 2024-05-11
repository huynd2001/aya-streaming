

.PHONY: clean

build-server-api:
	go build -C aya-backend/server-api -o ../../output/server-api

build-server-ws:
	go build -C aya-backend/server-ws -o ../../output/server-ws

build-db-migration:
	go build -C aya-backend/db-migration -o ../../output/db-migration

clean:
	rm -rf ./output
