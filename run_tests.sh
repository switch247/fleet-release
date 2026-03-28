#!/bin/sh
set -e

cd /app/backend
go test ./...

cd /app/tests
go test ./...
