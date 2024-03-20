#!/bin/bash

set -e
go build -C aya-db -o ./output/aya-db
./output/aya-db
