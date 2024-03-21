#!/bin/sh

set -e
go build -C aya-db-migration -o ../output/aya-db-migration
./output/aya-db
