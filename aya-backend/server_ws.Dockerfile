FROM golang:1.22-bookworm AS build-stage

WORKDIR /src

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

ENV CGO_ENABLED=1

RUN go build -C server-ws -o ../bin

FROM debian:bookworm

RUN apt-get -y update && \
    apt-get install -y ca-certificates

WORKDIR /app

COPY --from=build-stage /src/bin /app/

EXPOSE 8000

CMD ["./bin"]
