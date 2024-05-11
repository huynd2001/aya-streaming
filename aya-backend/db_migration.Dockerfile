FROM golang:1.22-bookworm AS build-stage

WORKDIR /src

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

ENV CGO_ENABLED=1

RUN go build -C db-migration -o ../bin

FROM debian:bookworm

WORKDIR /app

COPY --from=build-stage /src/bin /app/

CMD ["./bin"]
