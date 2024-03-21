#!/bin/sh

set -e
go build -C aya-backend -o ../output/aya-backend
./output/aya-backend
