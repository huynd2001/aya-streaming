.PHONY: build-server-api build-db-migration build-server-ws

build-server-api:
	go build -C aya-backend/server-api -o ../../output/server-api

build-server-ws:
	go build -C aya-backend/server-ws -o ../../output/server-ws

build-db-migration:
	go build -C aya-backend/db-migration -o ../../output/db-migration

install-web-app:
	cd aya-frontend && \
	npm install && \
	cd ..

build-web-app:
	cd aya-frontend && \
	npm run build && \
	cd ..

build-aya-streaming: build-server-ws build-db-migration build-server-api build-web-app

clean:
	rm -rf ./output
